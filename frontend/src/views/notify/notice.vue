<template>
    <div class="house-index notice-enhanced">
        <wrapper-content>
            <!-- 通知分类 Tab -->
            <a-tabs v-model="activeTab" @change="onTabChange" class="notice-tabs">
                <a-tab-pane key="all">
                    <span slot="tab">
                        <a-icon type="bell" /> 全部通知
                        <a-badge v-if="unreadTotal > 0" :count="unreadTotal" :numberStyle="{fontSize:'10px'}" />
                    </span>
                </a-tab-pane>
                <a-tab-pane key="task">
                    <span slot="tab">
                        <a-icon type="check-square" /> 任务通知
                    </span>
                </a-tab-pane>
                <a-tab-pane key="system">
                    <span slot="tab">
                        <a-icon type="notification" /> 系统公告
                    </span>
                </a-tab-pane>
            </a-tabs>

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
                    <a @click.stop="rowClick(record,'del')">删除</a>
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
                activeTab: 'all',
                unreadTotal: 0,
                detailVisible: false,
                currentNotice: {},
            }
        },
        created() {
            this.init();
        },
        methods: {
            init() {
                let app = this;
                if (app.activeTab === 'task') {
                    app.requestData.type = 'notice';
                    app.requestData.action = 'task';
                } else if (app.activeTab === 'system') {
                    app.requestData.type = 'system';
                    delete app.requestData.action;
                } else {
                    // 全部：不传 type，由后端返回所有通知和系统公告
                    app.requestData.type = 'all';
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
            onTabChange(key) {
                this.activeTab = key;
                this.pagination.page = 1;
                this.init();
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
                        this.$router.push('/project/space/task/' + projectCode + '/detail/' + taskCode).catch(() => {});
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
        .notice-tabs {
            margin-bottom: 16px;

            .ant-tabs-tab {
                font-size: 14px;
                font-weight: 500;
            }

            .ant-badge {
                margin-left: 4px;
            }
        }

        .notice-title-cell {
            display: flex;
            align-items: center;
            gap: 8px;
            cursor: pointer;

            .unread-title {
                font-weight: 600;
                color: rgba(0, 0, 0, 0.85);
            }
        }

        .notice-empty {
            text-align: center;
            padding: 60px 0;
            color: #999;

            p {
                margin-top: 12px;
                font-size: 14px;
            }
        }
    }

    // 未读行高亮
    /deep/ .notice-unread {
        background-color: #f0f7ff;

        &:hover > td {
            background-color: #e6f0fa !important;
        }
    }

    /deep/ .notice-read {
        td {
            color: rgba(0, 0, 0, 0.45);
        }
    }

    .notice-detail {
        .notice-detail-time {
            color: #999;
            font-size: 13px;
            margin-bottom: 16px;
        }

        .notice-detail-content {
            font-size: 14px;
            line-height: 1.8;
            color: #333;
        }

        .notice-detail-actions {
            margin-top: 20px;
            text-align: right;
        }
    }
</style>
