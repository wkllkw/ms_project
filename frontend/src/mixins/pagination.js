// pagination-mixin.vue
export default {
    data() {
        return {
            pagination: {
                current: 1,
                page: 1,
                pageSize: 10,
                total: 0,
                showTotal: (total, range) => `共 ${total} 条`,
                showSizeChanger: true,
            },
        }
    },
    computed:{
        requestData(){
            return {
                page: this.pagination.current || this.pagination.page,
                pageSize: this.pagination.pageSize,
            };
        }
    },
    methods: {
        init() {
        },
        pageChange(pagination) {
            this.pagination.current = pagination.current;
            this.pagination.page = pagination.current;
            this.pagination.pageSize = pagination.pageSize;
            this.init();
        },
    }
}
