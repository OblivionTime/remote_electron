<template>
    <div class="card">
        <div class="card-title">本设备</div>
        <div class="info">
            <div class="info-item">
                <div class="label">设备识别码</div>
                <div class="value">
                    {{ device.deviceInfo.identificationCode || '-------' }}
                    <el-icon v-show="device.deviceInfo.identificationCode" @click="copyText">
                        <DocumentCopy />
                    </el-icon>
                </div>
            </div>
            <div class="info-item">
                <div class="label">验证码</div>
                <div class="value">
                    {{ !hide && device.deviceInfo.verificationCode ? device.deviceInfo.verificationCode : '------' }}
                    <el-icon v-show="device.deviceInfo.verificationCode" @click="hide = !hide">
                        <View v-if="hide" />
                        <Hide v-else />
                    </el-icon>
                </div>
            </div>
        </div>
    </div>

    <div class="card">
        <div class="card-title">远程控制设备</div>
        <div class="form">
            <el-input v-model="formData.identificationCode" oninput="value=value.replace(/[^0-9]/g, '')"
                placeholder="请输入识别码" maxlength="9" />
            <el-input v-model="formData.verificationCode" placeholder="请输入验证码" maxlength="6" />
            <el-button :disabled="!formData.identificationCode || !formData.verificationCode" type="primary"
                :loading="tryConnect" @click="Connect">连接</el-button>
        </div>
        <el-radio-group v-model="mode" class="mode">
            <el-radio :label="'remote'">远程控制</el-radio>
            <el-radio :label="'file'">文件传输</el-radio>
        </el-radio-group>
    </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import clipboard3 from "vue-clipboard3"
const { toClipboard } = clipboard3();
import { useDeviceStore } from "@/store/index";
const device = useDeviceStore();
import { ipcRenderer } from "@/utils/electron";
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
    if (formData.value.identificationCode == device.deviceInfo.identificationCode) {
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
.info {
    display: flex;
    gap: 12px;

    .info-item {
        flex: 1;
        min-width: 0;
        padding: 14px 16px;
        background: #f7f9fc;
        border: 1px solid var(--border);
        border-radius: var(--radius-sm);

        .label {
            font-size: 12px;
            color: var(--text-secondary);
            margin-bottom: 8px;
        }

        .value {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 22px;
            font-weight: 700;
            letter-spacing: 2px;
            color: var(--text);

            .el-icon {
                font-size: 15px;
                letter-spacing: 0;
                cursor: pointer;
                color: var(--text-secondary);
                transition: color .2s;

                &:hover {
                    color: var(--primary);
                }
            }
        }
    }
}

.form {
    display: flex;
    gap: 10px;

    .el-input:nth-child(1) {
        flex: 3;
    }

    .el-input:nth-child(2) {
        flex: 1;
    }

    .el-button {
        width: 110px;
        flex-shrink: 0;
    }
}

.mode {
    margin-top: 16px;
}
</style>
