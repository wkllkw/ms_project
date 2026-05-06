<!--suppress ALL-->
<template>
    <div class="index">
        <a-spin :spinning="windowLoading">
            <a-layout id="layout" :class="layoutClass">
                <a-layout-header :class="{'collapsed':collapsed}">
                    <div class="logo" @click="toHome">
                        <img class="logo-img" src="../assets/image/common/logo.png" alt="">
                        <span class="title" v-if="system">
                            FlowHub
                             <span class="version">{{system.app_version}}</span>
                        </span>
                    </div>
                    <div class="right-menu">
                        <div class="m-r-lg" v-if="config.WS_URI">
                            <a-badge title="当前在线" :count="online" showZero :numberStyle="{backgroundColor: '#52c41a'} "
                                     :offset="[10,0]">
                                <a-icon type="team"/>
                            </a-badge>
                        </div>
                        <div class="action action-organization" v-if="organizationList && organizationList.length > 1">
                            <header-select></header-select>
                        </div>
                        <div class="action">
                            <header-notice></header-notice>
                        </div>
                        <div class="action action-avatar">
                            <header-avatar></header-avatar>
                        </div>
                    </div>
                </a-layout-header>
                <a-layout style="padding-top: 56px;">
                    <a-sider
                            mode="inline"
                            breakpoint="md"
                            collapsible
                            v-model="collapsed"
                            :trigger="null"
                    >
                        <div class="sider-inner">
                            <!-- 全量菜单树：一级菜单作为大分类，二级菜单作为子选项 -->
                            <div class="sider-menu-section" v-if="Array.isArray(menu) && menu.length > 0">
                                <a-menu
                                        :theme="theme"
                                        :openKeys="openKeys"
                                        v-model="selectedKeys"
                                        @click="siderMenuClick"
                                        @openChange="onOpenChange"
                                        mode="inline"
                                >
                                    <template v-for="item in menu">
                                        <!-- 一级菜单：有子菜单 → 展开分组（仅显示有可见子项的分组） -->
                                        <a-sub-menu
                                                v-if="item.children && item.children.filter(c => c.status && !c.is_inner).length > 0"
                                                :key="item.id.toString()"
                                        >
                                            <span slot="title">
                                                <a-icon v-if="item.icon" :type="item.icon"/>
                                                <span>{{item.title}}</span>
                                            </span>
                                            <template v-for="child in item.children">
                                                <!-- 二级菜单：有子菜单 → 继续展开（仅显示有可见子项的） -->
                                                <a-sub-menu
                                                        v-if="child.children && child.children.filter(gc => gc.status && !gc.is_inner).length > 0 && !child.is_inner && child.status"
                                                        :key="child.id.toString()"
                                                >
                                                    <span slot="title">
                                                        <a-icon v-if="child.icon" :type="child.icon"/>
                                                        <span>{{child.title}}</span>
                                                    </span>
                                                    <a-menu-item
                                                            v-for="grandChild in child.children"
                                                            v-if="!grandChild.is_inner && grandChild.status"
                                                            :key="grandChild.id.toString()"
                                                    >
                                                        <a-icon v-if="grandChild.icon" :type="grandChild.icon"/>
                                                        <span>{{grandChild.title}}</span>
                                                    </a-menu-item>
                                                </a-sub-menu>
                                                <!-- 二级菜单：无子菜单 → 直接显示菜单项（隐藏禁用项） -->
                                                <a-menu-item
                                                        v-else-if="!child.is_inner && child.status"
                                                        :key="child.id.toString()"
                                                >
                                                    <a-icon v-if="child.icon" :type="child.icon"/>
                                                    <span>{{child.title}}</span>
                                                </a-menu-item>
                                            </template>
                                        </a-sub-menu>
                                        <!-- 一级菜单：无子菜单 → 直接显示菜单项（隐藏禁用项） -->
                                        <a-menu-item
                                                v-else-if="item.status"
                                                :key="item.id.toString()"
                                        >
                                            <a-icon v-if="item.icon" :type="item.icon"/>
                                            <span>{{item.title}}</span>
                                        </a-menu-item>
                                    </template>
                                    <!-- 折叠按钮，放在菜单最下方 -->
                                    <a-menu-item class="menu-collapse-item" @click="()=> collapsed = !collapsed">
                                        <a-icon :type="collapsed ? 'menu-unfold' : 'menu-fold'"/>
                                        <span>{{ collapsed ? '展开侧边栏' : '收起侧边栏' }}</span>
                                    </a-menu-item>
                                </a-menu>
                            </div>

                            <!-- 菜单为空时的引导 -->
                            <div class="sider-empty" v-if="!menu || !menu.length">
                                <div class="sider-empty-icon" v-if="!collapsed">
                                    <a-icon type="appstore" />
                                </div>
                                <p class="sider-empty-text" v-if="!collapsed">暂无菜单数据</p>
                            </div>
                        </div>
                    </a-sider>
                    <a-layout
                            class="main-content"
                            :style="collapsed ? { paddingLeft: '72px'} : { paddingLeft: '240px'}">
                        <a-layout-content>
                            <transition name="router-fade" mode="out-in">
                                <a-spin :spinning="pageLoading">
                                    <router-view></router-view>
                                </a-spin>
                            </transition>
                        </a-layout-content>
                        <a-footer style="text-align: center">
                        </a-footer>
                    </a-layout>
                </a-layout>
            </a-layout>
            <Socket ref="socket" v-if="config.WS_URI"></Socket>
        </a-spin>
        <v-uploader></v-uploader>
        <command-palette ref="commandPalette"/>
        <ai-assistant v-if="logged"></ai-assistant>
    </div>
</template>
<script>
    import {mapState} from 'vuex'
    import ALayout from 'ant-design-vue/es/layout'
    import commonMixin from '../mixins/common';
    import HeaderNotice from '../components/layout/header/HeaderNotice';
    import HeaderAvatar from '../components/layout/header/HeaderAvatar';
    import HeaderSelect from '../components/layout/header/HeaderSelect';
    import CommandPalette from '../components/tools/CommandPalette';
    import VUploader from '../components/tools/VUploader';
    import AiAssistant from '../components/tools/AiAssistant';
    import Socket from '../components/websocket/socket';
    import config from "../config/config";
    import {notice} from "../assets/js/notice";
    import {getStore} from "../assets/js/storage";
    import {_getOrgList} from "../api/organization";

    const ASider = ALayout.Sider;
    const AFooter = ALayout.Footer;


    export default {
        name: 'index',
        mixins: [commonMixin],
        components: {
            HeaderNotice,
            HeaderAvatar,
            HeaderSelect,
            ALayout,
            ASider,
            AFooter,
            Socket,
            VUploader,
            CommandPalette,
            AiAssistant,
        },
        data() {
            return {
                collapsed: false,
                openKeys: [],
                openKeysTemp: [],
                selectedKeys: [],
                breadCrumbInfo: [],
                config: config,
                online: 0,
            }
        },
        computed: {
            ...mapState({
                theme: state => state.theme,
                logged: state => state.logged,
                menu: state => state.menu.menu,
                system: state => state.system,
                pageLoading: state => state.pageLoading,
                windowLoading: state => state.windowLoading,
                organizationList: state => state.organizationList,
                socketAction: state => state.socketAction,
            }),
            userInfo() {
                return getStore('userInfo', true) || {};
            },
            currentOrgName() {
                const org = getStore('currentOrganization', true);
                return org ? org.name : '';
            },
            layoutClass() {
                let className = 'layout-' + this.theme;
                const info = (this.$route.meta || {}).info || {};
                if (info.show_slider === false) {
                    className += ' hide';
                }
                return className;
            }
        },
        created() {
            if (this.logged) {
                this.ensureMenuReady();
            }
            if (this.$route.query.logged) {
                this.$store.dispatch('checkLogin');
            }
            if (this.$route.query.message) {
                notice({title: this.$route.query.message}, 'notice');
            }
            if (this.logged) {
                _getOrgList();
            }
        },

        watch: {
            $route: function (to, from) {
                this.checkLayout();
            },
            menu(val) {
                if (Array.isArray(val)) {
                    this.checkLayout();
                }
            },
            collapsed(v) {
                if (v) {
                    this.openKeysTemp = JSON.parse(JSON.stringify(this.openKeys));
                    this.openKeys = [];
                } else {
                    this.openKeys = JSON.parse(JSON.stringify(this.openKeysTemp));
                }
            },
            logged(val) {
                if (!val) {
                    this.$router.push({name: 'login'})
                }
            },
            socketAction(val) {
                if (val.action === 'connect' || val.action === 'onClose') {
                    this.online = val.data.online;
                }
            }
        },

        methods: {
            syncCurrentOrganizationWithRoute() {
                const routeCode = this.$route && this.$route.params ? this.$route.params.code : '';
                if (!routeCode) {
                    return;
                }
                const orgList = Array.isArray(this.organizationList) && this.organizationList.length
                    ? this.organizationList
                    : (getStore('organizationList', true) || []);
                if (!Array.isArray(orgList) || !orgList.length) {
                    return;
                }
                const matchedOrg = orgList.find(item => item && item.code === routeCode);
                if (!matchedOrg) {
                    return;
                }
                const currentOrganization = getStore('currentOrganization', true);
                if (!currentOrganization || currentOrganization.code !== matchedOrg.code) {
                    this.$store.dispatch('setCurrentOrganization', matchedOrg);
                }
            },
            ensureMenuReady() {
                if (Array.isArray(this.menu) && this.menu.length > 0) {
                    this.checkLayout();
                    // 确保权限节点已加载
                    if (!this.$store.state.permissionNodes || this.$store.state.permissionNodes.length === 0) {
                        this.$store.dispatch('FETCH_PERMISSION_NODES');
                    }
                    return;
                }
                this.$store.dispatch('GET_MENU').then(() => {
                    this.checkLayout();
                    // 获取权限节点
                    this.$store.dispatch('FETCH_PERMISSION_NODES');
                }).catch(() => {
                    this.checkLayout();
                });
            },

            checkLayout() {

                // 根据当前路由，高亮对应的侧边栏菜单项并展开其父级
                let that = this;
                const path = this.$route.path;
                const meta = this.$route.meta || {};
                const info = meta.info || {};

                that.breadCrumbInfo = [];

                if (!info.is_inner) {
                    that.selectedKeys = [];
                }

                if (!Array.isArray(that.menu)) {
                    // 没有菜单数据时，仅用路由 info 构建面包屑
                    if (info.title) {
                        that.breadCrumbInfo.push({title: info.title, 'path': info.fullUrl ? '/' + info.fullUrl : ''});
                    }
                    that.$store.dispatch('setBreadCrumbInfo', that.breadCrumbInfo);
                    return;
                }

                // 递归在菜单树中查找匹配当前路由的菜单项，同时记录祖先链（用于面包屑和展开）
                let foundItem = null;
                let parentIds = [];     // 祖先菜单的 id 列表（字符串），用于展开
                let ancestorItems = []; // 祖先菜单项对象列表，用于面包屑

                function findInTree(items, ancestorIds, ancestors) {
                    if (!items) return;
                    items.forEach(function (item) {
                        if (foundItem) return; // 已找到就不再匹配
                        // 匹配方式1: 通过 fullUrl 匹配路径
                        if (item.fullUrl && '/' + item.fullUrl === path) {
                            foundItem = item;
                            parentIds = ancestorIds.slice();
                            ancestorItems = ancestors.slice();
                        }
                        // 匹配方式2: 通过 meta.model 匹配一级菜单 id（用于内部页面，无 children 的情况）
                        if (!foundItem && meta.model && item.id == meta.model && !item.children) {
                            foundItem = item;
                            parentIds = ancestorIds.slice();
                            ancestorItems = ancestors.slice();
                        }
                        if (!foundItem && item.children) {
                            findInTree(
                                item.children,
                                ancestorIds.concat([item.id.toString()]),
                                ancestors.concat([item])
                            );
                        }
                    });
                }

                findInTree(that.menu, [], []);

                if (foundItem) {
                    if (!info.is_inner) {
                        that.selectedKeys = [foundItem.id.toString()];
                    }
                    // 展开所有祖先菜单
                    if (!that.collapsed) {
                        that.openKeys = parentIds;
                    } else {
                        that.openKeysTemp = parentIds;
                    }

                    // 构建面包屑：当前页面 title → 父级菜单 title → 顶级菜单 title
                    // WrapperContent 需要 breadCrumbInfo[0]（当前页）、[1]（父级）、[2]（顶级）
                    // 先 push 当前页面信息
                    let currentTitle = info.title || foundItem.title;
                    let currentPath = info.fullUrl ? '/' + info.fullUrl : '/' + (foundItem.fullUrl || '');
                    that.breadCrumbInfo.push({title: currentTitle, 'path': currentPath});

                    // 然后从祖先链倒序 push（最近的父级先 push，最顶级最后 push）
                    for (let i = ancestorItems.length - 1; i >= 0; i--) {
                        that.breadCrumbInfo.push({
                            title: ancestorItems[i].title,
                            'path': '/' + (ancestorItems[i].fullUrl || '')
                        });
                    }
                } else if (meta.model) {
                    // 内部页面（如任务看板）：高亮其所属菜单
                    let modelItem = null;
                    let modelAncestors = [];
                    function findModel(items, ancestorIds, ancestors) {
                        if (!items) return;
                        items.forEach(function (item) {
                            if (modelItem) return;
                            if (item.id == meta.model) {
                                modelItem = item;
                                parentIds = ancestorIds.slice();
                                modelAncestors = ancestors.slice();
                            }
                            if (!modelItem && item.children) {
                                findModel(
                                    item.children,
                                    ancestorIds.concat([item.id.toString()]),
                                    ancestors.concat([item])
                                );
                            }
                        });
                    }
                    findModel(that.menu, [], []);
                    if (modelItem) {
                        if (!modelItem.children) {
                            that.selectedKeys = [modelItem.id.toString()];
                        }
                        if (!that.collapsed) {
                            that.openKeys = parentIds.concat([modelItem.id.toString()]);
                        } else {
                            that.openKeysTemp = parentIds.concat([modelItem.id.toString()]);
                        }
                        // 面包屑：当前页 → modelItem → 顶级
                        let currentTitle = info.title || modelItem.title;
                        that.breadCrumbInfo.push({title: currentTitle, 'path': ''});
                        that.breadCrumbInfo.push({title: modelItem.title, 'path': '/' + (modelItem.fullUrl || '')});
                        for (let i = modelAncestors.length - 1; i >= 0; i--) {
                            that.breadCrumbInfo.push({
                                title: modelAncestors[i].title,
                                'path': '/' + (modelAncestors[i].fullUrl || '')
                            });
                        }
                    }
                } else {
                    // 未匹配到任何菜单项，用路由 info 构建面包屑
                    if (info.title) {
                        that.breadCrumbInfo.push({title: info.title, 'path': info.fullUrl ? '/' + info.fullUrl : ''});
                    }
                }

                that.$store.dispatch('setBreadCrumbInfo', that.breadCrumbInfo);
            },
            siderMenuClick(event) {
                // 点击侧边栏菜单项，根据 id 在菜单树中找到对应项并跳转
                let that = this;
                let clickedItem = null;

                function findById(items) {
                    if (!items) return;
                    items.forEach(function (item) {
                        if (item.id.toString() === event.key) {
                            clickedItem = item;
                        }
                        if (item.children) {
                            findById(item.children);
                        }
                    });
                }

                findById(that.menu);

                if (clickedItem && clickedItem.fullUrl) {
                    let turnPath = '/' + clickedItem.fullUrl;
                    if (turnPath === '/home') {
                        that.toHome();
                        return;
                    }
                    if (turnPath !== '/#' && that.$route.path !== turnPath) {
                        that.$router.push(turnPath);
                    }
                }
            },
            onOpenChange(openKeys) {
                // 手风琴模式：同层级只保留一个展开的 sub-menu
                if (!Array.isArray(this.menu)) {
                    this.openKeys = openKeys;
                    return;
                }
                const latestOpenKey = openKeys.find(key => this.openKeys.indexOf(key) === -1);
                // 检查是否是一级菜单 key
                const topLevelIds = this.menu.map(item => item.id.toString());
                if (topLevelIds.indexOf(latestOpenKey) !== -1) {
                    // 一级菜单：手风琴效果，只保留最新展开的一级 + 可能的二级
                    this.openKeys = latestOpenKey ? [latestOpenKey] : [];
                } else {
                    // 非一级菜单：正常展开
                    this.openKeys = openKeys;
                }
            },
        },
    }
</script>
<style lang="less">

</style>
