import Vue from 'vue'
import store from '@/store'
import Router from 'vue-router'
import Index from '@/views/index'
import Home from './home';
import {getStore, setStore} from "../assets/js/storage";
import {appendMenuRoutes, isTokenExpired} from "../assets/js/utils";
import config from "../config/config";
import {refreshAccessToken} from "../api/common/common";

let HOME_PAGE = config.HOME_PAGE;
const currentOrganization = getStore('currentOrganization', true);
if (currentOrganization) {
    HOME_PAGE = HOME_PAGE + '/' + currentOrganization.code;
}

Vue.use(Router);

// 全局处理受控导航失败，避免路由守卫重定向或重复跳转时抛出未捕获异常
const ignoredNavigationFailureNames = new Set([
    'NavigationDuplicated',
    'NavigationRedirected',
    'NavigationCancelled',
    'NavigationAborted',
]);

const originalPush = Router.prototype.push;
Router.prototype.push = function push(location) {
    return originalPush.call(this, location).catch(err => {
        if (!err || !ignoredNavigationFailureNames.has(err.name)) {
            throw err;
        }
    });
};

const originalReplace = Router.prototype.replace;
Router.prototype.replace = function replace(location) {
    return originalReplace.call(this, location).catch(err => {
        if (!err || !ignoredNavigationFailureNames.has(err.name)) {
            throw err;
        }
    });
};

const routes = [].concat(
    Home
);
// const router = new Router({
//     routes: routers
// });
const menu = getStore('menu', true);
if (menu) {
    appendMenuRoutes(routes, menu);
}
const router = new Router({
    routes: [
        {
            path: '/',
            name: 'index',
            component: Index,
            children: routes
        },
        // {
        //     name: 'login',
        //     path: '/login',
        //     component: resolve => require(['@/views/login'], resolve),
        //     meta: {model: 'Login'},
        // },
        {
            name: 'member',
            path: '/member',
            component: resolve => require(['@/components/layout/UserLayout'], resolve),
            meta: {model: 'Login'},
            children: [
                {
                    path: 'login',
                    name: 'login',
                    component: () => import(/* webpackChunkName: "user" */ '@/views/member/login'),
                    meta: {model: 'Login'},
                },
                {
                    path: 'register',
                    name: 'register',
                    component: () => import(/* webpackChunkName: "user" */ '@/views/member/Register'),
                    meta: {model: 'Login'},
                },
                {
                    path: 'forgot',
                    name: 'forgot',
                    component: () => import(/* webpackChunkName: "user" */ '@/views/member/forgot'),
                    meta: {model: 'Login'},
                },
                {
                    path: 'register-result',
                    name: 'registerResult',
                    component: () => import(/* webpackChunkName: "user" */ '@/views/member/RegisterResult'),
                    meta: {model: 'Login'},
                }
            ]
        },
        {
            name: 'resetEmail',
            path: '/reset/email',
            component: resolve => require(['@/views/reset/email'], resolve),
            meta: {model: 'error'},
        },
        {
            name: 'install',
            path: '/install',
            component: resolve => require(['@/views/error/install'], resolve),
            meta: {model: 'error'},
        },
        {
            name: '404',
            path: '/404',
            component: resolve => require(['@/views/error/404'], resolve),
            meta: {model: 'error'},
        },
        {
            name: '403',
            path: '/403',
            component: resolve => require(['@/views/error/403'], resolve),
            meta: {model: 'error'},
        },
        {
            name: '500',
            path: '/500',
            component: resolve => require(['@/views/error/500'], resolve),
            meta: {model: 'error'},
        },
    ]
});

router.beforeEach((to, from, next) => {
    let tokenList = getStore('tokenList', true);
    if (tokenList) {
        let refreshToken = tokenList.refreshToken;
        let accessTokenExp = tokenList.accessTokenExp;
        //判断accessToken即将到期后刷新token
        if (accessTokenExp && isTokenExpired(accessTokenExp)) {
            refreshAccessToken(refreshToken).then(res => {
                tokenList.accessToken = res.data.accessToken;
                tokenList.accessTokenExp = res.data.accessTokenExp;
                setStore('tokenList', tokenList);
            }).catch(() => {
                // token 刷新失败不在这里处理登出，避免在导航守卫中触发新导航
                // 登出逻辑交给后续请求的 401 响应处理
            });
        }
    }
    // 未登录时，只允许访问登录页和错误页
    if (!store.state.logged && to.meta.model !== 'Login' && to.meta.model !== 'error') {
        next({ name: 'login', query: {redirect: to.fullPath} });
        return;
    }
    // 已登录时，访问登录页则跳转首页
    if (to.meta.model === 'Login' && store.state.logged) {
        const org = getStore('currentOrganization', true);
        const homePath = config.HOME_PAGE + (org && org.code ? '/' + org.code : '');
        next({path: homePath});
        return;
    }
    //页面中转
    if (to.name === 'index' || to.path === '/index' || to.path === '/') {
        const org = getStore('currentOrganization', true);
        const homePath = config.HOME_PAGE + (org && org.code ? '/' + org.code : '');
        next({path: homePath});
        return;
    }
    // 权限节点校验：如果路由 meta 中定义了 permission，检查用户是否拥有该节点
    if (store.state.logged && to.meta && to.meta.permission) {
        const requiredNode = to.meta.permission;
        const permissionNodes = store.state.permissionNodes || getStore('permissionNodes', true) || [];
        // 如果有权限节点配置但用户不在权限列表中
        // 注意：当 permissionNodes 为空数组时，说明用户没有任何权限，应拒绝访问受保护页面
        if (!Array.isArray(permissionNodes) || !permissionNodes.includes(requiredNode)) {
            // 避免重定向死循环：如果来自错误页，不再重定向到403，而是中止导航
            if (from.path === '/403' || from.path === '/404' || from.path === '/500') {
                next(false);
                return;
            }
            next({path: '/403'});
            return;
        }
    }
    //无效页面跳转至404（仅已登录状态下判断）
    if (store.state.logged && !to.name && to.path !== HOME_PAGE) {
        next({path: '/404'});
        return;
    }
    next();
});
router.afterEach(route => {
    //预留
    // window.scrollTo(0,0)
});

// 全局捕获路由导航错误，忽略正常的导航中断（如守卫重定向、重复跳转）
const ignoredErrorNames = new Set([
    'NavigationDuplicated',
    'NavigationRedirected',
    'NavigationCancelled',
    'NavigationAborted',
]);
router.onError((error) => {
    if (error && ignoredErrorNames.has(error.name)) {
        return;
    }
    console.error('[Router Error]', error);
});

export default router
