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
                <span style="padding-left: 12px;" v-show="selectedRowKeys.length">已选择 <span class="text-warning">{{selectedRowKeys.length}}</span> 项</span>
            </div>
            <a-table :columns="columns" :dataSource="dataSource" :loading="loading" :showTotal="pagination.showTotal" :pagination="pagination"
                     @change="pageChange"
                     :rowSelection="{selectedRowKeys: selectedRowKeys, onChange: onSelectChange}"
                     :rowClassName="rowClassName"
                     :customRow="customRow"
                     rowKey="id">
                <template slot="title_col" slot-scope="text, record">
                    <div class="notice-title-cell">
                        <a-badge :status="record.is_read ? 'default' : 'processing'" />
                        <span :class="{'unread-title': !record.is_read}">{{ text }}</span>
                    </div>
                </template>
                <template slot="action" slot-scope="text,record,index">
                    <a @click.stop="rowClick(record,'del')">删除</a>
                </template>
            </a-table>

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
        title: '公告标题',
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
                app.requestData.type = 'system';
                app.loading = true;
                app.selectedRowKeys = [];
                list(app.requestData).then(res => {
                    app.dataSource = res.data.list;
                    app.pagination.total = res.data.total;
                    app.loading = false;
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
                }
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
                        app.$notice('操作成功', 'message', 'success');
                        break;
                    case 'delete':
                        this.$confirm({
                            title: '确认删除这些公告?',
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
                        title: '确认删除此公告?',
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
    }

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
    }
</style>
