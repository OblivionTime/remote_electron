"use strict";
const { app, BrowserWindow } = require("electron");
const { join } = require("path");
require("url");
process.env["ELECTRON_DISABLE_SECURITY_WARNINGS"] = "true";
const createWindow = () => {
  const win = new BrowserWindow({
    frame: false,
    width: 750,
    useContentSize: true,
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
  win.setMaximizable(false);
  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL);
  } else {
    win.setMenu(null);
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
