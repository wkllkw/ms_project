<template>
    <a-popover v-model="showNotice" overlayClassName="header-notice" trigger="click" placement="bottomRight">
        <template slot="content">
            <a-spin :spinning="loading">
                <a-tabs class="header-notice-content" :tabBarGutter="25">
                    <a-tab-pane key="1">
                        <span slot="tab">消息<span
                                v-if="total && totalSum['message']">({{totalSum['message']}})</span></span>
                        <template v-if="total && totalSum['message']">
                            <a-list>
                                <template v-for="item in list['message']">
                                    <a-list-item :key="item.id" @click="messageAction(item)">
                                        <a-list-item-meta :description="item.create_time">
                                             <span slot="title">
                                                    <p>{{item.title}}</p>
                                                    <p class="ant-list-item-meta-description" v-html="item.content"></p>
                                             </span>
                                            <a-avatar slot="avatar"
                                                      :src="item.avatar"/>
                                        </a-list-item-meta>
                                    </a-list-item>
                                </template>
                            </a-list>
                            <div class="footer-action">
                                <a class="item muted" @click="setRead('message')">清空消息</a>
                                <a class="item muted" @click="showMore('1')">查看更多</a>
                            </div>
                        </template>
                        <template v-else>
                            <div class="notFound">
                                <img src="../../../assets/image/notify/laba.svg" alt="not found">
                                <div>你已读完所有消息</div>
                            </div>
                        </template>
                    </a-tab-pane>
                    <a-tab-pane key="2">
                        <span slot="tab">任务通知<span
                                v-if="taskNoticeTotal">({{taskNoticeTotal}})</span></span>
                        <template v-if="taskNoticeList.length">
                            <a-list>
                                <template v-for="item in taskNoticeList">
                                    <a-list-item :key="item.id" @click="noticeAction(item)">
                                        <a-list-item-meta :description="item.create_time">
                                             <span slot="title">
                                                 <p v-html="item.title"></p>
                                                 <p class="ant-list-item-meta-description" v-html="item.content"></p>
                                             </span>
                                            <a-avatar style="background-color: white" slot="avatar"
                                                      src="https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png"/>
                                        </a-list-item-meta>
                                    </a-list-item>
                                </template>
                            </a-list>
                            <div class="footer-action">
                                <a class="item muted" @click="clearNotice('task')">清空任务通知</a>
                                <a class="item muted" @click="()=>{$router.push('/notify/notice?tab=task').catch(()=>{})}">查看更多</a>
                            </div>
                        </template>
                        <template v-else>
                            <div class="notFound">
                                <img src="../../../assets/image/notify/bell.svg" alt="not found">
                                <div>暂无任务通知</div>
                            </div>
                        </template>
                    </a-tab-pane>
                    <a-tab-pane key="3">
                        <span slot="tab">系统消息<span
                                v-if="systemNoticeTotal">({{systemNoticeTotal}})</span></span>
                        <template v-if="systemNoticeList.length">
                            <a-list>
                                <template v-for="item in systemNoticeList">
                                    <a-list-item :key="item.id" @click="noticeAction(item)">
                                        <a-list-item-meta :description="item.create_time">
                                             <span slot="title">
                                                 <p v-html="item.title"></p>
                                                 <p class="ant-list-item-meta-description" v-html="item.content"></p>
                                             </span>
                                            <a-avatar style="background-color: white" slot="avatar"
                                                      src="https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png"/>
                                        </a-list-item-meta>
                                    </a-list-item>
                                </template>
                            </a-list>
                            <div class="footer-action">
                                <a class="item muted" @click="clearNotice('system')">清空系统消息</a>
                                <a class="item muted" @click="()=>{$router.push('/notify/notice?tab=system').catch(()=>{})}">查看更多</a>
                            </div>
                        </template>
                        <template v-else>
                            <div class="notFound">
                                <img src="../../../assets/image/notify/laba.svg" alt="not found">
                                <div>暂无系统消息</div>
                            </div>
                        </template>
                    </a-tab-pane>
                </a-tabs>
            </a-spin>
        </template>
        <span>
          <a-badge :count="total">
            <a-icon class="action-item" type="bell"/>
          </a-badge>
        </span>
    </a-popover>
</template>

<script>
    import {mapState} from 'vuex'
    import moment from 'moment';
    import {_clearAll, noReads, setReadied} from "../../../api/notify";
    import {notice} from "../../../assets/js/notice";
    import {showMsgNotification} from "../../../assets/js/notify";
    import {selfList} from "../../../api/task";
    import {relativelyTime} from "../../../assets/js/dateTime";

    export default {
        name: 'HeaderNotice',
        data() {
            return {
                showNotice: false,
                loading: false,
                total: 0,
                messageTotal: 0,
                totalSum: {},
                list: {},
                task: {
                    page: 1,
                    pageSize: 5,
                    total: 0,
                    list: [],
                }
            }
        },
        computed: {
            ...mapState({
                socketAction: state => state.socketAction,
                currentOrganization: state => state.currentOrganization,
            }),
            // 从原始 notice 列表中提取任务通知（action 以 'task:' 开头）
            taskNoticeList() {
                const raw = (this.list && this.list['notice']) || [];
                return raw.filter(item => item.action && item.action.startsWith('task:'));
            },
            taskNoticeTotal() {
                return this.taskNoticeList.length;
            },
            // 从原始 notice 列表中提取系统消息（非任务通知）
            systemNoticeList() {
                const raw = (this.list && this.list['notice']) || [];
                return raw.filter(item => !item.action || !item.action.startsWith('task:'));
            },
            systemNoticeTotal() {
                return this.systemNoticeList.length;
            },
        },
        watch: {
            socketAction(val) {
                if (val.action === 'notice') {
                    this.init();
                } else if (val.action && val.action.startsWith('task:')) {
                    this.init();
                } else if (val.action === 'task' || val.action === 'events') {
                    this.init();
                    const permission = showMsgNotification(val.title, val.msg, {icon: val.data.notify.avatar});
                    if (permission === false) {
                        notice(val, 'notice', 'info', 10);
                    }
                }
            }
        },
        created() {
            this.init();
        },
        methods: {
            init() {
                this.fetchNotice();
            },
            fetchNotice() {
                let app = this;
                noReads().then(res => {
                    const data = res.data || {};
                    app.list = data.list || [];
                    app.messageTotal = data.total || 0;
                    app.total = app.messageTotal + app.task.total;
                    app.totalSum = data.totalSum || {};
                });
            },
            setRead(type) {
                if (this.list && this.list[type]) {
                    this.total -= this.list[type].length;
                    this.list[type] = [];
                }
                switch (type) {
                    case 'message':
                        this.totalSum.message = 0;
                        _clearAll();
                }
            },
            clearNotice(type) {
                if (type === 'task') {
                    // 清空任务通知
                    const taskItems = this.taskNoticeList;
                    this.total -= taskItems.length;
                    if (this.list && this.list['notice']) {
                        this.list['notice'] = this.list['notice'].filter(
                            item => !item.action || !item.action.startsWith('task:')
                        );
                    }
                } else if (type === 'system') {
                    // 清空系统消息
                    const sysItems = this.systemNoticeList;
                    this.total -= sysItems.length;
                    if (this.list && this.list['notice']) {
                        this.list['notice'] = this.list['notice'].filter(
                            item => item.action && item.action.startsWith('task:')
                        );
                    }
                }
            },
            showMore(key) {
                switch (key) {
                    default:
                        this.showNotice = false;
                        this.$router.push('/notify/notice').catch(() => {});
                }
            },
            getTasks() {
                selfList({page: this.task.page, pageSize: this.task.pageSize}).then(res => {
                    const data = res.data || {};
                    this.task.list = data.list || [];
                    this.task.total = data.total || 0;
                    this.total = this.messageTotal + this.task.total;
                })
            },
            messageAction(message) {
                const sendData = JSON.parse(message.send_data);
                this.showNotice = false;
                if (message.action === 'task') {
                    setReadied(JSON.stringify([message.id]));
                    this.$router.push(`/project/space/task/${sendData.project_code}/detail/${sendData.code}`).catch(() => {});
                }
                this.init();
            },
            noticeAction(item) {
                this.showNotice = false;
                setReadied(JSON.stringify([item.id]));
                if (item.action && item.action.startsWith('task:')) {
                    let taskCode = '';
                    let projectCode = '';
                    try {
                        const sendData = typeof item.send_data === 'string' ? JSON.parse(item.send_data) : item.send_data;
                        taskCode = sendData.taskCode || '';
                        projectCode = sendData.projectCode || '';
                    } catch (e) {}
                    if (taskCode && projectCode) {
                        const routeUrl = '/project/space/task/' + projectCode + '/detail/' + taskCode;
                        if (item.action === 'task:mention' || item.action === 'task:comment') {
                            this.$router.push({path: routeUrl, query: {scrollTo: 'comment'}}).catch(() => {});
                        } else {
                            this.$router.push(routeUrl).catch(() => {});
                        }
                    } else {
                        this.$router.push('/project/list/my').catch(() => {});
                    }
                    this.init();
                    return;
                }
                if (item.action && item.action !== '' && item.action.startsWith('/')) {
                    this.$router.push(item.action).catch(() => {});
                }
                this.init();
            },
            formatTime(time) {
                return moment(time).format('YY年MM月DD日 HH:mm');
            },
            showDiff(time, time2) {
                let diff = moment(time).diff(moment(time2), 'days');
                if (diff <= 0) {
                    diff = moment(time).diff(moment(time2), 'hours');
                    diff += '小时'
                } else {
                    diff += '天'
                }
                return diff;
            },
        }
    }
</script>

<style lang="less">
    .header-notice {
        .ant-popover-inner-content {
            padding: 0;

            .ant-tabs-bar {
                margin-bottom: 0 !important;
            }

            .ant-list {
                .ant-list-item {
                    padding: 12px 24px;
                    transition: all .3s;

                    &:hover {
                        background: #e6f7ff;
                    }
                }
            }
        }

        .header-notice-content {
            width: 340px;

            .ant-tabs-bar {
                text-align: center;
                margin-bottom: 5px;
            }

            .ant-list-item-meta-title {
                p {
                    margin-bottom: 0;
                    position: relative;
                }
            }

            .ant-list-item-meta-description {
                font-size: 12px;
            }

            .ant-list-item:hover {
                cursor: pointer;
            }

            .notFound {
                text-align: center;
                padding: 73px 0 88px 0;
                color: rgba(0, 0, 0, 0.45);
                height: 275px;

                img {
                    margin-bottom: 16px;
                }
            }

            .footer-action {
                border-top: 1px solid #e8e8e8;
                padding: 12px 0;

                .item {
                    width: 49%;
                    display: inline-block;
                    text-align: center;
                }
            }
        }
    }
</style>
