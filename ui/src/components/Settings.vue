<template>
    <div class="settings-page">
        <!-- 左：菜单 -->
        <div class="card menu-card">
            <div class="menu">
                <div class="menu-item" v-for="item in menus" :key="item.key" :class="{ active: active === item.key }"
                    @click="active = item.key">
                    <el-icon>
                        <component :is="item.icon" />
                    </el-icon>
                    <span>{{ item.label }}</span>
                </div>
            </div>
        </div>

        <!-- 右：内容 -->
        <div class="card content-card">
            <!-- 连接配置 -->
            <template v-if="active === 'conn'">
                <div class="card-title">连接配置</div>

                <div class="field">
                    <label>协议</label>
                    <el-select v-model="formData.https">
                        <el-option label="HTTP" :value="false" />
                        <el-option label="HTTPS" :value="true" />
                    </el-select>
                </div>

                <div class="field">
                    <label>服务器地址</label>
                    <el-input v-model="formData.remote_url" placeholder="例如 127.0.0.1:9998" />
                </div>

                <el-button type="primary" class="save" @click="SaveData">保存</el-button>
            </template>

            <!-- 远程画面 -->
            <template v-else-if="active === 'display'">
                <div class="card-title">远程画面</div>

                <div class="field">
                    <label>默认画质</label>
                    <el-radio-group v-model="quality" @change="onQualityChange">
                        <el-radio v-for="q in QUALITY_OPTIONS" :key="q.value" :label="q.value">{{ q.label }}</el-radio>
                    </el-radio-group>
                </div>

                <div class="field">
                    <label>默认屏幕比例</label>
                    <el-radio-group v-model="aspect" @change="onAspectChange">
                        <el-radio v-for="a in ASPECT_OPTIONS" :key="a.value" :label="a.value">{{ a.label }}</el-radio>
                    </el-radio-group>
                </div>

                <p class="hint">
                    画质会在远控时下发给被控端，真实影响编码码率与分辨率；原画质不限码率。<br>
                    屏幕比例只影响本机显示。这些默认值可在远控过程中临时覆盖。
                </p>
            </template>

            <!-- 窗口行为 -->
            <template v-else-if="active === 'window'">
                <div class="card-title">窗口行为</div>

                <div class="field">
                    <label>点击关闭按钮时</label>
                    <el-radio-group v-model="closeAction" @change="onCloseActionChange">
                        <el-radio label="ask">每次询问</el-radio>
                        <el-radio label="hide">隐藏到托盘</el-radio>
                        <el-radio label="quit">退出程序</el-radio>
                    </el-radio-group>
                </div>

                <p class="hint">隐藏到托盘后程序仍在后台运行，可以继续被远程控制。</p>
            </template>

            <!-- 关于 -->
            <template v-else>
                <div class="card-title">关于</div>

                <div class="about">
                    <img src="@/assets/logo.png" alt="">
                    <div>
                        <div class="name">山与路远程控制</div>
                        <div class="ver">V1.0.0</div>
                    </div>
                </div>

                <p class="hint">开源远程控制项目，请勿用于商业用途。</p>
            </template>
        </div>
    </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'
import { useDeviceStore } from "@/store/index";
import { getCloseAction, setCloseAction, CLOSE_ACTIONS } from "@/utils/electron";
import {
    QUALITY_OPTIONS, ASPECT_OPTIONS, loadRemotePrefs,
    saveDefaultQuality, saveDefaultAspect,
} from "@/utils/display";

const device = useDeviceStore();

const active = ref("conn")
const menus = [
    { key: "conn", label: "连接配置", icon: "Connection" },
    { key: "display", label: "远程画面", icon: "FullScreen" },
    { key: "window", label: "窗口行为", icon: "Monitor" },
    { key: "about", label: "关于", icon: "InfoFilled" },
]

let formData = ref({
    https: false,
    remote_url: "127.0.0.1:9998"
})

const remotePrefs = loadRemotePrefs()
const quality = ref(remotePrefs.quality)
const aspect = ref(remotePrefs.aspect)

const onQualityChange = (v) => {
    saveDefaultQuality(v)
    ElMessage({ showClose: true, message: '已保存', duration: 1500, type: 'success' })
}
const onAspectChange = (v) => {
    saveDefaultAspect(v)
    ElMessage({ showClose: true, message: '已保存', duration: 1500, type: 'success' })
}

const closeAction = ref("ask")
const onCloseActionChange = async (value) => {
    const saved = await setCloseAction(value)
    closeAction.value = CLOSE_ACTIONS.includes(saved) ? saved : "ask"
    ElMessage({ showClose: true, message: '已保存', duration: 1500, type: 'success' })
}

const SaveData = () => {
    if (!formData.value.remote_url) {
        return ElMessage({
            showClose: true,
            message: '请输入IP地址',
            duration: 1500,
            type: 'warning',
        })
    }
    localStorage.setItem("serverInfo", JSON.stringify(formData.value))
    ElMessageBox.alert('修改成功', '成功', {
        confirmButtonText: '确定',
        showClose: false,
        callback: () => {
            window.location.reload();
        },

    })
}

onMounted(async () => {
    formData.value = JSON.parse(JSON.stringify(device.serverInfo))
    closeAction.value = await getCloseAction()
})
</script>

<style lang="scss" scoped>
.settings-page {
    display: flex;
    gap: 12px;
    height: 100%;
    min-height: 0;

    /* 左栏菜单 */
    .menu-card {
        width: 150px;
        flex-shrink: 0;
        padding: 12px 10px;

        .menu {
            display: flex;
            flex-direction: column;
            gap: 4px;

            .menu-item {
                display: flex;
                align-items: center;
                gap: 10px;
                padding: 10px 12px;
                border-radius: var(--radius-sm);
                font-size: 13px;
                color: var(--text-secondary);
                cursor: pointer;
                transition: background .2s, color .2s;

                .el-icon {
                    font-size: 15px;
                }

                &:hover {
                    background: #f2f5fa;
                    color: var(--text);
                }

                &.active {
                    background: #ecf5ff;
                    color: var(--primary);
                    font-weight: 600;
                }
            }
        }
    }

    /* 右栏内容 */
    .content-card {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;

        .field {
            margin-bottom: 20px;

            label {
                display: block;
                font-size: 13px;
                color: var(--text-secondary);
                margin-bottom: 8px;
            }

            .el-select {
                width: 100%;
            }
        }

        .save {
            width: 100%;
        }

        .hint {
            font-size: 12px;
            line-height: 1.7;
            color: var(--text-secondary);
        }

        .about {
            display: flex;
            align-items: center;
            gap: 14px;
            margin-bottom: 18px;

            img {
                width: 48px;
                height: 48px;
                border-radius: 12px;
            }

            .name {
                font-size: 15px;
                font-weight: 700;
            }

            .ver {
                margin-top: 4px;
                font-size: 12px;
                color: var(--text-secondary);
            }
        }
    }
}
</style>
