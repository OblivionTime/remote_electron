<template>
    <div class="content">
        <ToolBar v-if="isShow" @myfresh="myfresh" @closeConnect="closeConnect"></ToolBar>
        <div class="myvideos" id="myvideos" ref="parent" :style="isShow ? '' : 'opacity:0;visibility:hidden'"
            @blur="release">
            <canvas id="action_area" ref="actionArea" @wheel="m_scroll" @mouseout="m_out" @mousedown="m_down"
                @mouseup="m_up" @mousemove="updatePosition" @mouseenter="m_enter"></canvas>
            <video id="other" ref="imgObj" autoplay></video>
        </div>

    </div>
    <Loading v-if="!isShow" @showVideo="showVideo"></Loading>
</template>

<script setup>
import Loading from '@/components/Loading.vue';
import ToolBar from '@/components/ToolBar.vue';
import { ElMessageBox } from "element-plus";
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
let globalDeviceInfo = ref({
    device: "",
    code: ""
})
let PeerConnection =
    window.PeerConnection ||
    window.webkitPeerConnection00 ||
    window.webkitRTCPeerConnection ||
    window.mozRTCPeerConnection;
let nativeRTCIceCandidate =
    window.mozRTCIceCandidate || window.RTCIceCandidate;
let nativeRTCSessionDescription =
    window.mozRTCSessionDescription || window.RTCSessionDescription;
//ice服务器地址
let iceServer = {
    iceServers: [],
};
let PC = ref(null)
//初始化PC源
function initPC() {
    let pc = new PeerConnection(iceServer);
    pc.onicecandidate = (evt) => {
        if (evt.candidate) {
            operateSocket.value.send(
                JSON.stringify({
                    name: `ice_candidate`,
                    data: {
                        id: evt.candidate.sdpMid,
                        label: evt.candidate.sdpMLineIndex,
                        sdpMLineIndex: evt.candidate.sdpMLineIndex,
                        candidate: evt.candidate.candidate,
                    },
                    device: globalDeviceInfo.device,
                    videoSender: true
                })
            );
        }
    };
    pc.onaddstream = (evt) => {

        let stream = evt.stream
        let video = document.getElementById('other')
        video.srcObject = stream
    };
    return pc
}
const initVideoSocket = () => {
    video_resize();
    operateSocket.value = new WebSocket(
        `ws://${import.meta.env.VITE_API_URL}/v1/api/remote/server/video_connect?device=${globalDeviceInfo.device}&code=${globalDeviceInfo.code}`
        // `ws://${import.meta.env.VITE_API_URL}/v1/api/remote/server/video_connect?device=829065585&code=W9jtX3`
    );
    operateSocket.value.onmessage = (msg) => {
        let data = JSON.parse(msg.data)
        switch (data.operation) {
            case "ice_server":
                iceServer.iceServers = data.iceservers
                PC.value = initPC()
                console.log(iceServer);
                break
            case "offer":

                //当收到对方接收请求后,设置音频源,并发送answer给对方
                PC.value.setRemoteDescription(new nativeRTCSessionDescription(data.data.sdp));
                PC.value.createAnswer((session_desc) => {
                    PC.value.setLocalDescription(session_desc);
                    operateSocket.value.send(
                        JSON.stringify({
                            name: "answer",
                            data: {
                                sdp: session_desc,
                            },
                            device: globalDeviceInfo.device,
                            videoSender: true
                        })
                    )

                }, (err) => {
                    console.log(err);
                })
                break;
            case "ice_candidate":
                //添加ice源,这一步很重要,如果没有接收ice则查看是否流程有问题
                var candidate = new nativeRTCIceCandidate(data.data);
                PC.value.addIceCandidate(candidate);
                break;
            case "disconnected":
                ElMessageBox.alert('对方断开连接', '连接断开', {
                    confirmButtonText: '确定',
                    showClose: false,
                    callback: () => {
                        release();
                        router.push("/");
                    },

                })
                break
        }
    };
    operateSocket.value.onerror = (err) => {
        console.log(err);
        if (operateSocket.value) {
            operateSocket.value = null
        }

    }
    operateSocket.value.onclose = () => {
        if (operateSocket.value) {
            operateSocket.value = null
        }
    }
};
//断开连接
const closeConnect = () => {
    ElMessageBox.confirm("确认退出吗?", "提示", {
        confirmButtonText: "确认",
        cancelButtonText: "取消",
    })
        .then(() => {
            release()
            if (operateSocket.value) {
                operateSocket.value.send(JSON.stringify({
                    "name": "disconnected",
                    device: globalDeviceInfo.device,
                    videoSender: true
                }))
            }

            router.push("/");
        })
        .catch(() => { });
}

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
    keyboardSocket.value.onerror = (err) => {
        console.log(err);

        ElMessageBox.alert("连接失败", '警告', {
            confirmButtonText: '确定',
            showClose: false,
            callback: () => {

                router.push("/");
            }
        })
    }
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
import { useRoute, useRouter } from "vue-router";
const route = useRoute()
const router = useRouter()
onMounted(() => {
    win.setSize(window.screen.availWidth, window.screen.availHeight);
    win.setPosition(0, 0, 0)
    window.addEventListener("resize", video_resize);
    globalDeviceInfo.device = route.query?.device
    globalDeviceInfo.code = route.query?.code
    setTimeout(() => {
        initVideoSocket()
        setTimeout(() => {
            initBlueToothSocket()
        }, 1000);
    }, 500);

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
    // -webkit-app-region: drag;
    width: 100vw;
    height: 100vh;
    position: relative;
    z-index: 99;
    display: flex;
    align-items: center;
    justify-content: center;
}

#action_area {
    -webkit-app-region: no-drag;
    position: absolute;
    user-select: none;
    z-index: 919;
    border-radius: 5px;
}

#other {
    -webkit-app-region: no-drag;
    border-radius: 5px;
    z-index: 98;
    object-fit: fill;
}

.content {
    width: 100vw;
    height: 100vh;
    background: #303133;
    -webkit-app-region: drag;
}
</style>