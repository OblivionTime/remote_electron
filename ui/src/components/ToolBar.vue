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
        <div class="tool_list" @click.stop=""  v-show="!flag">
            <div class="list_item" @click="refresh">
                <div class="item_icon"><img :src="RefreshIcon" alt="" /></div>
                <div class="text">刷新</div>
            </div>
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
const router = useRouter();

const props = defineProps({
    socket: Object
})

const flag = ref(true);

const changeBtn = () => {
    flag.value = false;
};

const emit = defineEmits(['myfresh'])


const close = () => {
    ElMessageBox.confirm("确认退出吗?", "提示", {
        confirmButtonText: "确认",
        cancelButtonText: "取消",
    })
        .then(() => {
            if (props.socket) {
                props.socket.send(JSON.stringify({
                    operation: "change",
                    status: false
                }))
            }

            router.push("/");
        })
        .catch(() => { });
};
//最小化
const minSize = () => {
    const remote = window.require("electron").remote;
    const win = remote.getCurrentWindow();
    win.minimize()
}

//刷新
const refresh = () => {
    emit('myfresh')
    window.location.reload();
};

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
    top: 0;
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;
    cursor: pointer;

    .down_btn {
        padding: 0 0.7rem;
        background: #67686b;
        color: white;

        i.el-icon:hover {
            background: black;
            border-radius: 14%;
        }
    }

    .down_btn:hover {
        cursor: pointer;
    }
}

.list_item {
    display: flex;
    justify-content: center;
    align-items: center;
    flex-direction: column;
    // padding: 0.2rem 1.6rem;
    width: 5.6rem;
    height: 3.4rem;

    .item_icon {
        width: 1.5rem;
        height: 1.5rem;
        margin: 0 auto;

        img {
            width: 100%;
        }
    }
}

.list_item:hover,
.el-dropdown:hover {
    background: gray;
    cursor: pointer;
}

.tool_list {
    -webkit-app-region: no-drag;
    position: absolute;
    top: 0;
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;
    display: flex;
    justify-content: center;
    align-items: center;
    box-sizing: border-box;
    background: #403f3f;
    color: white;

    .text {
        // height: 3rem;
        // line-height: 1.6rem;
        box-sizing: border-box;
        text-align: center;
        font-size: 0.8rem;
        color: white;
        padding-top: 0.4rem;
    }

    .el-dropdown {
        display: flex;
        justify-content: center;
        align-items: center;
        flex-direction: column;
        // padding: 0.2rem 1.6rem;
        width: 5.6rem;
        height: 3.4rem;

        .item_icon {
            width: 1.5rem;
            height: 1.5rem;
            margin: 0 auto;

            img {
                width: 100%;
            }
        }
    }
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