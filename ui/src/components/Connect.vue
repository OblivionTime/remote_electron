<template>
    <div class="title">
        允许控制本设备
    </div>
    <div class="info">
        <div class="info-left">
            <div class="info-title">本设备识别码<el-icon v-show="device.deviceInfo.identificationCode"
                    style="margin-left:5px; cursor: pointer;" @click="copyText">
                    <DocumentCopy />
                </el-icon></div>
            <div class="info-desc">{{ device.deviceInfo.identificationCode ? device.deviceInfo.identificationCode :
                '-------' }}
            </div>
        </div>
        <div class="info-right">
            <div class="info-title">验证码<el-icon v-show="device.deviceInfo.verificationCode"
                    style="margin-left:5px; cursor: pointer;">
                    <View v-if="hide" @click="hide = false" />
                    <Hide v-if="!hide" @click="hide = true" />
                </el-icon></div>
            <div class="info-desc">{{ !hide && device.deviceInfo.verificationCode ? device.deviceInfo.verificationCode :
                '------' }}
            </div>
        </div>
    </div>
    <div class="remote">
        <div class="remote-title">
            远程控制设备
        </div>
        <div class="remote-input">
            <div class="input">
                <el-input v-model="formData.identificationCode" oninput="value=value.replace(/[^0-9]/g, '')"
                    placeholder="请输入识别码" style="height:50px;flex:3" maxlength="9">
                </el-input>
                <el-input v-model="formData.verificationCode" placeholder="请输入验证码" style="height:50px;flex:1"
                    maxlength="6">
                </el-input>
            </div>
            <div class="btn">
                <el-button :disabled="!formData.identificationCode || !formData.verificationCode" type="primary"
                    style="height:50px;width:120px" @click="Connect" :loading="tryConnect">连接</el-button>
            </div>
        </div>
        <div style="margin-top:10px">
            <el-radio-group v-model="mode">
                <el-radio :label="'remote'">远程控制</el-radio>
                <el-radio :label="'file'">文件传输</el-radio>
            </el-radio-group>
        </div>

    </div>
</template>

<script setup>
import { Eleme } from '@element-plus/icons-vue'

import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import clipboard3 from "vue-clipboard3"
const { toClipboard } = clipboard3();
import { useDeviceStore } from "@/store/index";
const device = useDeviceStore();
const router = useRouter()
import { ConnectRemoteDevice } from '@/api/info.js';
let hide = ref(false)
let tryConnect = ref(false)
const formData = ref({
    identificationCode: "",
    verificationCode: ""
})
let mode = ref("remote")
const Connect = () => {
    if (formData.value.identificationCode?.length != 9 || formData.value.verificationCode?.length != 6) {
        return ElMessage({
            showClose: true,
            message: '识别码必须9位,验证码必须六位!!',
            duration: 1500,
            type: 'warning',
        })
    }
    if (formData.value.identificationCode?.length == device.deviceInfo.identificationCode) {
        return ElMessage({
            showClose: true,
            message: '不能输入自己的识别码!!!',
            duration: 1500,
            type: 'warning',
        })
    }
    tryConnect.value = true

    ConnectRemoteDevice(formData.value)
        .then((res) => {
            if (res.code == 0) {
                if (mode.value == "remote") {
                    router.push({ path: "/remote", query: { device: formData.value.identificationCode, code: formData.value.verificationCode } })
                } else {
                    ipcRenderer.send("createFileWindow", { ...formData.value })
                }
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
        })
        .finally(() => {
            tryConnect.value = false
        })

}
const ipcRenderer = window.require('electron').ipcRenderer;

const copyText = () => {
    toClipboard(device.deviceInfo.identificationCode).then(() => {
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
</script>

<style lang="scss" scoped>
.title {
    -webkit-app-region: no-drag;
    padding: 12vh 100px 5px 40px;
    font-size: 22px;
    margin-bottom: 50px;
}

.info {
    -webkit-app-region: no-drag;
    padding: 0 100px 0 40px;
    display: flex;
    align-items: center;
    justify-content: space-between;

    .info-left {
        .info-title {
            color: #999999;
            font-size: 14px;
            margin-bottom: 20px;
        }

        .info-desc {
            font-size: 25px;
            letter-spacing: 2px;
            font-weight: bold;
            color: #67C23A;
        }
    }

    .info-right {
        width: 120px;

        .info-title {
            color: #999999;
            font-size: 14px;
            margin-bottom: 20px;
        }

        .info-desc {
            font-size: 25px;
            letter-spacing: 2px;
            font-weight: bold;
        }
    }
}

.remote {
    padding: 12vh 20px 5px 40px;
    -webkit-app-region: no-drag;

    .remote-title {
        font-size: 22px;
        margin-bottom: 20px;
    }

    .remote-input {
        display: flex;
        gap: 10px;
        height: 50px;

        .input {
            flex: 6;
            height: 50px;
            display: flex;
        }

        .btn {
            flex: 1;

        }
    }
}
</style>