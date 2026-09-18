<template>
    <router-view></router-view>
</template>

<script setup>
import { ElMessageBox } from "element-plus";
import { useDeviceStore } from "@/store/index";
import { ConnectServer } from '@/api/info.js';
import { onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
const router = useRouter();

//#region 视频流 ------------------------------------------------------------


let selfPC = null
let videoSender = null
let rawVideoTrack = null
let grayPipeline = null

// 源端灰度：用 Insertable Streams 在采集侧把每一帧转成灰度，改的是真正传出去的画面。
// 不能用 canvas + requestAnimationFrame —— 被控期间主窗口是隐藏的，rAF 会被节流到 1fps
// （实测 2.5s 只画出 2 帧）。MediaStreamTrackProcessor 由流驱动，隐藏窗口下依然满帧。
const startGrayPipeline = () => {
    if (grayPipeline || !rawVideoTrack) return null
    try {
        const proc = new MediaStreamTrackProcessor({ track: rawVideoTrack })
        const gen = new MediaStreamTrackGenerator({ kind: "video" })
        const canvas = new OffscreenCanvas(2, 2)
        const ctx = canvas.getContext("2d")
        ctx.filter = "grayscale(1)"
        const transformer = new TransformStream({
            async transform(frame, controller) {
                try {
                    if (canvas.width !== frame.displayWidth || canvas.height !== frame.displayHeight) {
                        canvas.width = frame.displayWidth
                        canvas.height = frame.displayHeight
                    }
                    ctx.drawImage(frame, 0, 0)
                    controller.enqueue(new VideoFrame(canvas, { timestamp: frame.timestamp }))
                } catch (err) {
                    console.error("灰度转换失败", err)
                } finally {
                    frame.close()
                }
            },
        })
        proc.readable.pipeThrough(transformer).pipeTo(gen.writable)
            .catch((err) => console.error("灰度管线结束", err))
        grayPipeline = { proc, gen, canvas, ctx }
        return gen
    } catch (err) {
        console.error("初始化灰度管线失败", err)
        return null
    }
}

const stopGrayPipeline = () => {
    if (!grayPipeline) return
    const { proc, gen } = grayPipeline
    grayPipeline = null
    try {
        proc.readable.cancel()
    } catch (err) { /* 已经结束 */ }
    try {
        gen.stop()
    } catch (err) { /* 已经结束 */ }
}

const setSourceGrayscale = async (on) => {
    if (!videoSender || !rawVideoTrack) return false
    if (on) {
        const track = startGrayPipeline()
        if (!track) return false
        await videoSender.replaceTrack(track)
        return true
    }
    if (grayPipeline) {
        await videoSender.replaceTrack(rawVideoTrack)
        stopGrayPipeline()
    }
    return true
}

// 控制端下发的画面参数：画质落到编码器，黑白落到采集侧
const applyQuality = async (params) => {
    if (!params) return
    let ok = false
    let grayApplied = false
    try {
        grayApplied = await setSourceGrayscale(!!params.grayscale)
        if (videoSender) {
            const p = videoSender.getParameters()
            if (!p.encodings || !p.encodings.length) p.encodings = [{}]
            const enc = p.encodings[0]
            // setParameters 是合并语义：字段不写就保持原值，所以一律显式赋值
            enc.maxBitrate = params.maxBitrate > 0 ? params.maxBitrate : UNLIMITED_BITRATE
            enc.maxFramerate = params.maxFramerate || 30
            enc.scaleResolutionDownBy = params.scale > 1 ? params.scale : 1
            await videoSender.setParameters(p)
            ok = true
            console.log("已应用画面参数", params)
        }
    } catch (err) {
        console.error("应用画面参数失败", err)
    }
    // 回执给控制端，让它知道到底生效没有
    if (socket) {
        socket.send(JSON.stringify({
            name: "quality_ack",
            data: { ok, grayscale: grayApplied },
            device: connectDevice,
            videoSender: false,
        }))
    }
}

const handlerVideo = async (deviceID) => {
    let sourceId
    try {
        sourceId = await getScreenSourceId()
    } catch (err) {
        console.error("获取屏幕源失败", err)
        ElMessageBox.alert('获取屏幕源失败，请完全退出并重启应用后重试。\n' + (err?.message || err), '屏幕采集失败', {
            confirmButtonText: '确定',
            showClose: false,
        })
        return
    }
    if (!sourceId) {
        console.error("没有获取到屏幕源")
        return
    }
    try {
        const stream = await navigator.mediaDevices.getUserMedia({
            audio: false,
            video: {
                mandatory: {
                    chromeMediaSource: 'desktop',
                    chromeMediaSourceId: sourceId,
                    maxWidth: window.screen.width,
                    maxHeight: window.screen.height,
                }
            }
        })
        selfPC = createPeer()
        watchState(selfPC, () => {
            ipcRenderer.send("closeFloating")
            ElMessageBox.alert('对方断开连接', '连接断开', {
                confirmButtonText: '确定',
                showClose: false,
                callback: () => {
                    closePeer(selfPC)
                    selfPC = null
                },
            })
        })
        pipeCandidates(selfPC, (data) => {
            if (!socket) return
            socket.send(JSON.stringify({
                name: "ice_candidate",
                data,
                device: deviceID,
                videoSender: false
            }))
        })
        //添加音视频流
        stream.getTracks().forEach((track) => selfPC.addTrack(track, stream))
        videoSender = selfPC.getSenders().find((s) => s.track && s.track.kind === "video")
        rawVideoTrack = stream.getVideoTracks()[0]
        const offer = await createOffer(selfPC)
        socket.send(
            JSON.stringify({
                name: "offer",
                data: { sdp: offer },
                device: deviceID,
                videoSender: false
            })
        )
        connectDevice = deviceID
    } catch (err) {
        console.error("屏幕采集失败", err);
        ElMessageBox.alert('屏幕采集失败: ' + (err?.message || err), '屏幕采集失败', {
            confirmButtonText: '确定',
            showClose: false,
        })
    }
}
//#endregion -----------------------------------------------------------
import { ipcRenderer, getScreenSourceId } from "@/utils/electron";
import { createSocket } from "@/utils/ws";
import { isMainWindow } from "@/utils/window";
import { UNLIMITED_BITRATE } from "@/utils/display";
import {
    setIceServers, createPeer, pipeCandidates, createOffer,
    acceptAnswer, addRemoteCandidate, closePeer, watchState
} from "@/utils/video";
const device = useDeviceStore();
let connectDevice = ""
let socket = null
//建立连接（仅主窗口持有服务器连接，其它窗口各管各的，互不影响）
onMounted(() => {
    if (!isMainWindow) return
    ipcRenderer.on("video_disconnect", () => {
        console.log("video_disconnect");
        if (socket) {
            socket.send(JSON.stringify({
                name: "disconnected",
                device: connectDevice,
                videoSender: false
            }))
        }
    })
    ipcRenderer.on("errmsg", (e, data) => {
        ElMessageBox.alert(data, '警告', {
            confirmButtonText: '确定',
        })
    })
    ipcRenderer.on("connect", () => {
        if (navigator.onLine) {
            window.addEventListener("beforeunload", () => {
                if (socket) {
                    socket.close();
                    socket = null
                }

            });
            //联网的情况下,建立websocket
            ConnectServer(device.serverInfo)
                .then((res) => {
                    if (res.code == 0) {
                        device.deviceInfo = res.data
                        device.setOnline(true, "连接服务器成功")
                        //注册设备
                        socket = createSocket("/connect")

                        socket.onopen = (err) => {
                            //被控端监听剪贴板
                            ipcRenderer.on("clipboard", (e, data) => {
                                console.log("被控端监听剪贴板");
                                if (data.sender) {
                                    return
                                }
                                socket.send(JSON.stringify({
                                    name: "clipboard",
                                    clipboard_type: data.clipboard_type,
                                    clipboard_data: data.list
                                }))
                            })

                        }
                        socket.onmessage = (msg) => {
                            let data = JSON.parse(msg.data)
                            switch (data.operation) {
                                case "video":
                                    closePeer(selfPC)
                                    selfPC = null
                                    setIceServers(data.iceservers)
                                    handlerVideo(data.device)
                                    break;
                                case "answer":
                                    if (selfPC) {
                                        //设置邀请人发来的音频源
                                        acceptAnswer(selfPC, data.data.sdp)
                                            .catch((err) => console.error("设置 answer 失败", err))
                                        ipcRenderer.send("showFloating")
                                    }

                                    break;
                                case "ice_candidate":
                                    //添加ice源,这一步很重要,如果没有接收ice则查看是否流程有问题
                                    addRemoteCandidate(selfPC, data.data)
                                    break;
                                case "quality":
                                    applyQuality(data.data)
                                    break;
                                case "video_disconnected":
                                    ipcRenderer.send("closeFloating")
                                    closePeer(selfPC)
                                    selfPC = null
                                    break
                                case "disconnected":
                                    ipcRenderer.send("closeFloating")
                                    closePeer(selfPC)
                                    selfPC = null
                                    device.resetDevice()
                                    device.setOnline(false, "服务器断开连接")
                                    ElMessageBox.alert('服务器断开连接', '连接断开', {
                                        confirmButtonText: '确定',
                                        showClose: false,
                                        callback: () => {
                                            router.push("/index");
                                        },
                                    })
                                    break
                            }
                        }
                        socket.onclose = (err) => {
                            ipcRenderer.send("closeFloating")
                            console.log(err);
                            device.resetDevice()
                            device.setOnline(false, "服务器断开连接")
                        }
                        socket.onerror = (err) => {
                            ipcRenderer.send("closeFloating")
                            console.log(err);
                            device.resetDevice()
                            device.setOnline(false, "服务器断开连接")
                        }
                    } else {
                        device.resetDevice()
                        device.setOnline(false, res.msg)
                    }
                }).catch((err) => {

                });

        } else {
            device.resetDevice()
            device.setOnline(false, "当前网络环境处于离网模式")
        }
    })
})
onUnmounted(() => {
    if (!isMainWindow) return
    if (socket != null) {
        ipcRenderer.send("closeFloating")
        closePeer(selfPC)
        selfPC = null
        socket.close()
        socket = null
    }
})

</script>

<style lang="scss"></style>