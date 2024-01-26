<template>
    <div class="device-list">
        <div class="title">
            高级设置
        </div>
        <div class="settings">
            <div class="item">
                <div class="item-title">
                    协议:
                </div>
                <div>
                    <el-select v-model="formData.https" class="m-2" placeholder="Select" size="large">
                        <el-option label="HTTP" :value="false" />
                        <el-option label="HTTPS" :value="true" />
                    </el-select>
                </div>
            </div>
            <div class="item">
                <div class="item-title">
                    服务器地址:
                </div>
                <div>
                    <el-input v-model="formData.remote_url" placeholder="请输入服务器地址" style="height: 50px;" />
                </div>
            </div>
            <div class="item">
                <el-button type="primary" style="width:100%;height:40px" @click="SaveData">保存</el-button>
            </div>

        </div>
    </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'

const deviceList = ref([])
let formData = ref({
    https: false,
    remote_url: "127.0.0.1:9998"
})
const defaultProps = {
    children: 'children',
    label: 'label',
    value: "value"
}
import { useDeviceStore } from "@/store/index";
const device = useDeviceStore();
const SaveData = () => {
    if (!formData.value.remote_url) {
        return ElMessage({
            showClose: true,
            message: '请输入IP地址',
            duration: 1500,
            type: 'warning',
        })
    }
    localStorage.setItem("serverInfo", JSON.stringify(formData.value))
    ElMessageBox.alert('修改成功', '成功', {
        confirmButtonText: '确定',
        showClose: false,
        callback: () => {
            window.location.reload();
        },

    })
}
onMounted(() => {
    formData.value = JSON.parse(JSON.stringify(device.serverInfo))
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

    .settings {
        padding: 0 100px 0 40px;
        display: flex;
        flex-direction: column;
        gap: 20px;

        .item {
            position: relative;
            margin: 10px;

            .item-title {

                margin-bottom: 20px;
            }
        }
    }
}
</style>