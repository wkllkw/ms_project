<template>
    <a-modal
        :visible="visible"
        :footer="null"
        :closable="false"
        :maskClosable="true"
        :width="560"
        :bodyStyle="{ padding: 0 }"
        :getContainer="getContainer"
        :afterClose="reset"
        @cancel="close"
        class="command-palette-modal"
    >
        <div class="command-palette">
            <div class="palette-header">
                <a-icon type="search" class="search-icon"/>
                <input
                    ref="searchInput"
                    v-model="query"
                    class="search-input"
                    placeholder="搜索项目、任务、成员或输入命令..."
                    @input="handleSearch"
                    @keydown.down.prevent="navigateDown"
                    @keydown.up.prevent="navigateUp"
                    @keydown.enter.prevent="executeSelected"
                    @keydown.esc.prevent="close"
                />
                <span class="search-hint">ESC 关闭</span>
            </div>
            <div class="palette-results" v-if="filteredItems.length > 0">
                <div
                    v-for="(item, index) in filteredItems"
                    :key="item.key"
                    class="result-item"
                    :class="{ active: index === activeIndex }"
                    @click="executeItem(item)"
                    @mouseenter="activeIndex = index"
                >
                    <div class="result-icon">
                        <a-icon :type="item.icon" :style="{ color: item.color || '#1890ff' }"/>
                    </div>
                    <div class="result-content">
                        <div class="result-title" v-html="highlightMatch(item.title)"/>
                        <div class="result-desc" v-if="item.desc">{{ item.desc }}</div>
                    </div>
                    <div class="result-shortcut" v-if="item.shortcut">
                        <kbd v-for="key in item.shortcut" :key="key">{{ key }}</kbd>
                    </div>
                </div>
            </div>
            <div class="palette-empty" v-else-if="query && !searching">
                <a-icon type="search" style="font-size: 24px; color: #d9d9d9;"/>
                <p>未找到匹配结果</p>
            </div>
            <div class="palette-footer">
                <span><kbd>↑</kbd><kbd>↓</kbd> 导航</span>
                <span><kbd>Enter</kbd> 执行</span>
                <span><kbd>Esc</kbd> 关闭</span>
            </div>
        </div>
    </a-modal>
</template>

<script>
import {selfList as selfProjectList} from '@/api/project';
import {selfList as selfTaskList} from '@/api/task';
import {getStore} from '@/assets/js/storage';

export default {
    name: 'CommandPalette',
    data() {
        return {
            visible: false,
            query: '',
            activeIndex: 0,
            searching: false,
            projectList: [],
            recentCommands: [],
            commandItems: [
                {key: 'cmd:home', title: '回到首页', icon: 'home', color: '#1890ff', desc: '跳转到首页', shortcut: ['G', 'H'], action: () => this.$router.push('/home/' + (getStore('currentOrganization', true) || {}).code).catch(() => {})},
                {key: 'cmd:task', title: '我的任务', icon: 'check-circle', color: '#52c41a', desc: '查看我执行的任务', shortcut: ['G', 'T'], action: () => this.$router.push('/project/list/my').catch(() => {})},
                {key: 'cmd:project', title: '项目列表', icon: 'project', color: '#722ed1', desc: '查看所有项目', shortcut: ['G', 'P'], action: () => this.$router.push('/project/list/my').catch(() => {})},
                {key: 'cmd:notify', title: '通知中心', icon: 'bell', color: '#faad14', desc: '查看通知消息', shortcut: ['G', 'N'], action: () => this.$router.push('/notify/notice').catch(() => {})},
                {key: 'cmd:members', title: '成员管理', icon: 'team', color: '#13c2c2', desc: '查看团队成员', shortcut: ['G', 'M'], action: () => this.$router.push('/members').catch(() => {})},
                {key: 'cmd:theme', title: '切换主题', icon: 'bulb', color: '#eb2f96', desc: '切换亮色/暗色模式', shortcut: ['T'], action: () => {
                    const currentTheme = this.$store.state.theme;
                    this.$store.dispatch('setTheme', currentTheme === 'dark' ? 'light' : 'dark');
                }},
                {key: 'cmd:calendar', title: '日程', icon: 'calendar', color: '#fa8c16', desc: '查看日程安排', shortcut: ['G', 'C'], action: () => this.$router.push('/calendar')},
                {key: 'cmd:setting', title: '个人设置', icon: 'setting', color: '#595959', desc: '打开个人设置', shortcut: ['G', 'S'], action: () => this.$router.push('/account/setting/base')},
            ],
        };
    },
    computed: {
        filteredItems() {
            if (!this.query) {
                // 默认显示命令 + 最近项目
                const projectItems = this.projectList.slice(0, 5).map(p => ({
                    key: 'project:' + p.code,
                    title: p.name,
                    icon: 'project',
                    color: '#722ed1',
                    desc: '项目 · ' + (p.description || '点击进入项目'),
                    action: () => this.$router.push('/project/space/task/' + p.code),
                }));
                return [...this.commandItems, ...projectItems];
            }
            const q = this.query.toLowerCase();
            const results = [];
            // 搜索命令
            this.commandItems.forEach(cmd => {
                if (cmd.title.toLowerCase().includes(q) || (cmd.desc && cmd.desc.toLowerCase().includes(q))) {
                    results.push(cmd);
                }
            });
            // 搜索项目
            this.projectList.forEach(p => {
                if (p.name && p.name.toLowerCase().includes(q)) {
                    results.push({
                        key: 'project:' + p.code,
                        title: p.name,
                        icon: 'project',
                        color: '#722ed1',
                        desc: '项目',
                        action: () => this.$router.push('/project/space/task/' + p.code).catch(() => {}),
                    });
                }
            });
            return results;
        },
    },
    mounted() {
        this.loadProjects();
        document.addEventListener('keydown', this.handleGlobalKeydown);
    },
    beforeDestroy() {
        document.removeEventListener('keydown', this.handleGlobalKeydown);
    },
    methods: {
        getContainer() {
            return document.body;
        },
        handleGlobalKeydown(e) {
            // Ctrl+K 或 Cmd+K 打开命令面板
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                this.toggle();
                return;
            }
            // 快捷键导航 (G + 字母，类似 Vim)
            if (!this.visible && !e.ctrlKey && !e.metaKey && !e.altKey) {
                const target = e.target;
                if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) return;
                if (e.key === 't') {
                    e.preventDefault();
                    this.$store.dispatch('setTheme', this.$store.state.theme === 'dark' ? 'light' : 'dark');
                }
            }
        },
        toggle() {
            this.visible = !this.visible;
            if (this.visible) {
                this.$nextTick(() => {
                    if (this.$refs.searchInput) this.$refs.searchInput.focus();
                });
            }
        },
        open() {
            this.visible = true;
            this.$nextTick(() => {
                if (this.$refs.searchInput) this.$refs.searchInput.focus();
            });
        },
        close() {
            this.visible = false;
        },
        reset() {
            this.query = '';
            this.activeIndex = 0;
        },
        handleSearch() {
            this.activeIndex = 0;
        },
        navigateUp() {
            if (this.activeIndex > 0) {
                this.activeIndex--;
            }
        },
        navigateDown() {
            if (this.activeIndex < this.filteredItems.length - 1) {
                this.activeIndex++;
            }
        },
        executeSelected() {
            const item = this.filteredItems[this.activeIndex];
            if (item) this.executeItem(item);
        },
        executeItem(item) {
            this.close();
            if (item.action) {
                item.action();
            }
        },
        highlightMatch(text) {
            if (!this.query) return text;
            const regex = new RegExp(`(${this.query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
            return text.replace(regex, '<mark>$1</mark>');
        },
        loadProjects() {
            selfProjectList({pageSize: 50}).then(res => {
                const data = res.data;
                this.projectList = Array.isArray(data) ? data : (data && data.list) || [];
            }).catch(() => {});
        },
    },
};
</script>

<style lang="less">
.command-palette-modal {
    .ant-modal {
        top: 80px;
        padding-bottom: 0;
    }
    .ant-modal-body {
        border-radius: 10px;
        overflow: hidden;
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
    }
}

.command-palette {
    .palette-header {
        display: flex;
        align-items: center;
        padding: 12px 16px;
        border-bottom: 1px solid #f0f0f0;
        gap: 10px;

        .search-icon {
            font-size: 18px;
            color: #bfbfbf;
            flex-shrink: 0;
        }

        .search-input {
            flex: 1;
            border: none;
            outline: none;
            font-size: 15px;
            color: rgba(0, 0, 0, 0.85);
            background: transparent;

            &::placeholder {
                color: #bfbfbf;
            }
        }

        .search-hint {
            font-size: 12px;
            color: #bfbfbf;
            flex-shrink: 0;
        }
    }

    .palette-results {
        max-height: 360px;
        overflow-y: auto;
        padding: 4px;

        .result-item {
            display: flex;
            align-items: center;
            padding: 10px 12px;
            border-radius: 6px;
            cursor: pointer;
            transition: background-color 0.1s;
            gap: 12px;

            &:hover, &.active {
                background: #f5f5f5;
            }

            .result-icon {
                width: 32px;
                height: 32px;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: 6px;
                background: rgba(24, 144, 255, 0.08);
                flex-shrink: 0;
                font-size: 16px;
            }

            .result-content {
                flex: 1;
                min-width: 0;

                .result-title {
                    font-size: 14px;
                    color: rgba(0, 0, 0, 0.85);
                    line-height: 1.4;
                    mark {
                        background: #ffe58f;
                        padding: 0 2px;
                        border-radius: 2px;
                        color: rgba(0, 0, 0, 0.85);
                    }
                }

                .result-desc {
                    font-size: 12px;
                    color: rgba(0, 0, 0, 0.45);
                    margin-top: 2px;
                    white-space: nowrap;
                    overflow: hidden;
                    text-overflow: ellipsis;
                }
            }

            .result-shortcut {
                display: flex;
                gap: 4px;
                flex-shrink: 0;

                kbd {
                    display: inline-block;
                    padding: 1px 6px;
                    font-size: 11px;
                    font-family: inherit;
                    border: 1px solid #d9d9d9;
                    border-radius: 3px;
                    background: #fafafa;
                    color: rgba(0, 0, 0, 0.45);
                    line-height: 1.6;
                }
            }
        }
    }

    .palette-empty {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 40px 0;
        color: rgba(0, 0, 0, 0.25);

        p {
            margin-top: 12px;
            font-size: 13px;
        }
    }

    .palette-footer {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 16px;
        padding: 8px 16px;
        border-top: 1px solid #f0f0f0;
        font-size: 12px;
        color: rgba(0, 0, 0, 0.35);

        span {
            display: flex;
            align-items: center;
            gap: 4px;
        }

        kbd {
            display: inline-block;
            padding: 0 4px;
            font-size: 10px;
            font-family: inherit;
            border: 1px solid #d9d9d9;
            border-radius: 3px;
            background: #fafafa;
            color: rgba(0, 0, 0, 0.45);
            line-height: 1.6;
        }
    }
}
</style>
