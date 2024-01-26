<template>
    <div class="device-list">
        <div class="title">
            设备列表
        </div>
        <div class="tree">
            <el-tree :data="deviceList" :props="defaultProps" />

        </div>
    </div>
</template>

<script setup>
const deviceList = ref([])

const defaultProps = {
    children: 'children',
    label: 'label',
    value: "value"
}
import { useDeviceStore } from "@/store/index";
const device = useDeviceStore();
onMounted(() => {
    let list = device.deviceInfo.connectioned.map((item) => {
        return {
            label: item.device_id, value: {
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