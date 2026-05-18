/* eslint-disable no-console,no-unused-vars */
import Vue from 'vue'
import VueRouter from 'vue-router'
import Antd from "ant-design-vue";
import App from './App'
import store from './store/index'
import router from './router/index'
import 'ant-design-vue/dist/antd.css'
import vuescroll from 'vuescroll';
import 'vuescroll/dist/vuescroll.css';
import '@/assets/css/theme.less'
import '@/assets/icon/iconfont'
import WrapperContent from '@/components/layout/WrapperContent'
import {message, notification} from 'ant-design-vue'
import {notice, destroyNotice} from 'assets/js/notice'

import moment from 'moment';
import 'moment/locale/zh-cn';

import common from "./mixins/common";

import '@/utils/filter' // global filter


moment.locale('zh-cn');

Vue.use(VueRouter);
Vue.use(store);

Vue.config.productionTip = false;

// 全局静默 Vue Router 导航相关错误（NavigationCancelled, NavigationDuplicated 等）
// 这些是 Vue Router 2.x 的正常行为，当导航被守卫重定向或新导航取代旧导航时触发
const silentNavigationErrors = new Set([
    'NavigationCancelled',
    'NavigationDuplicated',
    'NavigationRedirected',
    'NavigationAborted',
]);
const originalErrorHandler = Vue.config.errorHandler;
Vue.config.errorHandler = function (err, vm, info) {
    if (err && (silentNavigationErrors.has(err.name) || err._isRouter)) {
        return;
    }
    if (originalErrorHandler) {
        return originalErrorHandler(err, vm, info);
    }
    console.error(err);
};

// 全局捕获未处理的 Promise rejection，静默路由导航中断错误
// 使用 capture 阶段注册，确保在 webpack-dev-server overlay 之前执行
// 调用 stopImmediatePropagation 阻止 overlay 捕获这些正常的导航中断错误
window.addEventListener('unhandledrejection', function (event) {
    const err = event.reason;
    if (err && silentNavigationErrors.has(err.name)) {
        event.preventDefault();
        event.stopImmediatePropagation();
    }
}, true);

// 全局捕获同步 error 事件，静默路由导航中断错误
// 当导航守卫重定向 (next('/xxx')) 时，Vue Router 内部会抛出同步的
// NavigationRedirected 错误，此 handler 防止其被 webpack-dev-server overlay 捕获
window.addEventListener('error', function (event) {
    const err = event.error;
    if (err && (silentNavigationErrors.has(err.name) || err._isRouter)) {
        event.preventDefault();
        event.stopImmediatePropagation();
        return false;
    }
}, true);
Vue.use(Antd);
Vue.component('WrapperContent', WrapperContent);

import VueClipboards from 'vue-clipboards';
Vue.use(VueClipboards);

import uploader from 'vue-simple-uploader'
Vue.use(uploader);

Vue.prototype.$message = message;
Vue.prototype.$notification = notification;
Vue.prototype.$notice = notice;
Vue.prototype.$destroyNotice = destroyNotice;

Vue.use(vuescroll);
Vue.prototype.$vuescrollConfig = {
    vuescroll: {
        mode: 'native'
    },
    scrollPanel: {
        scrollingX: true,
    },
    bar: {
        delayTime: 500,
        onlyShowBarOnScroll: false,
        background: "#cecece",
        keepShow: false
    }
};

// 全局快捷键管理
import ShortcutManager from '@/plugins/shortcuts';
ShortcutManager.install(Vue);

Vue.mixin(common);


new Vue({
    el: '#app',
    store,
    router,
    template: '<App/>',
    components: {App}
});
