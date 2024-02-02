<template>
    <div class="floatingBall">
        <img src="@/assets/logo.png" alt="" @dragstart="event => event.preventDefault()">


    </div>
</template>

<script setup>
const dropdown1 = ref()
const remote = window.require("electron").remote;
const ipcRenderer = window.require("electron").ipcRenderer;
const win = remote.getCurrentWindow();

onMounted(() => {
    //开启剪贴板监听
    ipcRenderer.send('startWatchClipboard', false);
    let biasX = 0;
    let biasY = 0;
    document.addEventListener('mousedown', function (e) {
        switch (e.button) {
            case 0:
                biasX = e.x;
                biasY = e.y;
                document.addEventListener('mousemove', moveEvent);
                break;
            case 2:
                ipcRenderer.send('createSuspensionMenu');
                break;
        }
    });
    document.addEventListener('mouseup', function () {
        biasX = 0;
        biasY = 0;
        document.removeEventListener('mousemove', moveEvent)
    });
    function moveEvent(e) {
        win.setPosition(e.screenX - biasX, e.screenY - biasY)
    }


})
onUnmounted(() => {
    //开启剪贴板监听
    ipcRenderer.send('stopWatchClipboard', false);
    window.removeEventListener("blur", () => {
        if (!dropdown1.value) return
        dropdown1.value.handleClose()
    });
})
</script>

<style lang="scss" scoped>
.floatingBall {
    width: 80px;
    height: 80px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    /* 初始边框颜色设为透明 */
    user-select: none;
    // -webkit-app-region: drag;

    img {
        -webkit-app-region: no-drag;
        cursor: pointer;
        width: 50px;
        height: 50px;
    }
}
</style>