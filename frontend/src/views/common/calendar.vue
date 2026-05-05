<template>
    <div class="common-calendar-page">
        <div class="calendar-header">
            <a-page-header title="日程日历" sub-title="查看和管理所有项目日程">
                <template slot="extra">
                    <a-radio-group v-model="viewMode" button-style="solid" size="small" @change="handleViewChange">
                        <a-radio-button value="calendar">日历视图</a-radio-button>
                        <a-radio-button value="list">列表视图</a-radio-button>
                    </a-radio-group>
                </template>
            </a-page-header>
        </div>
        <div class="calendar-body">
            <!-- 日历视图 -->
            <template v-if="viewMode === 'calendar'">
                <calendar ref="calendarComponent" v-if="calendarReady"></calendar>
            </template>
            <!-- 列表视图 -->
            <template v-else>
                <a-card :bordered="false" class="list-card">
                    <div class="list-toolbar">
                        <a-radio-group v-model="listFilter" size="small" @change="handleFilterChange">
                            <a-radio-button value="all">全部日程</a-radio-button>
                            <a-radio-button value="mine">我的日程</a-radio-button>
                        </a-radio-group>
                        <a-range-picker v-model="dateRange" size="small" @change="handleDateChange" style="margin-left: 16px;"/>
                    </div>
                    <a-list :loading="listLoading" :data-source="eventsList" size="small" class="events-list">
                        <a-list-item slot="renderItem" slot-scope="item">
                            <a-list-item-meta>
                                <div slot="title" style="display: flex; align-items: center; justify-content: space-between;">
                                    <span>
                                        <a-icon type="calendar" style="margin-right: 6px; color: #1890ff;"/>
                                        <router-link v-if="item.project_code"
                                            :to="`/project/space/events/${item.project_code}`">
                                            {{ item.title }}
                                        </router-link>
                                        <span v-else>{{ item.title }}</span>
                                        <a-tag v-if="item.all_day === 1" color="blue" style="margin-left: 6px;">全天</a-tag>
                                    </span>
                                    <span class="muted" style="font-size: 12px;">
                                        {{ formatEventTime(item) }}
                                    </span>
                                </div>
                                <div slot="description">
                                    <span v-if="item.project_name" class="muted">
                                        <a-icon type="folder"/> {{ item.project_name }}
                                    </span>
                                    <span v-if="item.memberList && item.memberList.length" class="muted" style="margin-left: 12px;">
                                        <a-icon type="team"/> {{ item.memberList.length }}人参与
                                    </span>
                                </div>
                            </a-list-item-meta>
                        </a-list-item>
                        <div v-if="eventsList.length === 0 && !listLoading" slot="footer" style="text-align: center;">
                            <a-empty description="暂无日程" :image="simpleImage"/>
                        </div>
                    </a-list>
                </a-card>
            </template>
        </div>
    </div>
</template>

<script>
import { Empty } from 'ant-design-vue'
import moment from 'moment'
import calendar from '@/components/project/calendar'
import { myList, getEventsListByCalendar } from '@/api/projectEvents'

export default {
    name: "CommonCalendar",
    components: {
        calendar,
    },
    data() {
        return {
            viewMode: 'calendar',
            calendarReady: true,
            listFilter: 'all',
            listLoading: false,
            eventsList: [],
            dateRange: null,
            simpleImage: Empty.PRESENTED_IMAGE_SIMPLE,
        }
    },
    created() {
        this.init();
    },
    methods: {
        init() {
            // 默认加载日历视图
        },
        handleViewChange() {
            if (this.viewMode === 'list') {
                this.loadEventsList();
            } else {
                // 切换回日历视图时，重新挂载组件确保数据刷新
                this.calendarReady = false;
                this.$nextTick(() => {
                    this.calendarReady = true;
                });
            }
        },
        handleFilterChange() {
            this.loadEventsList();
        },
        handleDateChange() {
            this.loadEventsList();
        },
        loadEventsList() {
            this.listLoading = true;
            const params = {
                page: 1,
                pageSize: 50,
            };
            if (this.dateRange && this.dateRange.length === 2) {
                params.beginTime = this.dateRange[0].format('YYYY-MM-DD');
                params.endTime = this.dateRange[1].format('YYYY-MM-DD');
            }

            let apiCall;
            if (this.listFilter === 'mine') {
                apiCall = myList(params);
            } else {
                // 全部日程 - 使用日历接口获取当月数据
                const now = moment();
                params.date = now.format('YYYY-MM-DD HH:mm:ss');
                apiCall = getEventsListByCalendar(params);
            }

            apiCall.then(res => {
                if (res.data) {
                    if (this.listFilter === 'mine' && res.data.list) {
                        this.eventsList = res.data.list;
                    } else if (res.data.list) {
                        // 日历接口返回的是按日期分组的数据，需要展平
                        const grouped = res.data.list;
                        let flatList = [];
                        Object.keys(grouped).forEach(dateKey => {
                            flatList = flatList.concat(grouped[dateKey]);
                        });
                        this.eventsList = flatList;
                    } else {
                        this.eventsList = [];
                    }
                } else {
                    this.eventsList = [];
                }
            }).catch(() => {
                this.eventsList = [];
            }).finally(() => {
                this.listLoading = false;
            });
        },
        formatEventTime(item) {
            if (!item.begin_time) return '';
            const begin = moment(item.begin_time);
            if (item.all_day === 1) {
                return begin.format('YYYY年MM月DD日');
            }
            const end = item.end_time ? moment(item.end_time) : null;
            if (end) {
                if (begin.isSame(end, 'day')) {
                    return begin.format('MM月DD日 HH:mm') + ' - ' + end.format('HH:mm');
                }
                return begin.format('MM月DD日 HH:mm') + ' - ' + end.format('MM月DD日 HH:mm');
            }
            return begin.format('MM月DD日 HH:mm');
        }
    }
}
</script>

<style scoped lang="less">
.common-calendar-page {
    .calendar-header {
        background: #fff;
        padding: 0 24px;
        border-bottom: 1px solid #e8e8e8;

        /deep/ .ant-page-header {
            padding: 16px 0;
        }

        /deep/ .ant-page-heading-title {
            font-size: 18px;
        }

        /deep/ .ant-page-heading-sub-title {
            font-size: 13px;
        }
    }

    .calendar-body {
        padding: 16px;
    }

    .list-card {
        border-radius: 4px;

        .list-toolbar {
            margin-bottom: 16px;
            display: flex;
            align-items: center;
        }

        .events-list {
            /deep/ .ant-list-item {
                padding: 12px 0;
            }
        }
    }

    .muted {
        color: #999;
    }
}
</style>
