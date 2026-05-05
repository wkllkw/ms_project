import {notice} from './notice';
import {getStore} from "./storage";

/**
 * 判断客户端返回状态
 * @param res
 * @param show_msg
 * @returns {boolean}
 */
export const checkResponse = (res, show_msg = false) => {
    const code = res.code;
    const msg = res.msg;
    if (code !== 200) {
        if (show_msg) {
            notice(msg);
        }
        return false
    } else {
        return true
    }
};
/**
 * 创建路由对象
 * @returns {boolean}
 * @param data
 */
const normalizeRoutePath = (path) => {
    if (!path || typeof path !== 'string') {
        return '';
    }
    return path.replace(/^\/+/, '').trim();
};

const resolveRoutePath = (data) => {
    if (!data) {
        return '';
    }
    if (data.fullUrl) {
        return normalizeRoutePath(data.fullUrl);
    }
    let path = data.url || '';
    if (path && data.params) {
        path += '/' + data.params;
    }
    return normalizeRoutePath(path);
};

const resolveRouteFilePath = (data) => {
    if (!data) {
        return '';
    }
    return normalizeRoutePath(data.file_path || data.filePath || data.url || data.fullUrl || '');
};

const resolveMenuId = (item, index = 0, parentKey = 'menu') => {
    if (!item || typeof item !== 'object') {
        return `${parentKey}-${index}`;
    }
    const rawId = item.id ?? item.code ?? item.node ?? item.fullUrl ?? item.url;
    if (rawId === undefined || rawId === null || rawId === '') {
        return `${parentKey}-${index}`;
    }
    return rawId;
};

export const normalizeMenuTree = (menu = [], parentId = 0, parentKey = 'menu') => {
    if (!Array.isArray(menu)) {
        return [];
    }
    return menu
        .filter(item => item && typeof item === 'object')
        .map((item, index) => {
            const id = resolveMenuId(item, index, parentKey);
            const currentParentId = item.pid ?? item.parentId ?? parentId ?? 0;
            return {
                ...item,
                id,
                pid: currentParentId,
                parentId: currentParentId,
                children: normalizeMenuTree(item.children, id, `${parentKey}-${id}`),
            };
        });
};

export const createRoute = (data) => {
    const path = resolveRoutePath(data);
    const filePath = resolveRouteFilePath(data);
    if (!path || !filePath) {
        return null;
    }
    return {
        name: String(data.id || data.code || path),
        path: path,
        component: resolve => require(['@/views/' + filePath], resolve),
        meta: {model: data.pid || data.parentId || 0, info: data},
    };
};

export const flattenMenuRoutes = (menu = []) => {
    const routes = [];
    const walk = (items) => {
        if (!Array.isArray(items)) {
            return;
        }
        items.forEach((item) => {
            const route = createRoute(item);
            if (route) {
                routes.push(route);
            }
            if (Array.isArray(item.children) && item.children.length) {
                walk(item.children);
            }
        });
    };
    walk(normalizeMenuTree(menu));
    return routes;
};

export const appendMenuRoutes = (routes = [], menu = []) => {
    const exists = new Set((routes || []).filter(route => route && route.path).map(route => route.path));
    flattenMenuRoutes(menu).forEach((route) => {
        if (!exists.has(route.path)) {
            routes.push(route);
            exists.add(route.path);
        }
    });
    return routes;
};

export const getBase64 = (img, callback) => {
    const reader = new FileReader();
    reader.addEventListener('load', () => callback(reader.result));
    reader.readAsDataURL(img);
};

/**
 * 操作确认
 * @param options
 * @param callback
 */
export const showWarConfirm = (options = {}, callback = function () {
}) => {
    Modal.confirm({
        title: options.title || '操作提示',
        content: '<p>' + options.content + '</p>',
        loading: true,
        onOk: () => {
            callback()
        }
    });
};

/**
 * 获取完整的api请求地址
 */
export const getFullUrl = (api) => {
    return process.env.VUE_APP_API_URL + '/' + api
};

export const getApiUrl = (api) => {
    if (
        process.env.VUE_APP_CROSS_DOMAIN === 'true'
        || process.env.VUE_APP_CROSS_DOMAIN === true
        || process.env.VUE_APP_CROSS_DOMAIN === ''
        || process.env.VUE_APP_CROSS_DOMAIN === null) {
        return process.env.VUE_APP_API_URL + '/' + api; //开启跨域直接返回
    }
    if (process.env.NODE_ENV === 'production') {
        return process.env.VUE_APP_API_URL + '/' + api;
    } else {
        return '/api/' + api;
    }
};

/**
 * 获取上传文件地址
 * @param api
 * @returns {string}
 */
export const getUploadUrl = (api) => {
    let baseUrl = '';
    // if (process.env.NODE_ENV === 'production') {
    //     baseUrl = process.env.VUE_APP_API_URL
    // }
    return baseUrl + getApiUrl(api)
};


export const format_date = (data, show = true) => {
    //格式化时间
    let now = null;
    if (isNaN(data)) {
        now = new Date(data * 1000);

    } else {
        now = new Date(data);
    }
    let year = now.getFullYear();
    let month = now.getMonth() + 1;
    let date = now.getDate();
    let hour = now.getHours();
    let minute = now.getMinutes();
    // let second = now.getSeconds();
    if (month < 10) {
        month = '0' + month;
    }
    if (date < 10) {
        date = '0' + date;
    }
    if (hour < 10) {
        hour = '0' + hour;
    }
    if (minute < 10) {
        minute = '0' + minute;
    }
    const finally_date = {
        year: year,
        month: month,
        day: date,
        hour: hour,
        minute: minute
    };
    if (show) {
        return year + "-" + month + "-" + date + "   " + hour + ":" + minute;
    } else {
        return finally_date
    }
};
/**
 * 转换时间为可阅读格式，传入date的time值
 * @param time
 * @returns {*}
 */
export const prettyTime2Chinese = (time) => {
    if (!time) {
        return '';
    }
    if (isNaN(time)) {
        return '格式不正确';
    }
    var minute = 60 * 1000, //1分钟
        hour = 60 * minute, //1小时
        day = 24 * hour, //1天
        month = 12 * day,//月
        year = 12 * month;//年

    var diff = new Date().getTime() - time;
    var r = 0;
    if (diff > year) {
        r = parseInt(diff / year);
        return r + "年前";
    }
    if (diff > month) {
        r = parseInt(diff / month);
        return r + "个月前";
    }
    if (diff > day) {
        r = parseInt(diff / day);
        if (r == 1) {
            return "昨天";
        }
        return r + "天前";
    }
    if (diff > hour) {
        r = parseInt(diff / hour);
        return r + "个小时前";
    }
    if (diff > minute) {
        r = parseInt(diff / minute);
        return r + "分钟前";
    }
    return "刚刚";
};

/**
 * 数字转英文字母
 * 如：1->A
 * @param num
 * @returns {string}
 */
export const convert = (num) => {
    let result = "";
    while (num) {
        result = String.fromCharCode(--num % 26 + 65) + result;
        num = Math.floor(num / 26)
    }
    return result
};

export function timeFix() {
    const time = new Date();
    const hour = time.getHours();
    return hour < 9 ? '早上好' : (hour <= 11 ? '上午好' : (hour <= 13 ? '中午好' : (hour < 20 ? '下午好' : '晚上好')))
}

export function getAuthorization() {
    let tokenList = getStore('tokenList', true);
    if (tokenList) {
        let accessToken = tokenList.accessToken;
        let tokenType = tokenList.tokenType;
        return {Authorization: `${tokenType} ${accessToken}`};
    }
    return {};
}

/**
 * 获取推送消息
 * 如：1->A
 * @param num
 * @returns {string}
 */
export const getPushData = (data) => {
    return JSON.parse(data)
};
// export const snail = (array) => {
//   let arrs = [];
//   for(var i=0,l=array.length;i<l;i++) {
//     if(!Array.isArray(array[i])) arrs.push(array[i]);
//     else arrs = arrs.concat(snail(array[i]) );
//   }
//   return arrs;
// }
export const snail = (arr, value) => {
    for (var a in arr) {
        if (typeof (arr[a]) == "object") {
            return snail(arr[a], value); //递归遍历
        } else {
            if (a === 'path') {
                if (arr[a] === value) {
                    return true;
                }
            }
        }
    }
};

// 判断参数是否是其中之一
export function oneOf(value, validList) {
    for (let i = 0; i < validList.length; i++) {
        if (value === validList[i]) {
            return true;
        }
    }
    return false;
}

/**
 * 根据class获取节点
 * @param className
 * @param tag
 * @returns {NodeList}
 */
export const getClassObj = (className, tag) => {
    tag = tag || document;
    className = className || '*';
    let findarr = [];
    if (document.getElementsByClassName) {
        return document.getElementsByClassName(className)
    }
    let el = document.getElementsByTagName(tag);
    let pattern = new RegExp('(^|\\s)' + className + '(\\s|$)');
    for (let i = 0; i < el.length; i++) {
        if (pattern.test(el[i].className)) {
            findarr.push(el[i])
        }
    }
    return findarr;
};

/**
 * 判断某dom是否有滚动条
 * @param el dom对象
 * @param direction 垂直滚动或水平滚动
 * @returns {boolean}
 */
export const hasScrolled = (el, direction = "vertical") => {
    if (!el) {
        return false;
    }
    let overflow = el.currentStyle ? el.currentStyle.overflow :
        window.getComputedStyle(el).getPropertyValue("overflow");
    if (overflow === "hidden") return false;

    if (direction === "vertical") {
        return el.scrollHeight > el.clientHeight;
    } else if (direction === "horizontal") {
        return el.scrollWidth > el.clientWidth;
    }
};
/*判断token是否过期*/
export const isTokenExpired = (timeStamp) => {
    let expiredTime = timeStamp;
    /*获取本地时间*/
    let nowTime = new Date().getTime() / 1000;
    /*如果 < 10分钟，则说明即将过期*/
    return (expiredTime - nowTime) < 60 * 10
};

/**
 * 遍历多维数组的工具函数
 * @param {Array} arr 要遍历的数组
 * @param {Function} fn 回调函数
 */
export const eachArray = (arr, fn) => {
    if (!arr || !arr.length || typeof fn !== 'function') {
        return;
    }
    for (let i = 0; i < arr.length; i++) {
        const e = arr[i];
        if (e && Array.isArray(e)) {
            eachArray(e, fn);
        } else {
            fn.call(e, e);
        }
    }
};
