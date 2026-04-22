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
    if (err && silentNavigationErrors.has(err.name)) {
        return;
    }
    if (originalErrorHandler) {
        return originalErrorHandler(err, vm, info);
    }
    console.error(err);
};

// 全局捕获未处理的 Promise rejection，静默路由导航中断错误
window.addEventListener('unhandledrejection', function (event) {
    const err = event.reason;
    if (err && silentNavigationErrors.has(err.name)) {
        event.preventDefault();
    }
});
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

Vue.mixin(common);


new Vue({
    el: '#app',
    store,
    router,
    template: '<App/>',
    components: {App}
});
