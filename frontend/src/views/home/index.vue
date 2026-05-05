<template>
    <div class="home-index">
        <div class="page-header enhanced-header">
            <p class="day-text muted">『 {{ yiyan.hitokoto }}』 —— 《{{ yiyan.from }}》 <a class="muted" @click="getYiYan">
                <a-icon type="reload"/>
            </a>
            </p>
            <div class="header-content" v-if="userInfo">
                <div class="left-content">
                    <div class="avatar">
                        <a-avatar :size="64" :src="userInfo.avatar">{{userInfo.name}}</a-avatar>
                    </div>
                    <div class="user-info">
                        <div class="title">{{helloTime}}{{ userInfo.name }}，祝你开心每一天！</div>
                        <div class="team muted" v-if="userInfo.position">{{userInfo.position}}
                            <template v-if="userInfo.department"> | {{ userInfo.department }}</template>
                        </div>
                    </div>
                </div>
                <div class="right-content">
                    <div class="content-item">
                        <div class="item-title muted">
                            <a-icon type="file-done" class="stat-icon" style="color:#1890ff;"/> 任务数
                        </div>
                        <div class="item-text">
                            <span>{{task.total}}</span>
                        </div>
                    </div>
                    <div class="content-item">
                        <div class="item-title muted">
                            <a-icon type="team" class="stat-icon" style="color:#52c41a;"/> 团队人数
                        </div>
                        <div class="item-text">
                            <span>{{accounts.total}}</span>
                        </div>
                    </div>
                    <div class="content-item">
                        <div class="item-title muted">
                            <a-icon type="appstore" class="stat-icon" style="color:#722ed1;"/> 项目总数
                        </div>
                        <div class="item-text">
                            <span>{{projectTotal}}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 快速操作区 + 逾期告警 -->
        <div class="quick-actions-bar">
            <div class="quick-actions-left">
                <a-button type="primary" icon="plus" @click="$router.push('/project/list/my')">创建项目</a-button>
                <a-button icon="calendar" @click="$router.push('/calendar')">我的日程</a-button>
                <a-button icon="bar-chart" @click="$router.push('/project/analysis')">数据分析</a-button>
            </div>
            <div class="quick-actions-right">
                <a-alert v-if="overdueCount > 0" :message="`你有 ${overdueCount} 个逾期任务需要处理！`" type="error" showIcon banner />
                <a-alert v-else-if="todayCount > 0" :message="`今日有 ${todayCount} 个任务待完成`" type="warning" showIcon banner />
            </div>
        </div>

        <div class="page-wrapper">
            <a-row class="page-wrapper-content" :gutter="24">
                <a-col class="project-list-content" :xl="16" :lg="24" :md="24" :sm="24" :xs="24">
                    <!-- 项目进度仪表盘 -->
                    <a-card class="progress-dashboard" :bordered="false" style="margin-bottom: 24px;" title="项目进度总览">
                        <router-link to="/project/analysis" slot="extra">详细分析</router-link>
                        <a-spin :spinning="dashboard.loading">
                            <div class="dashboard-stats">
                                <div class="dashboard-stat-item">
                                    <div class="stat-circle" style="border-color: #1890ff;">
                                        <span class="stat-value">{{ dashboard.totalProjects }}</span>
                                    </div>
                                    <div class="stat-label">总项目数</div>
                                </div>
                                <div class="dashboard-stat-item">
                                    <div class="stat-circle" style="border-color: #52c41a;">
                                        <span class="stat-value">{{ dashboard.avgProgress }}<small>%</small></span>
                                    </div>
                                    <div class="stat-label">平均进度</div>
                                </div>
                                <div class="dashboard-stat-item">
                                    <div class="stat-circle" style="border-color: #faad14;">
                                        <span class="stat-value">{{ dashboard.inProgress }}</span>
                                    </div>
                                    <div class="stat-label">进行中</div>
                                </div>
                                <div class="dashboard-stat-item">
                                    <div class="stat-circle" style="border-color: #f5222d;">
                                        <span class="stat-value">{{ dashboard.overdue }}</span>
                                    </div>
                                    <div class="stat-label">已逾期</div>
                                </div>
                            </div>
                            <div class="dashboard-project-bars" v-if="projectList.length">
                                <div class="project-bar-item" v-for="(item, i) in projectList.slice(0, 6)" :key="'bar-'+i">
                                    <div class="bar-info">
                                        <router-link :to="'/project/space/overview/' + item.code" class="bar-name">{{ item.name }}</router-link>
                                        <span class="bar-percent">{{ item.schedule }}%</span>
                                    </div>
                                    <a-progress :percent="item.schedule" :strokeWidth="8" :showInfo="false"
                                                :strokeColor="item.schedule >= 80 ? '#52c41a' : item.schedule >= 40 ? '#1890ff' : '#faad14'"/>
                                </div>
                            </div>
                            <div class="dashboard-empty" v-else-if="!dashboard.loading">
                                <a-icon type="pie-chart" style="font-size:32px;color:#d9d9d9;"/>
                                <p class="muted">暂无项目数据</p>
                            </div>
                        </a-spin>
                    </a-card>

                    <a-card
                            class="project-list"
                            :loading="loading"
                            style="margin-bottom: 24px;"
                            :bordered="false"
                            title="进行中的项目"
                            :body-style="{ padding: 0 }">
                        <router-link to="/project/list/my" slot="extra">全部项目</router-link>
                        <div style="display: flex; flex-direction: row; flex-wrap: wrap; justify-content: start;">
                            <a-card-grid class="project-card-grid" :key="i" v-for="(item, i) in projectList">
                                <a-card :bordered="false" :body-style="{ padding: 0 }" @click="routerLink('/project/space/task/' + item.code)">
                                    <img
                                        slot="cover"
                                        alt="项目封面"
                                        :src="item.cover || defaultCover"
                                        @error="onCoverError"
                                    />
                                    <a-card-meta>
                        <div slot="title" class="card-title">
                                            <router-link :to="'/project/space/task/' + item.code">
                                                <a-icon type="star" theme="filled" style="color: #ffaf38;margin-right: 6px;" v-show="item.collected"/>{{ item.name }}
                                            </router-link>
                                        </div>
                                        <div slot="description" class="card-description">
                                            <a-tooltip :mouseEnterDelay="0.3"
                                                       :title="item.description">
                                            <span class="description-text">
                                                <span v-if="item.description">{{ item.description }}</span>
                                                <span v-else>暂无介绍</span>
                                            </span>
                                            </a-tooltip>
                                            <a-tooltip placement="right" :mouseEnterDelay="0.3"
                                                       :title="`当前进度：${item.schedule}%`">
                                                <a-progress :strokeWidth="2" :showInfo="false"
                                                            :percent="item.schedule"/>
                                            </a-tooltip>
                                        </div>
                                    </a-card-meta>
                                    <div class="project-item">
                                        <a href="/#/">{{ item.owner_name }}</a>
                                        <span class="datetime">{{ formatTime(item.create_time) }}</span>
                                    </div>
                                </a-card>
                            </a-card-grid>
                            <p class="empty-state" v-if="!projectList.length">
                                <a-icon type="inbox" style="font-size:36px;color:#d9d9d9;display:block;margin-bottom:8px;"/>
                                <span>暂无进行中的项目</span>
                                <a-button type="link" size="small" @click="$router.push('/project/list/my')">去创建</a-button>
                            </p>
                        </div>
                    </a-card>
                    <a-card class="activities-list" :loading="loading" title="动态"  :bordered="false">
                        <a-list>
                            <a-list-item :key="index" v-for="(item, index) in activities">
                                <a-list-item-meta>
                                    <a-avatar slot="avatar" :src="item.member_avatar"/>
                                    <div slot="title">
                                        <span>{{ item.member_name }}</span>
                                        <span v-if="item.is_comment == 0">  <span v-html="item.remark "></span></span>&nbsp;
                                        <template v-if="item.is_comment == 1">发表了评论
                                            <p class="comment-text">{{ item.content }}</p>
                                        </template>
                                        <router-link v-if="item.action_type == 'task' && item.source_code" target="_blank"
                                                :to="`/project/space/task/${item.project_code}/detail/${item.source_code}`"
                                                class="right-item">「 {{ item.task_name || (item.sourceInfo && item.sourceInfo.name) }} 」
                                        </router-link>
                                    </div>
                                    <div slot="description">
                                        {{ formatTime(item.create_time) }} -
                                        <router-link v-if="item.project_code" target="_blank" :to="`/project/space/task/${item.project_code}`" class="muted">
                                            {{item.project_name || '项目动态'}}
                                        </router-link>
                                        <span v-else class="muted">{{item.project_name || '项目动态'}}</span>
                                    </div>
                                </a-list-item-meta>
                            </a-list-item>
                        </a-list>
                    </a-card>
                </a-col>
                <a-col
                        style="padding: 0 12px;flex: 1"
                        :xl="8"
                        :lg="24"
                        :md="24"
                        :sm="24"
                        :xs="24">
                    <a-card class="tasks-list" style="margin-bottom: 24px"
                            :bordered="false">
                        <div slot="title">
                            <div class="flex ant-row-flex-space-between ant-row-flex-middle">
                                <span>我的任务 · {{task.total}}</span>
                                <a-select v-model="task.done" @select="taskSelectChange" :defaultActiveFirstOption="false">
                                    <a-select-option :key="0">未完成</a-select-option>
                                    <a-select-option :key="1">已完成</a-select-option>
                                </a-select>
                            </div>
                        </div>
                        <a-tabs defaultActiveKey="1" :animated="false" @change="taskTabChange">
                            <a-tab-pane key="1">
                                <span slot="tab"><a-icon type="bars" />我执行的</span>
                            </a-tab-pane>
                            <a-tab-pane key="2">
                                <span slot="tab"><a-icon type="team" />我参与的</span>
                            </a-tab-pane>
                            <a-tab-pane key="3">
                                <span slot="tab"><a-icon type="rocket" />我创建的</span>
                            </a-tab-pane>
                        </a-tabs>
                        <a-list :loading="task.loading">
                            <a-list-item :key="index" v-for="(item, index) in task.list">
                                <a-list-item-meta>
                                    <div slot="title">
                                        <div style="display: flex;justify-content: space-between ">
                                            <a class="task-title-wrap">
                                                <a-tooltip title="优先级">
                                                    <a-tag :color="priColor(item.pri)">{{item.priText}}</a-tag>
                                                </a-tooltip>
                                                <a-tooltip placement="top">
                                                    <template slot="title">
                                                        <template v-if="item.pcode">
                                                            <span v-if="item.parentDone" style="font-size: 12px;">父任务已完成，无法重做子任务</span>
                                                            <span v-else-if="item.hasUnDone" style="font-size: 12px;">子任务尚未全部完成，无法完成父任务</span>
                                                        </template>
                                                        <template v-else>
                                                            <span v-if="item.hasUnDone" style="font-size: 12px;">子任务尚未全部完成，无法完成父任务</span>
                                                        </template>
                                                    </template>
                                                     <span class="check-box-wrapper task-item"
                                                           @click.stop="()=>{if(item.deleted || item.hasUnDone || (item.pcode && item.parentDone)) return false;taskDone(item.code, !item.done)}">
                                                        <a-icon class="check-box"
                                                                :class="{'disabled': item.deleted || item.parentDone || item.hasUnDone}"
                                                                :type="item.done ? 'check-square' : 'border'"
                                                                :style="{fontSize:'16px'}"/>
                                                </span>
                                                </a-tooltip>
                                                <a-tooltip :title="item.name">
                                                    <span @click="showTaskDetail = true;taskCode = item.code">{{ item.name }}</span>
                                                </a-tooltip>
                                            </a>
                                            <div>
                                                <a-tooltip title="任务开始 - 截止时间" v-if="item.end_time">
                                                    <span class="label m-r-sm" :class="showTimeLabel(item.end_time)">{{showTaskTime(item.begin_time, item.end_time)}}</span>
                                                </a-tooltip>
                                                <a-tooltip title="子任务" v-if="item.pcode">
                                                    <a-icon type="cluster" class="m-r-sm muted"/>
                                                </a-tooltip>
                                                <router-link target="_blank" class="muted" :to="'/project/space/task/' + item.projectInfo.code">
                                                    <a-tooltip title="所属项目">{{ item.projectInfo.name }}</a-tooltip>
                                                </router-link>
                                            </div>
                                        </div>
                                    </div>

                                </a-list-item-meta>
                            </a-list-item>
                        </a-list>
                        <a-pagination class="pull-right m-b" size="small" :defaultPageSize="task.pageSize" v-model="task.page" :total="task.total" @change="onLoadMoreTask"/>
                    </a-card>

                    <a-card class="events-list" :loading="events.loading" :title="`日程 · ${events.eventList.length}`" :bordered="false"  style="margin-bottom: 24px">
                        <router-link to="/calendar" slot="extra">日程日历</router-link>
                        <div class="list-content">
                            <a-list
                                :loading="events.loading"
                            >
                                <a-list-item class="list-item" :key="index" v-for="(item, index) in events.eventList">
                                    <a-list-item-meta>
                                        <div slot="title" style="display:flex;line-height: 20px;">
                                            <div class="info-item">
                                                <div class="text-center text-grey">
                                                    <div>{{ moment(item.begin_time).format('YYYY年MM月DD日 HH:mm') }}</div>
                                                    <div> ~</div>
                                                    <div>{{ moment(item.end_time).format('YYYY年MM月DD日 HH:mm') }}</div>
                                                </div>
                                            </div>
                                            <div class="info-item">
                                                <div class="line-item" style="font-size: 18px;margin-bottom: 20px;">
                                                    <span> {{ item.title }}</span>
                                                </div>
                                                <div class="line-item text-grey"> <a-icon type="environment" class="m-r-xs"/>{{ item.position }}</div>
                                                <template v-if="item.description">
                                                    <!--                                                <div class="line-item">备注</div>-->
                                                    <div class="line-item text-grey">{{item.description}}</div>
                                                </template>
                                                <div class="line-item">参与者 · {{item.memberList.length}}</div>
                                                <div class="line-item">
                                                    <template v-for="member in item.memberList">
                                                        <a-tooltip :title="`${member.memberInfo.name} ${member.is_owner ? ' · 组织者' : member.status ? member.status == 1 ? ' · 已接受' : ' · 已拒绝' : ' · 未响应'}`" :key="member.id">
                                                            <a-avatar :size="24" icon="user" :src="member.memberInfo.avatar"
                                                                      class="m-r-sm" />
                                                        </a-tooltip>
                                                    </template>
                                                </div>
                                                <template v-if="item.projectName">
                                                    <div class="line-item m-t text-grey" @click="routerLink('/project/space/events/' + item.project_code)">
                                                        <a-tag color="#52c41a" style="cursor: pointer;">{{item.projectName}}</a-tag>
                                                    </div>
                                                </template>
                                            </div>
                                            <div class="actions" style="position: absolute;right: 0;">
                                                <template v-if="item.waitConfirm">
                                                    <a-tooltip title="接受">
                                                        <a class="m-l-xs muted"><a-icon type="check"  @click="confirmJoinEvents(item, 1)"/></a>
                                                    </a-tooltip>
                                                    <a-tooltip title="拒绝">
                                                        <a class="m-l muted"> <a-icon type="close" @click="confirmJoinEvents(item, 2)"/></a>
                                                    </a-tooltip>
                                                </template>
                                            </div>
                                        </div>
                                    </a-list-item-meta>
                                    <div class="other-info muted">
                                    </div>
                                </a-list-item>
                                <div v-if="events.showLoadingMore" slot="loadMore"
                                     :style="{ textAlign: 'center', marginTop: '12px', height: '32px', lineHeight: '32px' }">
                                    <a-spin v-if="events.loadingMore"/>
                                    <a-button v-else @click="onLoadMoreEvents">查看更多日程</a-button>
                                </div>
                            </a-list>
                        </div>
                    </a-card>
                    <a-card :loading="loading" :title="'团队  · ' + accounts.total" :bordered="false">
                        <div class="members">
                            <a-row>
                                <a-col :span="8" v-for="(item, index) in accounts.list" :key="index">
                                    <a @click="routerLink('/members/profile/' + item.code + '?key=3')" style="display: flex;align-items: center"
                                    >
                                        <a-avatar size="small" :src="item.avatar"/>
                                        <span class="member">{{ item.name }}</span>
                                    </a>
                                </a-col>
                            </a-row>
                        </div>
                        <a-pagination class="pull-right m-b" :defaultPageSize="accounts.pageSize" size="small" v-show="accounts.total > accounts.pageSize" v-model="accounts.page" :total="accounts.total" @change="onLoadMoreAccounts"/>
                    </a-card>
                </a-col>
            </a-row>
        </div>

        <a-modal
            destroyOnClose
            class="task-detail-modal"
            width="min-content"
            :closable="false"
            :visible="showTaskDetail"
            title=""
            :footer="null"
            @cancel="detailClose"
        >
            <task-detail :taskCode="taskCode" @close="detailClose"></task-detail>

        </a-modal>
    </div>
</template>
<script>
    import {mapState} from 'vuex'
    import moment from "moment";
    import taskDetail from '../../components/project/taskDetail'
    import {getYiYan} from "@/api/other";
    import {formatTaskTime, relativelyTime, showHelloTime} from "assets/js/dateTime";
    import {selfList as getProjectList} from "../../api/project";
    import {list as accountList} from "../../api/user";
    import pagination from "../../mixins/pagination";
    import {getLogBySelfProject, selfList, taskDone} from "@/api/task";
    import task from "../project/space/task";
    import {confirmJoin, myList} from "@/api/projectEvents";
    import {checkResponse} from "assets/js/utils";

    export default {
        components: {
            taskDetail
        },
        mixins: [pagination],
        data() {
            return {
                moment,
                loading: false,
                yiyan: {},
                projectList: [],
                projectTotal: 0,
                defaultCover: 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="300" height="150"><defs><linearGradient id="d" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" style="stop-color:#e8e8e8"/><stop offset="100%" style="stop-color:#d9d9d9"/></linearGradient></defs><rect width="300" height="150" fill="url(#d)"/><rect x="120" y="55" width="60" height="40" rx="4" fill="#bbb"/><rect x="130" y="65" width="40" height="4" rx="1" fill="#999"/><rect x="135" y="75" width="30" height="4" rx="1" fill="#aaa"/></svg>'),
                activities: [],
                tasks: [],
                tasksTotal: 0,
                // accounts: [],
                accounts: {
                    list: [],
                    total: 0,
                    page: 1,
                    pageSize: 30,
                    loading: false,
                },
                task: {
                    list: [],
                    taskType: '1',
                    done: 0,
                    total: 0,
                    page: 1,
                    pageSize: 10,
                    loading: false,
                },
                showTaskDetail: false,
                taskCode: '',
                events: {
                    eventList: [],
                    showLoadingMore: false,
                    loadingMore: false,
                    total: 0,
                    page: 1,
                    pageSize: 10,
                    loading: false,
                },
                dashboard: {
                    loading: false,
                    totalProjects: 0,
                    avgProgress: 0,
                    inProgress: 0,
                    overdue: 0,
                }
            }
        },
        computed: {
            ...mapState({
                userInfo: state => state.userInfo,
                socketAction: state => state.socketAction,
            }),
            helloTime() {
                return showHelloTime()
            },
            overdueCount() {
                if (!this.task.list) return 0;
                return this.task.list.filter(item => {
                    if (item.done || !item.end_time) return false;
                    return moment(item.end_time).isBefore(moment(), 'day');
                }).length;
            },
            todayCount() {
                if (!this.task.list) return 0;
                return this.task.list.filter(item => {
                    if (item.done || !item.end_time) return false;
                    return moment(item.end_time).isSame(moment(), 'day');
                }).length;
            }
        },
        created() {
            this.getYiYan();
            this.init();
        },
        watch:{
            $route: function (to, from) {
                this.init();
            },
            socketAction(val) {
                if (val.action === 'organization:task') {
                    this.init(false, false);
                }
            },
        },
        methods: {
            init(reset = true, loading = true) {
                if (reset) {
                    this.projectList = [];
                    this.pagination.page = 1;
                    this.pagination.pageSize = 9;
                }
                this.getProjectList(loading);
                this.getTasks();
                this.getTaskLog();
                this.getAccountList();
                this.getEvents();
            },
            getProjectList(loading) {
                if (loading) {
                    this.loading = true;
                    this.dashboard.loading = true;
                }
                getProjectList(this.requestData).then(res => {
                    const data = res.data || {};
                    this.projectList = data.list || [];
                    this.projectTotal = data.total || 0;
                    this.loading = false;
                    // 计算仪表盘数据
                    this.dashboard.loading = false;
                    this.dashboard.totalProjects = data.total || 0;
                    let projects = data.list || [];
                    if (projects.length) {
                        let totalSchedule = 0;
                        let inProgress = 0;
                        let overdue = 0;
                        projects.forEach(p => {
                            totalSchedule += (p.schedule || 0);
                            if (p.schedule < 100) inProgress++;
                            if (p.end_time && moment(p.end_time).isBefore(moment(), 'day') && p.schedule < 100) {
                                overdue++;
                            }
                        });
                        this.dashboard.avgProgress = Math.round(totalSchedule / projects.length);
                        this.dashboard.inProgress = inProgress;
                        this.dashboard.overdue = overdue;
                    }
                });
            },
            getTaskLog() {
                getLogBySelfProject().then(res => {
                    const data = res.data || {};
                    this.activities = data.list || (Array.isArray(data) ? data : []);
                })
            },
            getAccountList() {
                this.accounts.loading = true;
                accountList({page: this.accounts.page, pageSize: this.accounts.pageSize}).then(res => {
                    const data = res.data || {};
                    this.accounts.loading = false;
                    this.accounts.list = data.list || [];
                    this.accounts.total = data.total || 0;
                })
            },
            getEvents() {
                let app = this;
                myList({page: this.events.page, pageSize: this.events.pageSize, deleted: 0}).then(res => {
                    const data = res.data || {};
                    app.events.eventList = app.events.eventList.concat(data.list || []);
                    app.events.total = data.total || 0;
                    app.events.showLoadingMore = app.events.total > app.events.eventList.length;
                    app.events.loading = false;
                    app.events.loadingMore = false
                })
            },
            getYiYan() {
                let app = this;
                getYiYan(function (data) {
                    app.yiyan = data
                }, 'd')
            },
            getTasks(reload = true) {
                if (reload) {
                    this.task.page = 1;
                }
                this.task.loading = true;
                selfList({page: this.task.page, pageSize: this.task.pageSize, taskType: this.task.taskType, type: this.task.done}).then(res => {
                    const data = res.data || {};
                    this.task.loading = false;
                    this.task.list = data.list || [];
                    // this.task.list =  this.task.list.concat(res.data.list);;
                    this.task.total = data.total || 0;
                })
            },
            taskTabChange(key) {
                this.task.taskType = key;
                this.task.loadingMore = true;
                this.getTasks();
            },
            taskSelectChange(value) {
                this.task.done = value;
                this.task.loadingMore = true;
                this.getTasks();
            },
            onLoadMoreTask(page, PageSize) {
                this.task.loadingMore = true;
                this.task.page = page;
                this.getTasks(false);
            },
            onLoadMoreAccounts(page, PageSize) {
                this.accounts.loadingMore = true;
                this.accounts.page = page;
                this.getAccountList();
            },
            detailClose() {
                this.taskCode = '';
                this.showTaskDetail = false;
                this.getTasks(false);
            },
            taskDone(taskCode, done) {
                done ? done = 1 : done = 0;
                taskDone(taskCode, done).then((res) => {
                    const result = checkResponse(res);
                    if (!result) {
                        return false;
                    }
                    this.getTasks(false);
                });
            },
            onLoadMoreEvents() {
                this.events.loadingMore = true;
                this.events.page++;
                this.getEvents();
            },
            confirmJoinEvents(events, status) {
                let app = this;
                confirmJoin({eventsCode: events.code, status: status}).then(res=>{
                    if (checkResponse(res)) {
                        events.waitConfirm = 0;
                        events.memberList.forEach(v => {
                            if (v.member_code == app.$store.state.userInfo.code ) {
                                v.status = status;
                            }
                        })
                    }
                });
            },
            priColor(pri) {
                switch (pri) {
                    case 1:
                        return '#ff9900';
                    case 2:
                        return '#ed3f14';
                    default:
                        return 'green';

                }
            },
            formatTime(time) {
                return relativelyTime(time);
            },
            onCoverError(e) {
                e.target.src = this.defaultCover;
            },
            showTaskTime(time, timeEnd) {
                return formatTaskTime(time, timeEnd);
            },
            showTimeLabel(time) {
                let str = 'label-primary';
                if (time == null) {
                    return str;
                }
                let cha = moment(moment(time).format("YYYY-MM-DD")).diff(moment().format("YYYY-MM-DD"), 'days');
                if (cha < 0) {
                    str = 'label-danger';
                } else if (cha == 0) {
                    str = 'label-warning';
                } else if (cha > 7) {
                    str = 'label-normal'
                }
                return str;
            },
        }
    }
</script>
<style lang="less">
    .home-index {

        // ===== 顶部欢迎区 - 极简白底 + 微装饰 =====
        .enhanced-header {
            background: #fff;
            border-bottom: 1px solid #f0f0f0;
            padding: 20px 28px;
            margin: 0;
            position: relative;
            overflow: hidden;

            &::after {
                content: '';
                position: absolute;
                top: 0;
                right: 0;
                width: 200px;
                height: 100%;
                background: linear-gradient(135deg, transparent 0%, rgba(58, 130, 248, 0.02) 100%);
                pointer-events: none;
            }

            .day-text {
                color: rgba(0, 0, 0, 0.35) !important;
                font-size: 13px;
                margin-bottom: 12px;

                a { color: rgba(0, 0, 0, 0.35) !important; }
            }

            .header-content {
                margin-bottom: 0;
                display: flex;
                justify-content: space-between;
                align-items: center;

                .left-content {
                    display: flex;
                    align-items: center;

                    .avatar {
                        .ant-avatar {
                            border: none;
                            box-shadow: none;
                        }
                    }

                    .user-info {
                        margin-left: 12px;

                        .title {
                            font-size: 18px;
                            font-weight: 600;
                            color: rgba(0, 0, 0, 0.85);
                        }

                        .team {
                            color: rgba(0, 0, 0, 0.4) !important;
                            font-size: 13px;
                        }
                    }
                }

                .right-content {
                    display: flex;

                    .content-item {
                        padding: 0 28px;
                        position: relative;

                        .item-title {
                            color: rgba(0, 0, 0, 0.4) !important;
                            font-size: 12px;
                            font-weight: 500;

                            .stat-icon {
                                color: rgba(0, 0, 0, 0.3) !important;
                                margin-right: 4px;
                            }
                        }

                        .item-text {
                            font-size: 26px;
                            font-weight: 600;
                            color: rgba(0, 0, 0, 0.85);

                            .small {
                                font-size: 18px;
                            }
                        }

                        &:after {
                            background-color: #f0f0f0;
                            position: absolute;
                            top: 4px;
                            right: 0;
                            width: 1px;
                            height: 36px;
                            content: "";
                        }

                        &:last-child {
                            &:after {
                                width: 0;
                            }
                        }
                    }
                }
            }
        }

        // ===== 快速操作栏 =====
        .quick-actions-bar {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 12px 28px;
            margin: 16px 28px 0;
            background: #fff;
            border-radius: 8px;
            border: 1px solid #f0f0f0;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
            transition: border-color 0.2s ease, box-shadow 0.25s ease;

            &:hover {
                border-color: rgba(58, 130, 248, 0.2);
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
            }

            .quick-actions-left {
                display: flex;
                gap: 8px;

                .ant-btn {
                    border-radius: 6px;
                    font-weight: 500;
                }
            }

            .quick-actions-right {
                flex: 0 0 auto;
                max-width: 400px;

                .ant-alert {
                    border-radius: 6px;
                    padding: 4px 12px;
                }
            }
        }

        // ===== 页面头部（通用） =====
        .page-header {
            .header-content {
                margin-bottom: 16px;
                display: flex;
                justify-content: space-between;

                .left-content {
                    display: flex;
                    align-items: center;

                    .user-info {
                        margin-left: 12px;
                        line-height: 33px;

                        .title {
                            font-size: 18px;
                        }
                    }
                }

                .right-content {
                    display: flex;

                    .content-item {
                        padding: 0 28px;
                        position: relative;

                        .item-text {
                            font-size: 26px;
                            font-weight: 600;

                            .small {
                                font-size: 18px;
                            }
                        }

                        &:after {
                            background-color: #f0f0f0;
                            position: absolute;
                            top: 4px;
                            right: 0;
                            width: 1px;
                            height: 36px;
                            content: "";
                        }

                        &:last-child {
                            &:after {
                                width: 0;
                            }
                        }
                    }
                }
            }
        }

        // ===== 页面主体 =====
        .page-wrapper {
            margin: 20px 28px;

            .page-wrapper-content {
                display: flex;
            }

            // 进度仪表盘
            .progress-dashboard {
                border-radius: 8px;
                overflow: hidden;
                border: 1px solid #f0f0f0;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
                transition: border-color 0.2s ease, box-shadow 0.25s ease;

                &:hover {
                    border-color: rgba(58, 130, 248, 0.25);
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
                }

                .dashboard-stats {
                    display: flex;
                    justify-content: space-around;
                    padding: 20px 0;

                    .dashboard-stat-item {
                        text-align: center;
                        transition: transform 0.2s ease;

                        &:hover {
                            transform: translateY(-2px);
                        }

                        .stat-circle {
                            width: 72px;
                            height: 72px;
                            border-radius: 50%;
                            border: 3px solid #3a82f8;
                            display: flex;
                            align-items: center;
                            justify-content: center;
                            margin: 0 auto 8px;
                            transition: border-color 0.2s ease, box-shadow 0.2s ease;
                            position: relative;
                            background: rgba(58, 130, 248, 0.03);

                            &:hover {
                                border-color: #5a96fa;
                                box-shadow: 0 2px 10px rgba(58, 130, 248, 0.15);
                            }

                            .stat-value {
                                font-size: 22px;
                                font-weight: 600;
                                color: rgba(0, 0, 0, 0.85);
                                animation: countUp 0.4s ease-out;

                                small {
                                    font-size: 13px;
                                    font-weight: 400;
                                }
                            }
                        }

                        .stat-label {
                            font-size: 12px;
                            color: rgba(0, 0, 0, 0.4);
                            font-weight: 500;
                        }
                    }
                }

                .dashboard-project-bars {
                    padding: 0 12px;

                    .project-bar-item {
                        margin-bottom: 10px;

                        .bar-info {
                            display: flex;
                            justify-content: space-between;
                            align-items: center;
                            margin-bottom: 4px;

                            .bar-name {
                                font-size: 13px;
                                color: rgba(0, 0, 0, 0.65);
                                overflow: hidden;
                                text-overflow: ellipsis;
                                white-space: nowrap;
                                max-width: 80%;

                                &:hover {
                                    color: #3a82f8;
                                }
                            }

                            .bar-percent {
                                font-size: 12px;
                                font-weight: 500;
                                color: rgba(0, 0, 0, 0.4);
                                flex-shrink: 0;
                            }
                        }

                        .ant-progress {
                            margin-bottom: 0;
                        }
                    }
                }

                .dashboard-empty {
                    text-align: center;
                    padding: 24px 0;

                    p {
                        margin-top: 8px;
                    }
                }
            }

            // 项目列表
            .project-list {
                border-radius: 8px;
                overflow: hidden;
                border: 1px solid #f0f0f0;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
                transition: border-color 0.2s ease, box-shadow 0.25s ease;

                &:hover {
                    border-color: rgba(58, 130, 248, 0.2);
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
                }

                .project-card-grid {
                    width: 25%;
                    padding: 10px;
                    cursor: pointer;
                    transition: background-color 0.15s ease, box-shadow 0.15s ease;

                    &:hover {
                        background: rgba(58, 130, 248, 0.03);
                    }
                }

                .ant-card-cover {
                    height: 110px;
                    overflow: hidden;

                    img {
                        display: block !important;
                        width: 100% !important;
                        height: 100% !important;
                        object-fit: cover !important;
                    }
                }

                .card-title {
                    font-size: 0;

                    a {
                        color: rgba(0, 0, 0, 0.85);
                        line-height: 24px;
                        height: 24px;
                        display: inline-block;
                        vertical-align: top;
                        font-size: 14px;
                        font-weight: 500;

                        &:hover {
                            color: #3a82f8;
                        }
                    }
                }

                .ant-card-meta-title {
                    margin-bottom: 0px;
                    margin-top: 8px;
                }

                .card-description {
                    color: rgba(0, 0, 0, 0.4);
                    height: 40px;
                    line-height: 20px;
                    overflow: hidden;
                    .description-text{
                        height: 20px;
                    }
                }

                .project-item {
                    display: flex;
                    margin-top: 6px;
                    overflow: hidden;
                    font-size: 12px;
                    height: 20px;
                    line-height: 20px;

                    a {
                        color: rgba(0, 0, 0, 0.4);
                        display: inline-block;
                        flex: 1 1 0;

                        &:hover {
                            color: #3a82f8;
                        }
                    }

                    .datetime {
                        color: rgba(0, 0, 0, 0.25);
                        flex: 0 0 auto;
                        float: right;
                    }
                }

                .ant-card-meta-description {
                    color: rgba(0, 0, 0, 0.4);
                    height: 40px;
                    line-height: 20px;
                    overflow: hidden;
                }
            }

            .empty-state {
                text-align: center;
                padding: 40px 0;
                color: #999;

                .ant-btn-link {
                    margin-top: 4px;
                }
            }

            // 活动列表
            .activities-list {
                border-radius: 8px;
                overflow: hidden;
                border: 1px solid #f0f0f0;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
                transition: border-color 0.2s ease, box-shadow 0.25s ease;

                &:hover {
                    border-color: rgba(58, 130, 248, 0.2);
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
                }

                .ant-list-item-meta-title {
                    position: relative;
                }

                .comment-text {
                    margin-bottom: 0;
                }

                .right-item {
                    float: right;
                    position: absolute;
                    right: 0;
                    top: 0;
                }
            }

            // 任务列表
            .tasks-list {
                border-radius: 8px;
                overflow: hidden;
                border: 1px solid #f0f0f0;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
                transition: border-color 0.2s ease, box-shadow 0.25s ease;

                &:hover {
                    border-color: rgba(58, 130, 248, 0.2);
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
                }

                .ant-card-body {
                    padding: 6px 20px;

                    .ant-list-item-meta, .ant-list-item-meta-content{
                        width: 100%;
                    }

                    .task-title-wrap{
                        flex: 1;
                        overflow: hidden;
                        text-overflow: ellipsis;
                        white-space: nowrap;
                        padding-right: 10px;

                        .check-box-wrapper {
                            text-align: center;
                            margin: 11px 2px 0 0;
                            padding: 10px 0;
                            transition: background 0.15s;
                            border-radius: 3px;
                            .check-box {
                                color: #b0b0b0;
                                cursor: pointer;
                                border-radius: 3px;
                                margin: 5px;
                            }
                            &:hover {
                                .check-box {
                                    color: rgba(0, 0, 0, 0.45);
                                }

                                background: #f5f5f5;
                            }
                        }
                    }
                }
            }

            // 日程列表
            .events-list {
                border-radius: 8px;
                overflow: hidden;
                border: 1px solid #f0f0f0;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
                transition: border-color 0.2s ease, box-shadow 0.25s ease;

                &:hover {
                    border-color: rgba(58, 130, 248, 0.2);
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
                }

                .ant-card-body {
                    padding: 0px 6px;

                    .ant-list-item-meta, .ant-list-item-meta-content{
                        width: 100%;
                    }
                }
                .list-content {

                    .list-item-title {
                        padding: 10px 20px;

                        .ant-list-item-action {
                            li {
                                color: #fff;
                            }

                            em {
                                width: 0;
                            }
                        }
                    }

                    .list-item {
                        margin-top: 10px;
                        border-bottom: none;
                        margin-bottom: 2px;
                        border-bottom: 1px solid #f5f5f5;
                        padding: 10px 20px;
                        transition: background-color 0.15s;

                        .ant-list-item-meta-title {
                            overflow: hidden;
                            text-overflow: ellipsis;
                            white-space: nowrap;
                            position: relative;
                            margin-bottom: 0;
                            line-height: 32px;
                        }

                        .ant-list-item-action {
                            em {
                                width: 0;
                            }
                        }
                    }

                    .info-item {
                        margin-right: 35px;
                    }

                    .line-item {
                        margin-bottom: 10px;
                    }

                    .other-info {
                        display: flex;

                        .info-item {
                            display: flex;
                            flex-direction: column;
                            padding-left: 0;
                            width: 105px;
                            text-align: right;
                        }

                        .schedule {
                            width: 250px;
                        }
                    }
                }
            }

            .item-group {
                padding: 20px 0 8px 24px;
                font-size: 0;

                a {
                    color: rgba(0, 0, 0, 0.65);
                    display: inline-block;
                    font-size: 14px;
                    margin-bottom: 13px;
                    width: 25%;
                }
            }

            .members {
                a {
                    display: block;
                    margin: 12px 0;
                    line-height: 24px;
                    height: 24px;

                    .member {
                        font-size: 14px;
                        color: rgba(0, 0, 0, .65);
                        line-height: 24px;
                        max-width: 100px;
                        vertical-align: top;
                        margin-left: 6px;
                        transition: color 0.15s;
                        display: inline-block;
                    }

                    &:hover {
                        span {
                            color: #3a82f8;
                        }
                    }
                }
            }
        }
    }
</style>
