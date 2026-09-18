const electron = window.require("electron");
const path = window.require("path");

export const ipcRenderer = electron.ipcRenderer;
export const basename = path.basename;

// Electron 34 起 desktopCapturer 只能在主进程使用，渲染进程通过 IPC 取屏幕源 id
export const getScreenSourceId = () => ipcRenderer.invoke("get-screen-source");

// 关闭按钮行为（ask 每次询问 / hide 隐藏到托盘 / quit 退出程序）
export const CLOSE_ACTIONS = ["ask", "hide", "quit"];
export const getCloseAction = () => ipcRenderer.invoke("get-close-action");
export const setCloseAction = (action) => ipcRenderer.invoke("set-close-action", action);

export const minimizeWindow = () => ipcRenderer.send("window-minimize");
export const closeWindow = () => ipcRenderer.send("window-close");
export const restoreWindow = () => ipcRenderer.send("window-restore");
export const centerWindow = () => ipcRenderer.send("window-center");
export const setWindowSize = (width, height) =>
  ipcRenderer.send("window-set-size", width, height);
export const setWindowPosition = (x, y) =>
  ipcRenderer.send("window-set-position", x, y);
