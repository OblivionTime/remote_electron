"use strict";
require("console");
const {
  app,
  BrowserWindow,
  Tray,
  Menu,
  nativeImage,
  globalShortcut
} = require("electron");
const { join } = require("path");
require("url");
process.env["ELECTRON_DISABLE_SECURITY_WARNINGS"] = "true";
const createWindow = () => {
  const icon = nativeImage.createFromPath(join(__dirname, "logo.png"));
  console.log(__dirname);
  const win = new BrowserWindow({
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
  tray = new Tray(icon);
  const contextMenu = Menu.buildFromTemplate([
    { label: "主界面", click: () => win.show() },
    { label: "退出", click: () => app.exit() }
  ]);
  tray.on("double-click", () => {
    win.show();
  });
  tray.setToolTip("山与路远程控制");
  tray.setContextMenu(contextMenu);
  win.on("close", (event) => {
    win.hide();
    win.setSkipTaskbar(true);
    event.preventDefault();
  });
  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL);
  } else {
    win.loadFile(join(__dirname, "../dist/index.html"));
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
  if (process.platform !== "darwin")
    app.quit();
});
