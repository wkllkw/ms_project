<template>
    <div class="project-space-gantt" :class="project.task_board_theme">
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
                    <li><a class="app" data-app="tasks"
                           @click="$router.push('/project/space/task/' + code)">任务</a></li>
                    <li><a class="app" data-app="works"
                           @click="$router.push('/project/space/files/' + code)">
                        文件</a>
                    <li><a class="app" data-app="build"
                           @click="$router.push('/project/space/overview/' + code)">
                        概览</a>
                    </li>
                    <li class=""><a class="app" data-app="build"
                                    @click="$router.push('/project/space/features/' + code)">
                        版本</a>
                    </li>
                    <li><a class="app" data-app="build"
                           @click="$router.push('/project/space/events/' + code)">
                        日程</a>
                    </li>
                    <li class="actives"><a class="app" data-app="gantt">
                        甘特图</a>
                    </li>
                </ul>
            </section>
        </div>
        <wrapper-content :showHeader="false">
            <div class="gantt-wrapper">
                <div class="gantt-toolbar">
                    <div class="toolbar-left">
                        <a-radio-group v-model="viewMode" @change="refreshChart" button-style="solid" size="small">
                            <a-radio-button value="day">日</a-radio-button>
                            <a-radio-button value="week">周</a-radio-button>
                            <a-radio-button value="month">月</a-radio-button>
                        </a-radio-group>
                        <a-select v-model="filterStage" style="width: 160px; margin-left: 12px;" allowClear placeholder="按任务列表筛选" @change="refreshChart" size="small">
                            <a-select-option :key="'all'" :value="'all'">全部列表</a-select-option>
                            <a-select-option :key="stage.code" v-for="stage in taskStages" :value="stage.code">{{stage.name}}</a-select-option>
                        </a-select>
                    </div>
                    <div class="toolbar-right">
                        <a-tooltip title="刷新数据">
                            <a-button icon="reload" size="small" @click="init" :loading="loading"/>
                        </a-tooltip>
                        <a-tooltip title="只显示有时间的任务">
                            <a-switch checked-children="有时间" un-checked-children="全部" v-model="onlyWithTime" @change="refreshChart" style="margin-left: 8px;"/>
                        </a-tooltip>
                    </div>
                </div>
                <a-spin :spinning="loading">
                    <div v-if="ganttTasks.length === 0 && !loading" class="gantt-empty">
                        <a-empty description="暂无可展示的任务数据，请先为任务设置开始时间和截止时间"/>
                    </div>
                    <div v-else ref="ganttChart" class="gantt-chart" :style="{height: chartHeight + 'px'}"></div>
                </a-spin>
            </div>
        </wrapper-content>
    </div>
</template>

<script>
    import {mapState} from 'vuex'
    import moment from 'moment'
    import echarts from 'echarts'
    import projectSelect from '@/components/project/projectSelect'
    import {list as getTaskStages, tasks as getTasks} from "@/api/taskStages"
    import {read as getProject} from "@/api/project"
    import {collect} from "@/api/projectCollect"
    import {checkResponse} from "@/assets/js/utils"

    // 优先级颜色映射
    const PRI_COLORS = {
        0: '#1890ff',  // 普通 - 蓝色
        1: '#faad14',  // 紧急 - 橙色
        2: '#f5222d',  // 非常紧急 - 红色
    };
    const DONE_COLOR = '#52c41a'; // 已完成 - 绿色
    const NO_TIME_COLOR = '#d9d9d9'; // 无时间 - 灰色

    export default {
        name: "project-space-gantt",
        components: {
            projectSelect,
        },
        data() {
            return {
                code: this.$route.params.code,
                loading: false,
                project: {task_board_theme: 'simple'},
                taskStages: [],
                ganttTasks: [], // 处理后的甘特图任务数据
                viewMode: 'day', // day / week / month
                filterStage: 'all',
                onlyWithTime: true,
                chartInstance: null,
            }
        },
        computed: {
            ...mapState({
                userInfo: state => state.userInfo,
            }),
            chartHeight() {
                const taskCount = this.ganttTasks.length;
                return Math.max(400, taskCount * 36 + 100);
            },
        },
        watch: {
            $route(to) {
                if (this.code !== to.params.code) {
                    this.code = to.params.code;
                    this.getProject();
                    this.init();
                }
            },
        },
        created() {
            this.getProject();
            this.init();
        },
        mounted() {
            window.addEventListener('resize', this.handleResize);
        },
        beforeDestroy() {
            window.removeEventListener('resize', this.handleResize);
            if (this.chartInstance) {
                this.chartInstance.dispose();
            }
        },
        methods: {
            getProject() {
                getProject(this.code).then((res) => {
                    this.project = res.data;
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
            init() {
                this.loading = true;
                getTaskStages({projectCode: this.code, pageSize: 30}).then((res) => {
                    this.taskStages = res.data.list || [];
                    this.loadAllTasks();
                });
            },
            loadAllTasks() {
                if (!this.taskStages.length) {
                    this.loading = false;
                    this.ganttTasks = [];
                    return;
                }
                let promises = this.taskStages.map(stage => {
                    return getTasks({stageCode: stage.code}).then(res => {
                        return (res.data || []).map(task => ({
                            ...task,
                            stageName: stage.name,
                            stageCode: stage.code,
                        }));
                    });
                });
                Promise.all(promises).then(results => {
                    let allTasks = [];
                    results.forEach(tasks => {
                        allTasks = allTasks.concat(tasks);
                    });
                    this.processGanttData(allTasks);
                    this.loading = false;
                    this.$nextTick(() => {
                        this.renderChart();
                    });
                }).catch(() => {
                    this.loading = false;
                });
            },
            processGanttData(allTasks) {
                let tasks = allTasks;
                // 按列表筛选
                if (this.filterStage && this.filterStage !== 'all') {
                    tasks = tasks.filter(t => t.stageCode === this.filterStage);
                }
                // 只显示有时间的
                if (this.onlyWithTime) {
                    tasks = tasks.filter(t => t.begin_time || t.end_time);
                }
                // 按开始时间排序
                tasks.sort((a, b) => {
                    const aTime = a.begin_time || a.end_time || '9999-12-31';
                    const bTime = b.begin_time || b.end_time || '9999-12-31';
                    return moment(aTime).valueOf() - moment(bTime).valueOf();
                });
                this.ganttTasks = tasks;
            },
            refreshChart() {
                // 重新处理和渲染
                this.loading = true;
                getTaskStages({projectCode: this.code, pageSize: 30}).then((res) => {
                    this.taskStages = res.data.list || [];
                    this.loadAllTasks();
                });
            },
            renderChart() {
                if (!this.$refs.ganttChart || this.ganttTasks.length === 0) return;

                if (this.chartInstance) {
                    this.chartInstance.dispose();
                }
                this.chartInstance = echarts.init(this.$refs.ganttChart);

                const tasks = this.ganttTasks;
                const categories = tasks.map((t, i) => t.name || `任务${i + 1}`);

                // 计算时间范围
                let minTime = Infinity, maxTime = -Infinity;
                const now = moment();
                tasks.forEach(t => {
                    const start = t.begin_time ? moment(t.begin_time).valueOf() : null;
                    const end = t.end_time ? moment(t.end_time).valueOf() : null;
                    if (start && start < minTime) minTime = start;
                    if (end && end > maxTime) maxTime = end;
                    if (start && start > maxTime) maxTime = start;
                    if (end && end < minTime) minTime = end;
                });

                if (minTime === Infinity) {
                    minTime = now.clone().subtract(7, 'days').valueOf();
                    maxTime = now.clone().add(7, 'days').valueOf();
                }

                // 扩展时间范围
                const range = maxTime - minTime;
                minTime -= range * 0.05 || 86400000;
                maxTime += range * 0.05 || 86400000;

                // 构建数据
                const seriesData = tasks.map((t, i) => {
                    const start = t.begin_time ? moment(t.begin_time).valueOf() : null;
                    const end = t.end_time ? moment(t.end_time).valueOf() : null;

                    let barStart, barEnd;
                    if (start && end) {
                        barStart = start;
                        barEnd = end;
                    } else if (start) {
                        barStart = start;
                        barEnd = start + 86400000; // 默认1天
                    } else if (end) {
                        barStart = end - 86400000;
                        barEnd = end;
                    } else {
                        barStart = now.valueOf();
                        barEnd = now.valueOf() + 86400000;
                    }

                    let color;
                    if (t.done) {
                        color = DONE_COLOR;
                    } else if (!start && !end) {
                        color = NO_TIME_COLOR;
                    } else {
                        color = PRI_COLORS[t.pri] || PRI_COLORS[0];
                    }

                    return {
                        name: t.name,
                        value: [i, barStart, barEnd, t.done ? 100 : (t.schedule || 0)],
                        itemStyle: {color: color},
                        task: t,
                    };
                });

                // 今日线
                const todayLine = now.valueOf();

                const option = {
                    tooltip: {
                        formatter: function (params) {
                            const task = params.data.task;
                            const start = task.begin_time ? moment(task.begin_time).format('YYYY-MM-DD') : '未设置';
                            const end = task.end_time ? moment(task.end_time).format('YYYY-MM-DD') : '未设置';
                            const status = task.done ? '<span style="color:#52c41a">已完成</span>' : '<span style="color:#1890ff">进行中</span>';
                            const pri = ['普通', '紧急', '非常紧急'][task.pri] || '普通';
                            const executor = task.executor ? task.executor.name : '未指派';
                            return `<div style="padding:4px 0">
                                <div style="font-weight:bold;margin-bottom:4px">${task.name}</div>
                                <div>列表：${task.stageName}</div>
                                <div>执行者：${executor}</div>
                                <div>优先级：${pri}</div>
                                <div>状态：${status}</div>
                                <div>开始：${start}</div>
                                <div>截止：${end}</div>
                            </div>`;
                        },
                        backgroundColor: '#fff',
                        textStyle: {color: '#333'},
                        borderWidth: 1,
                        borderColor: '#e8e8e8',
                    },
                    grid: {
                        left: '220',
                        right: '40',
                        top: '40',
                        bottom: '50',
                    },
                    xAxis: {
                        type: 'time',
                        min: minTime,
                        max: maxTime,
                        axisLabel: {
                            formatter: (value) => {
                                if (this.viewMode === 'month') {
                                    return moment(value).format('YYYY-MM');
                                } else if (this.viewMode === 'week') {
                                    return moment(value).format('MM-DD') + '\n第' + moment(value).isoWeek() + '周';
                                }
                                return moment(value).format('MM-DD');
                            },
                        },
                        splitLine: {show: true, lineStyle: {type: 'dashed', color: '#f0f0f0'}},
                    },
                    yAxis: {
                        type: 'category',
                        data: categories,
                        inverse: true,
                        axisLabel: {
                            formatter: function (value) {
                                return value.length > 14 ? value.substring(0, 14) + '...' : value;
                            },
                            width: 200,
                            overflow: 'truncate',
                        },
                        splitLine: {show: true, lineStyle: {type: 'dashed', color: '#f0f0f0'}},
                    },
                    dataZoom: [
                        {
                            type: 'slider',
                            xAxisIndex: 0,
                            filterMode: 'none',
                            bottom: 10,
                            height: 20,
                        },
                        {
                            type: 'inside',
                            xAxisIndex: 0,
                            filterMode: 'none',
                        }
                    ],
                    series: [
                        {
                            type: 'custom',
                            renderItem: function (params, api) {
                                const categoryIndex = api.value(0);
                                const start = api.coord([api.value(1), categoryIndex]);
                                const end = api.coord([api.value(2), categoryIndex]);
                                const height = api.size([0, 1])[1] * 0.6;

                                const rectShape = echarts.graphic.clipRectByRect({
                                    x: start[0],
                                    y: start[1] - height / 2,
                                    width: Math.max(end[0] - start[0], 6),
                                    height: height,
                                }, {
                                    x: params.coordSys.x,
                                    y: params.coordSys.y,
                                    width: params.coordSys.width,
                                    height: params.coordSys.height,
                                });

                                if (rectShape) {
                                    // 进度条
                                    const progress = api.value(3) / 100;
                                    const progressWidth = Math.max((end[0] - start[0]) * progress, 0);

                                    const group = {
                                        type: 'group',
                                        children: [
                                            // 背景条
                                            {
                                                type: 'rect',
                                                shape: rectShape,
                                                style: api.style({
                                                    fill: api.visual('color'),
                                                    opacity: 0.35,
                                                }),
                                            },
                                            // 进度条
                                            {
                                                type: 'rect',
                                                shape: {
                                                    ...rectShape,
                                                    width: Math.min(progressWidth, rectShape.width),
                                                },
                                                style: api.style({
                                                    fill: api.visual('color'),
                                                    opacity: 1,
                                                }),
                                            }
                                        ]
                                    };
                                    return group;
                                }
                            },
                            encode: {
                                x: [1, 2],
                                y: 0,
                            },
                            data: seriesData,
                        },
                        // 今日标记线
                        {
                            type: 'line',
                            markLine: {
                                silent: true,
                                symbol: 'none',
                                lineStyle: {
                                    color: '#f5222d',
                                    type: 'dashed',
                                    width: 1,
                                },
                                data: [
                                    {xAxis: todayLine, label: {formatter: '今日', position: 'start'}}
                                ],
                            },
                            data: [],
                        }
                    ]
                };

                this.chartInstance.setOption(option);

                // 点击事件 - 跳转任务详情
                this.chartInstance.off('click');
                this.chartInstance.on('click', (params) => {
                    if (params.data && params.data.task) {
                        const task = params.data.task;
                        this.$router.push(`/project/space/task/${this.code}/detail/${task.code}`);
                    }
                });
            },
            handleResize() {
                if (this.chartInstance) {
                    this.chartInstance.resize();
                }
            },
        }
    }
</script>

<style lang="less">
    .project-space-gantt {
        .gantt-wrapper {
            padding: 16px;
        }

        .gantt-toolbar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
            padding: 8px 12px;
            background: #fafafa;
            border-radius: 4px;
            border: 1px solid #e8e8e8;

            .toolbar-left {
                display: flex;
                align-items: center;
            }

            .toolbar-right {
                display: flex;
                align-items: center;
            }
        }

        .gantt-chart {
            width: 100%;
            min-height: 400px;
        }

        .gantt-empty {
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 400px;
        }
    }
</style>
