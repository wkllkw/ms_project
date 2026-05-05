<template>
    <div class="register-result-wrapper">
        <div class="register-result-content">
            <a-result
                status="success"
                title="注册成功"
                :subTitle="subTitle"
            >
                <template slot="extra">
                    <a-button type="primary" @click="goLogin">
                        立即登录
                    </a-button>
                    <a-button style="margin-left: 8px" @click="goHome">
                        返回首页
                    </a-button>
                </template>
            </a-result>
            <div class="result-tips" v-if="account">
                <a-alert type="info" showIcon>
                    <template slot="message">
                        <span>您的账号为 <strong>{{ account }}</strong>，请妥善保管</span>
                    </template>
                </a-alert>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: 'RegisterResult',
    data() {
        return {
            account: '',
        }
    },
    computed: {
        subTitle() {
            if (this.account) {
                return `账号 ${this.account} 已注册成功，请登录使用系统。`;
            }
            return '您的账号已注册成功，请登录使用系统。';
        }
    },
    created() {
        // 从路由参数中获取注册账号
        const query = this.$route.query;
        if (query && query.account) {
            this.account = query.account;
        }
    },
    methods: {
        goLogin() {
            this.$router.push('/member/login');
        },
        goHome() {
            this.$router.push('/');
        }
    }
}
</script>

<style scoped lang="less">
.register-result-wrapper {
    width: 100%;
    min-height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 32px 0;

    .register-result-content {
        width: 500px;

        .result-tips {
            margin-top: 16px;
            text-align: left;
        }
    }
}
</style>
