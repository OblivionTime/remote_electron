const { clipboard } = require("electron");
const { formProps } = require("element-plus");

let timer = null;
let clipboardContent = "";
let fileList = [];
// 处理文件的逻辑
function sendData(data, MainWindow) {
  MainWindow.webContents.send("clipboard", data);
}
function getClipboardFiles() {
  const clipboardEx = require("electron-clipboard-ex");
  filePath = clipboardEx.readFilePaths();
  if (filePath.length != 0) {
    if (filePath.join(",") != fileList.join(",")) {
      clipboardContent = "";
      fileList = filePath;
      return fileList;
    }
  }
  return [];
}

exports.startWatchClipboard = (MainWindow, sender) => {
  
  let filePath = [];
  // 启动一个循环，监听剪贴板的变化
  timer = setInterval(() => {
    let clipboardText = clipboard.readText();
    if (clipboardText && clipboardContent != clipboardText) {
      clipboardContent = clipboardText;
      fileList = [];
      sendData(
        { clipboard_type: "text", list: [clipboardContent], sender: sender },
        MainWindow
      );
    } else {
      filePath = getClipboardFiles();
      if (filePath.length > 0) {
        sendData(
          { clipboard_type: "file", list: filePath, sender: sender },
          MainWindow
        );
      }
    }
  }, 1000); // 每秒检查一次
};
exports.stopWatchClipboard = () => {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
};
