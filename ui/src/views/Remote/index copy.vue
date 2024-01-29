<template>
    <div class="content">
        <ToolBar v-if="isShow" @myfresh="myfresh" :socket="keyboardSocket"></ToolBar>
        <div class="myvideos" id="myvideos" ref="parent" :style="isShow ? '' : 'opacity:0;visibility:hidden'"
            @blur="release">
            <canvas id="action_area" ref="actionArea" @wheel="m_scroll" @mouseout="m_out" @mousedown="m_down"
                @mouseup="m_up" @mousemove="updatePosition" @mouseenter="m_enter"></canvas>
            <canvas id="other" ref="imgObj"></canvas>
        </div>

    </div>
    <Loading v-if="!isShow" @showVideo="showVideo"></Loading>
</template>

<script setup>
import Loading from '@/components/Loading.vue';
import ToolBar from '@/components/ToolBar.vue';
const remote = window.require("electron").remote;
const win = remote.getCurrentWindow();
const isShow = ref(false)
const myfresh = () => {
    window.removeEventListener("blur", () => {
        release();
    });
    if (operateSocket.value) {
        operateSocket.value.close();
        operateSocket.value = null;
    }
    if (keyboardSocket.value) {
        keyboardSocket.value.close();
        keyboardSocket.value = null;
    }
};
const showVideo = () => {
    isShow.value = true
}
//#region 图片缩放 ------------------------------------------------------------
let attribute = ref({})
const video_resize = () => {
    let parent = document.getElementById("myvideos");
    if (!parent) {
        return
    }
    let disp_ratio = 16 / 8;
    let rw = 1;
    let rh = 1;
    let pre;
    let myvideos = parent.getBoundingClientRect();
    let width = parseInt(myvideos.width);
    let height = parseInt(myvideos.height);
    let wratio = width / height;
    if (disp_ratio == 0) {
        rw = 1;
        rh = 1;
    } else if (disp_ratio > wratio) {
        pre = (disp_ratio - wratio) * 50;
        rw = 1;
        rh = (100 - pre) * 0.01;
    } else {
        pre = (wratio - disp_ratio) * 50;
        rw = (100 - pre) * 0.01;
        rh = 1;
    }
    let o = document.getElementById("other");
    o.width = rw * width;
    o.height = rh * height;
    actionArea.value.width = o.width;
    actionArea.value.height = o.height;
    attribute.value.width = o.width;
    attribute.value.height = o.height;
};
//#endregion -----------------------------------------------------------
//#region 视频连接 ------------------------------------------------------------
const imgObj = ref(null);
const actionArea = ref(null);
const operateSocket = ref(null)
let ctx = ref(null);
let globalDeviceInfo = ref({
    device: "",
    code: ""
})
const initVideoSocket = () => {
    imgObj.value = document.getElementById("other");
    ctx.value = imgObj.value.getContext("2d");
    video_resize();
    operateSocket.value = new WebSocket(
        `ws://${import.meta.env.VITE_API_URL}/v1/api/remote/server/video_connect?device=${globalDeviceInfo.device}&code=${globalDeviceInfo.code}`
        // `ws://${import.meta.env.VITE_API_URL}/v1/api/remote/server/video_connect?device=829065585&code=W9jtX3`
    );
    operateSocket.value.onmessage = (msgs) => {
        try {
            const blob = new Blob([msgs.data]);
            let img = new Image();
            const objectURL = URL.createObjectURL(blob);
            img.src = objectURL;
            img.onload = () => {
                ctx.value.drawImage(img, 0, 0, attribute.value.width, attribute.value.height);
            };
        } catch (error) { }
    };
};


//#endregion -----------------------------------------------------------
//#region  ------------------------------------------------------------
let keyboardSocket = ref("");
//模式 键盘初始化连接
const initBlueToothSocket = () => {
    keyboardSocket.value = new WebSocket(
        `ws://${import.meta.env.VITE_API_URL}/v1/api/remote/server/keyboard_connect?device=${globalDeviceInfo.device}`
    );
    keyboardSocket.value.onmessage = (msgs) => {
        // console.log(msgs);
    };
};
//释放鼠标
const release = () => {
    let sendData = {
        operation: "change",
        status: false,
    };
    if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
        keyboardSocket.value.send(JSON.stringify(sendData));
    }
};
//鼠标抬起
const m_up = (e) => {
    let sendData = {
        operation: "send",
        data: {
            method: "MouseUp",
            arg1: e.which,
        },
    };
    if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
        keyboardSocket.value.send(JSON.stringify(sendData));
    }
};
//鼠标按下
const m_down = (e) => {
    let sendData = {
        operation: "send",
        data: {
            method: "MouseDown",
            arg1: e.which,
        },
    };
    if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
        keyboardSocket.value.send(JSON.stringify(sendData));
    }
};
//鼠标移除
const m_out = () => {
    release();
};
//更新位置
const updatePosition = (e) => {
    let tt = actionArea.value.getBoundingClientRect();
    let def_x = Math.floor(e.clientX - tt.left);
    let def_y = Math.floor(e.clientY - tt.top);
    if (def_y < 0) {
        def_y = 0;
    }
    if (def_x < 0) {
        def_x = 0;
    }
    if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
        keyboardSocket.value.send(
            JSON.stringify({
                operation: "send",
                data: {
                    method: "MouseMove",
                    arg1: def_x,
                    arg2: def_y,
                    width: parseInt(tt.width),
                    height: parseInt(tt.height),
                },
            })
        );
    }
};

//鼠标移入
const m_enter = () => {
    let sendData = {
        operation: "change",
        status: true,
    };
    if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
        keyboardSocket.value.send(JSON.stringify(sendData));
    }
};
//监听滚轮
const m_scroll = (e) => {
    if (e.deltaY > 0) {
        if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
            keyboardSocket.value.send(
                JSON.stringify({
                    operation: "send",
                    data: {
                        method: "ScrollMouse",
                        arg1: -1,
                    },
                })
            );
        }
    } else {
        if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
            keyboardSocket.value.send(
                JSON.stringify({
                    operation: "send",
                    data: {
                        method: "ScrollMouse",
                        arg1: 1,
                    },
                })
            );
        }
    }
};
//#endregion -----------------------------------------------------------
import { useRoute } from "vue-router";
const route = useRoute()
onMounted(() => {
    win.maximize()
    globalDeviceInfo.device = route.query?.device
    globalDeviceInfo.code = route.query?.code
    setTimeout(() => {
        initVideoSocket()
        initBlueToothSocket()
    }, 2000);

})
onUnmounted(() => {
    window.addEventListener("blur", () => {
        release();
    });
    if (operateSocket.value) {
        operateSocket.value.close();
        operateSocket.value = null;
    }
    if (keyboardSocket.value) {
        keyboardSocket.value.close()
        keyboardSocket.value = null
    }
})
</script>

<style lang="scss" scoped>
#myvideos {
    width: 100vw;
    height: 100vh;
    position: relative;
    z-index: 99;
    display: flex;
    align-items: center;
    justify-content: center;
}

#action_area {
    position: absolute;
    user-select: none;
    z-index: 99;
    border-radius: 5px;
}

#other {
    border-radius: 5px;
    z-index: 98;
}

.content {
    width: 100vw;
    height: 100vh;
    background: #303133;
    // -webkit-app-region: no-drag;
}
</style>