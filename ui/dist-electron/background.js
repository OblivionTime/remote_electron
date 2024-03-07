"use strict";
const {
  app,
  BrowserWindow,
  Tray,
  Menu,
  nativeImage,
  globalShortcut,
  screen,
  ipcMain,
  dialog
} = require("electron");
const { startWatchClipboard, stopWatchClipboard } = require("./clipboard.js");
const { join, dirname } = require("path");
const { spawn } = require("child_process");
let exampleProcess;
require("url");
process.env["ELECTRON_DISABLE_SECURITY_WARNINGS"] = "true";
let MainWindow;
let MainTray;
const icon = nativeImage.createFromPath(join(__dirname, "logo.png"));
ipcMain.on("showFloating", () => {
  if (!FloatingWin) {
    createFloatingWindow();
  }
});
ipcMain.on("closeFloating", () => {
  if (FloatingWin) {
    FloatingWin.close();
    FloatingWin = null;
    MainWindow.show();
    MainTray.destroy();
    MainTray = null;
    CreateMainTray();
  }
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
      }
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
          detail: ""
        };
        dialog.showMessageBox(null, options).then((response) => {
          if (response.response === 1) {
            FloatingWin.close();
            FloatingWin = null;
            MainWindow.show();
            MainTray.destroy();
            MainTray = null;
            CreateMainTray();
            MainWindow.webContents.send("video_disconnect", "发送成功");
          }
        });
      }
    }
  ]);
  rightM.popup({});
});
const createWindow = () => {
  MainWindow = new BrowserWindow({
    frame: false,
    width: 750,
    useContentSize: true,
    icon,
    height: 510,
    minWidth: 750,
    minHeight: 510,
    center: true,
    useContenRtSize: true,
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: true,
      enableRemoteModule: true,
      contextIsolation: false
    }
  });
  CreateMainTray();
  MainWindow.webContents.on("did-finish-load", () => {
    MainWindow.webContents.send("connect", "发送成功");
  });
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
  MainTray = new Tray(icon);
  const contextMenu = Menu.buildFromTemplate([
    { label: "主界面", click: () => MainWindow.show() },
    { label: "退出", click: () => app.exit() }
  ]);
  MainTray.on("double-click", () => {
    MainWindow.show();
  });
  MainTray.setToolTip("山与路远程控制");
  MainTray.setContextMenu(contextMenu);
  MainWindow.on("close", (event) => {
    MainWindow.hide();
    MainWindow.setSkipTaskbar(true);
    event.preventDefault();
  });
};
let FloatingWin = null;
const createFloatingWindow = () => {
  MainWindow.hide();
  if (MainTray) {
    MainTray.destroy();
    MainTray = null;
  }
  FloatingWin = new BrowserWindow({
    frame: false,
    width: 80,
    alwaysOnTop: true,
    useContentSize: true,
    icon,
    height: 300,
    maxWidth: 80,
    maxHeight: 300,
    resizable: false,
    useContenRtSize: true,
    autoHideMenuBar: true,
    transparent: true,
    // 设置为透明
    type: "toolbar",
    webPreferences: {
      nodeIntegration: true,
      enableRemoteModule: true,
      contextIsolation: false
    }
  });
  const { left, top } = {
    left: screen.getPrimaryDisplay().workAreaSize.width - 160,
    top: 80
  };
  FloatingWin.setPosition(left, top);
  FloatingWin.setMaximizable(false);
  MainTray = new Tray(icon);
  const contextMenu = Menu.buildFromTemplate([
    { label: "显示", click: () => FloatingWin.show() },
    {
      label: "退出",
      click: () => {
        const options = {
          type: "question",
          buttons: ["取消", "确定"],
          defaultId: 1,
          title: "警告",
          message: "点击退出后,目标将无法在远控当前设备",
          detail: ""
        };
        dialog.showMessageBox(null, options).then((response) => {
          if (response.response === 1) {
            FloatingWin.close();
            FloatingWin = null;
            MainWindow.show();
            MainTray.destroy();
            MainTray = null;
            CreateMainTray();
            MainWindow.webContents.send("video_disconnect", "发送成功");
          }
        });
      }
    }
  ]);
  MainTray.on("double-click", () => {
    FloatingWin.show();
  });
  MainTray.setToolTip("你当前正在被控制中...");
  MainTray.setContextMenu(contextMenu);
  if (process.env.VITE_DEV_SERVER_URL) {
    FloatingWin.loadURL(process.env.VITE_DEV_SERVER_URL + "#/controlledEnd");
  } else {
    FloatingWin.loadFile(join(__dirname, "../dist/index.html"));
    FloatingWin.webContents.executeJavaScript(
      `window.location.hash = '#/controlledEnd';`
    );
  }
};
let FileWindow;
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
  if (FileWindow) {
    MainWindow.webContents.send("errmsg", "目前只支持远程一个目标进行传输文件");
    return;
  }
  FileWindow = new BrowserWindow({
    width: 850,
    useContentSize: true,
    icon,
    height: 610,
    minWidth: 850,
    minHeight: 610,
    center: true,
    title: "文件传输",
    useContenRtSize: true,
    autoHideMenuBar: true,
    webPreferences: {
      nodeIntegration: true,
      enableRemoteModule: true,
      contextIsolation: false
    }
  });
  FileWindow.on("close", (event) => {
    if (FileWindow) {
      FileWindow = null;
    }
  });
  if (process.env.VITE_DEV_SERVER_URL) {
    FileWindow.loadURL(
      process.env.VITE_DEV_SERVER_URL + `#/fileManage?device=${device}&code=${code}`
    );
  } else {
    FileWindow.loadFile(join(__dirname, "../dist/index.html"));
    FileWindow.webContents.executeJavaScript(
      `window.location.hash = '#/fileManage?device=${device}&code=${code}';`
    );
  }
};
app.whenReady().then(() => {
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0)
      createWindow();
  });
});
app.on("window-all-closed", () => {
  try {
    stopWatchClipboard();
    if (exampleProcess) {
      process.kill(exampleProcess.pid);
    }
  } catch (error) {
  }
  if (process.platform !== "darwin")
    app.quit();
});
