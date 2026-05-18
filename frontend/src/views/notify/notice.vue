<template>
    <div class="house-index notice-enhanced">
        <wrapper-content>
            <div class="page-search">
                <a-form
                        layout="inline"
                        :autoFormCreate="(form)=>{this.searchForm = form}"
                        @submit.prevent="handleSearchSubmit"
                >
                    <a-form-item
                            label='关键字'
                            fieldDecoratorId="title"
                    >
                        <a-input placeholder='请输入关键字'/>
                    </a-form-item>
                    <a-form-item
                            label='日期'
                            fieldDecoratorId="date"
                    >
                        <a-range-picker :placeholder="['开始日期','结束日期']"></a-range-picker>
                    </a-form-item>
                    <a-form-item
                    >
                        <a-button icon="search" type="primary" htmlType='submit'
                                :loading="searchLoading"
                        >搜索
                        </a-button>
                    </a-form-item>
                </a-form>
            </div>
            <div class="action">
                <a-button icon="check" class="m-r-sm" :disabled="!selectedRowKeys.length" @click="listAction({key:'setReadied'})">
                    <span>批量标记已读</span>
                </a-button>
                <a-button icon="delete" type="danger" :disabled="!selectedRowKeys.length" @click="listAction({key:'delete'})">
                    <span>批量删除</span>
                </a-button>
                <a-button icon="check-circle" class="m-l-sm" @click="markAllRead" :disabled="unreadTotal === 0">
                    <span>全部已读</span>
                </a-button>
                <span style="padding-left: 12px;" v-show="selectedRowKeys.length">已选择 <span class="text-warning">{{selectedRowKeys.length}}</span> 项</span>
            </div>
            <a-table :columns="columns" :dataSource="dataSource" :loading="loading" :showTotal="pagination.showTotal" :pagination="pagination"
                     @change="pageChange"
                     :rowSelection="{selectedRowKeys: selectedRowKeys, onChange: onSelectChange}"
                     :rowClassName="rowClassName"
                     :customRow="customRow"
                     rowKey="id">
                <template slot="title_col" slot-scope="text, record">
                    <div class="notice-title-cell" @click="showDetail(record)">
                        <a-badge :status="record.is_read ? 'default' : 'processing'" />
                        <span :class="{'unread-title': !record.is_read}">{{ text }}</span>
                    </div>
                </template>
                <template slot="action" slot-scope="text,record,index">
                    <span class="table-action-btn" @click.stop="rowClick(record,'del')">
                        <a-icon type="delete" />
                    </span>
                </template>
            </a-table>

            <!-- 空状态 -->
            <div class="notice-empty" v-if="!loading && dataSource.length === 0">
                <a-icon type="inbox" style="font-size:48px;color:#d9d9d9;" />
                <p>暂无通知消息</p>
            </div>

            <!-- 详情弹窗 -->
            <a-modal
                v-model="detailVisible"
                :title="currentNotice.title"
                :footer="null"
                width="500px"
            >
                <div class="notice-detail">
                    <div class="notice-detail-time">
                        <a-icon type="clock-circle" /> {{ currentNotice.create_time }}
                    </div>
                    <div class="notice-detail-content" v-html="currentNotice.content"></div>
                    <div class="notice-detail-actions" v-if="currentNotice.action">
                        <a-button type="primary" @click="goToAction(currentNotice)">
                            查看详情
                        </a-button>
                    </div>
                </div>
            </a-modal>
        </wrapper-content>
    </div>
</template>
<script>
    import {list, del, setReadied} from '@/api/notify';
    import {checkResponse} from '@/assets/js/utils';
    import {batchDel} from "../../api/notify";
    import pagination from "@/mixins/pagination";

    const columns = [{
        title: '消息标题',
        dataIndex: 'title',
        width: '30%',
        scopedSlots: {
            customRender: 'title_col'
        }
    },{
        title: '时间',
        dataIndex: 'create_time',
        width: '30%',
    },{
        title: '操作',
        width: '10%',
        scopedSlots: {
            customRender: 'action'
        }
    }];

    export default {
        mixins: [pagination],
        data() {
            return {
                columns,
                dataSource: [],
                selectedRowKeys: [],
                loading: true,
                searchForm: {},
                searchLoading: false,
                activeTab: 'task',
                unreadTotal: 0,
                detailVisible: false,
                currentNotice: {},
            }
        },
        created() {
            this.loadByRoute();
        },
        watch: {
            // 同一组件切换路由时（/notify/notice ↔ /notify/system），Vue 复用实例，
            // created 不会重复触发，需要 watch $route 来重新加载数据
            '$route.path'() {
                this.loadByRoute();
            }
        },
        methods: {
            loadByRoute() {
                this.activeTab = this.$route.path === '/notify/system' ? 'system' : 'task';
                this.pagination.page = 1;
                this.init();
            },
            init() {
                let app = this;
                if (app.activeTab === 'task') {
                    app.requestData.type = 'notice';
                    app.requestData.action = 'task';
                } else {
                    // 系统消息
                    app.requestData.type = 'system';
                    delete app.requestData.action;
                }
                app.loading = true;
                app.selectedRowKeys = [];
                list(app.requestData).then(res => {
                    app.dataSource = res.data.list;
                    app.pagination.total = res.data.total;
                    app.loading = false;
                    // 计算未读数
                    app.unreadTotal = (res.data.list || []).filter(item => !item.is_read).length;
                });
            },
            onSelectChange(selectedRowKeys) {
                this.selectedRowKeys = selectedRowKeys
            },
            rowClassName(record) {
                return record.is_read ? 'notice-read' : 'notice-unread';
            },
            customRow(record) {
                return {
                    on: {
                        click: () => {
                            this.showDetail(record);
                        }
                    }
                };
            },
            showDetail(record) {
                this.currentNotice = record;
                this.detailVisible = true;
                // 标记为已读
                if (!record.is_read) {
                    setReadied(JSON.stringify([record.id]));
                    record.is_read = 1;
                    this.unreadTotal = this.dataSource.filter(v => !v.is_read).length;
                }
            },
            goToAction(record) {
                this.detailVisible = false;
                // 处理任务相关通知（task:mention, task:comment, task:done, task:redo, task:assign）
                if (record.action && record.action.startsWith('task:')) {
                    let taskCode = '';
                    let projectCode = '';
                    try {
                        const sendData = typeof record.send_data === 'string' ? JSON.parse(record.send_data) : record.send_data;
                        taskCode = sendData.taskCode || '';
                        projectCode = sendData.projectCode || '';
                    } catch (e) {}
                    if (taskCode && projectCode) {
                        const routeUrl = '/project/space/task/' + projectCode + '/detail/' + taskCode;
                        // task:mention 和 task:comment 通知直接滚动到评论区
                        if (record.action === 'task:mention' || record.action === 'task:comment') {
                            this.$router.push({path: routeUrl, query: {scrollTo: 'comment'}}).catch(() => {});
                        } else {
                            this.$router.push(routeUrl).catch(() => {});
                        }
                        return;
                    }
                    if (taskCode) {
                        // 没有 projectCode 时，尝试通过任务列表跳转
                        this.$router.push('/project/list/my').catch(() => {});
                        return;
                    }
                }
                if (record.action && record.action !== '' && record.action.startsWith('/')) {
                    this.$router.push(record.action).catch(() => {});
                }
            },
            markAllRead() {
                const unreadIds = this.dataSource.filter(v => !v.is_read).map(v => v.id);
                if (unreadIds.length === 0) return;
                setReadied(JSON.stringify(unreadIds));
                this.dataSource.forEach((v) => {
                    v.is_read = 1;
                });
                this.unreadTotal = 0;
                this.$notice('已全部标记为已读', 'message', 'success');
            },
            listAction(type) {
                let app = this;
                switch (type.key) {
                    case 'setReadied':
                        setReadied(JSON.stringify(app.selectedRowKeys));
                        app.dataSource.forEach(function (v, k) {
                            const index = app.selectedRowKeys.find(item => item == v.id);
                            if (index) {
                                app.dataSource[k].is_read = 1;
                            }
                        });
                        app.selectedRowKeys = [];
                        app.unreadTotal = app.dataSource.filter(v => !v.is_read).length;
                        app.$notice('操作成功', 'message', 'success');
                        break;
                    case 'delete':
                        this.$confirm({
                            title: '确认删除这些消息?',
                            content: '删除后不可恢复',
                            okText: '删除',
                            okType: 'danger',
                            cancelText: '取消',
                            onOk() {
                                batchDel(JSON.stringify(app.selectedRowKeys)).then(res => {
                                    if (checkResponse(res)) {
                                        app.init();
                                        app.notice('操作成功', 'message', 'success');
                                        app.selectedRowKeys = [];
                                    }
                                });
                                return Promise.resolve();
                            }
                        });
                        break;
                }
            },
            rowClick(record, action) {
                let app = this;
                if (action == 'del') {
                    this.$confirm({
                        title: '确认删除此消息?',
                        content: '删除后不可恢复',
                        okText: '删除',
                        okType: 'danger',
                        cancelText: '取消',
                        onOk() {
                            del(record.id).then(res => {
                                app.init();
                            });
                            return Promise.resolve();
                        }
                    });
                }

            },
            handleSearchSubmit() {
                let app = this;
                app.searchForm.validateFields(
                    (err, values) => {
                        if (!err) {
                            app.search();
                        }
                    })
            },
            search(){
                let obj = this.searchForm.getFieldsValue();
                if (obj.date && obj.date.length > 0) {
                    const beginDate = obj.date[0].format('YYYY-MM-DD');
                    const endDate = obj.date[1].format('YYYY-MM-DD');
                    obj.date = beginDate + '~' + endDate;
                }
                Object.assign(this.requestData, obj);
                this.init();
            }
        }
    }
</script>
<style lang="less" scoped>
    .notice-enhanced {
        .action {
            padding: 8px 0 20px;

            .ant-btn {
                height: 34px;
                border-radius: 8px;
                padding: 0 14px;
                font-size: 13px;
                font-weight: 500;

                &:not(.ant-btn-primary) {
                    border-color: #e4e7eb;
                    color: #5f6773;
                    &:hover {
                        border-color: #4f8cff;
                        color: #4f8cff;
                        background: #eef4ff;
                    }
                }

                &.ant-btn-danger {
                    color: #ef4444;
                    border-color: transparent;
                    &:hover {
                        background: rgba(239, 68, 68, 0.06);
                        color: #ef4444;
                    }
                }
            }
        }

        .notice-title-cell {
            display: flex;
            align-items: center;
            gap: 10px;
            cursor: pointer;
            font-weight: 450;

            .ant-badge-status {
                .ant-badge-status-dot {
                    width: 6px;
                    height: 6px;
                }
            }

            .unread-title {
                font-weight: 600;
                color: #1a1d23;
            }
        }

        // 表格操作图标按钮
        /deep/ .ant-table {
            border: none;
            border-radius: 0;
            overflow: visible;

            .ant-table-thead > tr > th {
                background: #f7f8fa;
                color: #9ca3af;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.3px;
                padding: 10px 16px;
                border-bottom: 1px solid #eef0f2;
                border-top: none;
            }

            .ant-table-tbody > tr > td {
                padding: 14px 16px;
                font-size: 13px;
                color: #5f6773;
                border-bottom: 1px solid #eef0f2;
                transition: background 0.12s ease;
            }

            .ant-table-tbody > tr:hover > td {
                background: #f7f8fa !important;
            }
        }

        // 未读行
        /deep/ .notice-unread {
            background: rgba(79, 140, 255, 0.03) !important;

            > td:first-child {
                position: relative;
                &::after {
                    content: '';
                    position: absolute;
                    left: 0;
                    top: 50%;
                    transform: translateY(-50%);
                    width: 4px;
                    height: 4px;
                    border-radius: 50%;
                    background: #4f8cff;
                }
            }

            &:hover > td {
                background: #f7f8fa !important;
            }
        }

        /deep/ .notice-read {
            td {
                color: #9ca3af;
            }
        }

        .notice-empty {
            text-align: center;
            padding: 80px 24px;

            .anticon {
                color: #d4d7dc;
            }

            p {
                margin-top: 16px;
                font-size: 14px;
                color: #9ca3af;
                font-weight: 450;
            }
        }

        .table-action-btn {
            width: 32px;
            height: 32px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border-radius: 6px;
            color: #9ca3af;
            transition: all 0.15s ease;
            cursor: pointer;
            font-size: 14px;

            &:hover {
                background: #f7f8fa;
                color: #ef4444;
            }
        }
    }

    // ========== 详情弹窗 ==========
    .notice-detail {
        .notice-detail-time {
            font-size: 12px;
            color: #9ca3af;
            margin-bottom: 20px;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .notice-detail-content {
            font-size: 14px;
            line-height: 1.8;
            color: #1a1d23;
            background: #f7f8fa;
            border-radius: 8px;
            padding: 16px;
        }

        .notice-detail-actions {
            margin-top: 24px;
            text-align: right;
        }
    }
</style>
