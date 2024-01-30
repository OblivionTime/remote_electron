<template>
    <div class="device-list">
        <div class="title">
            设备列表
        </div>
        <div class="tree">
            <el-tree :data="deviceList" :props="defaultProps" @nodeClick="HandlerNode" />

        </div>
    </div>
</template>

<script setup>
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
import { GetConnectDevice } from '@/api/info.js';
const HandlerNode = (item) => {
    if (item.value == "") {
        return
    }
    ElMessageBox.confirm(`你确定要连接<span style='color:#F56C6C'>${item.label}</span>吗?`, "提示", {
        confirmButtonText: "确认",
        cancelButtonText: "取消",
        dangerouslyUseHTMLString: true,

    })
        .then(() => {
            ConnectRemoteDevice(item.value)
                .then((res) => {
                    if (res.code == 0) {
                        router.push({ path: "/remote", query: { device: item.value.identificationCode, code:item.value.verificationCode } })
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
    console.log(item);
}
onMounted(() => {
    GetConnectDevice()
        .then((res) => {
            try {
                let connectioned = JSON.parse(res.data)

                let list = connectioned.map((item) => {
                    return {
                        label: item.identificationCode, value: {
                            identificationCode: item.identificationCode,
                            verificationCode: item.verificationCode,
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

    .tree {
        padding: 0 100px 0 40px;
    }
}
</style>