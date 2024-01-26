<template>
    <router-view></router-view>
</template>

<script setup>
import { useDeviceStore } from "@/store/index";
import { ConnectServer } from '@/api/info.js';
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
                        msg = JSON.parse(msg)
                        switch (msg.operation) {
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