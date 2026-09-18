<template>
    <div class="floatingBall">
        <img src="@/assets/logo.png" alt="" @dragstart="event => event.preventDefault()">


    </div>
</template>

<script setup>
import { ipcRenderer, setWindowPosition } from "@/utils/electron";
const dropdown1 = ref()

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
        setWindowPosition(e.screenX - biasX, e.screenY - biasY)
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
    background: #fff;
    border: 2px solid var(--primary);
    box-shadow: 0 4px 16px rgba(0, 0, 0, .18);
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