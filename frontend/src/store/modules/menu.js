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
                    const menuData = Array.isArray(res.data)
                        ? res.data
                        : (Array.isArray(res.data && res.data.menusFormat) ? res.data.menusFormat : []);
                    setStore('menu', menuData);
                    commit('SET_MENU', menuData);
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
