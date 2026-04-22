<template>
    <div class="project-list-index">
        <wrapper-content :showHeader="false">
            <div class="page-search">
                <a-form
                        layout="inline"
                        @submit.prevent="handleSearchSubmit"
                >
                    <a-form-item label='名称'>
                        <a-input v-model="searchParams.name" placeholder='请输入项目名称'/>
                    </a-form-item>
                    <a-form-item label='创建日期'>
                        <a-range-picker v-model="searchParams.dateRange" :placeholder="['开始日期','结束日期']"></a-range-picker>
                    </a-form-item>
                    <a-form-item>
                        <a-button icon="search" type="primary" htmlType='submit'>搜索</a-button>
                        <a-button style="margin-left: 8px" @click="resetSearch">重置</a-button>
                    </a-form-item>
                </a-form>
            </div>
            <a-tabs v-model="selectBy" @change="selectChange" :animated="false">
                <a-tab-pane key="my" tab="全部项目"></a-tab-pane>
                <a-tab-pane key="collect" tab="我的收藏"></a-tab-pane>
                <a-tab-pane key="archive" tab="已归档"></a-tab-pane>
                <a-tab-pane key="deleted" tab="回收站"></a-tab-pane>
                <a-button slot="tabBarExtraContent" type="primary" icon="plus" @click="doAction(null,'new')">创建新项目</a-button>
            </a-tabs>
            
            <!-- 批量操作工具栏 -->
            <div class="batch-toolbar" v-if="selectedRowKeys.length > 0 && (selectBy === 'archive' || selectBy === 'deleted')">
                <div class="selected-info">
                    <a-icon type="info-circle" />
                    <span>已选择 {{ selectedRowKeys.length }} 个项目</span>
                </div>
                <div class="batch-actions">
                    <a-button type="primary" @click="batchRecovery" :loading="batchRecoveryLoading">
                        <a-icon type="undo" />{{ selectBy === 'archive' ? '批量取消归档' : '批量恢复' }}
                    </a-button>
                    <a-button type="danger" @click="batchDelete" :loading="batchDeleteLoading" v-if="selectBy === 'deleted'">
                        <a-icon type="delete" />批量彻底删除
                    </a-button>
                    <a-button @click="clearSelection">取消选择</a-button>
                </div>
            </div>
            
            <a-list
                    class="project-list"
                    :loading="loading"
                    itemLayout="horizontal"
                    :dataSource="dataSource"
            >
                <div v-if="showLoadingMore" slot="loadMore"
                     :style="{ textAlign: 'center', marginTop: '12px', height: '32px', lineHeight: '32px' }">
                    <a-spin v-if="loadingMore"/>
                    <a-button v-else @click="onLoadMore">查看更多项目</a-button>
                </div>
                <a-list-item slot="renderItem" slot-scope="item,index">
                    <!-- 选择框 (仅在归档和回收站tab显示) -->
                    <a-checkbox 
                        v-if="selectBy === 'archive' || selectBy === 'deleted'"
                        :checked="selectedRowKeys.includes(item.code)"
                        @change="e => onSelectChange(e, item.code)"
                        style="margin-right: 12px;"
                    />
                    
                    <a-list-item-meta
                            :description="item.description"
                    >
                        <div slot="title">
                            <router-link :to="'/project/space/task/' + item.code">{{item.name}}</router-link>
                            <a-tag color="green" class="m-l" v-show="!item.private">公开</a-tag>
                        </div>
                        <a-avatar slot="avatar" icon="user" :src="item.cover"/>
                    </a-list-item-meta>
                    <div class="ant-list-item-content">
                        <div class="other-info muted">
                            <div class="info-item">
                                <span>创建日期</span>
                                <span>{{moment(item.create_time).format('YYYY-MM-DD')}}</span>
                            </div>
                            <div class="info-item">
                                <span>创建人</span>
                                <span>{{item.owner_name}}</span>
                            </div>
                            <div class="info-item schedule">
                                <span>进度</span>
                                <a-progress :strokeWidth="5" :percent="item.schedule"/>
                            </div>
                        </div>
                    </div>
                    <template v-if="selectBy === 'my' || selectBy === 'collect'">
                        <span slot="actions" @click="inviteProjectMember(item)">
                             <a-tooltip title="添加成员">
                                 <a-icon type="user-add"/>
                             </a-tooltip>
                        </span>
                        <span slot="actions" @click="doAction(item,'edit',index)">
                             <a-tooltip title="项目设置">
                                  <a-icon type="setting"/>
                             </a-tooltip>
                        </span>
                        <span slot="actions" @click="saveProjectAsTemplate(item)">
                             <a-tooltip title="保存为模板">
                                  <a-icon type="save"/>
                             </a-tooltip>
                        </span>
                        <span slot="actions">
                             <a-tooltip :title="item.collected ? '取消收藏' : '加入收藏'"
                                        @click="doAction(item,'collect',index)">
                                 <a-icon type="star" v-show="!item.collected"/>
                                 <a-icon type="star" theme="filled" style="color: #ffaf38;" v-show="item.collected"/>
                             </a-tooltip>
                        </span>
                    </template>
                    <template v-if="selectBy === 'archive'">
                         <span slot="actions">
                         <a-tooltip title="恢复项目" @click="doAction(item,'recoveryArchive',index)">
                             <a-icon type="undo"/>
                         </a-tooltip>
                    </span>
                    </template>
                    <template v-if="selectBy === 'archive' || selectBy === 'my' || selectBy === 'collect'">
                             <span slot="actions" @click="doAction(item,'del',index)">
                             <a-tooltip title="移至回收站">
                                  <a-icon type="delete"/>
                             </a-tooltip>
                        </span>
                    </template>
                    <template v-if="selectBy === 'deleted'">
                        <span slot="actions">
                             <a-tooltip title="恢复项目" @click="doAction(item,'recovery',index)">
                                 <a-icon type="undo"/>
                             </a-tooltip>
                        </span>
                        <span slot="actions">
                             <a-popconfirm
                                 title="确定要彻底删除此项目吗？此操作不可恢复！"
                                 ok-text="彻底删除"
                                 cancel-text="取消"
                                 ok-type="danger"
                                 @confirm="doAction(item,'permanentDel',index)"
                             >
                                 <a-tooltip title="彻底删除">
                                     <a-icon type="delete" style="color: #ff4d4f;"/>
                                 </a-tooltip>
                             </a-popconfirm>
                        </span>
                    </template>
                </a-list-item>
            </a-list>
            
            <!-- 全选操作栏 -->
            <div class="select-all-bar" v-if="dataSource.length > 0 && (selectBy === 'archive' || selectBy === 'deleted')">
                <a-checkbox 
                    :indeterminate="selectedRowKeys.length > 0 && selectedRowKeys.length < dataSource.length"
                    :checked="selectedRowKeys.length === dataSource.length && dataSource.length > 0"
                    @change="onSelectAll"
                >
                    全选当前页 ({{ dataSource.length }} 个项目)
                </a-checkbox>
            </div>
        </wrapper-content>
        <a-modal
                destroyOnClose
                :width="360"
                v-model="actionInfo.modalStatus"
                :title="actionInfo.modalTitle"
                :bodyStyle="{paddingBottom:'1px'}"
                :footer="null"
        >
            <a-form
                    @submit.prevent="handleSubmit"
                    :form="form"
            >
                <a-form-item
                >
                    <a-input placeholder='项目名称（必填）'
                             v-decorator="[
                                            'name',
                                            {rules: [{ required: true, message: '请输入项目名称' }]}
                                            ]"/>
                </a-form-item>
                <a-form-item
                >
                    <a-select
                            placeholder='项目模板'
                            v-decorator="[
                                            'templateCode',
                                            ]"
                    >
                        <a-select-option :key="0">
                            空白项目
                        </a-select-option>
                        <a-select-option :key="template.code" v-for="template in templateList">
                            {{template.name}}
                        </a-select-option>
                    </a-select>
                </a-form-item>
                <a-form-item
                >
                    <a-textarea placeholder='项目简介'
                                :rows="2"
                                v-decorator="['description']"
                    />
                </a-form-item>
                <a-form-item
                >
                    <div class="action-btn">
                        <a-button type="primary" htmlType='submit'
                                  block
                                  size="large"
                                  :loading="actionInfo.confirmLoading"
                                  class="middle-btn">完成并创建
                        </a-button>
                    </div>
                </a-form-item>
            </a-form>
        </a-modal>
        <a-modal
                destroyOnClose
                class="project-config-modal"
                :width="800"
                v-model="projectModal.modalStatus"
                :title="projectModal.modalTitle"
                :footer="null"
        >
            <project-config :code="currentProjectCode" @update="updateProject" @complete="projectModal.modalStatus = false;init(true)"></project-config>
        </a-modal>
        <invite-project-member v-model="showInviteMember" :project-code="currentProjectCode"
                               v-if="showInviteMember"></invite-project-member>
        <!-- 保存为模板弹窗 -->
        <a-modal
                destroyOnClose
                :width="400"
                v-model="templateModal.visible"
                title="保存为项目模板"
                :confirmLoading="templateModal.loading"
                @ok="handleSaveTemplate"
        >
            <a-form :form="templateForm" @submit.prevent="handleSaveTemplate">
                <a-form-item label="模板名称">
                    <a-input placeholder="请输入模板名称"
                             v-decorator="['templateName', {rules: [{required: true, message: '请输入模板名称'}]}]"/>
                </a-form-item>
                <a-form-item label="模板说明">
                    <a-textarea placeholder="请输入模板说明（选填）" :rows="3"
                                v-decorator="['templateDescription']"/>
                </a-form-item>
            </a-form>
            <a-alert message="将基于此项目的任务列表结构创建模板，方便后续快速创建同类项目。" type="info" showIcon style="margin-top:8px;"/>
        </a-modal>
    </div>
</template>
<script>
    import inviteProjectMember from '../../../components/project/inviteProjectMember'
    import projectConfig from '@/components/project/projectConfig'
    import {list, doData, recycle, del, recovery, recoveryArchive, batchDelete, batchRecovery} from '@/api/project';
    import {checkResponse} from '@/assets/js/utils';
    import pagination from "@/mixins/pagination";
    import moment from 'moment';
    import {collect} from "../../../api/projectCollect";
    import {list as projectTemplates, doData as doTemplateData} from "../../../api/projectTemplate";
    import {list as getTaskStages} from "../../../api/taskStages";

    export default {
        components: {
            inviteProjectMember,
            projectConfig
        },
        mixins: [pagination],
        data() {
            return {
                selectBy: this.$route.params.type || 'my',
                dataSource: [],
                loading: true,
                showLoadingMore: false,
                loadingMore: false,
                showInviteMember: false,
                currentProject: {},
                currentProjectCode: 0,
                currentProjectIndex: 0,
                newData: {
                    code: ''
                },
                form: this.$form.createForm(this),
                searchParams: {
                    name: '',
                    dateRange: [],
                },
                actionInfo: {
                    modalStatus: false,
                    confirmLoading: false,
                    modalTitle: '编辑项目',
                },
                /*项目设置*/
                projectModal: {
                    modalStatus: false,
                    modalTitle: '项目设置',
                },
                templateList: [],
                templateForm: this.$form.createForm(this, {name: 'templateForm'}),
                templateModal: {
                    visible: false,
                    loading: false,
                    projectCode: '',
                    projectName: '',
                },
                // 批量操作相关
                selectedRowKeys: [],
                batchDeleteLoading: false,
                batchRecoveryLoading: false,
            }
        },
        watch: {
            $route: function () {
                this.selectBy =  this.$route.params.type || 'my'
                this.init();
            },
        },
        created() {
            this.init();
            this.projectTemplates();
        },
        methods: {
            moment,
            init(reset = true) {
                let app = this;
                if (reset) {
                    this.dataSource = [];
                    this.pagination.page = 1;
                    this.pagination.pageSize = 100;
                    this.showLoadingMore = false;
                }
                this.requestData.selectBy = this.selectBy;
                app.loading = true;
                list(app.requestData).then(res => {
                    app.dataSource = app.dataSource.concat(res.data.list);
                    app.pagination.total = res.data.total;
                    app.showLoadingMore = app.pagination.total > app.dataSource.length;
                    app.loading = false;
                    app.loadingMore = false
                }).catch(err => {
                    app.loading = false;
                    app.loadingMore = false;
                    console.error('获取项目列表失败:', err);
                });
            },

            projectTemplates() {
                projectTemplates({pageSize: 100, viewType: -1}).then(res => {
                    this.templateList = res.data.list;
                });
            },
            onLoadMore() {
                this.loadingMore = true;
                this.pagination.page++;
                this.init(false);
            },
            inviteProjectMember(item) {
                this.currentProject = item;
                this.currentProjectCode = item.code;
                this.showInviteMember = true;
            },
            doAction(record, action, index) {
                this.currentProject = record;
                this.currentProjectIndex = index;
                let app = this;
                app.newData = {id: 0};
                if (action == 'new') {
                    setTimeout(function () {
                        app.form && app.form.resetFields();
                    }, 0);
                    app.actionInfo.modalStatus = true;
                    app.actionInfo.modalTitle = '创建项目';
                } else if (action == 'edit') {
                    app.currentProjectCode = record.code;
                    app.projectModal.modalStatus = true;
                } else if (action == 'del') {
                    this.$confirm({
                        title: '确定放入回收站？',
                        content: `一旦将项目「${this.currentProject.name}」放入回收站，所有与项目有关的信息将会被放入回收站`,
                        okText: '放入回收站',
                        okType: 'danger',
                        cancelText: '再想想',
                        onOk() {
                            recycle(record.code).then(() => {
                                app.dataSource.splice(index, 1);
                            });
                            return Promise.resolve();
                        }
                    });
                } else if (action == 'collect') {
                    const type = record.collected ? 'cancel' : 'collect';
                    collect(record.code, type).then(() => {
                        app.$set(app.dataSource[index], 'collected', !record.collected);
                        if (this.requestData.type == 'collect') {
                            app.dataSource.splice(index, 1);
                        }
                    });
                } else if (action == 'recoveryArchive') {
                    this.$confirm({
                        title: '取消归档项目？',
                        content: `取消归档「${this.currentProject.name}」后就可以正常使用了`,
                        okText: '取消归档',
                        okType: 'primary',
                        cancelText: '再想想',
                        onOk() {
                            recoveryArchive(record.code).then(()=>{
                                app.dataSource.splice(index, 1);
                            });
                            return Promise.resolve();
                        }
                    });
                } else if (action == 'recovery') {
                    this.$confirm({
                        title: '确定恢复项目？',
                        content: `恢复「${this.currentProject.name}」后就可以正常使用了`,
                        okText: '恢复项目',
                        okType: 'primary',
                        cancelText: '再想想',
                        onOk() {
                            recovery(record.code).then(()=>{
                                app.dataSource.splice(index, 1);
                                app.removeFromSelection(record.code);
                            });
                            return Promise.resolve();
                        }
                    });
                } else if (action == 'permanentDel') {
                    del(record.code).then(res => {
                        if (checkResponse(res)) {
                            app.dataSource.splice(index, 1);
                            app.removeFromSelection(record.code);
                            app.$message.success('项目已彻底删除');
                        }
                    });
                }
            },
            updateProject(data) {
                this.dataSource[this.currentProjectIndex] = JSON.parse(JSON.stringify(data));
            },
            handleSubmit() {
                let app = this;
                app.form.validateFields(
                    (err) => {
                        if (!err) {
                            app.handleOk();
                        }
                    })
            },
            handleOk() {
                let app = this;
                app.actionInfo.confirmLoading = true;
                let obj = app.form.getFieldsValue();
                if (app.newData.code) {
                    //编辑
                    obj.projectCode = app.newData.code;
                } else {
                    //新增
                    Object.assign(obj, app.newData);
                }

                doData(obj).then(res => {
                    app.actionInfo.confirmLoading = false;
                    if (!checkResponse(res, true)) {
                        return;
                    }
                    app.form.resetFields();
                    app.newData = {id: 0};
                    app.actionInfo.modalStatus = false;
                    // 刷新项目列表
                    app.init();
                    // 显示成功提示
                    app.$message.success('项目创建成功');
                }).catch((err) => {
                    app.actionInfo.confirmLoading = false;
                    console.error('创建项目失败:', err);
                });

            },
            handleSearchSubmit() {
                this.search();
            },
            resetSearch() {
                this.searchParams = { name: '', dateRange: [] };
                this.requestData.name = '';
                this.requestData.startTime = '';
                this.requestData.endTime = '';
                this.init();
            },
            search() {
                this.requestData.name = this.searchParams.name || '';
                if (this.searchParams.dateRange && this.searchParams.dateRange.length === 2) {
                    this.requestData.startTime = this.searchParams.dateRange[0].format('YYYY-MM-DD');
                    this.requestData.endTime = this.searchParams.dateRange[1].format('YYYY-MM-DD');
                } else {
                    this.requestData.startTime = '';
                    this.requestData.endTime = '';
                }
                this.init();
            },
            selectChange() {
                this.selectedRowKeys = [];
                this.init();
            },
            // 批量操作相关方法
            onSelectChange(e, code) {
                if (e.target.checked) {
                    if (!this.selectedRowKeys.includes(code)) {
                        this.selectedRowKeys.push(code);
                    }
                } else {
                    this.removeFromSelection(code);
                }
            },
            onSelectAll(e) {
                if (e.target.checked) {
                    this.selectedRowKeys = this.dataSource.map(item => item.code);
                } else {
                    this.selectedRowKeys = [];
                }
            },
            removeFromSelection(code) {
                const index = this.selectedRowKeys.indexOf(code);
                if (index > -1) {
                    this.selectedRowKeys.splice(index, 1);
                }
            },
            clearSelection() {
                this.selectedRowKeys = [];
            },
            batchRecovery() {
                if (this.selectedRowKeys.length === 0) {
                    this.$message.warning('请选择要恢复的项目');
                    return;
                }
                const action = this.selectBy === 'archive' ? '取消归档' : '恢复';
                this.$confirm({
                    title: `批量${action}确认`,
                    content: `确定要${action}选中的 ${this.selectedRowKeys.length} 个项目吗？`,
                    okText: `确认${action}`,
                    okType: 'primary',
                    cancelText: '取消',
                    onOk: () => {
                        this.batchRecoveryLoading = true;
                        batchRecovery(this.selectedRowKeys).then(res => {
                            this.batchRecoveryLoading = false;
                            if (checkResponse(res)) {
                                this.$message.success(res.data.message || `批量${action}成功`);
                                this.selectedRowKeys = [];
                                this.init();
                            }
                        }).catch(() => {
                            this.batchRecoveryLoading = false;
                        });
                        return Promise.resolve();
                    }
                });
            },
            batchDelete() {
                if (this.selectedRowKeys.length === 0) {
                    this.$message.warning('请选择要删除的项目');
                    return;
                }
                this.$confirm({
                    title: '批量删除确认',
                    content: `确定要彻底删除选中的 ${this.selectedRowKeys.length} 个项目吗？此操作不可恢复！`,
                    okText: '彻底删除',
                    okType: 'danger',
                    cancelText: '取消',
                    onOk: () => {
                        this.batchDeleteLoading = true;
                        batchDelete(this.selectedRowKeys).then(res => {
                            this.batchDeleteLoading = false;
                            if (checkResponse(res)) {
                                this.$message.success(res.data.message || '批量删除成功');
                                this.selectedRowKeys = [];
                                this.init();
                            }
                        }).catch(() => {
                            this.batchDeleteLoading = false;
                        });
                        return Promise.resolve();
                    }
                });
            },
            saveProjectAsTemplate(item) {
                let app = this;
                app.templateModal.projectCode = item.code;
                app.templateModal.projectName = item.name;
                app.templateModal.visible = true;
                app.$nextTick(() => {
                    app.templateForm && app.templateForm.resetFields();
                    app.$nextTick(() => {
                        app.templateForm.setFieldsValue({
                            templateName: item.name + ' - 模板',
                            templateDescription: item.description || '',
                        });
                    });
                });
            },
            handleSaveTemplate() {
                let app = this;
                app.templateForm.validateFields((err) => {
                    if (err) return;
                    app.templateModal.loading = true;
                    let formData = app.templateForm.getFieldsValue();
                    // 先创建模板
                    doTemplateData({
                        name: formData.templateName,
                        description: formData.templateDescription,
                        cover: app.currentProject ? app.currentProject.cover : '',
                    }).then(templateRes => {
                        if (!checkResponse(templateRes)) {
                            app.templateModal.loading = false;
                            return;
                        }
                        // 获取项目的任务列表（stages），复制到模板中
                        let templateCode = templateRes.data.code || templateRes.data;
                        getTaskStages({projectCode: app.templateModal.projectCode}).then(stagesRes => {
                            let stages = stagesRes.data || [];
                            if (stages.length === 0) {
                                app.templateModal.loading = false;
                                app.templateModal.visible = false;
                                app.$message.success('模板创建成功（项目暂无任务列表）');
                                return;
                            }
                            // 逐个创建模板任务阶段
                            let promises = stages.map((stage, index) => {
                                return new Promise((resolve) => {
                                    // 使用 taskStagesTemplate API 保存阶段
                                    import('../../../api/taskStagesTemplate').then(module => {
                                        module.doData({
                                            templateCode: templateCode,
                                            name: stage.name,
                                            sort: index,
                                        }).then(resolve).catch(resolve);
                                    });
                                });
                            });
                            Promise.all(promises).then(() => {
                                app.templateModal.loading = false;
                                app.templateModal.visible = false;
                                app.$message.success(`模板「${formData.templateName}」创建成功，已包含 ${stages.length} 个任务列表`);
                                app.projectTemplates();
                            });
                        }).catch(() => {
                            app.templateModal.loading = false;
                            app.templateModal.visible = false;
                            app.$message.success('模板创建成功');
                        });
                    }).catch(() => {
                        app.templateModal.loading = false;
                    });
                });
            },
        }
    }
</script>
<style lang="less">
    @import "~ant-design-vue/lib/style/themes/default";

    .project-list-index {
        .batch-toolbar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 16px;
            background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
            border-radius: 8px;
            margin-bottom: 16px;
            
            .selected-info {
                display: flex;
                align-items: center;
                gap: 8px;
                font-weight: 500;
                color: #1890ff;
            }
            
            .batch-actions {
                display: flex;
                gap: 8px;
            }
        }
        
        .select-all-bar {
            padding: 12px 0;
            border-top: 1px solid #f0f0f0;
            margin-top: 16px;
        }
        
        .page-search {
            background: linear-gradient(135deg, #fafafa 0%, #f0f2f5 100%);
            padding: 20px 24px;
            border-radius: 12px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
            
            .ant-form-inline .ant-form-item {
                margin-bottom: 0;
            }
            
            .ant-input,
            .ant-calendar-picker {
                border-radius: 8px;
                transition: all 0.3s ease;
                
                &:hover, &:focus {
                    border-color: @primary-color;
                    box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.1);
                }
            }
            
            .ant-btn {
                border-radius: 8px;
                height: 36px;
                font-weight: 500;
                transition: all 0.3s ease;
                
                &:hover {
                    transform: translateY(-1px);
                    box-shadow: 0 4px 12px rgba(24, 144, 255, 0.25);
                }
            }
        }
        
        .ant-tabs {
            .ant-tabs-bar {
                border-bottom: 2px solid #f0f0f0;
                margin-bottom: 20px;
            }
            
            .ant-tabs-tab {
                font-size: 15px;
                font-weight: 500;
                padding: 12px 20px;
                transition: all 0.3s ease;
                
                &:hover {
                    color: @primary-color;
                }
                
                &-active {
                    font-weight: 600;
                }
            }
            
            .ant-tabs-ink-bar {
                height: 3px;
                border-radius: 2px;
            }
            
            .ant-tabs-tabpane {
                animation: fadeIn 0.4s ease;
            }
        }
        
        .ant-btn-primary {
            border-radius: 8px;
            height: 36px;
            font-weight: 500;
            background: linear-gradient(135deg, @primary-color 0%, #096dd9 100%);
            border: none;
            box-shadow: 0 4px 12px rgba(24, 144, 255, 0.35);
            transition: all 0.3s ease;
            
            &:hover {
                transform: translateY(-1px);
                box-shadow: 0 6px 16px rgba(24, 144, 255, 0.45);
                background: linear-gradient(135deg, #40a9ff 0%, @primary-color 100%);
            }
            
            &:active {
                transform: translateY(0);
            }
        }

        .project-list {
            .ant-list-item {
                background: #fff;
                border-radius: 12px;
                padding: 20px 24px;
                margin-bottom: 16px;
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
                border: 1px solid #f0f0f0;
                transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
                
                &:hover {
                    transform: translateY(-2px);
                    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
                    border-color: rgba(24, 144, 255, 0.2);
                }
                
                &-meta {
                    align-items: center;
                }
                
                &-meta-avatar {
                    .ant-avatar {
                        width: 56px;
                        height: 56px;
                        border-radius: 10px;
                        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                    }
                }
                
                &-meta-title {
                    font-size: 16px;
                    font-weight: 600;
                    margin-bottom: 6px;
                    
                    a {
                        color: rgba(0, 0, 0, 0.85);
                        transition: color 0.3s ease;
                        
                        &:hover {
                            color: @primary-color;
                        }
                    }
                    
                    .ant-tag {
                        font-size: 11px;
                        padding: 0 8px;
                        height: 20px;
                        line-height: 18px;
                        border-radius: 4px;
                        margin-left: 8px;
                    }
                }
                
                &-meta-description {
                    color: rgba(0, 0, 0, 0.45);
                    font-size: 13px;
                    line-height: 1.5;
                    max-width: 400px;
                }
            }

            .ant-list-item-content {
                display: flex;
                flex: 1;
                justify-content: flex-end;

                .other-info {
                    display: flex;
                    align-items: center;

                    .info-item {
                        display: flex;
                        flex-direction: column;
                        padding-left: 40px;
                        
                        span:first-child {
                            font-size: 12px;
                            color: rgba(0, 0, 0, 0.45);
                            margin-bottom: 4px;
                            text-transform: uppercase;
                            letter-spacing: 0.5px;
                        }
                        
                        span:last-child {
                            font-size: 14px;
                            color: rgba(0, 0, 0, 0.65);
                            font-weight: 500;
                        }
                    }

                    .schedule {
                        width: 200px;
                        
                        .ant-progress {
                            margin-top: 4px;
                            
                            &-text {
                                font-size: 12px;
                                color: rgba(0, 0, 0, 0.45);
                            }
                        }
                    }
                }
            }

            .ant-list-item-action {
                margin-left: 40px;
                
                li {
                    padding: 0 8px;
                }
                
                .anticon {
                    font-size: 18px;
                    padding: 8px;
                    border-radius: 8px;
                    transition: all 0.3s ease;
                    
                    &:hover {
                        background: rgba(24, 144, 255, 0.1);
                        
                        svg {
                            color: @primary-color;
                        }
                    }
                }
            }
        }
        
        // 模态框美化
        .ant-modal {
            &-content {
                border-radius: 16px;
                overflow: hidden;
            }
            
            &-header {
                background: linear-gradient(135deg, #fafafa 0%, #f5f5f5 100%);
                border-bottom: 1px solid #f0f0f0;
                padding: 20px 24px;
            }
            
            &-title {
                font-weight: 600;
                font-size: 17px;
            }
            
            &-body {
                padding: 24px;
            }
            
            .ant-input,
            .ant-select-selection,
            .ant-input-affix-wrapper {
                border-radius: 8px;
            }
            
            .ant-btn-primary.middle-btn {
                border-radius: 8px;
                height: 44px;
                font-size: 15px;
                font-weight: 500;
            }
        }
    }
    
    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
</style>
