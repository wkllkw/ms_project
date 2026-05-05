<template>
    <div class="admin-auth-apply">
        <wrapper-content>
            <div slot="page-action">

            </div>
            <div style="padding: 12px;">
                <Button @click="back">返回</Button>
                <Button type="primary" @click="saveAuth" :style="{marginLeft: '4px'}">保存</Button>
            </div>
            <Tree
                    checkable
                    :defaultExpandAll="defaultExpandAll"
                    :autoExpandParent="autoExpandParent"
                    v-if="nodeList.length"
                    :treeData="nodeList"
                    :checkedKeys="checkedList"
                    @check="onCheck"
            >
            </Tree>
        </wrapper-content>
    </div>
</template>
<script>
    import {Button, Tree} from 'ant-design-vue';
    import {apply} from '@/api/auth';
    import {getNodeList} from '@/api/node';
    import {checkResponse} from '@/assets/js/utils';

    export default {
        components: {
            Button,
            Tree
        },
        data() {
            return {
                nodeList: [],
                checkedList: [],
                defaultExpandAll: true,
                autoExpandParent: true,
            }
        },
        created() {
            this.init();
        },
        methods: {
            init() {
                let app = this;
                // 同时加载节点树和已选中的节点
                Promise.all([
                    getNodeList(),
                    apply(app.$route.params.id)
                ]).then(([nodeRes, authRes]) => {
                    // 转换节点树数据为 Tree 组件需要的格式
                    let nodes = [];
                    if (nodeRes && nodeRes.data && nodeRes.data.nodes) {
                        nodes = app.transformNodeTree(nodeRes.data.nodes);
                    }
                    app.nodeList = nodes;

                    // 获取已选中的节点列表
                    let checked = [];
                    if (authRes && authRes.data) {
                        if (authRes.data.checkedList && Array.isArray(authRes.data.checkedList)) {
                            checked = authRes.data.checkedList;
                        } else if (authRes.data.list && Array.isArray(authRes.data.list)) {
                            checked = authRes.data.list;
                        }
                    }
                    app.checkedList = checked;
                });
            },
            // 将后端节点数据转换为 Tree 组件格式
            transformNodeTree(nodes) {
                if (!nodes || !Array.isArray(nodes)) return [];
                return nodes.map(node => {
                    let item = {
                        title: node.title || '',
                        key: node.node || '',
                        value: node.node || '',
                    };
                    if (node.children && node.children.length > 0) {
                        item.children = this.transformNodeTree(node.children);
                    }
                    return item;
                });
            },
            onCheck(checkedKeys) {
                this.checkedList = checkedKeys;
            },
            saveAuth() {
                let app = this;
                apply(app.$route.params.id, JSON.stringify(app.checkedList), 'save').then(res => {
                    if (checkResponse(res)) {
                        app.$message.success('保存成功');
                    }
                });
            },
            back(){
                this.$router.push('/system/account/auth')
            }
        }
    }
</script>
<style lang="less">
    .admin-auth-apply {
        .ant-tree li ul {
            margin: 12px 0;
        }
        .ant-tree-child-tree-open {
            .ant-tree-child-tree-open li {
                display: inline;
                white-space: initial;
            }
        }
    }
</style>
