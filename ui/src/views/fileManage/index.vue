<template>
    <div class="page">
        <!-- 顶部：两侧目录 -->
        <div class="card toolbar">
            <div class="side">
                <div class="side-head">
                    <span class="badge local">本地</span>
                    <div class="crumbs">
                        <span class="crumb" v-for="(seg, i) in localCrumbs" :key="seg.path"
                            :class="{ last: i === localCrumbs.length - 1 }" :title="seg.path"
                            @click="jumpTo('local', seg.path)">{{ seg.label }}</span>
                    </div>
                    <div class="icons">
                        <el-icon title="上级目录" @click="goParent('local')">
                            <ArrowUpBold />
                        </el-icon>
                        <el-icon title="刷新" @click="refresh('local')">
                            <RefreshRight />
                        </el-icon>
                    </div>
                </div>
                <el-select v-model="local.jump" size="small" placeholder="快速跳转" @change="(v) => onJump('local', v)">
                    <el-option v-for="d in local.basic" :key="d.value" :label="d.label" :value="d.value" />
                </el-select>
            </div>

            <div class="ops">
                <el-button type="primary" :icon="Upload" @click="enqueue('local')">上传</el-button>
                <el-button :icon="Download" @click="enqueue('remote')">下载</el-button>
            </div>

            <div class="side">
                <div class="side-head">
                    <span class="badge remote">远程</span>
                    <div class="crumbs">
                        <span class="crumb" v-for="(seg, i) in remoteCrumbs" :key="seg.path"
                            :class="{ last: i === remoteCrumbs.length - 1 }" :title="seg.path"
                            @click="jumpTo('remote', seg.path)">{{ seg.label }}</span>
                    </div>
                    <div class="icons">
                        <el-icon title="上级目录" @click="goParent('remote')">
                            <ArrowUpBold />
                        </el-icon>
                        <el-icon title="刷新" @click="refresh('remote')">
                            <RefreshRight />
                        </el-icon>
                    </div>
                </div>
                <el-select v-model="remote.jump" size="small" placeholder="快速跳转"
                    @change="(v) => onJump('remote', v)">
                    <el-option v-for="d in remote.basic" :key="d.value" :label="d.label" :value="d.value" />
                </el-select>
            </div>
        </div>

        <!-- 两侧文件列表 -->
        <div class="lists">
            <div class="card panel" @contextmenu.prevent="(e) => openMenu('local', null, e)">
                <div class="panel-head">
                    <el-checkbox :model-value="allChecked('local')" :indeterminate="someChecked('local')"
                        @change="(v) => toggleAll('local', v)" />
                    <span class="dot local"></span>
                    <span class="panel-title">本地文件</span>
                    <span class="count">{{ local.files.length }}</span>
                    <span class="tip">Ctrl 点选 · Shift 范围选 · 右键更多</span>
                    <a class="link" @click="createDir('local')">新建文件夹</a>
                </div>
                <el-table :data="local.files" size="small" height="340" empty-text="空目录"
                    :row-class-name="(o) => rowClass('local', o)"
                    @row-click="(row, col, e) => onRowClick('local', row, e)"
                    @row-dblclick="(row) => onRowDblClick('local', row)"
                    @row-contextmenu="(row, col, e) => openRowMenu('local', row, e)">
                    <el-table-column label="名称" min-width="180" show-overflow-tooltip>
                        <template #default="scope">
                            <div class="file-cell">
                                <img :src="scope.row.is_dir ? FolderIcon : FileIcon" alt="">
                                <span>{{ scope.row.file_name }}</span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column label="大小" width="80" align="right">
                        <template #default="scope">
                            {{ scope.row.is_dir ? '—' : formatBytes(scope.row.file_size) }}
                        </template>
                    </el-table-column>
                    <el-table-column property="update_time" label="修改时间" width="140" />
                </el-table>
            </div>

            <div class="card panel" @contextmenu.prevent="(e) => openMenu('remote', null, e)">
                <div class="panel-head">
                    <el-checkbox :model-value="allChecked('remote')" :indeterminate="someChecked('remote')"
                        @change="(v) => toggleAll('remote', v)" />
                    <span class="dot remote"></span>
                    <span class="panel-title">远程文件</span>
                    <span class="count">{{ remote.files.length }}</span>
                    <span class="tip">双击进入目录 · 右键更多</span>
                    <a class="link" @click="createDir('remote')">新建文件夹</a>
                </div>
                <el-table :data="remote.files" size="small" height="340" empty-text="空目录"
                    :row-class-name="(o) => rowClass('remote', o)"
                    @row-click="(row, col, e) => onRowClick('remote', row, e)"
                    @row-dblclick="(row) => onRowDblClick('remote', row)"
                    @row-contextmenu="(row, col, e) => openRowMenu('remote', row, e)">
                    <el-table-column label="名称" min-width="180" show-overflow-tooltip>
                        <template #default="scope">
                            <div class="file-cell">
                                <img :src="scope.row.is_dir ? FolderIcon : FileIcon" alt="">
                                <span>{{ scope.row.file_name }}</span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column label="大小" width="80" align="right">
                        <template #default="scope">
                            {{ scope.row.is_dir ? '—' : formatBytes(scope.row.file_size) }}
                        </template>
                    </el-table-column>
                    <el-table-column property="update_time" label="修改时间" width="140" />
                </el-table>
            </div>
        </div>

        <!-- 传输记录 -->
        <div class="card panel">
            <div class="panel-head">
                <span class="dot transfer"></span>
                <span class="panel-title">传输记录</span>
                <span class="count">{{ transfers.length }}</span>
                <span class="tip">断开后重试会自动校验内容并续传</span>
                <a class="link" v-if="finishedCount" @click="clearFinished">清空已完成（{{ finishedCount }}）</a>
            </div>
            <el-table ref="singleTableRef" :data="transfers" size="small" height="150" empty-text="还没有传输任务">
                <el-table-column label="名称" min-width="170" show-overflow-tooltip>
                    <template #default="scope">
                        <div class="file-cell">
                            <img :src="scope.row.is_dir ? FolderIcon : FileIcon" alt="">
                            <span>{{ scope.row.filename }}</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column label="方向" width="70">
                    <template #default="scope">
                        <span class="dir-tag" :class="scope.row.op">{{ scope.row.op === 'upload' ? '上传' : '下载' }}</span>
                    </template>
                </el-table-column>
                <el-table-column label="状态" width="190">
                    <template #default="scope">
                        <div class="status">
                            <el-progress :percentage="Number(scope.row.progress) || 0"
                                :status="progressStatus(scope.row.status)" :stroke-width="5" :show-text="false" />
                            <el-tooltip :content="scope.row.errmsg" placement="top" :disabled="!scope.row.errmsg">
                                <span class="status-text" :class="'s' + scope.row.status">
                                    {{ statusText(scope.row) }}
                                </span>
                            </el-tooltip>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column property="filesize" label="大小" width="80" align="right" />
                <el-table-column label="目标路径" min-width="160" show-overflow-tooltip>
                    <template #default="scope">{{ scope.row.remotepath }}</template>
                </el-table-column>
                <el-table-column label="操作" width="132" align="center">
                    <template #default="scope">
                        <div class="ops-cell">
                            <el-tooltip content="重试（自动续传）" placement="top"
                                :disabled="scope.row.status !== 3">
                                <el-icon v-if="scope.row.status === 3" @click="retryTransfer(scope.row)">
                                    <RefreshRight />
                                </el-icon>
                            </el-tooltip>
                            <el-tooltip content="继续（从断点续传）" placement="top"
                                :disabled="scope.row.status !== 4">
                                <el-icon v-if="scope.row.status === 4" @click="resumeTransfer(scope.row)">
                                    <VideoPlay />
                                </el-icon>
                            </el-tooltip>
                            <el-tooltip content="暂停（保留已传部分）" placement="top"
                                :disabled="scope.row.status !== 0 && scope.row.status !== 1">
                                <el-icon v-if="scope.row.status === 0 || scope.row.status === 1"
                                    @click="pauseTransfer(scope.row)">
                                    <VideoPause />
                                </el-icon>
                            </el-tooltip>
                            <el-tooltip content="取消并移出列表" placement="top">
                                <el-icon @click="cancelTransfer(scope.row)">
                                    <Close />
                                </el-icon>
                            </el-tooltip>
                        </div>
                    </template>
                </el-table-column>
            </el-table>
        </div>

        <!-- 右键菜单 -->
        <div class="ctx-menu" v-show="menu.visible" :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
            @click.stop @contextmenu.prevent>
            <div class="ctx-item" v-if="menu.side === 'local'"
                :class="{ disabled: !selectedRows('local').length }" @click="menuUpload">
                <el-icon>
                    <Upload />
                </el-icon>上传到远程
            </div>
            <div class="ctx-item" v-if="menu.side === 'remote'"
                :class="{ disabled: !selectedRows('remote').length }" @click="menuDownload">
                <el-icon>
                    <Download />
                </el-icon>下载到本地
            </div>
            <div class="ctx-item danger" :class="{ disabled: !selectedRows(menu.side).length }"
                @click="menuDelete">
                <el-icon>
                    <Delete />
                </el-icon>删除
            </div>
            <div class="ctx-divider"></div>
            <div class="ctx-item" @click="menuMkdir">
                <el-icon>
                    <FolderAdd />
                </el-icon>新建文件夹
            </div>
            <div class="ctx-item" @click="menuRefresh">
                <el-icon>
                    <RefreshRight />
                </el-icon>刷新
            </div>
        </div>
    </div>
</template>

<script setup>
import {
    Upload, Download, ArrowUpBold, RefreshRight, VideoPause, VideoPlay, Close,
    Delete, FolderAdd,
} from '@element-plus/icons-vue'
import FileIcon from '@/assets/file.svg';
import FolderIcon from '@/assets/folder.svg';
import { ElMessage, ElMessageBox } from "element-plus";
import { ipcRenderer, basename } from "@/utils/electron";
import { createSocket } from "@/utils/ws";
import { useRoute } from "vue-router";

const route = useRoute()
const fileSocket = ref(null)

//#region 路径工具 ------------------------------------------------------------
// 统一成 / 分隔；Windows 盘符也走同一套逻辑
const normalize = (p) => String(p || "").replace(/\\/g, "/")
const isDrive = (p) => /^[A-Za-z]:$/.test(p)

// 把路径拆成可点击的面包屑片段。
// 盘符统一成 "C:/" 而不是 "C:"，因为 "C:" 在 Windows 上是"该盘的当前目录"，
// 交给 Go 侧 os.ReadDir 会列错目录。
const splitPath = (p) => {
    const norm = normalize(p)
    if (!norm) return []
    const abs = norm.startsWith("/")
    const out = []
    let acc = ""
    for (const seg of norm.split("/").filter(Boolean)) {
        if (acc === "") {
            acc = isDrive(seg) ? seg + "/" : (abs ? "/" + seg : seg)
        } else {
            acc = acc.replace(/\/+$/, "") + "/" + seg
        }
        out.push({ label: seg, path: acc })
    }
    return out
}

// 真正的上级目录（不是访问历史）
const parentPath = (p) => {
    const norm = normalize(p)
    if (norm === "/" || norm === "") return p
    const parts = splitPath(p)
    if (parts.length <= 1) {
        // 已经是盘符根目录 / POSIX 根目录，不能再往上
        if (isDrive(parts[0]?.path?.replace(/\/$/, ""))) return norm
        return norm.startsWith("/") ? "/" : p
    }
    const parent = parts[parts.length - 2].path
    return isDrive(parent) ? parent + "/" : parent
}

const joinPath = (dir, name) => `${normalize(dir).replace(/\/+$/, "")}/${name}`

// 保留两位小数
const formatBytes = (bytes) => {
    const n = Number(bytes) || 0;
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    let v = n;
    let i = 0;
    while (v >= k && i < sizes.length - 1) {
        v /= k;
        i++;
    }
    return v.toFixed(2) + ' ' + sizes[i];
}
//#endregion -----------------------------------------------------------

//#region 两侧状态 ------------------------------------------------------------
const makeSide = () => reactive({
    path: "未获取",
    files: [],
    basic: [],
    jump: "",
    selected: [],
    anchor: null,
})
const local = makeSide()
const remote = makeSide()
const sides = { local, remote }
const activeSide = ref("local")

const localCrumbs = computed(() => splitPath(local.path))
const remoteCrumbs = computed(() => splitPath(remote.path))
//#endregion -----------------------------------------------------------

//#region 传输队列 ------------------------------------------------------------
const transfers = ref([])
const singleTableRef = ref()
const finishedCount = computed(() => transfers.value.filter((t) => t.status === 2 || t.status === 3).length)

// 未完成的任务记在 localStorage 里（按对端设备区分）。
// 断点本身的"真相"在磁盘上——对端的半成品文件 + CRC，这里只记任务清单，
// 所以不需要 sqlite 这类额外依赖。
const storageKey = () => `fileTransfers:${route.query?.device || ""}`
const persist = () => {
    try {
        localStorage.setItem(storageKey(), JSON.stringify(transfers.value.filter((t) => t.status !== 2)))
    } catch (err) {
        console.error("保存传输队列失败", err)
    }
}
const restore = () => {
    try {
        const saved = JSON.parse(localStorage.getItem(storageKey()) || "[]")
        transfers.value = saved.map((t) => ({
            ...t,
            // 上次正在传的，重连后当等待处理（连接建立时会自动重发）
            status: t.status === 1 ? 0 : t.status,
            speed: "",
        }))
    } catch (err) {
        transfers.value = []
    }
}
//#endregion -----------------------------------------------------------

//#region 通信 ------------------------------------------------------------
const send = (payload) => {
    const s = fileSocket.value
    if (s && s.readyState === 1) s.send(JSON.stringify(payload))
}
const lsOp = (name) => (name === "local" ? "local_ls" : "remote_ls")
const lsField = (name) => (name === "local" ? "localpath" : "remotepath")

const sendLs = (name) => {
    sides[name].jump = ""
    send({ op: lsOp(name), [lsField(name)]: sides[name].path })
}
const refresh = (name) => sendLs(name)

const jumpTo = (name, path) => {
    if (!path || path === sides[name].path) return
    sides[name].path = path
    sendLs(name)
}
const goParent = (name) => jumpTo(name, parentPath(sides[name].path))
const onJump = (name, value) => {
    if (value) jumpTo(name, value)
}

const handleMessage = (msg) => {
    switch (msg.op) {
        case "connection":
            send({ op: "remote_basic" })
            send({ op: "local_basic" })
            // 数据通道（重新）建立：自动续传中断的任务，以及上次没传完的。
            // 此刻不可能有任务正在这条新通道上跑，所以等待/传输中的都重发一遍。
            transfers.value
                .filter((t) => t.status === 0 || t.status === 1)
                .forEach((t) => {
                    t.status = 0
                    dispatch(t)
                })
            break
        case "disconnected":
            ElMessageBox.alert(msg.errmsg, '连接断开', {
                confirmButtonText: '确定',
                showClose: false,
                callback: () => ipcRenderer.send("closeFileDialog"),
            })
            break
        case "local_ls":
            local.files = msg.local_file_list || []
            local.selected = []
            local.anchor = null
            break
        case "remote_ls":
            remote.files = msg.remote_file_list || []
            remote.selected = []
            remote.anchor = null
            break
        case "local_basic": {
            const list = msg.local_basic_list || {}
            local.basic = Object.entries(list).map(([label, value]) => ({ label, value }))
            local.path = list["桌面"] || Object.values(list)[0] || local.path
            sendLs("local")
            break
        }
        case "remote_basic": {
            const list = msg.remote_basic_list || {}
            remote.basic = Object.entries(list).map(([label, value]) => ({ label, value }))
            remote.path = list["桌面"] || Object.values(list)[0] || remote.path
            sendLs("remote")
            break
        }
        case "error":
            ElMessage({ showClose: true, message: msg.errmsg || "操作失败", duration: 2500, type: 'error' })
            break
        case "upload":
        case "download": {
            // 合并而不是整行替换，保留 op / size_raw 等本地字段（重试要用）
            const idx = transfers.value.findIndex((t) => t.filepath === msg.filepath)
            if (idx < 0) break
            const cur = transfers.value[idx]
            // 用户刚暂停，别被还在路上的进度回包盖掉
            if (cur.status === 4 && (msg.status === 0 || msg.status === 1)) break
            transfers.value[idx] = { ...cur, ...msg, filesize: msg.filesize || cur.filesize }
            persist()
            break
        }
    }
}
//#endregion -----------------------------------------------------------

//#region 选择 ------------------------------------------------------------
const keyOf = (row) => row.file_path
const isSelected = (name, row) => sides[name].selected.includes(keyOf(row))
const rowClass = (name, { row }) => (isSelected(name, row) ? "is-selected" : "")
const selectedRows = (name) => {
    const s = sides[name]
    return s.files.filter((f) => s.selected.includes(keyOf(f)))
}

const onRowClick = (name, row, event) => {
    const s = sides[name]
    activeSide.value = name
    const key = keyOf(row)

    if (event.shiftKey && s.anchor) {
        const a = s.files.findIndex((f) => keyOf(f) === s.anchor)
        const b = s.files.findIndex((f) => keyOf(f) === key)
        if (a > -1 && b > -1) {
            const [from, to] = a < b ? [a, b] : [b, a]
            s.selected = s.files.slice(from, to + 1).map(keyOf)
            return
        }
    }
    if (event.ctrlKey || event.metaKey) {
        s.selected = s.selected.includes(key) ? s.selected.filter((k) => k !== key) : [...s.selected, key]
        s.anchor = key
        return
    }
    s.selected = [key]
    s.anchor = key
}

const onRowDblClick = (name, row) => {
    if (row.is_dir) {
        jumpTo(name, row.file_path)
        return
    }
    sides[name].selected = [keyOf(row)]
    enqueue(name)
}

const toggleAll = (name, checked) => {
    const s = sides[name]
    s.selected = checked ? s.files.map(keyOf) : []
    s.anchor = null
}
const allChecked = (name) => {
    const files = sides[name].files
    return files.length > 0 && files.every((f) => sides[name].selected.includes(keyOf(f)))
}
const someChecked = (name) => sides[name].selected.length > 0 && !allChecked(name)
//#endregion -----------------------------------------------------------

//#region 传输 ------------------------------------------------------------
// name = local 表示 本地 → 远程（upload）；remote 表示 远程 → 本地（download）
// 目录会带上 fileop:"dir"，由 Go 侧递归展开；文件夹整体只占一行，进度按总字节聚合
const enqueue = (name) => {
    const rows = selectedRows(name)
    if (!rows.length) {
        ElMessage({
            showClose: true,
            message: name === "local" ? "请先在左侧选择要上传的文件或文件夹" : "请先在右侧选择要下载的文件或文件夹",
            duration: 2000,
            type: 'warning',
        })
        return
    }
    const op = name === "local" ? "upload" : "download"
    const targetDir = name === "local" ? remote.path : local.path
    let queued = 0
    for (const row of rows) {
        const busy = transfers.value.some((t) => t.filepath === row.file_path && t.status !== 2 && t.status !== 3)
        if (busy) continue
        const item = {
            filename: row.is_dir ? row.file_name : basename(row.file_path),
            filepath: row.file_path,
            remotepath: joinPath(targetDir, row.file_name),
            size_raw: row.is_dir ? 0 : row.file_size,
            filesize: row.is_dir ? '文件夹' : formatBytes(row.file_size),
            is_dir: !!row.is_dir,
            op,
            progress: 0,
            speed: "0 MB/s",
            status: 0,
        }
        transfers.value.unshift(item)
        dispatch(item)
        queued++
    }
    if (!queued) {
        ElMessage({ showClose: true, message: "所选内容已在队列中", duration: 1500, type: 'warning' })
    }
    persist()
}

// 重发一条任务。接收端会回执已落盘的字节数并按 CRC 校验，所以重试/继续都是续传
const dispatch = (row) => {
    send({
        op: row.op,
        fileop: row.is_dir ? "dir" : undefined,
        remotepath: row.remotepath,
        filepath: row.filepath,
        filesize: row.size_raw,
    })
}

const retryTransfer = (row) => {
    row.progress = 0
    row.status = 0
    row.speed = "0 MB/s"
    dispatch(row)
}

// 暂停：发送端会停下，已落盘的部分保留，继续时从断点续
const pauseTransfer = (row) => {
    if (row.status !== 0 && row.status !== 1) return
    row.status = 4
    row.speed = ""
    send({ op: "pause", remotepath: row.remotepath, filepath: row.filepath })
    persist()
}

const resumeTransfer = (row) => {
    row.status = 0
    row.progress = row.progress || 0
    row.speed = "0 MB/s"
    dispatch(row)
}

const cancelTransfer = (row) => {
    send({ op: "cancel", remotepath: row.remotepath, filepath: row.filepath })
    transfers.value = transfers.value.filter((t) => t !== row)
    persist()
}

const clearFinished = () => {
    transfers.value = transfers.value.filter((t) => t.status !== 2 && t.status !== 3)
    persist()
}

// 进度条状态：0 等待 / 1 传输中 / 2 成功 / 3 失败 / 4 已暂停
const progressStatus = (status) => {
    if (status == 2) return 'success'
    if (status == 3) return 'exception'
    if (status == 4) return 'warning'
    return ''
}
const statusText = (row) => {
    if (row.status == 0) return '等待传输'
    if (row.status == 1) return row.speed || '传输中'
    if (row.status == 2) return '传输成功'
    if (row.status == 4) return '已暂停'
    return '传输失败'
}
//#endregion -----------------------------------------------------------

//#region 右键菜单 / 新建 / 删除 ------------------------------------------------------------
const menu = reactive({ visible: false, x: 0, y: 0, side: "local" })

// 行上右键：阻止冒泡，否则会再触发外层容器的"空白右键"，菜单被重置
const openRowMenu = (side, row, event) => {
    event.stopPropagation()
    openMenu(side, row, event)
}

const openMenu = (side, row, event) => {
    event.preventDefault()
    activeSide.value = side
    if (row) {
        // 右键的行不在选中集里就先选中它，和其它文件管理器一致
        if (!isSelected(side, row)) {
            sides[side].selected = [keyOf(row)]
            sides[side].anchor = keyOf(row)
        }
    } else {
        // 空白处右键：清空选择，避免误删上一批选中的内容
        sides[side].selected = []
        sides[side].anchor = null
    }
    // 贴边时把菜单收回窗口内
    const w = 152
    const h = 200
    menu.side = side
    menu.x = Math.min(event.clientX, window.innerWidth - w - 8)
    menu.y = Math.min(event.clientY, window.innerHeight - h - 8)
    menu.visible = true
}
const closeMenu = () => {
    menu.visible = false
}

// 把操作发给对应一侧：local 由本机 agent 执行，remote 走数据通道转发到对端
const fsOp = (side, payload) => {
    closeMenu()
    send({
        op: payload.op,
        side,
        target: payload.target || "",
        paths: payload.paths || [],
        localpath: local.path,
        remotepath: remote.path,
    })
}

const menuUpload = () => {
    closeMenu()
    enqueue("local")
}
const menuDownload = () => {
    closeMenu()
    enqueue("remote")
}
const menuRefresh = () => {
    const side = menu.side
    closeMenu()
    refresh(side)
}
const menuMkdir = () => {
    const side = menu.side
    createDir(side)
}

const menuDelete = async () => {
    const side = menu.side
    closeMenu()
    const rows = selectedRows(side)
    if (!rows.length) return
    const where = side === "local" ? "本地" : "远程"
    const preview = rows.slice(0, 5).map((r) => r.file_name).join("、")
    const more = rows.length > 5 ? ` 等 ${rows.length} 项` : ""
    try {
        await ElMessageBox.confirm(
            `<div style="line-height:1.8">确定删除${where}的 <b>${rows.length}</b> 项吗？</div>` +
            `<div style="color:#f56c6c;word-break:break-all;margin-top:6px">${preview}${more}</div>` +
            `<div style="color:#909399;font-size:12px;margin-top:8px">文件夹会连同里面的内容一起删除，且无法恢复。</div>`,
            "删除确认",
            { confirmButtonText: "删除", cancelButtonText: "取消", type: "warning", dangerouslyUseHTMLString: true }
        )
    } catch (err) {
        return
    }
    fsOp(side, { op: "fs_delete", paths: rows.map((r) => r.file_path) })
}

const createDir = async (side) => {
    closeMenu()
    let name
    try {
        const res = await ElMessageBox.prompt("请输入文件夹名称", `在${side === "local" ? "本地" : "远程"}目录下新建`, {
            confirmButtonText: "创建",
            cancelButtonText: "取消",
            inputPattern: /^[^\\/:*?"<>|]+$/,
            inputErrorMessage: '名称不能包含 \\ / : * ? " < > |',
        })
        name = res.value
    } catch (err) {
        return
    }
    if (!name) return
    fsOp(side, { op: "fs_mkdir", target: joinPath(sides[side].path, name) })
}
//#endregion -----------------------------------------------------------

const onKeydown = (e) => {
    if (e.key === "Escape") {
        closeMenu()
        return
    }
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "a") {
        const tag = (e.target?.tagName || "").toLowerCase()
        if (tag === "input" || tag === "textarea") return
        e.preventDefault()
        toggleAll(activeSide.value, true)
    }
}

onMounted(() => {
    window.addEventListener("keydown", onKeydown)
    window.addEventListener("click", closeMenu)
    window.addEventListener("resize", closeMenu)
    window.addEventListener("blur", closeMenu)
    // 恢复上次没传完的队列，等数据通道建立后自动续传
    restore()
    const device = route.query?.device
    const code = route.query?.code
    setTimeout(() => {
        fileSocket.value = createSocket("/file_connect");
        fileSocket.value.onopen = () => {
            send({ op: "connection", device, code })
        }
        fileSocket.value.onmessage = (msg) => handleMessage(JSON.parse(msg.data))
        fileSocket.value.onclose = () => {
            ElMessageBox.alert('检测到连接断开!', '连接断开', {
                confirmButtonText: '确定',
                showClose: false,
                callback: () => ipcRenderer.send("closeFileDialog"),
            })
        }
        fileSocket.value.onerror = (err) => {
            console.log(err);
        }
    }, 1000);
})

onUnmounted(() => {
    window.removeEventListener("keydown", onKeydown)
    window.removeEventListener("click", closeMenu)
    window.removeEventListener("resize", closeMenu)
    window.removeEventListener("blur", closeMenu)
    if (fileSocket.value) {
        fileSocket.value.close()
        fileSocket.value = null
    }
})
</script>

<style lang="scss" scoped>
.page {
    height: 100vh;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    overflow-y: auto;
    background: var(--bg);
}

/* ---------------- 顶部 ---------------- */
.toolbar {
    display: flex;
    align-items: flex-end;
    gap: 16px;
    padding: 14px 16px;

    .side {
        flex: 1;
        min-width: 0;

        .side-head {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 10px;

            .badge {
                flex-shrink: 0;
                padding: 2px 9px;
                font-size: 11px;
                font-weight: 600;
                border-radius: 20px;

                &.local {
                    color: #2f7f43;
                    background: #e8f6ec;
                }

                &.remote {
                    color: #1f6fb8;
                    background: #e8f2fd;
                }
            }

            .crumbs {
                flex: 1;
                min-width: 0;
                display: flex;
                align-items: center;
                overflow-x: auto;
                white-space: nowrap;
                scrollbar-width: none;

                &::-webkit-scrollbar {
                    display: none;
                }

                .crumb {
                    flex-shrink: 0;
                    padding: 3px 6px;
                    font-size: 12px;
                    color: var(--text-secondary);
                    border-radius: 5px;
                    cursor: pointer;
                    transition: background .2s, color .2s;

                    &+.crumb::before {
                        content: "/";
                        margin-right: 8px;
                        color: #cfd6e4;
                    }

                    &:hover {
                        color: var(--primary);
                        background: #eef4fd;
                    }

                    &.last {
                        color: var(--text);
                        font-weight: 600;
                        cursor: default;
                        background: transparent;
                    }
                }
            }

            .icons {
                flex-shrink: 0;
                display: flex;
                gap: 4px;

                .el-icon {
                    padding: 4px;
                    font-size: 13px;
                    border-radius: 6px;
                    color: var(--text-secondary);
                    cursor: pointer;
                    transition: background .2s, color .2s;

                    &:hover {
                        color: var(--primary);
                        background: #eef4fd;
                    }
                }
            }
        }

        .el-select {
            width: 100%;
        }
    }

    .ops {
        display: flex;
        flex-direction: column;
        gap: 8px;
        padding-bottom: 1px;

        .el-button {
            width: 96px;
            margin-left: 0;
        }
    }
}

.lists {
    display: flex;
    gap: 12px;

    .panel {
        flex: 1;
        min-width: 0;
    }
}

/* ---------------- 面板 ---------------- */
.panel {
    padding: 12px 14px;

    .panel-head {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 10px;

        .el-checkbox {
            margin-right: -2px;
        }

        .dot {
            width: 6px;
            height: 6px;
            flex-shrink: 0;
            border-radius: 50%;

            &.local {
                background: var(--success);
            }

            &.remote {
                background: var(--primary);
            }

            &.transfer {
                background: #e6a23c;
            }
        }

        .panel-title {
            font-size: 13px;
            font-weight: 600;
        }

        .count {
            padding: 1px 7px;
            font-size: 11px;
            color: var(--text-secondary);
            background: #f0f2f5;
            border-radius: 20px;
        }

        .tip {
            margin-left: auto;
            font-size: 11px;
            color: var(--text-secondary);
        }
    }

    .file-cell {
        display: flex;
        align-items: center;
        gap: 6px;
        min-width: 0;

        img {
            width: 14px;
            flex-shrink: 0;
        }

        span {
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
        }
    }

    .dir-tag {
        padding: 1px 8px;
        font-size: 11px;
        border-radius: 20px;

        &.upload {
            color: #1f6fb8;
            background: #e8f2fd;
        }

        &.download {
            color: #2f7f43;
            background: #e8f6ec;
        }
    }

    .status {
        display: flex;
        align-items: center;
        gap: 8px;

        .el-progress {
            flex: 1;
            min-width: 0;
        }

        .status-text {
            flex-shrink: 0;
            width: 54px;
            font-size: 11px;

            &.s0 {
                color: #909399;
            }

            &.s1 {
                color: var(--primary);
            }

            &.s2 {
                color: var(--success);
            }

            &.s3 {
                color: var(--danger);
            }
        }
    }

    .link {
        font-size: 12px;
        color: var(--primary);
        cursor: pointer;

        &:hover {
            text-decoration: underline;
        }
    }

    .ops-cell {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 10px;

        .el-icon {
            font-size: 14px;
            color: var(--text-secondary);
            cursor: pointer;
            transition: color .2s;

            &:hover {
                color: var(--primary);
            }
        }
    }

    .muted {
        color: #c4cad6;
    }
}

/* ---------------- 表格 ---------------- */
:deep(.el-table) {
    background: transparent;

    .cell {
        font-size: 12px;
        line-height: 20px;
    }

    .el-table__cell {
        padding: 6px 0;
    }

    .el-table__body td.el-table__cell {
        border-bottom: 0;
    }

    .el-table__inner-wrapper::before {
        display: none;
    }

    .el-table__empty-text {
        font-size: 12px;
        color: var(--text-secondary);
    }

    /* 多选高亮 */
    .el-table__row.is-selected>td.el-table__cell {
        background: #ecf5ff;
    }

    .el-table__row.is-selected:hover>td.el-table__cell {
        background: #e0eefd;
    }
}
/* ---------------- 右键菜单 ---------------- */
.ctx-menu {
    position: fixed;
    z-index: 3000;
    min-width: 152px;
    padding: 5px;
    background: #fff;
    border: 1px solid var(--border);
    border-radius: 10px;
    box-shadow: 0 8px 28px rgba(16, 24, 40, .16);
    user-select: none;

    .ctx-item {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 10px;
        font-size: 13px;
        color: var(--text);
        border-radius: 6px;
        cursor: pointer;
        transition: background .15s, color .15s;

        .el-icon {
            font-size: 14px;
            color: var(--text-secondary);
            transition: color .15s;
        }

        &:hover {
            background: #f2f5fa;

            .el-icon {
                color: var(--primary);
            }
        }

        &.danger {
            color: var(--danger);

            .el-icon {
                color: var(--danger);
            }

            &:hover {
                background: #fef0f0;
            }
        }

        &.disabled {
            color: #c4cad6;
            cursor: not-allowed;

            .el-icon {
                color: #c4cad6;
            }

            &:hover {
                background: transparent;
            }
        }
    }

    .ctx-divider {
        height: 1px;
        margin: 5px 6px;
        background: var(--border);
    }
}

</style>
