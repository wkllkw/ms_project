<template>
    <div class="project-space-index" :class="project.task_board_theme">
        <div class="project-navigation">
            <div class="project-nav-header">
                <a-breadcrumb>
                    <a-breadcrumb-item>
                        <router-link to="/home">
                            <a-icon type="home"/>
                            首页
                        </router-link>
                    </a-breadcrumb-item>
                    <a-breadcrumb-item>
                        <project-select class="nav-title" style="display: inline-block" :code="code"></project-select>
                        <span class="actions">
                             <a-tooltip :mouseEnterDelay="0.3" :title="project.collected ? '取消收藏' : '加入收藏'"
                                        @click="collectProject">
                            <a-icon type="star" theme="filled" style="color: grey;" v-show="!project.collected"/>
                            <a-icon type="star" theme="filled" style="color: #ffaf38;" v-show="project.collected"/>
                        </a-tooltip>
                        </span>
                        <span class="label label-normal" v-if="project.private === 0"><a-icon type="global"/> 公开</span>
                    </a-breadcrumb-item>
                </a-breadcrumb>
            </div>
            <section class="nav-body">
                <ul class="nav-wrapper nav nav-underscore pull-left">
                    <li class="actives"><a class="app" data-app="home">概览</a></li>
                    <li><a class="app" data-app="tasks"
                           @click="$router.push('/project/space/task/' + code)">任务</a></li>
                    <li><a class="app" data-app="works"
                           @click="$router.push('/project/space/files/' + code)">文件</a>
                    </li>
                    <li><a class="app" data-app="build"
                           @click="$router.push('/project/space/overview/' + code)">报表</a>
                    </li>
                    <li><a class="app" data-app="build"
                           @click="$router.push('/project/space/features/' + code)">版本</a>
                    </li>
                    <li><a class="app" data-app="build"
                           @click="$router.push('/project/space/events/' + code)">日程</a>
                    </li>
                    <li><a class="app" data-app="gantt"
                           @click="$router.push('/project/space/gantt/' + code)">甘特图</a>
                    </li>
                </ul>
            </section>
        </div>
        <wrapper-content :showHeader="false">
            <div class="overview-dashboard">
                <!-- 项目基本信息卡片 -->
                <div class="dashboard-header">
                    <div class="project-info-card">
                        <div class="project-cover">
                            <a-avatar :size="64" :src="project.cover" icon="project"/>
                        </div>
                        <div class="project-meta">
                            <h2>{{ project.name || '项目概览' }}</h2>
                            <p class="project-desc muted">{{ project.description || '暂无项目描述' }}</p>
                            <div class="project-tags">
                                <a-tag color="blue" v-if="project.owner_name">
                                    <a-icon type="user"/> {{ project.owner_name }}
                                </a-tag>
                                <a-tag color="green" v-if="project.create_time">
                                    <a-icon type="clock-circle"/> {{ formatDate(project.create_time) }}
                                </a-tag>
                                <a-tag :color="project.private === 0 ? 'cyan' : 'orange'">
                                    <a-icon :type="project.private === 0 ? 'global' : 'lock'"/>
                                    {{ project.private === 0 ? '公开' : '私有' }}
                                </a-tag>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- 统计卡片 -->
                <a-row :gutter="16" class="stats-row">
                    <a-col :span="6" v-for="(stat, index) in projectStats" :key="index">
                        <a-card class="stat-card" :bordered="false">
                            <a-statistic
                                :title="stat.title"
                                :value="stat.number"
                                :value-style="{ color: stat.color, fontWeight: 'bold' }"
                            >
                                <template slot="prefix">
                                    <a-icon :type="stat.icon"/>
                                </template>
                            </a-statistic>
                            <a-progress
                                v-if="stat.schedule !== undefined"
                                :strokeColor="stat.color"
                                :strokeWidth="3"
                                :showInfo="false"
                                :percent="stat.schedule"
                                size="small"
                            />
                        </a-card>
                    </a-col>
                </a-row>

                <!-- 最近动态 & 我的任务 -->
                <a-row :gutter="16" class="content-row">
                    <a-col :span="14">
                        <a-card title="最近动态" :bordered="false" class="content-card">
                            <a-list :loading="activitiesLoading" :data-source="activities" size="small">
                                <a-list-item slot="renderItem" slot-scope="item">
                                    <a-list-item-meta>
                                        <a-avatar slot="avatar" :src="item.member_avatar" size="small"/>
                                        <div slot="title">
                                            <span>{{ item.member_name }}</span>
                                            <span v-if="item.is_comment == 0" v-html="item.remark"></span>
                                            <template v-if="item.is_comment == 1">发表了评论</template>
                                            <span class="muted" style="float: right; font-size: 12px;">{{ formatTime(item.create_time) }}</span>
                                        </div>
                                        <div slot="description">
                                            <template v-if="item.action_type == 'task' && item.sourceInfo">
                                                <router-link class="muted"
                                                    :to="`/project/space/task/${item.project_code}/detail/${item.source_code}`">
                                                    {{ item.sourceInfo.name }}
                                                </router-link>
                                            </template>
                                        </div>
                                    </a-list-item-meta>
                                </a-list-item>
                                <div v-if="activities.length === 0 && !activitiesLoading" slot="footer" style="text-align:center;">
                                    <span class="muted">暂无动态</span>
                                </div>
                            </a-list>
                        </a-card>
                    </a-col>
                    <a-col :span="10">
                        <a-card title="我的待办" :bordered="false" class="content-card">
                            <a-list :loading="myTasksLoading" :data-source="myTasks" size="small">
                                <a-list-item slot="renderItem" slot-scope="item">
                                    <a-list-item-meta>
                                        <div slot="title">
                                            <a-icon type="check-circle" class="m-r-xs"
                                                :style="{color: item.done ? '#52c41a' : '#d9d9d9'}"
                                                @click="toggleTaskDone(item)"/>
                                            <router-link :to="`/project/space/task/${item.project_code}/detail/${item.code}`">
                                                {{ item.name }}
                                            </router-link>
                                        </div>
                                        <div slot="description">
                                            <span v-if="item.end_time" :class="getTimeClass(item.end_time)">
                                                <a-icon type="clock-circle"/> {{ formatDate(item.end_time) }}
                                            </span>
                                        </div>
                                    </a-list-item-meta>
                                </a-list-item>
                                <div v-if="myTasks.length === 0 && !myTasksLoading" slot="footer" style="text-align:center;">
                                    <span class="muted">暂无待办任务</span>
                                </div>
                            </a-list>
                        </a-card>
                    </a-col>
                </a-row>
            </div>
        </wrapper-content>
    </div>
</template>

<script>
import moment from 'moment'
import {read as getProject, _projectStats} from '@/api/project'
import {collect} from '@/api/projectCollect'
import {selfList as getMyTasks, taskDone, getLogBySelfProject} from '@/api/task'
import {relativelyTime} from '@/assets/js/dateTime'
import {checkResponse} from '@/assets/js/utils'
import projectSelect from '@/components/project/projectSelect'

export default {
    name: "project-space-index",
    components: {
        projectSelect,
    },
    data() {
        return {
            code: this.$route.params.code,
            project: {task_board_theme: 'simple'},
            activitiesLoading: false,
            myTasksLoading: false,
            activities: [],
            myTasks: [],
            projectStats: [
                {title: '未完成', key: 'unDone', number: 0, schedule: 0, color: '#1890ff', icon: 'undo'},
                {title: '已完成', key: 'done', number: 0, schedule: 0, color: '#52c41a', icon: 'check-circle'},
                {title: '已逾期', key: 'overdue', number: 0, schedule: 0, color: '#f5222d', icon: 'warning'},
                {title: '待认领', key: 'toBeAssign', number: 0, schedule: 0, color: '#faad14', icon: 'user-add'},
            ]
        }
    },
    watch: {
        '$route.params.code'(val) {
            if (val) {
                this.code = val;
                this.init();
            }
        }
    },
    created() {
        this.init();
    },
    methods: {
        init() {
            this.getProject();
            this.getProjectStats();
            this.getActivities();
            this.getMyTasks();
        },
        getProject() {
            if (!this.code) return;
            getProject(this.code).then(res => {
                if (res.data) {
                    this.project = res.data;
                }
            });
        },
        getProjectStats() {
            if (!this.code) return;
            _projectStats({projectCode: this.code}).then(res => {
                if (res.data) {
                    const taskStats = res.data;
                    const total = taskStats['total'] || 0;
                    this.projectStats.forEach(v => {
                        v.number = taskStats[v.key] || 0;
                        v.schedule = total ? parseInt(v.number / total * 100) : 0;
                    });
                }
            });
        },
        getActivities() {
            if (!this.code) return;
            this.activitiesLoading = true;
            getLogBySelfProject({projectCode: this.code, page: 1, pageSize: 10}).then(res => {
                if (res.data && res.data.list) {
                    this.activities = res.data.list.slice(0, 8);
                }
            }).finally(() => {
                this.activitiesLoading = false;
            });
        },
        getMyTasks() {
            this.myTasksLoading = true;
            getMyTasks({page: 1, pageSize: 8, done: 0}).then(res => {
                if (res.data && res.data.list) {
                    this.myTasks = res.data.list.slice(0, 8);
                }
            }).finally(() => {
                this.myTasksLoading = false;
            });
        },
        toggleTaskDone(task) {
            const newDone = task.done ? 0 : 1;
            taskDone(task.code, newDone).then(res => {
                if (checkResponse(res)) {
                    this.getMyTasks();
                    this.getProjectStats();
                }
            });
        },
        collectProject() {
            const type = this.project.collected ? 'cancel' : 'collect';
            collect(this.project.code, type).then((res) => {
                if (!checkResponse(res)) {
                    return;
                }
                this.project.collected = !this.project.collected;
            })
        },
        formatDate(time) {
            if (!time) return '';
            return moment(time).format('MM月DD日');
        },
        formatTime(time) {
            return relativelyTime(time);
        },
        getTimeClass(endTime) {
            if (!endTime) return '';
            const diff = moment(endTime).diff(moment(), 'days');
            if (diff < 0) return 'text-error';
            if (diff === 0) return 'text-warning';
            return 'muted';
        }
    }
}
</script>

<style scoped lang="less">
.project-space-index {
    .project-navigation {
        top: 65px;
        z-index: 4;
    }

    .overview-dashboard {
        padding: 16px;
        width: 1100px;
        margin: 12px auto auto;
    }

    .dashboard-header {
        margin-bottom: 16px;

        .project-info-card {
            display: flex;
            align-items: center;
            background: #fff;
            padding: 20px 24px;
            border-radius: 4px;

            .project-cover {
                margin-right: 20px;
            }

            .project-meta {
                h2 {
                    margin: 0 0 6px 0;
                    font-size: 20px;
                }

                .project-desc {
                    margin-bottom: 8px;
                    font-size: 13px;
                }

                .project-tags {
                    .ant-tag {
                        margin-right: 6px;
                    }
                }
            }
        }
    }

    .stats-row {
        margin-bottom: 16px;

        .stat-card {
            text-align: center;
            border-radius: 4px;

            /deep/ .ant-statistic-title {
                font-size: 13px;
                color: #999;
            }

            /deep/ .ant-statistic-content {
                font-size: 28px;
            }
        }
    }

    .content-row {
        .content-card {
            border-radius: 4px;

            /deep/ .ant-card-head {
                border-bottom: 1px solid #f0f0f0;
                font-size: 15px;
            }
        }
    }

    .text-error {
        color: #f5222d;
    }

    .text-warning {
        color: #faad14;
    }

    .muted {
        color: #999;
    }
}
</style>
