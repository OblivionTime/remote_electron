const {
  app,
  BrowserWindow,
  Tray,
  Menu,
  nativeImage,
  globalShortcut,
  screen,
  ipcMain,
  dialog,
  desktopCapturer,
} = require("electron");
const { startWatchClipboard, stopWatchClipboard } = require("./clipboard.js");
//#region 启动后端项目 ------------------------------------------------------------
const { join, dirname } = require("path");
const { spawn } = require("child_process");
const fs = require("fs");
let exampleProcess;

//#endregion -----------------------------------------------------------

const { format } = require("url");

let MainWindow;
let MainTray;
let FloatingWin = null;
let FileWindow;
const icon = nativeImage.createFromPath(join(__dirname, "logo.png"));

//#region 关闭行为 / 退出 ------------------------------------------------------------
// ask = 每次询问；hide = 隐藏到托盘；quit = 直接退出
const CLOSE_ACTIONS = ["ask", "hide", "quit"];
let closeAction = "ask";
let isQuitting = false;

const closeActionFile = () => join(app.getPath("userData"), "close-action.json");

const loadCloseAction = () => {
  try {
    const saved = JSON.parse(fs.readFileSync(closeActionFile(), "utf8")).action;
    closeAction = CLOSE_ACTIONS.includes(saved) ? saved : "ask";
  } catch (error) {
    closeAction = "ask";
  }
};

const saveCloseAction = (action) => {
  closeAction = CLOSE_ACTIONS.includes(action) ? action : "ask";
  try {
    fs.writeFileSync(closeActionFile(), JSON.stringify({ action: closeAction }));
  } catch (error) {
    console.error("保存关闭行为失败", error);
  }
};

const hideMainWindow = () => {
  if (!MainWindow) return;
  MainWindow.hide();
  MainWindow.setSkipTaskbar(true);
};

const showMainWindow = () => {
  if (!MainWindow) return;
  if (MainWindow.isMinimized()) MainWindow.restore();
  MainWindow.show();
  MainWindow.setSkipTaskbar(false);
  MainWindow.focus();
};

const quitApp = () => {
  isQuitting = true;
  app.quit();
};
//#endregion -----------------------------------------------------------

//#region ipMain相关操作 ------------------------------------------------------------
ipcMain.on("showFloating", () => {
  if (!FloatingWin) {
    createFloatingWindow();
  }
});
ipcMain.on("closeFloating", () => {
  if (FloatingWin) {
    FloatingWin.close();
    FloatingWin = null;
  }
  //被控结束，恢复主界面
  showMainWindow();
});
ipcMain.on("startWatchClipboard", (e, sender) => {
  stopWatchClipboard();
  startWatchClipboard(MainWindow, sender);
});
ipcMain.on("stopWatchClipboard", () => {
  stopWatchClipboard();
});
ipcMain.on("createSuspensionMenu", (e) => {
  const rightM = Menu.buildFromTemplate([
    {
      label: "隐藏",
      click: () => {
        FloatingWin.hide();
      },
    },
    {
      label: "退出远控",
      click: () => {
        const options = {
          type: "question",
          buttons: ["取消", "确定"],
          defaultId: 1,
          title: "警告",
          message: "点击退出后,目标将无法在远控当前设备",
          detail: "",
        };

        dialog.showMessageBox(null, options).then((response) => {
          if (response.response === 1) {
            FloatingWin.close();
            FloatingWin = null;
            showMainWindow();
            if (MainWindow) {
              MainWindow.webContents.send("video_disconnect", "发送成功");
            }
          }
        });
      },
    },
  ]);
  rightM.popup({});
});
// desktopCapturer 在 Electron 34 只能用于主进程，渲染进程通过 IPC 取屏幕源 id
ipcMain.handle("get-screen-source", async () => {
  const sources = await desktopCapturer.getSources({ types: ["screen"] });
  return sources.length > 0 ? sources[0].id : null;
});
ipcMain.handle("get-close-action", () => closeAction);
ipcMain.handle("set-close-action", (e, action) => {
  saveCloseAction(action);
  return closeAction;
});
const senderWindow = (e) => BrowserWindow.fromWebContents(e.sender);
ipcMain.on("window-minimize", (e) => { senderWindow(e)?.minimize(); });
ipcMain.on("window-close", (e) => { senderWindow(e)?.close(); });
ipcMain.on("window-restore", (e) => { senderWindow(e)?.restore(); });
ipcMain.on("window-set-size", (e, w, h) => { senderWindow(e)?.setSize(w, h); });
ipcMain.on("window-center", (e) => { senderWindow(e)?.center(); });
ipcMain.on("window-set-position", (e, x, y) => { senderWindow(e)?.setPosition(x, y); });

//#endregion -----------------------------------------------------------
//#region 主窗口 ------------------------------------------------------------
const createWindow = () => {
  MainWindow = new BrowserWindow({
    frame: false,
    width: 750,
    useContentSize: true,
    icon: icon,
    height: 510,
    minWidth: 750,
    minHeight: 510,
    center: true,
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
      additionalArguments: ["--window-role=main"],
    },
  });
  CreateMainTray();
  MainWindow.on("close", async (event) => {
    if (isQuitting) return;
    event.preventDefault();

    if (closeAction === "quit") {
      quitApp();
      return;
    }
    if (closeAction === "hide") {
      hideMainWindow();
      return;
    }

    const { response, checkboxChecked } = await dialog.showMessageBox(MainWindow, {
      type: "question",
      buttons: ["隐藏到托盘", "退出程序", "取消"],
      defaultId: 0,
      cancelId: 2,
      noLink: true,
      title: "关闭窗口",
      message: "关闭窗口后要做什么？",
      detail: "隐藏到托盘：程序继续在后台运行，仍可被远程控制\n退出程序：完全退出，不再响应远程控制",
      checkboxLabel: "记住我的选择（可在「高级设置 - 窗口行为」中修改）",
      checkboxChecked: true,
    });

    if (response === 2) return;

    const action = response === 1 ? "quit" : "hide";
    if (checkboxChecked) saveCloseAction(action);

    if (action === "quit") {
      quitApp();
    } else {
      hideMainWindow();
    }
  });
  MainWindow.webContents.on("did-finish-load", () => {
    MainWindow.webContents.send("connect", "发送成功");
  });

  // development模式
  if (process.env.VITE_DEV_SERVER_URL) {
    MainWindow.loadURL(process.env.VITE_DEV_SERVER_URL + "#/index");
  } else {
    MainWindow.setMenu(null);

    const appPath = app.getAppPath();
    process.chdir(appPath + "/../../");
    exampleProcess = spawn("remote.exe");
    MainWindow.loadFile(join(__dirname, "../dist/index.html"));
    MainWindow.webContents.executeJavaScript(
      `window.location.hash = '#/index';`
    );
  }
};
const CreateMainTray = () => {
  if (MainTray) return;
  MainTray = new Tray(icon);

  const contextMenu = Menu.buildFromTemplate([
    { label: "主界面", click: () => showMainWindow() },
    { label: "退出", click: () => quitApp() },
  ]);
  MainTray.on("double-click", () => {
    showMainWindow();
  });
  MainTray.setToolTip("山与路远程控制");
  MainTray.setContextMenu(contextMenu);
};
//#endregion -----------------------------------------------------------

//#region 悬浮球相关 ------------------------------------------------------------
const createFloatingWindow = () => {
  //被控期间隐藏主界面，只留悬浮球
  hideMainWindow();
  FloatingWin = new BrowserWindow({
    frame: false,
    width: 80,
    alwaysOnTop: true,
    useContentSize: true,
    icon: icon,
    height: 300,
    maxWidth: 80,
    maxHeight: 300,
    resizable: false,
    autoHideMenuBar: true,
    transparent: true, // 设置为透明
    type: "toolbar",
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
      additionalArguments: ["--window-role=float"],
    },
  });
  const { left, top } = {
    left: screen.getPrimaryDisplay().workAreaSize.width - 160,
    top: 80,
  };
  FloatingWin.setPosition(left, top); //设置悬浮球位置

  FloatingWin.setMaximizable(false); // 禁止双击最大化

  // development模式
  if (process.env.VITE_DEV_SERVER_URL) {
    FloatingWin.loadURL(process.env.VITE_DEV_SERVER_URL + "#/controlledEnd");
    // 开启调试台
    // FloatingWin.webContents.openDevTools()
  } else {
    // FloatingWin.setMenu(null);
    FloatingWin.loadFile(join(__dirname, "../dist/index.html"));
    FloatingWin.webContents.executeJavaScript(
      `window.location.hash = '#/controlledEnd';`
    );
  }
};
//#endregion -----------------------------------------------------------
//#region 文件相关 ------------------------------------------------------------
ipcMain.on("closeFileDialog", () => {
  if (FileWindow) {
    FileWindow.close();
    FileWindow = null;
  }
});
ipcMain.on("createFileWindow", (e, item) => {
  const { identificationCode, verificationCode } = item;
  createFileWindow(identificationCode, verificationCode);
});
const createFileWindow = (device, code) => {
  // 已经开过就复用，不再开第二个
  if (FileWindow) {
    if (FileWindow.isMinimized()) FileWindow.restore();
    FileWindow.show();
    FileWindow.focus();
    return;
  }
  FileWindow = new BrowserWindow({
    width: 1100,
    useContentSize: true,
    icon: icon,
    height: 760,
    minWidth: 940,
    minHeight: 620,
    center: true,
    title: "文件传输",
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
      additionalArguments: ["--window-role=file"],
    },
  });
  FileWindow.on("close", (event) => {
    if (FileWindow) {
      FileWindow = null;
    }
  });
  // development模式
  if (process.env.VITE_DEV_SERVER_URL) {
    FileWindow.loadURL(
      process.env.VITE_DEV_SERVER_URL +
        `#/fileManage?device=${device}&code=${code}`
    );
  } else {
    // FloatingWin.setMenu(null);
    FileWindow.loadFile(join(__dirname, "../dist/index.html"));
    FileWindow.webContents.executeJavaScript(
      `window.location.hash = '#/fileManage?device=${device}&code=${code}';`
    );
  }
};
//#endregion -----------------------------------------------------------
// 已经有实例在跑时就复用（聚焦已有窗口），不再开第二个
if (!app.requestSingleInstanceLock()) {
  app.quit();
} else {
  app.on("second-instance", () => {
    if (FloatingWin) {
      FloatingWin.show();
      return;
    }
    showMainWindow();
  });

  app.on("before-quit", () => {
    isQuitting = true;
  });

  app.whenReady().then(() => {
    loadCloseAction();
    createWindow();

    app.on("activate", () => {
      if (BrowserWindow.getAllWindows().length === 0) createWindow();
    });
  });

  app.on("window-all-closed", () => {
    try {
      stopWatchClipboard();
      if (exampleProcess) {
        process.kill(exampleProcess.pid);
      }
    } catch (error) {}
    if (process.platform !== "darwin") app.quit();
  });
}
