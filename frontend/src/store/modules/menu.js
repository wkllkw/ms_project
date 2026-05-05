import {getMenuForUser} from "@/api/menu";
import {getStore, setStore} from '@/assets/js/storage';
import {checkResponse, normalizeMenuTree} from '@/assets/js/utils';

const common = {
    state: {
        menu: getStore('menu', true),
        breadCrumbInfo: []
    },
    mutations: {
        SET_MENU(state, data) {
            state.menu = data;
        },
        setBreadCrumbInfo(state, data) {
            state.breadCrumbInfo = data;
        }
    },
    actions: {
        GET_MENU({commit}) {
            return getMenuForUser().then(res => {
                if (checkResponse(res)) {
                    // 菜单API现在返回 { menus: [...], nodes: [...] }
                    let menuData = [];
                    let nodes = [];
                    if (res.data && Array.isArray(res.data.menus)) {
                        menuData = res.data.menus;
                        nodes = res.data.nodes || [];
                    } else if (Array.isArray(res.data)) {
                        // 兼容旧格式：直接返回菜单数组
                        menuData = res.data;
                    } else if (res.data && Array.isArray(res.data.menusFormat)) {
                        menuData = res.data.menusFormat;
                    }
                    setStore('menu', menuData);
                    commit('SET_MENU', menuData);
                    // 存储权限节点
                    if (nodes.length > 0) {
                        setStore('permissionNodes', nodes);
                        commit('SET_PERMISSION_NODES', nodes);
                    }
                }
                return res;
            })
        },
        SET_MENU({commit},data) {
            const menuData = normalizeMenuTree(Array.isArray(data) ? data : []);
            setStore('menu', menuData);
            commit('SET_MENU', menuData);
        },
        setBreadCrumbInfo({commit}, data) {
            commit('setBreadCrumbInfo', data);
        }
    }

};
export default common
