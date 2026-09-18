<template>
    <div class="app">
        <aside class="sidebar card">
            <div class="brand">
                <img src="@/assets/logo.png" alt="">
                <div>
                    <div class="name">山与路远程控制</div>
                    <div class="version">V1.0.0</div>
                </div>
            </div>

            <nav class="nav">
                <div class="nav-item" v-for="item in Menu" :key="item.action" @click="active = item.action"
                    :class="{ 'nav-item-active': active == item.action }">
                    <img :src="item.icon" alt="">
                    <span>{{ item.name }}</span>
                </div>
            </nav>

            <div class="status">
                <span class="dot" :class="deivce.online.status ? 'dot-ok' : 'dot-off'"></span>
                <span>{{ deivce.online.message }}</span>
            </div>
        </aside>

        <section class="main">
            <header class="topbar">
                <span class="tip">免责声明：本作品为开源项目，请勿用于商业用途</span>
                <div class="win-btns">
                    <button class="win-btn" @click="minSize">
                        <el-icon>
                            <Minus />
                        </el-icon>
                    </button>
                    <button class="win-btn win-btn-close" @click="closeWindow">
                        <el-icon>
                            <Close />
                        </el-icon>
                    </button>
                </div>
            </header>

            <div class="page">
                <Connect v-if="active == 'connect'"></Connect>
                <DeviceList v-else-if="active == 'device_list'"></DeviceList>
                <Settings v-else-if="active == 'settings'"></Settings>
            </div>
        </section>
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
import { minimizeWindow, closeWindow as closeMainWindow, restoreWindow, setWindowSize, centerWindow } from "@/utils/electron";
//最小化
const minSize = () => minimizeWindow()
const closeWindow = () => closeMainWindow()
onMounted(() => {
    restoreWindow()
    setWindowSize(750, 510)
    centerWindow()
})
</script>

<style lang="scss" scoped>
.app {
    height: 100vh;
    padding: 12px;
    display: flex;
    gap: 12px;
    background: var(--bg);
    -webkit-app-region: drag;
}

.sidebar {
    width: 196px;
    flex-shrink: 0;
    padding: 18px 14px;
    display: flex;
    flex-direction: column;
    -webkit-app-region: no-drag;

    .brand {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 0 6px 16px;
        border-bottom: 1px solid var(--border);
        margin-bottom: 14px;

        img {
            width: 32px;
        }

        .name {
            font-size: 14px;
            font-weight: 600;
        }

        .version {
            font-size: 11px;
            color: var(--text-secondary);
            margin-top: 2px;
        }
    }

    .nav {
        display: flex;
        flex-direction: column;
        gap: 6px;
        flex: 1;

        .nav-item {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px 12px;
            border-radius: var(--radius-sm);
            cursor: pointer;
            color: var(--text-secondary);
            transition: background .2s, color .2s;

            img {
                width: 16px;
                opacity: .7;
            }

            &:hover {
                background: #f2f5fa;
                color: var(--text);
            }

            &-active {
                background: #ecf5ff;
                color: var(--primary);
                font-weight: 600;

                img {
                    opacity: 1;
                }
            }
        }
    }

    .status {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 10px 6px 2px;
        border-top: 1px solid var(--border);
        font-size: 12px;
        color: var(--text-secondary);

        .dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            flex-shrink: 0;

            &-ok {
                background: var(--success);
            }

            &-off {
                background: var(--danger);
            }
        }
    }
}

.main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
    -webkit-app-region: no-drag;

    .topbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 2px 4px 2px 8px;
        -webkit-app-region: drag;

        .tip {
            font-size: 12px;
            color: var(--text-secondary);
        }

        .win-btns {
            display: flex;
            gap: 6px;
            -webkit-app-region: no-drag;

            .win-btn {
                width: 28px;
                height: 28px;
                display: flex;
                align-items: center;
                justify-content: center;
                border: none;
                background: transparent;
                cursor: pointer;
                border-radius: var(--radius-sm);
                color: var(--text-secondary);
                transition: background .2s, color .2s;

                &:hover {
                    background: #e8ecf2;
                    color: var(--text);
                }

                &-close:hover {
                    background: #fef0f0;
                    color: var(--danger);
                }
            }
        }
    }

    .page {
        flex: 1;
        overflow-y: auto;
        padding-right: 2px;
    }
}
</style>
