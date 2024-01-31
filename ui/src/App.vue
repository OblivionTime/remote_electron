<template>
    <router-view></router-view>
</template>

<script setup>
//#region 视频流 ------------------------------------------------------------


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
const iceServer = {
    iceServers: []
};
//初始化PC源
function initPC(deviceID) {
    let pc = new PeerConnection(iceServer);
    pc.onicecandidate = (evt) => {
        if (evt.candidate) {
            socket.send(
                JSON.stringify({
                    name: `ice_candidate`,
                    data: {
                        id: evt.candidate.sdpMid,
                        label: evt.candidate.sdpMLineIndex,
                        sdpMLineIndex: evt.candidate.sdpMLineIndex,
                        candidate: evt.candidate.candidate,
                    },
                    device: deviceID,
                    videoSender: false
                })
            );
        }
    };
    pc.oniceconnectionstatechange = (evt) => {
        setTimeout(() => {
            let connectionState = evt.target.connectionState
            if (connectionState == "failed" || connectionState == "disconnected" || connectionState == "closed") {
                ElMessageBox.alert('对方断开连接', '连接断开', {
                    confirmButtonText: '确定',
                    showClose: false,
                    callback: () => {
                        ipcRenderer.send("closeFloating")
                        selfPC.close()
                        selfPC = null
                    },

                })
            }
        }, 1000);

    }
    return pc
}
var selfPC = null
const handlerVideo = async (deviceID) => {
    let sources = await desktopCapturer.getSources({ types: ['screen'] })
    navigator.mediaDevices.getUserMedia({
        audio: false,
        video: {
            mandatory: {
                chromeMediaSource: 'desktop',
                maxWidth: window.screen.width,
                maxHeight: window.screen.height,
                chromeMediaSourceId: sources[0].id

            }
        }

    })
        .then((stream) => {
            //初始化PC源
            selfPC = initPC(deviceID)
            //添加音视频流
            selfPC.addStream(stream)
            selfPC.createOffer((session_desc) => {
                selfPC.setLocalDescription(session_desc);
                socket.send(
                    JSON.stringify({
                        name: "offer",
                        data: {
                            sdp: session_desc,
                        },
                        device: deviceID,
                        videoSender: false
                    })
                )
                connectDevice = deviceID
            }, (err) => {
                console.log(err);
            });
        }).catch(error => console.log(error));
}
//#endregion -----------------------------------------------------------
import { useDeviceStore } from "@/store/index";
import { ConnectServer } from '@/api/info.js';
const remote = window.require('electron').remote;
const ipcRenderer = window.require('electron').ipcRenderer;
const desktopCapturer = window.require('electron').desktopCapturer;
const win = remote.getCurrentWindow();
const device = useDeviceStore();
let connectDevice = ""
let socket = null
//建立连接
onMounted(() => {
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
                        device.deviceInfo.connectioned = JSON.parse(res.data.connectioned)
                        device.online = {
                            status: true,
                            message: "连接服务器成功",
                        }
                        //注册设备
                        socket = new WebSocket("ws://127.0.0.1:3002/v1/api/remote/server/connect")
                        socket.onmessage = (msg) => {
                            let data = JSON.parse(msg.data)
                            switch (data.operation) {
                                case "video":
                                    if (selfPC) {
                                        selfPC.close()
                                        selfPC = null
                                    }
                                    iceServer.iceServers = data.iceservers
                                    handlerVideo(data.device)
                                case "answer":
                                    if (selfPC) {
                                        //设置邀请人发来的音频源
                                        selfPC.setRemoteDescription(new nativeRTCSessionDescription(data.data.sdp));
                                        ipcRenderer.send("showFloating")
                                    }

                                    break;
                                case "ice_candidate":
                                    if (selfPC) {
                                        //添加ice源,这一步很重要,如果没有接收ice则查看是否流程有问题
                                        var candidate = new nativeRTCIceCandidate(data.data);
                                        selfPC.addIceCandidate(candidate);
                                    }
                                    break;
                                case "video_disconnected":
                                    ipcRenderer.send("closeFloating")
                                    selfPC.close()
                                    selfPC = null
                                    break
                                case "disconnected":
                                    device.deviceInfo = {
                                        device_id: "",
                                        identificationCode: "",
                                        verificationCode: "",
                                        connectioned: [],
                                    }
                                    device.online = {
                                        status: false,
                                        message: "服务器断开连接",
                                    }
                                    ElMessageBox.alert('服务器断开连接', '连接断开', {
                                        confirmButtonText: '确定',
                                        showClose: false,
                                        callback: () => {
                                            router.push("/");
                                        },
                                    })
                                    break
                            }
                        }
                        socket.onclose = (err) => {
                            ipcRenderer.send("closeFloating")
                            console.log(err);
                            device.deviceInfo = {
                                device_id: "",
                                identificationCode: "",
                                verificationCode: "",
                                connectioned: [],
                            }
                            device.online = {
                                status: false,
                                message: "服务器断开连接",
                            }
                        }
                        socket.onerror = (err) => {
                            ipcRenderer.send("closeFloating")
                            console.log(err);
                            device.deviceInfo = {
                                device_id: "",
                                identificationCode: "",
                                verificationCode: "",
                                connectioned: [],
                            }
                            device.online = {
                                status: false,
                                message: "服务器断开连接",
                            }
                        }
                    } else {
                        device.deviceInfo = {
                            device_id: "",
                            identificationCode: "",
                            verificationCode: "",
                            connectioned: [],
                        }
                        device.online = {
                            status: false,
                            message: res.msg,
                        }
                    }
                }).catch((err) => {

                });

        } else {
            device.deviceInfo = {
                device_id: "",
                identificationCode: "",
                verificationCode: "",
                connectioned: [],
            }
            device.online = {
                status: false,
                message: "当前网络环境处于离网模式",
            }
        }
    })
})
onUnmounted(() => {
    if (socket != null) {
        ipcRenderer.send("closeFloating")
        if (selfPC) {
            selfPC.close()
            selfPC = null
        }
        socket.close()
        socket = null
    }
})

</script>

<style lang="scss" ></style>