<template>
    <div class="device-page">
        <!-- 左：设备列表 -->
        <div class="card list-card">
            <div class="card-title">设备列表</div>

            <div class="list" v-if="devices.length">
                <div class="device" v-for="item in devices" :key="item.label"
                    :class="{ active: DeviceDetail.device_id === item.label }" @click="selectDevice(item)">
                    <div class="avatar">{{ item.label.slice(0, 1).toUpperCase() }}</div>
                    <div class="meta">
                        <div class="name">{{ item.label }}</div>
                        <div class="sub">{{ item.value.note || '未备注' }}</div>
                    </div>
                </div>
            </div>

            <div class="empty" v-else>
                <el-icon>
                    <Box />
                </el-icon>
                <p>还没有连接过设备</p>
            </div>
        </div>

        <!-- 右：设备详情 -->
        <div class="card detail-card">
            <template v-if="DeviceDetail.device_id">
                <div class="head">
                    <div class="avatar lg">{{ DeviceDetail.device_id.slice(0, 1).toUpperCase() }}</div>
                    <div class="head-meta">
                        <div class="name">{{ DeviceDetail.note || DeviceDetail.device_id }}</div>
                        <div class="sub">识别码 {{ DeviceDetail.identificationCode }}</div>
                    </div>
                </div>

                <div class="rows">
                    <div class="row" @click="copy(DeviceDetail.identificationCode)">
                        <span class="label">设备识别码</span>
                        <span class="value">
                            {{ DeviceDetail.identificationCode }}
                            <el-icon>
                                <DocumentCopy />
                            </el-icon>
                        </span>
                    </div>
                    <div class="row" @click="DeviceDetail.hide = !DeviceDetail.hide">
                        <span class="label">验证码</span>
                        <span class="value">
                            {{ DeviceDetail.hide ? '••••••' : DeviceDetail.verificationCode }}
                            <el-icon>
                                <View v-if="DeviceDetail.hide" />
                                <Hide v-else />
                            </el-icon>
                        </span>
                    </div>
                    <div class="row" @click="copy(DeviceDetail.note)">
                        <span class="label">备注</span>
                        <span class="value">
                            {{ DeviceDetail.note || '未设置' }}
                            <el-icon v-if="DeviceDetail.note">
                                <DocumentCopy />
                            </el-icon>
                        </span>
                    </div>
                </div>

                <div class="actions">
                    <el-button type="primary" @click="StartRemote">远程控制</el-button>
                    <el-button @click="StartFile">文件传输</el-button>
                </div>
            </template>

            <div class="empty" v-else>
                <el-icon>
                    <Monitor />
                </el-icon>
                <p>从左侧选择一个设备</p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'
import clipboard3 from 'vue-clipboard3'
const { toClipboard } = clipboard3();
import { ipcRenderer } from '@/utils/electron'
import { ConnectRemoteDevice, GetConnectDevice } from '@/api/info.js';
import { useRouter } from 'vue-router'

const router = useRouter()

const devices = ref([])
const DeviceDetail = ref({
    device_id: "",
    identificationCode: "",
    note: "",
    verificationCode: "",
    hide: true,
})

const selectDevice = (item) => {
    DeviceDetail.value = {
        device_id: item.label,
        identificationCode: item.value.identificationCode,
        verificationCode: item.value.verificationCode,
        note: item.value.note,
        hide: true,
    }
}

const copy = (text) => {
    if (!text) return
    toClipboard(text).then(() => {
        ElMessage({ showClose: true, message: '已复制', duration: 1500, type: 'success' })
    }).catch(() => {
        ElMessage({ showClose: true, message: '复制失败，请手动复制', duration: 1500, type: 'warning' })
    })
}

const StartRemote = () => {
    ElMessageBox.confirm(`你确定要远程控制<span style='color:#F56C6C'>${DeviceDetail.value.device_id}</span>吗?`, "提示", {
        confirmButtonText: "确认",
        cancelButtonText: "取消",
        dangerouslyUseHTMLString: true,

    })
        .then(() => {

            ConnectRemoteDevice(DeviceDetail.value)
                .then((res) => {
                    if (res.code == 0) {
                        router.push({ path: "/remote", query: { device: DeviceDetail.value.identificationCode, code: DeviceDetail.value.verificationCode } })
                    } else {
                        ElMessageBox.alert(res.msg, '警告', {
                            confirmButtonText: '确定',
                            showClose: false,
                        })
                    }
                }).catch((err) => {
                    console.log(err);
                    ElMessageBox.alert('请检测网络是否连接,可以尝试重启应用', '出现错误', {
                        confirmButtonText: '确定',
                        showClose: false,

                    })
                });
        })
        .catch(() => { });
}
const StartFile = () => {
    ElMessageBox.confirm(`你确定要<span style='color:#F56C6C'>${DeviceDetail.value.device_id}</span>进行文件传输吗?`, "提示", {
        confirmButtonText: "确认",
        cancelButtonText: "取消",
        dangerouslyUseHTMLString: true,

    })
        .then(() => {

            ConnectRemoteDevice(DeviceDetail.value)
                .then((res) => {
                    if (res.code == 0) {
                        ipcRenderer.send("createFileWindow", { ...DeviceDetail.value })
                    } else {
                        ElMessageBox.alert(res.msg, '警告', {
                            confirmButtonText: '确定',
                            showClose: false,
                        })
                    }
                }).catch((err) => {
                    console.log(err);
                    ElMessageBox.alert('请检测网络是否连接,可以尝试重启应用', '出现错误', {
                        confirmButtonText: '确定',
                        showClose: false,

                    })
                });
        })
        .catch(() => { });
}

onMounted(() => {
    GetConnectDevice()
        .then((res) => {
            devices.value = (res.data || []).map((item) => ({
                label: item.device_id,
                value: {
                    identificationCode: item.identificationCode,
                    verificationCode: item.verificationCode,
                    note: item.note,
                },
            }))
        })
        .catch((err) => {
            console.log(err);
        });
})
</script>

<style lang="scss" scoped>
.device-page {
    display: flex;
    gap: 12px;
    height: 100%;
    min-height: 0;

    .avatar {
        width: 34px;
        height: 34px;
        flex-shrink: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 50%;
        font-size: 15px;
        font-weight: 600;
        color: var(--primary);
        background: #ecf5ff;
        transition: background .2s, color .2s;

        &.lg {
            width: 44px;
            height: 44px;
            font-size: 19px;
            color: #fff;
            background: linear-gradient(135deg, #6db4ff, var(--primary));
        }
    }

    .empty {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 10px;
        color: var(--text-secondary);
        font-size: 13px;

        .el-icon {
            font-size: 30px;
            opacity: .45;
        }
    }

    /* 左栏 */
    .list-card {
        width: 190px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        min-height: 0;

        .list {
            flex: 1;
            overflow-y: auto;
            display: flex;
            flex-direction: column;
            gap: 4px;
            margin: 0 -6px;

            .device {
                display: flex;
                align-items: center;
                gap: 10px;
                padding: 9px 10px;
                border-radius: var(--radius-sm);
                cursor: pointer;
                transition: background .2s;

                .meta {
                    min-width: 0;
                }

                .name {
                    font-size: 13px;
                    font-weight: 600;
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;
                }

                .sub {
                    font-size: 11px;
                    color: var(--text-secondary);
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;
                }

                &:hover {
                    background: #f2f5fa;
                }

                &.active {
                    background: #ecf5ff;

                    .avatar {
                        color: #fff;
                        background: var(--primary);
                    }

                    .name {
                        color: var(--primary);
                    }
                }
            }
        }
    }

    /* 右栏 */
    .detail-card {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;

        .head {
            display: flex;
            align-items: center;
            gap: 12px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--border);

            .head-meta {
                min-width: 0;
            }

            .name {
                font-size: 16px;
                font-weight: 700;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .sub {
                margin-top: 4px;
                font-size: 12px;
                color: var(--text-secondary);
                letter-spacing: 1px;
            }
        }

        .rows {
            padding: 6px 0;

            .row {
                display: flex;
                align-items: center;
                justify-content: space-between;
                gap: 12px;
                padding: 11px 10px;
                border-radius: var(--radius-sm);
                cursor: pointer;
                transition: background .2s;

                &:hover {
                    background: #f7f9fc;

                    .value .el-icon {
                        color: var(--primary);
                    }
                }

                .label {
                    flex-shrink: 0;
                    font-size: 12px;
                    color: var(--text-secondary);
                }

                .value {
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    min-width: 0;
                    font-size: 13px;
                    font-weight: 600;
                    letter-spacing: 1px;
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;

                    .el-icon {
                        flex-shrink: 0;
                        font-size: 14px;
                        letter-spacing: 0;
                        color: var(--text-secondary);
                        transition: color .2s;
                    }
                }
            }
        }

        .actions {
            margin-top: auto;
            padding-top: 16px;
            display: flex;
            gap: 10px;

            .el-button {
                flex: 1;
                margin-left: 0;
            }
        }
    }
}
</style>
