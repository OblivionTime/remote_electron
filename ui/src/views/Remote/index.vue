<template>
    <div class="content">
        <ToolBar v-if="isShow" @myfresh="myfresh" @closeConnect="closeConnect" @file="openFile"
            :quality="quality" :grayscale="grayscale" :aspect="aspect" @update:quality="setQuality"
            @update:grayscale="setGrayscale" @update:aspect="setAspect"></ToolBar>
        <div class="myvideos" id="myvideos" ref="parent" :style="isShow ? '' : 'opacity:0;visibility:hidden'"
            @blur="release">
            <div class="stage" :style="stageStyle">
                <video id="other" ref="imgObj" autoplay :class="{ grayscale }"
                    @loadedmetadata="video_resize"></video>
                <canvas id="action_area" ref="actionArea" @wheel="m_scroll" @mouseout="m_out" @mousedown="m_down"
                    @mouseup="m_up" @mousemove="updatePosition" @mouseenter="m_enter"></canvas>
                <!-- 隐藏本地光标，改画对端光标，和市面远控一致 -->
                <div class="remote-cursor" v-show="cursorVisible" :style="cursorStyle"></div>
            </div>
            <!-- 当前画面设置，便于确认设置是否生效 -->
            <div class="display-badge">
                {{ qualityLabel(quality) }}
                <span v-if="grayscale"> · 黑白</span>
                <span> · {{ aspectLabel(aspect) }}</span>
            </div>
        </div>

    </div>
    <Loading v-if="!isShow" @showVideo="showVideo"></Loading>
</template>

<script setup>
import Loading from '@/components/Loading.vue';
import ToolBar from '@/components/ToolBar.vue';
import { ElMessage, ElMessageBox } from "element-plus";
import { ipcRenderer, setWindowSize, setWindowPosition } from "@/utils/electron";
import { createSocket } from "@/utils/ws";
import {
    setIceServers, createPeer, pipeCandidates, acceptOffer,
    addRemoteCandidate, closePeer, watchState
} from "@/utils/video";
import { QUALITY_PRESETS, loadRemotePrefs, qualityLabel, aspectLabel } from "@/utils/display";
const isShow = ref(false)

const myfresh = () => {
    window.removeEventListener("blur", () => {
        release();
    });
    if (operateSocket.value) {
        operateSocket.value.close();
        serverClose = false
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
//#region 画面尺寸 / 比例 ------------------------------------------------------------
// stage 是"视频内容区"：canvas 和虚拟光标都相对它定位，
// 所以坐标映射和光标绘制天然对齐（比例模式也一致）。
const stageSize = reactive({ w: 0, h: 0 })
const stageStyle = computed(() => ({ width: stageSize.w + "px", height: stageSize.h + "px" }))

const video_resize = () => {
    const parent = document.getElementById("myvideos");
    const video = document.getElementById("other");
    if (!parent || !video) return

    const box = parent.getBoundingClientRect()
    let w = Math.floor(box.width)
    let h = Math.floor(box.height)

    if (aspect.value === "original") {
        // 保持原始比例：按流的分辨率等比缩放并居中
        const vw = video.videoWidth || 16
        const vh = video.videoHeight || 9
        const scale = Math.min(box.width / vw, box.height / vh)
        w = Math.floor(vw * scale)
        h = Math.floor(vh * scale)
    }

    stageSize.w = w
    stageSize.h = h
    if (actionArea.value) {
        actionArea.value.width = w
        actionArea.value.height = h
    }
}
//#endregion -----------------------------------------------------------

//#region 画面设置（画质 / 黑白 / 比例）------------------------------------------------------------
const prefs = loadRemotePrefs()
const quality = ref(prefs.quality)
const aspect = ref(prefs.aspect)
const grayscale = ref(false)

// 画质和黑白都必须作用在被控端，一起下发：
// 画质 → 被控端 RTCRtpSender 编码参数（真实降带宽）
// 黑白 → 被控端采集侧转灰度（改的是真正传出去的画面，顺带省掉色度编码）
const syncDisplay = () => {
    const preset = QUALITY_PRESETS[quality.value] || QUALITY_PRESETS.high
    if (!operateSocket.value || operateSocket.value.readyState !== 1) return
    operateSocket.value.send(JSON.stringify({
        name: "quality",
        data: {
            maxBitrate: preset.maxBitrate,
            maxFramerate: preset.maxFramerate,
            scale: preset.scale,
            grayscale: grayscale.value,
        },
        device: globalDeviceInfo.device,
        videoSender: true,
    }))
}
const setQuality = (v) => {
    quality.value = v
    syncDisplay()
}
const setGrayscale = (v) => {
    grayscale.value = v
    syncDisplay()
}
const setAspect = (v) => {
    aspect.value = v
    nextTick(video_resize)
}

//#region 虚拟光标 ------------------------------------------------------------
// 显示本地光标会和对端实际生效位置对不上（分辨率/DPI 不同），
// 所以隐藏本地箭头，改在画面上按"我们发过去的映射坐标"画一个光标。
const cursorVisible = ref(false)
const cursorPos = reactive({ x: 0, y: 0 })
const cursorStyle = computed(() => ({ left: cursorPos.x + "px", top: cursorPos.y + "px" }))
//#endregion -----------------------------------------------------------
//#endregion -----------------------------------------------------------
//#region 视频连接 ------------------------------------------------------------
const imgObj = ref(null);
const actionArea = ref(null);
const operateSocket = ref(null)
let globalDeviceInfo = ref({
    device: "",
    code: ""
})
let PC = ref(null)
//初始化PC源
function initPC() {
    const pc = createPeer()
    pc.ontrack = (evt) => {
        const video = document.getElementById('other')
        if (!video) {
            console.error("找不到 video 元素")
            return
        }
        video.srcObject = evt.streams[0] || new MediaStream([evt.track])
        video.play().catch((err) => console.error("视频播放失败", err))
        console.log("收到远端视频轨道", evt.track.kind)
    }
    watchState(pc, () => {
        ElMessageBox.alert('对方断开连接', '连接断开', {
            confirmButtonText: '确定',
            showClose: false,
            callback: () => {
                release();
                router.push("/index");
            },
        })
    })
    pipeCandidates(pc, (data) => {
        if (!operateSocket.value) return
        operateSocket.value.send(JSON.stringify({
            name: "ice_candidate",
            data,
            device: globalDeviceInfo.device,
            videoSender: true
        }))
    })
    return pc
}
const initVideoSocket = () => {
    video_resize();
    operateSocket.value = createSocket("/video_connect", {
        device: globalDeviceInfo.device,
        code: globalDeviceInfo.code,
    });
    operateSocket.value.onmessage = (msg) => {
        let data = JSON.parse(msg.data)
        switch (data.operation) {
            case "ice_server":
                setIceServers(data.iceservers)
                PC.value = initPC()
                break
            case "offer":
                //当收到对方接收请求后,设置音频源,并发送answer给对方
                if (!PC.value) {
                    // ice_server 还没到就收到 offer 时兜底，避免 PC.value 为空直接抛错
                    PC.value = initPC()
                }
                acceptOffer(PC.value, data.data.sdp)
                    .then((answer) => {
                        operateSocket.value.send(
                            JSON.stringify({
                                name: "answer",
                                data: { sdp: answer },
                                device: globalDeviceInfo.device,
                                videoSender: true
                            })
                        )
                        // 把当前的默认画质/黑白同步给被控端
                        syncDisplay()
                    }).catch((err) => {
                        console.error("处理 offer 失败", err);
                    })
                break;
            case "ice_candidate":
                //添加ice源,这一步很重要,如果没有接收ice则查看是否流程有问题
                addRemoteCandidate(PC.value, data.data)
                break;
            case "disconnected":
                ElMessageBox.alert('对方断开连接', '连接断开', {
                    confirmButtonText: '确定',
                    showClose: false,
                    callback: () => {
                        release();
                        router.push("/index");
                    },

                })
                break
            case "quality_ack":
                // 被控端回执：告诉用户到底生效没有，避免"设了没反应"的困惑
                if (data.data?.ok) {
                    ElMessage({ showClose: true, message: '画面设置已生效', duration: 1500, type: 'success' })
                } else {
                    ElMessage({ showClose: true, message: '被控端应用画面设置失败', duration: 2000, type: 'warning' })
                }
                if (grayscale.value && data.data && data.data.grayscale === false) {
                    ElMessage({ showClose: true, message: '被控端不支持源端黑白，已回退为本地黑白', duration: 2500, type: 'warning' })
                }
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
//打开文件传输窗口（已经开过会复用同一个窗口，不会重复开）
const openFile = () => {
    ipcRenderer.send("createFileWindow", {
        identificationCode: globalDeviceInfo.device,
        verificationCode: globalDeviceInfo.code,
    })
}

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

            router.push("/index");
        })
        .catch(() => { });
}

//#endregion -----------------------------------------------------------
//#region  ------------------------------------------------------------
let keyboardSocket = ref("");
var serverClose = true
//模式 键盘初始化连接
const initBlueToothSocket = () => {
    keyboardSocket.value = createSocket("/keyboard_connect", {
        device: globalDeviceInfo.device,
    });
    keyboardSocket.value.onopen = () => {
        //开启剪贴板监听
        ipcRenderer.send('startWatchClipboard', true);
        ipcRenderer.on("clipboard", (e, data) => {
            console.log("clipboard", e, data);
            if (keyboardSocket.value && keyboardSocket.value.readyState == 1) {
                keyboardSocket.value.send(JSON.stringify({
                    operation: "send",

                    data: {
                        method: "clipboard",
                        clipboard_type: data.clipboard_type,
                        clipboard_data: data.list,
                    },
                }));
            }
        })
    }


    keyboardSocket.value.onerror = (err) => {
        console.log(err);


    }
    operateSocket.value.onclose = () => {
        //关闭剪贴板监听
        ipcRenderer.send('stopWatchClipboard');
        if (operateSocket.value) {
            operateSocket.value = null
        }
        closePeer(PC.value)
        PC.value = null
        if (serverClose) {
            ElMessageBox.alert("服务器断开!!", '警告', {
                confirmButtonText: '确定',
                showClose: false,
                callback: () => {

                    router.push("/index");
                }
            })
        }

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
    cursorVisible.value = false;
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
    // 虚拟光标跟着走（对端的光标就在这个映射位置上）
    cursorPos.x = def_x;
    cursorPos.y = def_y;
    cursorVisible.value = true;
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
    setWindowSize(window.screen.availWidth, window.screen.availHeight);
    setWindowPosition(0, 0)
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
    window.removeEventListener("resize", video_resize);
    if (operateSocket.value) {
        operateSocket.value.close();
        serverClose = false
        operateSocket.value = null;
    }
    if (keyboardSocket.value) {
        keyboardSocket.value.close()
        keyboardSocket.value = null
    }
    closePeer(PC.value)
    PC.value = null

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
    /* 隐藏本地光标，改画对端光标（见 .remote-cursor） */
    cursor: none;

    .display-badge {
        position: absolute;
        left: 12px;
        bottom: 12px;
        z-index: 1300;
        padding: 4px 10px;
        font-size: 12px;
        color: #fff;
        background: rgba(0, 0, 0, .55);
        border-radius: 20px;
        backdrop-filter: blur(8px);
        pointer-events: none;
    }

    .stage {
        position: relative;

        #other {
            -webkit-app-region: no-drag;
            display: block;
            width: 100%;
            height: 100%;
            border-radius: 5px;
            object-fit: fill;

            &.grayscale {
                filter: grayscale(1);
            }
        }

        #action_area {
            -webkit-app-region: no-drag;
            position: absolute;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            user-select: none;
            z-index: 919;
            border-radius: 5px;
            cursor: none;
        }

        .remote-cursor {
            position: absolute;
            z-index: 1200;
            width: 20px;
            height: 20px;
            /* 箭头尖端在 SVG 的 (3,1)，把元素偏移过去让尖端对准坐标点 */
            margin: -1px 0 0 -3px;
            pointer-events: none;
            background-repeat: no-repeat;
            background-size: contain;
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='20' height='20' viewBox='0 0 20 20'%3E%3Cpath d='M3 1l12 9-5 .6 3 6.4-2.4 1.1-3-6.4L3 16z' fill='%23fff' stroke='%23000' stroke-width='1.1' stroke-linejoin='round'/%3E%3C/svg%3E");
            filter: drop-shadow(0 1px 2px rgba(0, 0, 0, .45));
        }
    }
}

.content {
    width: 100vw;
    height: 100vh;
    background: #303133;
    -webkit-app-region: drag;
}
</style>