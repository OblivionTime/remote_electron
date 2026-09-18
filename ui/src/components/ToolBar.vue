<template>
    <div class="main">
        <div class="toolNar_top" v-show="flag">
            <el-tooltip content="展开" placement="bottom" effect="light" popper-class="popper">
                <div class="down_btn" @click.stop="changeBtn">
                    <el-icon size="15">
                        <ArrowDownBold />
                    </el-icon>
                </div>
            </el-tooltip>
        </div>
        <div class="tool_list" @click.stop="" v-show="!flag">
            <div class="list_item" @click="refresh">
                <div class="item_icon"><img :src="RefreshIcon" alt="" /></div>
                <div class="text">刷新</div>
            </div>
            <div class="list_item" @click="openFile">
                <div class="item_icon"><img :src="FileIcon" alt="" /></div>
                <div class="text">文件</div>
            </div>
            <el-popover placement="bottom" :width="252" trigger="click" popper-class="screen-popover">
                <template #reference>
                    <div class="list_item" @click.stop>
                        <div class="item_icon"><img :src="ScreenIcon" alt="" /></div>
                        <div class="text">画面</div>
                    </div>
                </template>
                <div class="screen-settings">
                    <div class="row">
                        <span class="label">画质</span>
                        <el-radio-group size="small" :model-value="quality"
                            @change="(v) => emit('update:quality', v)">
                            <el-radio-button v-for="q in QUALITY_OPTIONS" :key="q.value" :label="q.value">
                                {{ q.label }}
                            </el-radio-button>
                        </el-radio-group>
                    </div>
                    <div class="row">
                        <span class="label">颜色</span>
                        <el-radio-group size="small" :model-value="grayscale ? 'gray' : 'color'"
                            @change="(v) => emit('update:grayscale', v === 'gray')">
                            <el-radio-button label="color">彩色</el-radio-button>
                            <el-radio-button label="gray">黑白</el-radio-button>
                        </el-radio-group>
                    </div>
                    <div class="row">
                        <span class="label">比例</span>
                        <el-radio-group size="small" :model-value="aspect"
                            @change="(v) => emit('update:aspect', v)">
                            <el-radio-button v-for="a in ASPECT_OPTIONS" :key="a.value" :label="a.value">
                                {{ a.label }}
                            </el-radio-button>
                        </el-radio-group>
                    </div>
                    <div class="hint">画质会实时下发给被控端，影响码率与分辨率</div>
                </div>
            </el-popover>
            <div class="list_item" @click="minSize">
                <div class="item_icon"><img :src="minSizeIcon" alt="" /></div>
                <div class="text">最小化</div>
            </div>
            <div class="list_item" @click="close">
                <div class="item_icon"><img :src="closeIcon" alt="" /></div>
                <div class="text">退出</div>
            </div>
        </div>
    </div>
</template>
  
<script setup>
import { useRouter } from "vue-router";
import { onMounted, ref } from "vue";
import { ElMessageBox } from "element-plus";
import RefreshIcon from "@/assets/refresh.svg"
import minSizeIcon from "@/assets/minSize.svg"
import closeIcon from "@/assets/close.svg"
import FileIcon from "@/assets/file.svg"
import ScreenIcon from "@/assets/remote.svg"
import { minimizeWindow } from "@/utils/electron";
import { QUALITY_OPTIONS, ASPECT_OPTIONS } from "@/utils/display";
const router = useRouter();

const props = defineProps({
    quality: { type: String, default: "high" },
    grayscale: { type: Boolean, default: false },
    aspect: { type: String, default: "fit" },
})

const flag = ref(true);

const changeBtn = () => {
    flag.value = false;
};

const emit = defineEmits(['myfresh', "closeConnect", "file", "update:quality", "update:grayscale", "update:aspect"])


const close = () => {
    emit("closeConnect")
};
//最小化
const minSize = () => minimizeWindow();

//刷新
const refresh = () => {
    emit('myfresh')
    window.location.reload();
};

//文件传输（已经开过会复用同一个窗口）
const openFile = () => emit('file')

onMounted(() => {
    document.addEventListener("click", (e) => {
        flag.value = true;
    });
});
onUnmounted(() => {
    document.removeEventListener("click", (e) => {
        flag.value = true;
    });
})
</script>
  
<style lang="scss" scoped>
.toolNar_top {
    -webkit-app-region: no-drag;
    position: absolute;
    top: 12px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;
    cursor: pointer;

    .down_btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 44px;
        height: 26px;
        color: #fff;
        background: rgba(0, 0, 0, .45);
        border-radius: 13px;
        backdrop-filter: blur(8px);
        transition: background .2s;

        &:hover {
            background: rgba(0, 0, 0, .68);
        }
    }
}

.list_item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    width: 64px;
    height: 56px;
    border-radius: 8px;
    color: #fff;
    cursor: pointer;
    transition: background .2s;

    &:hover {
        background: rgba(255, 255, 255, .14);
    }

    .item_icon {
        width: 18px;
        height: 18px;

        img {
            width: 100%;
        }
    }

    .text {
        font-size: 11px;
    }
}

.tool_list {
    -webkit-app-region: no-drag;
    position: absolute;
    top: 12px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;
    display: flex;
    gap: 4px;
    padding: 4px;
    background: rgba(0, 0, 0, .45);
    border-radius: 12px;
    backdrop-filter: blur(10px);
}

.isTransform {
    // transform: translateX(13rem);
    // transition: all 0.2s linear;
}

.el-dropdown-link {
    width: 100%;
}
</style>
  
<style lang="scss" >
/* 画面设置弹层走 popper 挂到 body，scoped 样式命不中，放这里 */
.screen-popover {
    padding: 12px 14px;

    .screen-settings {
        display: flex;
        flex-direction: column;
        gap: 12px;

        .row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;

            .label {
                flex-shrink: 0;
                font-size: 12px;
                color: #606266;
            }

            .el-radio-group {
                flex-wrap: nowrap;
            }
        }

        .hint {
            font-size: 11px;
            line-height: 1.6;
            color: #909399;
        }
    }
}

li.selected-item.el-dropdown-menu__item {
    background: rgb(163, 158, 158);
}

.el-message-box {
    width: 23rem;
    // height: 11.5rem;
    max-width: 23rem;

    .el-message-box__header {
        height: 2rem;
        padding: 0.8rem;

        .el-message-box__title {
            font-size: 1.4rem;
        }

        .el-message-box__headerbtn {
            font-size: 1.6rem;
        }
    }

    .el-message-box__content {
        padding: 0.7rem;
        font-size: 1rem;
        padding-bottom: 2rem;
    }

    .el-message-box__btns {
        padding: 0.13rem 0.4rem 0.2rem;

        .el-button {
            font-size: 1rem;
            padding: 1rem;
        }
    }

    .el-message-box__btns button:nth-child(2) {
        margin-left: 0.7rem;
    }
}

.el-dropdown-menu__item {
    font-size: 1rem;
    padding: 0.8rem;
}

.el-dropdown__popper .el-dropdown-menu {
    background: #403f3f;
}

.el-dropdown__popper.el-popper {
    border: 0.02667rem solid #403f3f;
}

.el-dropdown-menu__item {
    font-size: 0.8rem;
    padding: 0.7rem;
    color: white;
}

.el-dropdown-menu__item:hover {
    background: gray !important;
    color: white !important;
}
</style>