<template>
    <div id="app" style="height: 100%">
        <a-config-provider :getPopupContainer="getPopupContainer" :locale="zh_CN">
            <transition name="router-fades" mode="out-in">
                <router-view></router-view>
            </transition>
        </a-config-provider>
    </div>
</template>
<script>
    import zh_CN from 'ant-design-vue/lib/locale-provider/zh_CN';
    import 'moment/locale/zh-cn';
    export default {
        name: 'app',
        data() {
            return {
                zh_CN
            }
        },
        watch: {},
        methods: {
            getPopupContainer(el, dialogContext) {
                if (el) {
                    // 如果触发元素在 .task-detail 内，渲染到 body 以避免被 vue-scroll 裁剪
                    if (el.closest && el.closest('.task-detail')) {
                        return document.body;
                    }
                }
                if (dialogContext) {
                    return dialogContext.getDialogWrap();
                } else {
                    return document.body;
                }
            },
        },
    }
</script>
