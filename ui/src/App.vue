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
    iceServers: [
        {
            url: "turn:42.192.40.58:3478?transport=udp",
            username: "ddssingsong",
            credential: "123456",
        },
        {
            url: "turn:42.192.40.58:3478?transport=tcp",
            username: "ddssingsong",
            credential: "123456",
        },
    ],
};
//初始化PC源
function initPC(device) {
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
                    device: device,
                    videoSender: false
                })
            );
        }
    };

    return pc
}
var selfPC = null
const handlerVideo = async (device) => {
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
            selfPC = initPC(device)
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
                        device: device,
                        videoSender: false
                    })
                )
            }, (err) => {
                console.log(err);
            });
        }).catch(error => console.log(error));
}
//#endregion -----------------------------------------------------------
import { useDeviceStore } from "@/store/index";
import { ConnectServer } from '@/api/info.js';
const remote = window.require('electron').remote;
const desktopCapturer = window.require('electron').desktopCapturer;
const win = remote.getCurrentWindow();
const device = useDeviceStore();
let socket = null
//建立连接
onMounted(() => {
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
                        console.log("🚀 ~ .then ~ msg:", msg)
                        let data = JSON.parse(msg.data)
                        switch (data.operation) {
                            case "video":
                                handlerVideo(data.device)
                            case "answer":
                                if (selfPC) {
                                    //设置邀请人发来的音频源
                                    selfPC.setRemoteDescription(new nativeRTCSessionDescription(data.data.sdp));
                                }

                                break;
                            case "ice_candidate":
                                if (selfPC) {
                                    //添加ice源,这一步很重要,如果没有接收ice则查看是否流程有问题
                                    var candidate = new nativeRTCIceCandidate(data.data);
                                    selfPC.addIceCandidate(candidate);
                                }
                                break;

                            case "disconnect":
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
                                break
                        }
                    }
                    socket.onclose = (err) => {
                        console.log(err);
                    }
                    socket.onerror = (err) => {
                        console.log(err);
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
onUnmounted(() => {
    if (socket != null) {
        socket.close()
        socket = null
    }
})

</script>

<style lang="scss" ></style>