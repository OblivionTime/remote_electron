<template>
    <div class="content">
        <div class="main-left">
            <div class="version">
                山与路远程控制 V1.0.0
            </div>
            <div class="operation-list">
                <div class="operation-item" v-for="item in Menu" :key="item.action" @click="active = item.action"
                    :class="active == item.action ? 'operation-item-active' : ''">
                    <div><img :src="item.icon" alt=""> </div>
                    <span class="desc">{{ item.name }}</span>
                </div>

            </div>
            <div class="logo">
                <img src="@/assets/logo.png" alt="">
            </div>
            <div class="online">
                <div class="status" :class="deivce.online.status ? 'success' : 'error'"></div>
                <div class="desc">{{ deivce.online.message }}</div>
            </div>
        </div>
        <div class="main-right">
            <div class="header">
                <div class="tip">
                    免责声明:本作品为开源项目,请勿用于商业用途
                </div>
                <div class="btns">
                    <div class="semi">
                        <el-icon style="font-size: 25px;">
                            <Minus />
                        </el-icon>
                    </div>
                    <div class="close">
                        <el-icon style="font-size: 25px;">
                            <Close />
                        </el-icon>
                    </div>
                </div>


            </div>
            <Connect v-if="active == 'connect'"></Connect>
            <DeviceList v-else-if="active == 'device_list'"></DeviceList>
            <Settings v-else-if="active == 'settings'"></Settings>
        </div>
    </div>
</template>

<script setup>
import Connect from "@/components/Connect.vue"
import DeviceList from "@/components/DeviceList.vue"
import Settings from "@/components/Settings.vue"
import { useDeviceStore } from "@/store/index";
import RemoteIco from '@/assets/remote.svg';
import DeviceIco from '@/assets/device.svg';
import SettingsIco from '@/assets/settings.svg';
let Menu = ref([
    {
        icon: RemoteIco,
        name: "远程控制",
        action: "connect"
    },
    {
        icon: DeviceIco,
        name: "设备列表",
        action: "device_list"
    },
    {
        icon: SettingsIco,
        name: "高级设置",
        action: "settings"
    },
])
const deivce = useDeviceStore();
const active = ref("connect")
const remote = window.require("electron").remote;
const win = remote.getCurrentWindow();
onMounted(() => {
    win.restore()
})
</script>

<style lang="scss" scoped>
.content {
    min-height: 100vh;
    box-sizing: border-box;
    display: flex;
    position: relative;

    .main-left {
        flex-basis: 200px;

        background: #F0F3F8;

        .version {
            margin: 10px;
            font-size: 12px;
            letter-spacing: 1px;
            color: #6F788D;
            margin-bottom: 60px;
        }

        .operation-list {
            -webkit-app-region: no-drag;
            margin-left: 10px;

            .operation-item {
                display: flex;
                align-items: center;
                cursor: pointer;
                box-sizing: border-box;
                width: 95%;
                margin: 20px 0;
                padding: 15px;
                box-shadow: 0px 0px 12px rgba(0, 0, 0, .12);
                border-radius: 10px;
                gap: 20px;

                .desc {
                    font-size: 14px;
                }

                img {
                    width: 20px;
                }
            }

            .operation-item-active {
                background: #CCDFF8;
                color: 0070F9
            }



        }

        .logo {
            position: absolute;
            bottom: 30px;
            -webkit-app-region: no-drag;

            img {
                width: 180px;
            }
        }

        .online {
            position: absolute;
            bottom: 10px;
            left: 20px;
            -webkit-app-region: no-drag;
            display: flex;
            align-items: center;
            font-size: 13px;
            gap: 20px;
            color: #999999;

            .status {
                width: 10px;
                height: 10px;
                border-radius: 50%;
            }

            .success {
                background: #67C23A;
            }

            .error {
                background: #F56C6C;
            }
        }

    }

    .main-right {

        flex: auto;
        // background: #2ed573;
        box-sizing: border-box;

        .header {
            justify-content: space-between;
            display: flex;
            margin-top: 10px;

            .tip {
                color: #E6A23C;
                font-size: 14px;
                margin-left: 5vw;
            }

            .btns {
                display: flex;

                .semi {
                    cursor: pointer;
                    margin-right: 10px;
                    width: 30px;
                    height: 30px;
                    line-height: 30px;
                    text-align: center;

                    &:hover {
                        background: #909399;
                        border-radius: 5px;
                    }
                }

                .close {
                    cursor: pointer;
                    margin-right: 10px;
                    width: 30px;
                    height: 30px;
                    line-height: 30px;
                    text-align: center;

                    &:hover {
                        background: #F56C6C;
                        border-radius: 5px;
                    }
                }
            }

        }
    }

}</style>