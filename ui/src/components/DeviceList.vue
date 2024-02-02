<template>
    <div class="device-list">
        <div class="title">
            设备列表
        </div>
        <div class="device-show">
            <el-tree class="tree" :data="deviceList" :props="defaultProps" @nodeClick="HandlerNode" />
            <div class="divisionLine"></div>
            <div class="detail">
                <div class="detail-code"> {{ DeviceDetail.device_id }}</div>
                <div class="detail-items" v-if="DeviceDetail.device_id">
                    <div class="detail-item">
                        设备识别码
                        <p>
                            {{ DeviceDetail.identificationCode }}
                            <el-icon style="margin-left:5px; cursor: pointer;" @click="copyText">
                                <DocumentCopy />
                            </el-icon>
                        </p>
                    </div>
                    <div class="detail-item">
                        验证码
                        <p>
                            {{ DeviceDetail.hide ? "***********" : DeviceDetail.verificationCode }}
                            <el-icon style="margin-left:5px; cursor: pointer;" >
                                <Hide v-if="DeviceDetail.hide"  @click="DeviceDetail.hide = false"/>
                                <View v-if="!DeviceDetail.hide" @click="DeviceDetail.hide = true"/>
                            </el-icon>
                        </p>
                    </div>
                    <div class="detail-item">
                        备注
                        <p>
                            {{ DeviceDetail.note ? DeviceDetail.note : DeviceDetail.device_id }}
                            <el-icon style="margin-left:5px; cursor: pointer;" @click="copyText">
                                <Edit />
                            </el-icon>
                        </p>
                    </div>
                    <div class="detail-btn">
                        <el-button type="primary" @click="StartRemote" style=" width: 100%;">远程控制</el-button>

                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import RemoteIco from '@/assets/remote.svg';

const deviceList = ref([])
import { ElMessage, ElMessageBox } from 'element-plus'
import { ConnectRemoteDevice } from '@/api/info.js';

import { useRouter } from 'vue-router'

const router = useRouter()

const defaultProps = {
    children: 'children',
    label: 'label',
    value: "value"
}
const DeviceDetail = ref({
    device_id: "",
    identificationCode: "",
    note: "",
    verificationCode: "",
    hide: true,
})
import { GetConnectDevice } from '@/api/info.js';
const HandlerNode = (item) => {
    if (item.value == "") {
        return
    }
    DeviceDetail.value = {
        device_id: item.label,
        identificationCode: item.value.identificationCode,
        note: item.value.note,
        verificationCode: item.value.verificationCode,
        hide: true
    }
}
const StartRemote = () => {
    ElMessageBox.confirm(`你确定要连接<span style='color:#F56C6C'>${DeviceDetail.value.device_id}</span>吗?`, "提示", {
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
const copyText = () => {
    toClipboard(DeviceDetail.value.identificationCode).then(() => {
        // 复制成功时的回调
        ElMessage({
            showClose: true,
            message: '复制成功!!!',
            duration: 1500,
            type: 'success',
        })
    }).catch(() => {
        ElMessage({
            showClose: true,
            message: '复制失败，请手动复制!!',
            duration: 1500,
            type: 'warning',
        })
        // 复制失败时的回调
    });
}
onMounted(() => {
    GetConnectDevice()
        .then((res) => {
            try {
                let connectioned = res.data

                let list = connectioned.map((item) => {
                    return {
                        label: item.device_id, value: {
                            identificationCode: item.identificationCode,
                            verificationCode: item.verificationCode,
                            note: item.note
                        }, children: []
                    }
                })
                if (list.length == 0) {
                    list = [{
                        label: "你还未连接设备",
                        value: '',
                        children: []
                    }]
                }
                deviceList.value.push({
                    label: '连接过的设备',
                    value: "",
                    children: list,
                })
            } catch (error) {
                list = [{
                    label: "你还未连接设备",
                    value: '',
                    children: []
                }]
                deviceList.value.push({
                    label: '连接过的设备',
                    value: "",
                    children: list,
                })
            }

        }).catch((err) => {

        });

})
</script>

<style lang="scss" scoped>
.device-list {
    -webkit-app-region: no-drag;

    .title {
        padding: 5vh 100px 0px 40px;
        font-size: 22px;
        margin-bottom: 20px;
    }

    .device-show {
        padding: 0 40px 0 40px;
        display: flex;
        gap: 20px;
        height: 75vh;

        .tree {
            flex-basis: 200px;
        }

        .divisionLine {
            background: #747d8c;
            width: 1px;
            height: auto;
        }

        .detail {
            display: flex;
            flex-direction: column;
            gap: 10px;
            flex: auto;

            .detail-code {
                font-size: 18px;
                font-weight: bold;
            }

            .detail-items {
                .detail-item {
                    margin-top: 20px;
                    display: flex;
                    justify-content: space-between;
                    font-size: 14px;
                    color: #999999;
                    letter-spacing: 1px;
                }

                .detail-btn {
                    margin-top: 40px;
                    width: 100%;

                }
            }
        }
    }
}
</style>