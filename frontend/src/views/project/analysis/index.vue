<template>
    <div class="analysis-index">
        <div class="page-wrapper">
            <!-- 页面标题和导出按钮 -->
            <div class="analysis-header">
                <h2 class="analysis-title">
                    <a-icon type="bar-chart" /> 数据分析
                </h2>
                <div class="analysis-actions">
                    <a-dropdown>
                        <a-button icon="download">
                            导出报表 <a-icon type="down" />
                        </a-button>
                        <a-menu slot="overlay" @click="handleExport">
                            <a-menu-item key="png"><a-icon type="file-image" /> 导出为 PNG</a-menu-item>
                            <a-menu-item key="print"><a-icon type="printer" /> 打印报表</a-menu-item>
                        </a-menu>
                    </a-dropdown>
                    <a-button icon="reload" @click="init()" :loading="loading" class="m-l-sm">刷新数据</a-button>
                </div>
            </div>
            <a-row :gutter="24">
                <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
                    <chart-card :loading="loading" title="项目总数" :total="projectData.count | NumberFormat">
                        <a-tooltip title="所有参与的项目总数" slot="action">
                            <a-icon type="info-circle-o"/>
                        </a-tooltip>
                        <div class="chart-wrapper">
                            <ve-histogram
                                    :data="projectData.chartData"
                                    :settings="projectData.chartSettings"
                                    :extend="chartExtend"
                                    :legend-visible="false"
                                    height="55px"></ve-histogram>
                        </div>
                        <template slot="footer">本月立项 <span>{{ projectData.monthCount | NumberFormat }}</span></template>
                    </chart-card>
                </a-col>
                <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
                    <chart-card :loading="loading" title="任务总数" :total="taskData.count | NumberFormat">
                        <a-tooltip title="所有任务总数（含已完成）" slot="action">
                            <a-icon type="info-circle-o"/>
                        </a-tooltip>
                        <div>
                            <div class="chart-wrapper">
                                <ve-line
                                        :data="taskData.chartData"
                                        :settings="taskData.chartSettings"
                                        :extend="chartExtend"
                                        :legend-visible="false"
                                        height="55px"></ve-line>
                            </div>
                        </div>
                        <template slot="footer">已完成<span> {{ taskData.taskDoneCount | NumberFormat }}</span></template>
                    </chart-card>
                </a-col>
                <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
                    <chart-card :loading="loading" title="逾期任务" :total="taskData.taskOverdueCount | NumberFormat">
                        <a-tooltip title="超过截止日期仍未完成的任务" slot="action">
                            <a-icon type="info-circle-o"/>
                        </a-tooltip>
                        <div>
                            <trend flag="up" style="margin-right: 16px;">
                                <span slot="term">周同比</span>
                                {{ taskData.weekRate }}%
                            </trend>
                            <trend flag="down">
                                <span slot="term">日同比</span>
                                {{ taskData.dayRate }}%
                            </trend>
                        </div>
                        <template slot="footer">逾期率 <span>{{taskData.taskOverduePercent}}%</span></template>
                    </chart-card>
                </a-col>
                <a-col :sm="24" :md="12" :xl="6" :style="{ marginBottom: '24px' }">
                    <chart-card :loading="loading" title="整体进度" :total="`${projectData.projectSchedule}%`">
                        <a-tooltip title="所有项目的整体完成进度" slot="action">
                            <a-icon type="info-circle-o"/>
                        </a-tooltip>
                        <div>
                            <mini-progress color="#ffd401" :target="80" :percentage="projectData.projectSchedule" height="8px"/>
                        </div>
                        <template slot="footer">
                            <trend flag="down" style="margin-right: 16px;">
                                <span slot="term">完成率</span>
                                {{ taskData.donePercent }}%
                            </trend>
                            <trend flag="up">
                                <span slot="term">进行中</span>
                                {{ taskData.doingPercent }}%
                            </trend>
                        </template>
                    </chart-card>
                </a-col>
            </a-row>
            <a-card :loading="loading" :bordered="false" :body-style="{padding: '0'}">
                <div class="salesCard">
                    <a-tabs default-active-key="1" size="large"
                            :tab-bar-style="{marginBottom: '24px', paddingLeft: '16px'}">
                        <div class="extra-wrapper" slot="tabBarExtraContent">
                            <div class="extra-item">
                                <a :class="{'active-time': activeTimeRange === 'today'}" @click="changeTimeRange('today')">今日</a>
                                <a :class="{'active-time': activeTimeRange === 'week'}" @click="changeTimeRange('week')">本周</a>
                                <a :class="{'active-time': activeTimeRange === 'month'}" @click="changeTimeRange('month')">本月</a>
                                <a :class="{'active-time': activeTimeRange === 'year'}" @click="changeTimeRange('year')">本年</a>
                            </div>
                            <a-range-picker :style="{width: '256px'}" @change="onDateRangeChange"/>
                        </div>
                        <a-tab-pane forceRender tab="项目数" key="1">
                            <a-row>
                                <a-col :xl="16" :lg="12" :md="12" :sm="24" :xs="24">
                                    <div class="chart-wrappers-single">
                                        <ve-histogram
                                                :data="projectTotalData.chartData"
                                                :settings="projectTotalData.chartSettings"
                                                :extend="projectTotalData.chartExtend"
                                                :legend-visible="false"
                                                height="300px"></ve-histogram>
                                    </div>
                                </a-col>
                                <a-col :xl="8" :lg="12" :md="12" :sm="24" :xs="24">
                                    <rank-list title="项目数排行榜" :list="rankList"/>
                                </a-col>
                            </a-row>
                        </a-tab-pane>
                        <a-tab-pane forceRender tab="任务数" key="2">
                            <a-row>
                                <a-col :xl="16" :lg="12" :md="12" :sm="24" :xs="24">
                                    <div class="chart-wrappers-single">
                                        <ve-histogram
                                                :data="taskTotalData.chartData"
                                                :settings="taskTotalData.chartSettings"
                                                :extend="taskTotalData.chartExtend"
                                                :legend-visible="false"
                                                height="300px"></ve-histogram>
                                    </div>
                                </a-col>
                                <a-col :xl="8" :lg="12" :md="12" :sm="24" :xs="24">
                                    <rank-list title="任务数排行榜" :list="taskRankList"/>
                                </a-col>
                            </a-row>
                        </a-tab-pane>
                    </a-tabs>
                </div>
            </a-card>
            <a-row :gutter="12">
                <a-col :xl="12" :lg="24" :md="24" :sm="24" :xs="24">
                    <a-card :loading="loading" :bordered="false" title="我的项目" :style="{ marginTop: '24px' }">
                        <a-dropdown :trigger="['click']" placement="bottomLeft" slot="extra">
                            <a class="ant-dropdown-link" href="#">
                                <a-icon type="ellipsis"/>
                            </a>
                            <a-menu slot="overlay">
                                <a-menu-item @click="init()">
                                    <a-icon type="reload"/>
                                    <a>刷新数据</a>
                                </a-menu-item>
                                <a-menu-item @click="$router.push('/project/list/my')">
                                    <a-icon type="appstore"/>
                                    <a>查看全部项目</a>
                                </a-menu-item>
                            </a-menu>
                        </a-dropdown>
                        <div class="project-list-wrapper">
                            <a-list
                                    :loading="projectLoading"
                                    itemLayout="horizontal"
                                    :dataSource="projectList"
                            >
                                <a-list-item slot="renderItem" slot-scope="project" class="project-list-item" @click="$router.push('/project/space/task/' + project.code)">
                                    <a-list-item-meta :description="project.description || '暂无描述'">
                                        <span slot="title">{{ project.name }}</span>
                                        <a-avatar slot="avatar" :style="{backgroundColor: project.cover ? '' : '#1890ff'}" :src="project.cover">
                                            {{ project.cover ? '' : (project.name ? project.name.charAt(0) : 'P') }}
                                        </a-avatar>
                                    </a-list-item-meta>
                                    <div slot="actions">
                                        <a-tag v-if="project.collected" color="orange"><a-icon type="star" theme="filled"/> 收藏</a-tag>
                                    </div>
                                </a-list-item>
                            </a-list>
                        </div>
                        <a-pagination v-model="pagination.page" :total="projectTotal" size="small" @change="pageChange" style="margin-top: 12px;"/>
                    </a-card>
                </a-col>
                <a-col :xl="12" :lg="24" :md="24" :sm="24" :xs="24">
                    <a-card :loading="loading" :bordered="false" title="任务优先级分布" :style="{ marginTop: '24px' }">
                        <a-dropdown :trigger="['click']" placement="bottomLeft" slot="extra">
                            <a class="ant-dropdown-link" href="#">
                                <a-icon type="ellipsis"/>
                            </a>
                            <a-menu slot="overlay">
                                <a-menu-item @click="init()">
                                    <a-icon type="reload"/>
                                    <a>刷新数据</a>
                                </a-menu-item>
                            </a-menu>
                        </a-dropdown>
                        <div ref="priChart" style="height: 260px;"></div>
                    </a-card>
                </a-col>
            </a-row>
            <a-row :gutter="12">
                <a-col :xl="12" :lg="24" :md="24" :sm="24" :xs="24">
                    <a-card :loading="loading" :bordered="false" title="执行者任务分布" :style="{ marginTop: '24px' }">
                        <a-dropdown :trigger="['click']" placement="bottomLeft" slot="extra">
                            <a class="ant-dropdown-link" href="#">
                                <a-icon type="ellipsis"/>
                            </a>
                            <a-menu slot="overlay">
                                <a-menu-item @click="init()">
                                    <a-icon type="reload"/>
                                    <a>刷新数据</a>
                                </a-menu-item>
                            </a-menu>
                        </a-dropdown>
                        <div ref="executorChart" style="height: 260px;"></div>
                    </a-card>
                </a-col>
                <a-col :xl="12" :lg="24" :md="24" :sm="24" :xs="24">
                    <a-card :loading="loading" :bordered="false" title="任务完成趋势" :style="{ marginTop: '24px' }">
                        <a-dropdown :trigger="['click']" placement="bottomLeft" slot="extra">
                            <a class="ant-dropdown-link" href="#">
                                <a-icon type="ellipsis"/>
                            </a>
                            <a-menu slot="overlay">
                                <a-menu-item @click="init()">
                                    <a-icon type="reload"/>
                                    <a>刷新数据</a>
                                </a-menu-item>
                            </a-menu>
                        </a-dropdown>
                        <div ref="trendChart" style="height: 260px;"></div>
                    </a-card>
                </a-col>
            </a-row>
        </div>
    </div>
</template>
<script>
    import {mapState} from 'vuex'
    import moment from "moment"
    import echarts from 'echarts'
    import VeLine from 'v-charts/lib/line.common'
    import VeHistogram from 'v-charts/lib/histogram.common'
    import ChartCard from '@/components/chart/ChartCard'
    import Trend from '@/components/Trend'
    import MiniProgress from '@/components/chart/MiniProgress'
    import RankList from '@/components/chart/RankList'
    import pagination from "@/mixins/pagination"
    import {analysis, selfList as getProjectList, _projectStats} from "@/api/project"

    export default {
        components: {
            VeLine,
            VeHistogram,
            ChartCard,
            MiniProgress,
            Trend,
            RankList,
        },
        mixins: [pagination],
        data() {
            return {
                loading: false,
                activeTimeRange: 'year',
                rankList: [],
                taskRankList: [],
                chartExtend: {
                    grid: {
                        left: '-25',
                        right: '0',
                        top: '10',
                        bottom: '-15'
                    },
                    series: {
                        barWidth: 15,
                    },
                    xAxis: {
                        show: false,
                    },
                    yAxis: {
                        show: false,
                    },
                    tooltip: {
                        backgroundColor: '#fff',
                        textStyle: {
                            color: '#333'
                        },
                        borderWidth: 1,
                        borderColor: '#e8e8e8',
                    },
                    axisPointer: {
                        lineStyle: {
                            width: 0
                        }
                    }
                },
                projectData: {
                    count: 0,
                    monthCount: 0,
                    projectSchedule: 0,
                    chartData: {
                        columns: ['日期', '数量'],
                        rows: []
                    },
                    chartSettings: {
                        itemStyle: {
                            color: '#1890ff'
                        },
                    },
                },
                taskData: {
                    count: 0,
                    taskDoneCount: 0,
                    taskOverdueCount: 0,
                    taskOverduePercent: 0,
                    weekRate: 0,
                    dayRate: 0,
                    donePercent: 0,
                    doingPercent: 0,
                    chartData: {
                        columns: ['日期', '任务'],
                        rows: []
                    },
                    chartSettings: {
                        area: true,
                        itemStyle: {
                            color: '#b68eec'
                        },
                        areaStyle: {
                            color: '#b68eec'
                        }
                    },
                },
                projectTotalData: {
                    chartData: {
                        columns: ['日期', '数量'],
                        rows: []
                    },
                    chartSettings: {
                        itemStyle: {
                            color: '#1890ff'
                        },
                    },
                    chartExtend: {
                        grid: {
                            left: '30',
                            right: '0',
                            top: '15',
                            bottom: '0'
                        },
                        series: {
                            barWidth: 45,
                        },
                    }
                },
                taskTotalData: {
                    chartData: {
                        columns: ['日期', '数量'],
                        rows: []
                    },
                    chartSettings: {
                        itemStyle: {
                            color: '#b68eec'
                        },
                    },
                    chartExtend: {
                        grid: {
                            left: '30',
                            right: '0',
                            top: '15',
                            bottom: '0'
                        },
                        series: {
                            barWidth: 45,
                        },
                    }
                },
                // 优先级分布数据
                priData: {
                    normal: 0,
                    urgent: 0,
                    veryUrgent: 0,
                },
                // 执行者分布数据
                executorData: [],
                // 完成趋势数据
                trendData: [],
                projectList: [],
                projectTotal: 0,
                projectLoading: false,
                // ECharts 实例
                priChartInstance: null,
                executorChartInstance: null,
                trendChartInstance: null,
            }
        },
        computed: {
            ...mapState({
                userInfo: state => state.userInfo,
            }),
        },
        created() {
            this.init();
        },
        mounted() {
            window.addEventListener('resize', this.handleResize);
        },
        beforeDestroy() {
            window.removeEventListener('resize', this.handleResize);
            if (this.priChartInstance) this.priChartInstance.dispose();
            if (this.executorChartInstance) this.executorChartInstance.dispose();
            if (this.trendChartInstance) this.trendChartInstance.dispose();
        },
        methods: {
            init(reset = true) {
                this.loading = true;
                analysis({type: 1}).then(res => {
                    const data = res.data || {};
                    this.projectData.count = data.projectCount || 0;
                    this.projectData.monthCount = data.monthProjectCount || 0;
                    this.projectData.projectSchedule = Math.round((data.projectSchedule || 0) * 100) / 100;
                    this.projectData.chartData.rows = data.projectList || [];
                    this.projectTotalData.chartData.rows = data.projectList || [];

                    this.taskData.count = data.taskCount || 0;
                    this.taskData.taskDoneCount = data.taskDoneCount || 0;
                    this.taskData.taskOverdueCount = data.taskOverdueCount || 0;
                    this.taskData.taskOverduePercent = Math.round((data.taskOverduePercent || 0) * 100) / 100;
                    this.taskData.chartData.rows = data.taskList || [];
                    this.taskTotalData.chartData.rows = data.taskList || [];
                    this.taskData.weekRate = data.weekRate || 0;
                    this.taskData.dayRate = data.dayRate || 0;

                    // 计算完成率和进行中比例
                    const totalTasks = data.taskCount || 1;
                    this.taskData.donePercent = totalTasks > 0 ? Math.round((data.taskDoneCount || 0) / totalTasks * 100) : 0;
                    this.taskData.doingPercent = totalTasks > 0 ? (100 - this.taskData.donePercent) : 0;

                    // 优先级分布
                    this.priData.normal = data.normalCount || data.taskCount - (data.urgentCount || 0) - (data.veryUrgentCount || 0) || 0;
                    this.priData.urgent = data.urgentCount || 0;
                    this.priData.veryUrgent = data.veryUrgentCount || 0;

                    // 执行者分布
                    this.executorData = data.executorList || [];

                    // 完成趋势
                    this.trendData = data.trendList || data.taskList || [];

                    // 排行榜
                    this.rankList = (data.projectRankList || data.projectList || []).map((item, i) => ({
                        name: item['日期'] || item.name || `项目 ${i + 1}`,
                        total: parseInt(item['数量'] || item.total || 0)
                    }));
                    this.taskRankList = (data.taskRankList || data.taskList || []).map((item, i) => ({
                        name: item['日期'] || item.name || `成员 ${i + 1}`,
                        total: parseInt(item['任务'] || item.total || 0)
                    }));

                    this.loading = false;
                    this.$nextTick(() => {
                        this.renderPriChart();
                        this.renderExecutorChart();
                        this.renderTrendChart();
                    });
                }).catch(() => {
                    this.loading = false;
                    // 如果API没有返回完整数据，使用fallback渲染
                    this.$nextTick(() => {
                        this.renderPriChart();
                        this.renderExecutorChart();
                        this.renderTrendChart();
                    });
                });
                if (reset) {
                    this.pagination.page = 1;
                    this.pagination.pageSize = 10;
                }
                this.getProjectList(true);
            },
            getProjectList(loading) {
                if (loading) {
                    this.projectLoading = true;
                }
                getProjectList(this.requestData).then(res => {
                    const data = res.data || {};
                    this.projectList = data.list || [];
                    this.projectTotal = data.total || 0;
                    this.projectLoading = false;
                }).catch(() => {
                    this.projectList = [];
                    this.projectTotal = 0;
                    this.projectLoading = false;
                });
            },
            pageChange(page) {
                this.pagination.page = page;
                this.getProjectList(true);
            },
            changeTimeRange(range) {
                this.activeTimeRange = range;
                // 可以根据时间范围重新请求数据
                this.init(false);
            },
            onDateRangeChange(dates) {
                if (dates && dates.length === 2) {
                    this.activeTimeRange = '';
                    this.init(false);
                }
            },
            renderPriChart() {
                if (!this.$refs.priChart) return;
                if (this.priChartInstance) this.priChartInstance.dispose();
                this.priChartInstance = echarts.init(this.$refs.priChart);

                const total = this.priData.normal + this.priData.urgent + this.priData.veryUrgent;
                const data = [
                    {value: this.priData.normal || (total === 0 ? 1 : 0), name: '普通'},
                    {value: this.priData.urgent, name: '紧急'},
                    {value: this.priData.veryUrgent, name: '非常紧急'},
                ];

                this.priChartInstance.setOption({
                    tooltip: {
                        trigger: 'item',
                        formatter: '{b}: {c} ({d}%)',
                        backgroundColor: '#fff',
                        textStyle: {color: '#333'},
                        borderWidth: 1,
                        borderColor: '#e8e8e8',
                    },
                    legend: {
                        orient: 'vertical',
                        right: 10,
                        top: 'center',
                        data: ['普通', '紧急', '非常紧急'],
                    },
                    color: ['#1890ff', '#faad14', '#f5222d'],
                    series: [{
                        type: 'pie',
                        radius: ['40%', '70%'],
                        center: ['40%', '50%'],
                        avoidLabelOverlap: false,
                        label: {
                            show: true,
                            formatter: '{b}\n{d}%',
                        },
                        emphasis: {
                            label: {show: true, fontSize: '14', fontWeight: 'bold'},
                        },
                        data: data,
                    }]
                });
            },
            renderExecutorChart() {
                if (!this.$refs.executorChart) return;
                if (this.executorChartInstance) this.executorChartInstance.dispose();
                this.executorChartInstance = echarts.init(this.$refs.executorChart);

                let data = this.executorData.map(item => ({
                    value: item.count || item.total || 0,
                    name: item.name || item.executor || '未指派',
                }));

                // 如果没有执行者数据，显示占位
                if (data.length === 0) {
                    data = [{value: 1, name: '暂无数据'}];
                }

                const colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#b68eec', '#13c2c2', '#eb2f96', '#722ed1'];

                this.executorChartInstance.setOption({
                    tooltip: {
                        trigger: 'item',
                        formatter: '{b}: {c} ({d}%)',
                        backgroundColor: '#fff',
                        textStyle: {color: '#333'},
                        borderWidth: 1,
                        borderColor: '#e8e8e8',
                    },
                    legend: {
                        orient: 'vertical',
                        right: 10,
                        top: 'center',
                        data: data.map(d => d.name),
                    },
                    color: colors,
                    series: [{
                        type: 'pie',
                        radius: ['40%', '70%'],
                        center: ['40%', '50%'],
                        avoidLabelOverlap: false,
                        label: {
                            show: true,
                            formatter: '{b}\n{d}%',
                        },
                        emphasis: {
                            label: {show: true, fontSize: '14', fontWeight: 'bold'},
                        },
                        data: data,
                    }]
                });
            },
            renderTrendChart() {
                if (!this.$refs.trendChart) return;
                if (this.trendChartInstance) this.trendChartInstance.dispose();
                this.trendChartInstance = echarts.init(this.$refs.trendChart);

                let xData = [];
                let yData = [];

                if (this.trendData.length > 0) {
                    this.trendData.forEach(item => {
                        xData.push(item['日期'] || item.date || '');
                        yData.push(parseInt(item['任务'] || item['数量'] || item.count || 0));
                    });
                } else {
                    // Fallback: 生成近7天数据
                    for (let i = 6; i >= 0; i--) {
                        xData.push(moment().subtract(i, 'days').format('MM-DD'));
                        yData.push(0);
                    }
                }

                this.trendChartInstance.setOption({
                    tooltip: {
                        trigger: 'axis',
                        backgroundColor: '#fff',
                        textStyle: {color: '#333'},
                        borderWidth: 1,
                        borderColor: '#e8e8e8',
                    },
                    grid: {
                        left: '40',
                        right: '20',
                        top: '20',
                        bottom: '30',
                    },
                    xAxis: {
                        type: 'category',
                        data: xData,
                        axisLine: {lineStyle: {color: '#d9d9d9'}},
                        axisLabel: {color: '#666'},
                    },
                    yAxis: {
                        type: 'value',
                        axisLine: {show: false},
                        axisTick: {show: false},
                        splitLine: {lineStyle: {type: 'dashed', color: '#f0f0f0'}},
                        axisLabel: {color: '#666'},
                    },
                    series: [{
                        data: yData,
                        type: 'line',
                        smooth: true,
                        symbol: 'circle',
                        symbolSize: 6,
                        lineStyle: {color: '#1890ff', width: 2},
                        itemStyle: {color: '#1890ff'},
                        areaStyle: {
                            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                {offset: 0, color: 'rgba(24,144,255,0.3)'},
                                {offset: 1, color: 'rgba(24,144,255,0.05)'}
                            ]),
                        },
                    }]
                });
            },
            handleResize() {
                if (this.priChartInstance) this.priChartInstance.resize();
                if (this.executorChartInstance) this.executorChartInstance.resize();
                if (this.trendChartInstance) this.trendChartInstance.resize();
            },
            handleExport({key}) {
                if (key === 'png') {
                    // 导出当前页面为PNG
                    const el = document.querySelector('.analysis-index .page-wrapper');
                    if (!el) return;
                    this.$message.loading('正在生成报表图片...', 0);
                    // 使用 canvas 截图方式（简化版：直接导出图表）
                    const charts = [this.priChartInstance, this.executorChartInstance, this.trendChartInstance].filter(Boolean);
                    if (charts.length > 0) {
                        const chart = charts[0];
                        const url = chart.getDataURL({type: 'png', pixelRatio: 2, backgroundColor: '#fff'});
                        const link = document.createElement('a');
                        link.download = `数据分析报表_${new Date().toLocaleDateString()}.png`;
                        link.href = url;
                        link.click();
                    }
                    this.$message.destroy();
                    this.$message.success('报表已导出');
                } else if (key === 'print') {
                    window.print();
                }
            },
        }
    }
</script>
<style lang="less">
    .analysis-index {
        .page-wrapper {
            margin: 24px;

            .analysis-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: 24px;

                .analysis-title {
                    margin: 0;
                    font-size: 22px;
                    font-weight: 600;
                    color: rgba(0, 0, 0, 0.85);

                    .anticon {
                        margin-right: 8px;
                        color: #1890ff;
                    }
                }

                .analysis-actions {
                    display: flex;
                    align-items: center;

                    .ant-btn {
                        border-radius: 6px;
                    }
                }
            }

            .ant-card {
                border-radius: 8px;
                overflow: hidden;
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
                transition: box-shadow 0.3s;

                &:hover {
                    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
                }
            }

            .extra-wrapper {
                line-height: 55px;
                padding-right: 24px;

                .extra-item {
                    display: inline-block;
                    margin-right: 24px;

                    a {
                        margin-left: 24px;
                        color: rgba(0, 0, 0, 0.65);
                        padding: 4px 12px;
                        border-radius: 4px;
                        transition: all 0.3s;

                        &:hover {
                            background: #f0f7ff;
                        }

                        &.active-time {
                            color: #1890ff;
                            font-weight: 500;
                            background: #e6f7ff;
                        }
                    }
                }
            }

            .chart-wrapper {
                position: absolute;
                bottom: -10px;
                width: 100%;
            }

            .chart-wrappers-single {
                div {
                    width: auto !important;
                }
            }

            .project-list-wrapper {
                .project-list-item {
                    cursor: pointer;
                    transition: all 0.3s;
                    border-radius: 6px;
                    padding: 8px 12px;

                    &:hover {
                        background-color: #f0f7ff;
                        transform: translateX(4px);
                    }
                }
            }
        }
    }

    @media print {
        .analysis-header .analysis-actions {
            display: none !important;
        }
    }
</style>
