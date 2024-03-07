<template>
    <div class="container">
        <div class="header">
            <div class="title">
                本地设备
            </div>
            <div class="title" style="text-align:right">
                远程设备
            </div>
        </div>
        <div class="path">
            <p>{{ LocalPath }}</p>
            <p style="text-align:right">{{ RemotePath }}</p>
        </div>
        <div class="middle">
            <div class="item">
                <div class="icon" @click="backLocalPath"><el-icon>
                        <ArrowLeftBold />
                    </el-icon></div>
                <div class="icon" @click="RefeshLocalPath"><el-icon>
                        <RefreshRight />
                    </el-icon></div>
                <div class="icon"><el-icon>
                        <FolderAdd />
                    </el-icon></div>
                <select class="item-select" v-model="localSelect" placeholder="请选择目录" @change="Handlerlocal">
                    <option :value="item.value" v-for="item in LocalDirOptions" :key="item.key">{{ item.label }}
                    </option>
                </select>
            </div>
            <el-button type="primary" size="small" :icon="Connection" @click="UploadFile">传输</el-button>
            <div class="item">
                <div class="icon" @click="backRemotePath"><el-icon>
                        <ArrowLeftBold />
                    </el-icon></div>
                <div class="icon" @click="RefeshRemotePath"><el-icon>
                        <RefreshRight />
                    </el-icon></div>
                <div class="icon"><el-icon>
                        <FolderAdd />
                    </el-icon></div>
                <select class="item-select" v-model="remoteSelect" placeholder="请选择目录" @change="HandlerRemote">
                    <option :value="item.value" v-for="item in RemoteDirOptions" :key="item.key">{{ item.label }}
                    </option>
                </select>
            </div>
        </div>
        <div class="main">
            <div class="main-left">
                <el-table ref="localTableRef" :data="LocalFileList" highlight-current-row @row-click="HandlerLocalClick"
                    @row-dblclick="HandlerLocalDBClick" height="230">
                    <el-table-column property="file_name" show-overflow-tooltip label="名称">
                        <template #default="scope">
                            <img :src="scope.row.is_dir ? FolderIcon : FileIcon" alt=""
                                style="width:13px;margin-right:5px">
                            {{ scope.row.file_name }}
                        </template>
                    </el-table-column>
                    <el-table-column property="" label="大小" width="80">

                        <template #default="scope">

                            {{ formatBytes(scope.row.file_size) }}
                        </template>
                    </el-table-column>
                    <el-table-column property="update_time" label="修改时间" width="150" />
                </el-table>
            </div>
            <div class="main-right">
                <el-table ref="remoteTableRef" :data="RemoteFileList" highlight-current-row
                    @row-click="HandlerRemoteClick" @row-dblclick="HandlerRemoteDBClick" height="230">
                    <el-table-column property="file_name" show-overflow-tooltip label="名称">

                        <template #default="scope">
                            <img :src="scope.row.is_dir ? FolderIcon : FileIcon" alt=""
                                style="width:13px;margin-right:5px">
                            {{ scope.row.file_name }}
                        </template>
                    </el-table-column>
                    <el-table-column property="" label="大小" width="80">

                        <template #default="scope">
                            {{ formatBytes(scope.row.file_size) }}
                        </template>
                    </el-table-column>
                    <el-table-column property="update_time" label="修改时间" width="150" />
                </el-table>
            </div>
        </div>
        <div class="send">
            <div class="title">文件传输</div>
            <el-table ref="singleTableRef" :data="tableData2" highlight-current-row height="130">
                <el-table-column property="filename" show-overflow-tooltip label="名称" />
                <el-table-column label="状态" width="200">

                    <template #default="scope">
                        <el-progress :percentage="parseInt(scope.row.progress)">
                            <template #default="">

                                <p style="color: #409EFF;font-size:10px" v-if="scope.row.status == 0">
                                    等待传输
                                </p>
                                <p style="color: #909399;font-size:10px" v-else-if="scope.row.status == 1">{{
                scope.row.speed }}</p>
                                <p style="color: #67C23A;font-size:10px" v-else-if="scope.row.status == 2">传输成功</p>
                                <p style="color: #F56C6C;font-size:10px" v-else>
                                    <el-tooltip :content="scope.row.errmsg" effect="customized">
                                        传输失败
                                    </el-tooltip>
                                </p>
                            </template>
                        </el-progress>
                    </template>
                </el-table-column>
                <el-table-column property="filesize" label="大小" width="80" show-overflow-tooltip />
                <el-table-column property="filepath" label="发送路径" show-overflow-tooltip />
                <el-table-column property="remotepath" label="接收路径" show-overflow-tooltip />
            </el-table>
        </div>
    </div>
</template>

<script setup>
import { Connection } from '@element-plus/icons-vue'
import FileIcon from '@/assets/file.svg';
import FolderIcon from '@/assets/folder.svg';
const currentRow = ref()

//#region 本地相关参数 ------------------------------------------------------------
const LocalPath = ref("未获取")
var LocalFileList = ref([])
var LocalDirOptions = ref([])
var localSelect = ref("")
var historyDir = []
var uploadFileInfo = {}
var localTableRef = ref()
//切换目录
function Handlerlocal() {
    historyDir.push(LocalPath.value)
    LocalPath.value = localSelect.value
    localSelect.value = ""
    fileSocket.value.send(JSON.stringify({
        "op": "local_ls",
        "localpath": LocalPath.value,
    }))
}
//返回
function backLocalPath() {
    if (historyDir.length == 0) {
        return
    }
    LocalPath.value = historyDir.pop()
    localSelect.value = ""
    fileSocket.value.send(JSON.stringify({
        "op": "local_ls",
        "localpath": LocalPath.value,
    }))
}
//刷新
function RefeshLocalPath() {
    fileSocket.value.send(JSON.stringify({
        "op": "local_ls",
        "localpath": LocalPath.value,
    }))
}
//处理单击事件
function HandlerLocalClick(item) {
    if (!item.is_dir) {
        uploadFileInfo = item
        localTableRef.value.setCurrentRow(item)
        uploadRemoteFileInfo = null
        remoteTableRef.value.setCurrentRow(null)
    } else {
        uploadFileInfo = null
    }


}
//处理双击事件
function HandlerLocalDBClick(item) {
    if (item.is_dir) {
        uploadFileInfo = null
        localTableRef.value.setCurrentRow(uploadFileInfo)
        historyDir.push(LocalPath.value)
        LocalPath.value = item.file_path
        localSelect.value = ""
        fileSocket.value.send(JSON.stringify({
            "op": "local_ls",
            "localpath": LocalPath.value,
        }))

    } else {
        if (!tableData2.value.some((e) => e.filepath == item.file_path)) {
            tableData2.value.unshift({
                filename: getFileName(item.file_path),
                filepath: item.file_path,
                filesize: formatBytes(item.file_size),
                op: "download",
                progress: 0,
                remotepath: RemotePath.value + "/" + item.file_name,
                speed: "0 MB/s",
                status: 0,
            })
            fileSocket.value.send(JSON.stringify({
                op: "upload",
                remotepath: RemotePath.value + "/" + item.file_name,
                filepath: item.file_path,
                filesize: item.file_size,
            }))
        }
    }

}
//#endregion -----------------------------------------------------------
//#region 远程相关参数 ------------------------------------------------------------
const RemotePath = ref("未获取")
var RemoteFileList = ref([])
var RemoteDirOptions = ref([])
var remoteSelect = ref("")
var historyRemoteDir = []
var uploadRemoteFileInfo = {}
var remoteTableRef = ref()
//切换目录
function HandlerRemote() {
    historyRemoteDir.push(RemotePath.value)
    RemotePath.value = remoteSelect.value
    remoteSelect.value = ""
    fileSocket.value.send(JSON.stringify({
        "op": "remote_ls",
        "remotepath": RemotePath.value,
    }))
}
//返回
function backRemotePath() {
    if (historyRemoteDir.length == 0) {
        console.log("🚀 ~ backRemotePath ~ historyRemoteDir:", historyRemoteDir)
        return
    }
    RemotePath.value = historyRemoteDir.pop()
    remoteSelect.value = ""
    fileSocket.value.send(JSON.stringify({
        "op": "remote_ls",
        "remotepath": RemotePath.value,
    }))
}
//刷新
function RefeshRemotePath() {
    fileSocket.value.send(JSON.stringify({
        "op": "remote_ls",
        "remotepath": RemotePath.value,
    }))
}
//处理单击事件
function HandlerRemoteClick(item) {
    if (!item.is_dir) {
        uploadRemoteFileInfo = item
        remoteTableRef.value.setCurrentRow(uploadRemoteFileInfo)
        localTableRef.value.setCurrentRow(null)
        uploadFileInfo = null
    } else {
        uploadRemoteFileInfo = null
    }

}
//处理双击事件
function HandlerRemoteDBClick(item) {
    if (item.is_dir) {
        uploadRemoteFileInfo = null
        remoteTableRef.value.setCurrentRow(null)
        historyRemoteDir.push(RemotePath.value)
        RemotePath.value = item.file_path
        remoteSelect.value = ""
        fileSocket.value.send(JSON.stringify({
            "op": "remote_ls",
            "remotepath": RemotePath.value,
        }))

    } else {
        if (!tableData2.value.some((e) => e.filepath == item.file_path)) {
            tableData2.value.unshift({
                filename: getFileName(item.file_path),
                filesize: formatBytes(item.file_size),
                op: "download",
                progress: 0,
                remotepath: LocalPath.value + "/" + item.file_name,
                filepath: item.file_path,
                speed: "0 MB/s",
                status: 0,
            })
            fileSocket.value.send(JSON.stringify({
                op: "download",
                remotepath: LocalPath.value + "/" + item.file_name,
                filepath: item.file_path,
                filesize: item.file_size,
            }))
        }

    }

}
//#endregion -----------------------------------------------------------
//#region 文件相关操作和参数 ------------------------------------------------------------
const tableData2 = ref([])
var fileSocket = ref(null)
//#endregion -----------------------------------------------------------
const UploadFile = () => {
    if (!uploadFileInfo && !uploadRemoteFileInfo) {
        ElMessage({
            showClose: true,
            message: '请选择一个文件后在进行传输.',
            type: 'warning',
        })
        return
    }
    let item = uploadFileInfo ? uploadFileInfo : uploadRemoteFileInfo
    if (uploadFileInfo) {
        if (!tableData2.value.some((e) => e.filepath == item.file_path)) {

            tableData2.value.unshift({
                filename: getFileName(item.file_path),
                filepath: item.file_path,
                filesize: formatBytes(item.file_size),
                op: "download",
                progress: 0,
                remotepath: RemotePath.value + "/" + item.file_name,
                speed: "0 MB/s",
                status: 0,
            })
            fileSocket.value.send(JSON.stringify({
                op: "upload",
                remotepath: RemotePath.value + "/" + item.file_name,
                filepath: item.file_path,
                filesize: item.file_size,
            }))
        }
    } else {
        tableData2.value.unshift({
            filename: getFileName(item.file_path),
            filesize: formatBytes(item.file_size),
            op: "download",
            progress: 0,
            remotepath: LocalPath.value + "/" + item.file_name,
            filepath: item.file_path,
            speed: "0 MB/s",
            status: 0,
        })
        fileSocket.value.send(JSON.stringify({
            op: "download",
            remotepath: LocalPath.value + "/" + item.file_name,
            filepath: item.file_path,
            filesize: item.file_size,
        }))
    }
    clearChoose()

}




//#region 方法 ------------------------------------------------------------
const path = window.require('path');
const getFileName = (p) => {
    return path.basename(p);

}
function formatBytes(bytes) {
    if (bytes === 0 || !bytes) return '0 B';

    const k = 1024;
    const dm = 0

    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}
function clearChoose() {
    uploadRemoteFileInfo = null
    remoteTableRef.value.setCurrentRow(null)
    localTableRef.value.setCurrentRow(null)
    uploadFileInfo = null
}
//#endregion -----------------------------------------------------------
const ipcRenderer = window.require('electron').ipcRenderer;
import { ElMessageBox } from "element-plus";
const route = useRoute()
onMounted(() => {
    let device = route.query?.device
    let code = route.query?.code
    setTimeout(() => {
        fileSocket.value = new WebSocket(
            `ws://${import.meta.env.VITE_API_URL}/v1/api/remote/server/file_connect`
        );
        fileSocket.value.onopen = () => {
            fileSocket.value.send(JSON.stringify({
                "op": "connection",
                "device": device,
                "code": code,
            }))

        }

        fileSocket.value.onmessage = (msg) => {
            msg = JSON.parse(msg.data)
            switch (msg.op) {
                case "connection":
                    fileSocket.value.send(JSON.stringify({
                        "op": "remote_basic",
                    }))
                    fileSocket.value.send(JSON.stringify({
                        "op": "local_basic",
                    }))
                    break
                case "disconnected":
                    ElMessageBox.alert(msg.errmsg, '连接断开', {
                        confirmButtonText: '确定',
                        showClose: false,
                        callback: () => {
                            ipcRenderer.send("closeFileDialog")
                        },
                    })
                    break
                case "local_ls":
                    if (msg.local_file_list) {
                        LocalFileList.value = msg.local_file_list
                    } else {
                        LocalFileList.value = []
                    }

                    break
                case "local_basic":
                    for (const key in msg.local_basic_list) {
                        LocalDirOptions.value.push({ label: key, value: msg.local_basic_list[key] })
                    }
                    LocalPath.value = msg.local_basic_list["桌面"]
                    fileSocket.value.send(JSON.stringify({
                        "op": "local_ls",
                        "localpath": LocalPath.value,
                    }))
                    break
                case "remote_ls":
                    if (msg.remote_file_list) {
                        RemoteFileList.value = msg.remote_file_list
                    } else {
                        RemoteFileList.value = []
                    }
                    break
                case "remote_basic":
                    for (const key in msg.remote_basic_list) {
                        RemoteDirOptions.value.push({ label: key, value: msg.remote_basic_list[key] })
                    }
                    RemotePath.value = msg.remote_basic_list["桌面"]
                    fileSocket.value.send(JSON.stringify({
                        "op": "remote_ls",
                        "remotepath": RemotePath.value,
                    }))
                    break
                case "upload":
                case "download":
                    let index = tableData2.value.findIndex((item) => {
                        return item.filepath == msg.filepath;
                    });
                    tableData2.value[index] = msg
                    break
            }
        }

        fileSocket.value.onclose = () => {
            ElMessageBox.alert('检测到连接断开!', '连接断开', {
                confirmButtonText: '确定',
                showClose: false,
                callback: () => {
                    ipcRenderer.send("closeFileDialog")
                },
            })
        }
        fileSocket.value.onerror = (err) => {
            console.log(err);

        }
    }, 1000);

})
</script>

<style lang="scss" scoped>
.container {
    padding: 10px;
    box-sizing: border-box;
    background: #F4F4F4;
    height: 100vh;

    .header {
        display: flex;
        justify-content: space-between;
        font-size: 20px;
        font-weight: bolder;
        background: white;
        padding: 10px 20px;
        border-radius: 10px;

        .title {
            flex: 1;

        }

    }

    .path {
        display: flex;
        justify-content: space-between;
        box-sizing: border-box;
        margin: 15px 0;
        font-size: 12px;
        padding: 0 20px;

        color: #636e72;

        p {
            flex: 1;
        }
    }

    .middle {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .item {
            background: white;
            padding: 5px 15px 5px 20px;
            display: flex;
            gap: 20px;
            justify-content: flex-start;
            border-radius: 5px;

            .icon {
                cursor: pointer;
                font-size: 14px;
            }

            .item-select {
                width: 200px;
            }
        }

    }

    .main {
        margin-top: 20px;
        width: 100%;
        display: flex;
        justify-content: space-between;

        .main-left {
            width: 48vw;
        }

        .main-right {
            width: 48vw;
        }
    }

    .send {
        .title {
            background: white;
            padding: 10px 20px;
            margin: 10px 0;
            font-size: 20px;
            font-weight: bolder;
            border-radius: 10px;
        }
    }

    :deep(.el-table .cell) {
        font-size: 12px;
        line-height: 20px;
    }

    :deep(.el-table .el-table__cell) {
        padding: 6px 0;
    }

    :deep(.el-table__body td.el-table__cell) {
        border-bottom: 0px;
    }
}

:deep(.el-popper.is-customized) {
    /* Set padding to ensure the height is 32px */
    padding: 6px 12px;
    background: linear-gradient(90deg, rgb(159, 229, 151), rgb(204, 229, 129));

    .el-popper__arrow::before {
        background: linear-gradient(45deg, #b2e68d, #bce689);
        right: 0;
    }
}
</style>