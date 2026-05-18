import {message,} from 'ant-design-vue'
import Axios from "axios"
import * as utils from './utils'
import {getStore} from './storage'
import $store from '../../store/index';
import $router from '../../router/index';
import {notice} from './notice';
import config from "../../config/config";

let HOME_PAGE = config.HOME_PAGE;
const currentOrganization = getStore('currentOrganization', true);
if (currentOrganization) {
    HOME_PAGE = HOME_PAGE + '/' + currentOrganization.code;
}
const crossDomain = config.crossDomain;
let axiosConfig = {};
if (crossDomain) {
    axiosConfig.withCredentials = true;
    axiosConfig.crossDomain = true;
}
const $http = Axios.create(axiosConfig);

// Before request
$http.interceptors.request.use(
    config => {
        //正在请求更新token时，其他接口等待
        config.url = utils.getApiUrl(config.url);
        if (config.method === 'post') {
            const params = new URLSearchParams();
            if (config.data && typeof config.data === 'object') {
                Object.keys(config.data).forEach(key => {
                    if (config.data[key] !== undefined && config.data[key] !== null) {
                        // 处理数组参数
                        if (Array.isArray(config.data[key])) {
                            config.data[key].forEach(item => {
                                params.append(key, item);
                            });
                        } else {
                            params.append(key, config.data[key]);
                        }
                    }
                });
                config.data = params.toString();
            }
        }
        let tokenList = getStore('tokenList', true);
        if (tokenList) {
            let accessToken = tokenList.accessToken;
            let tokenType = tokenList.tokenType;
            config.headers.Authorization = `${tokenType} ${accessToken}`;
        }
        let organization = getStore('currentOrganization', true);
        if (organization) {
            config.headers.organizationCode = organization.code;
        }
        return config;
    },
    error => {
        return Promise.reject(error);
    }
);
// After request
$http.interceptors.response.use(
    response => {
        response = response.data;
        response.code = Number(response.code);
        // 确保 data 字段不为 null，防止组件中 res.data.xxx 崩溃
        if (response.data === null || response.data === undefined) {
            response.data = {};
        }
        switch (response.code) {
            case 200:
                return Promise.resolve(response);
            case 401:
                $router.replace('/member/login?redirect=' + $router.currentRoute.fullPath).catch(() => {});
                $store.dispatch('SET_LOGOUT');
                return Promise.resolve(response);
            case 403:
                notice({
                    title: response.msg !== '' ? response.msg : '无权限操作资源，访问被拒绝',
                }, 'notice', 'error', 5);
                return Promise.resolve(response);
            case 4031:
                //无权限操作资源
                notice({
                    title: response.msg !== '' ? response.msg : '无权限操作资源，访问被拒绝',
                }, 'notice', 'error', 5);
                if ($router.currentRoute.path !== HOME_PAGE) {
                    const permNodes = getStore('permissionNodes', true) || [];
                    const org = getStore('currentOrganization', true);
                    const fallback = utils.getFirstAvailableRoute(permNodes, org);
                    $router.replace(fallback).catch(() => {});
                }
                return Promise.resolve(response);
            case 4041:
                //资源不存在
                notice({
                    title: response.msg !== '' ? response.msg : '资源不存在',
                }, 'notice', 'warning', 5);
                if ($router.currentRoute.path !== HOME_PAGE) {
                    const permNodes = getStore('permissionNodes', true) || [];
                    const org = getStore('currentOrganization', true);
                    const fallback = utils.getFirstAvailableRoute(permNodes, org);
                    $router.replace(fallback).catch(() => {});
                }
                return Promise.resolve(response);
        }
        if (response.code === 200) {
            notice({
                title: '请求错误 ' + response.code,
                desc: response.msg
            }, 'notice', 'warning', 5);
            return Promise.resolve(response);
        } else {
            response.msg !== '' && notice({
                title: response.msg,
            }, 'notice', 'error', 5);
            return Promise.resolve(response);
        }
    },
    error => {
        const raw = error && error.response ? error.response.data : null;
        const response = (raw && typeof raw === 'object') ? raw : {
            code: error && error.response ? error.response.status : 500,
            msg: typeof raw === 'string' ? raw : '请求出现错误，请稍后再试'
        };
        response.code = Number(response.code);
        message.destroy();
        // HTTP 404（路由不存在）不弹出通知，避免退出登录等场景的干扰
        if ((error && error.response && error.response.status === 404) || response.code === 404) {
            return Promise.reject(error);
        }
        switch (response.code) {
            default:
                response.msg !== '' && notice({
                    title: response.msg,
                }, 'notice', 'error', 5);
                return Promise.reject(error);
        }
    }
);

export default $http;
