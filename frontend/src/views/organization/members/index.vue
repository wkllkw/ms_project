<template>
    <div class="org-members-index">
        <wrapper-content>
            <div class="header">
                <div class="title">组织成员管理</div>
                <div class="header-right">
                    <a-input-search
                        v-model="keyword"
                        placeholder="搜索成员名称/账号/邮箱"
                        style="width: 280px"
                        @search="onSearch"
                    />
                    <span class="m-l-md muted">共 {{ pagination.total }} 人</span>
                </div>
            </div>

            <a-table
                :columns="columns"
                :dataSource="members"
                :loading="loading"
                rowKey="id"
                :pagination="pagination"
                @change="pageChange"
            >
                <template slot="avatar" slot-scope="text, record">
                    <a-avatar :src="record.avatar" :size="32" />
                    <span class="m-l-sm">{{ record.name }}</span>
                    <a-tag class="m-l-xs" v-if="record.is_owner" color="orange">拥有者</a-tag>
                </template>

                <template slot="authorize" slot-scope="text, record">
                    <a-select
                        v-if="!record.is_owner && record.id != userInfo.id"
                        :value="record.authorize || 0"
                        style="width: 140px"
                        :loading="record._saving"
                        @change="(val) => handleRoleChange(record, val)"
                    >
                        <a-select-option :value="0" :key="0">
                            <span class="muted">无角色</span>
                        </a-select-option>
                        <a-select-option
                            v-for="role in authList"
                            :value="role.id"
                            :key="role.id"
                        >
                            {{ role.title }}
                            <span class="muted m-l-sm" style="font-size:12px">{{ role.desc }}</span>
                        </a-select-option>
                    </a-select>
                    <span v-else>
                        <a-tag v-if="record.is_owner" color="orange">组织拥有者</a-tag>
                        <a-tag v-else-if="record.id == userInfo.id" color="blue">我自己</a-tag>
                        <span v-else class="muted">-</span>
                    </span>
                </template>

                <template slot="departments" slot-scope="text, record">
                    <template v-if="record.departments && record.departments.length">
                        <a-tag v-for="(d, idx) in record.departments" :key="idx" class="m-r-xs">{{ d }}</a-tag>
                    </template>
                    <span v-else class="muted">未分配</span>
                </template>
            </a-table>
        </wrapper-content>
    </div>
</template>

<script>
    import { mapState } from 'vuex'
    import { _listMembers, _setMemberAuth } from '@/api/organization'
    import { checkResponse } from '@/assets/js/utils'
    import { notice } from '@/assets/js/notice'
    import pagination from '@/mixins/pagination'

    const columns = [{
        title: '成员',
        dataIndex: 'name',
        width: '25%',
        scopedSlots: { customRender: 'avatar' },
    }, {
        title: '邮箱',
        dataIndex: 'email',
        width: '20%',
    }, {
        title: '所属部门',
        dataIndex: 'departments',
        width: '25%',
        scopedSlots: { customRender: 'departments' },
    }, {
        title: '组织角色',
        dataIndex: 'authorize',
        width: '20%',
        scopedSlots: { customRender: 'authorize' },
    }]

    export default {
        name: 'organizationMembers',
        mixins: [pagination],
        data() {
            return {
                columns,
                members: [],
                authList: [],
                loading: false,
                keyword: '',
            }
        },
        computed: {
            ...mapState({
                userInfo: state => state.userInfo,
                currentOrganization: state => state.currentOrganization,
            })
        },
        created() {
            this.pagination.pageSize = 20
            this.init()
        },
        methods: {
            init() {
                if (!this.currentOrganization || !this.currentOrganization.code) {
                    notice({ title: '请先选择一个组织' }, 'message', 'warning')
                    return
                }
                this.loading = true
                const data = {
                    organizationCode: this.currentOrganization.code,
                    page: this.pagination.page,
                    pageSize: this.pagination.pageSize,
                }
                if (this.keyword) {
                    data.keyword = this.keyword
                }
                _listMembers(data).then(res => {
                    if (checkResponse(res)) {
                        this.members = (res.data.list || []).map(m => ({ ...m, _saving: false }))
                        this.pagination.total = res.data.total || 0
                        this.authList = res.data.authList || []
                    }
                    this.loading = false
                }).catch(() => {
                    this.loading = false
                })
            },
            onSearch() {
                this.pagination.page = 1
                this.init()
            },
            pageChange(pagination) {
                this.pagination.page = pagination.current
                this.init()
            },
            handleRoleChange(record, authId) {
                const app = this
                record._saving = true
                _setMemberAuth({
                    organizationCode: app.currentOrganization.code,
                    memberCode: record.code,
                    authId: authId,
                }).then(res => {
                    if (checkResponse(res)) {
                        record.authorize = authId
                        const role = app.authList.find(r => r.id === authId)
                        notice({
                            title: `已设置「${role ? role.title : '无角色'}」`
                        }, 'notice', 'success')
                    }
                    record._saving = false
                }).catch(() => {
                    record._saving = false
                })
            }
        }
    }
</script>

<style lang="less" scoped>
    .org-members-index {
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;

            .title {
                font-size: 18px;
                font-weight: 500;
            }

            .header-right {
                display: flex;
                align-items: center;
            }
        }
    }
</style>
