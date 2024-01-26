<template>
    <div class="progress">
        <el-progress type="circle" :colors="colors" :width="125" :percentage="percentage" />
        <div class="title">
            <p>已进入安全加密通道</p>
            <p>已通过身份验证</p>
            正在初始化相关数据.... <el-icon class="is-loading">
                <Loading />
            </el-icon>
        </div>
    </div>
</template>

<script setup>
const colors = [
    { color: '#f56c6c', percentage: 20 },
    { color: '#e6a23c', percentage: 40 },
    { color: '#5cb87a', percentage: 60 },
    { color: '#1989fa', percentage: 80 },
    { color: '#6f7ad3', percentage: 100 },
]
const emit = defineEmits(['showVideo'])
const timer = ref(null)
const percentage = ref(0)
onMounted(() => {

    timer.value = setInterval(() => {
        var randomNumber = Math.floor(Math.random() * 20) + 5;
        if (percentage.value + randomNumber >= 100) {
            percentage.value = 100
            clearInterval(timer.value)
            timer.value = null
            emit("showVideo")

        } else {
            percentage.value += randomNumber
        }
    }, 500);
})
onUnmounted(() => {
    if (timer.value) {
        clearInterval(timer.value)
        timer.value = null
    }
})
</script>

<style lang="scss" scoped>
.progress {
    -webkit-app-region: no-drag;
    position: absolute;
    top: 50%;
    left: 50%;
    -webkit-transform: translate(-50%, -50%);
    -moz-transform: translate(-50%, -50%);
    -ms-transform: translate(-50%, -50%);
    -o-transform: translate(-50%, -50%);
    transform: translate(-50%, -50%);

    .title {
        margin-top: 50px;
        text-align: center;
        font-size: 20px;
        color: #ced6e0;
        line-height: 30px;
    }

}

:deep(.el-progress-circle) {
    width: 20vw !important;
    height: initial !important;

}

:deep(.el-progress__text) {
    color: #ced6e0;
    font-size: 20px !important;
}
</style>