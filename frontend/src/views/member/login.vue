<template>
    <div class="main">
        <a-spin class="text-center" :spinning="oauthLoading">
            <span v-show="oauthLoading">正在登陆，请稍后...</span>
        </a-spin>
        <a-form
                v-show="!oauthLoading"
                class="user-layout-login"
                ref="formLogin"
                id="formLogin"
                :form="form"
        >
            <a-tabs
                    :activeKey="customActiveKey"
                    :tabBarStyle="{ textAlign: 'center', borderBottom: 'unset' }"
                    @change="handleTabClick"
            >
                <a-tab-pane key="tab1" tab="账号密码登录">
                    <a-form-item>
                        <a-input size="large" type="text" placeholder="帐户名或邮箱地址"
                                 v-decorator="[
                                'account',
                                {rules: [{ required: true, message: '请输入帐户名或邮箱地址' },{ validator: this.handleUsernameOrEmail }], validateTrigger: 'blur'}
                            ]">
                            <a-icon slot="prefix" type="user" :style="{ color: 'rgba(0,0,0,.25)' }"/>
                        </a-input>
                    </a-form-item>

                    <a-form-item
                    >
                        <a-input size="large" type="password" autocomplete="false" placeholder="密码"
                                 v-decorator="[
                                'password',
                                {rules: [{ required: true, message: '请输入密码' }], validateTrigger: 'blur'}
                            ]">
                            <a-icon slot="prefix" type="lock" :style="{ color: 'rgba(0,0,0,.25)' }"/>
                        </a-input>
                    </a-form-item>
                </a-tab-pane>
                <a-tab-pane key="tab2" tab="手机号登录">
                    <a-form-item
                    >
                        <a-input size="large" type="text" placeholder="手机号"
                                 v-decorator="[
                                'mobile',
                                {rules: [{ required: true, pattern: /^1[34578]\d{9}$/, message: '请输入正确的手机号' },{ validator: this.handleUsernameOrEmail }], validateTrigger: 'change'}
                            ]">
                            <a-icon slot="prefix" type="mobile" :style="{ color: 'rgba(0,0,0,.25)' }"/>
                        </a-input>
                    </a-form-item>

                    <a-row :gutter="16">
                        <a-col class="gutter-row" :span="16">
                            <a-form-item
                            >
                                <a-input size="large" type="text" placeholder="验证码"
                                         v-decorator="[
                                'captcha',
                                {rules: [{ required: true, message: '请输入验证码' }], validateTrigger: 'blur'}
                            ]">
                                    <a-icon slot="prefix" type="mail" :style="{ color: 'rgba(0,0,0,.25)' }"/>
                                </a-input>
                            </a-form-item>
                        </a-col>
                        <a-col class="gutter-row" :span="8">
                            <a-button
                                    class="getCaptcha"
                                    tabindex="-1"
                                    :disabled="state.smsSendBtn"
                                    @click.stop.prevent="getCaptcha"
                                    v-text="!state.smsSendBtn && '获取验证码' || (state.time+' s')"
                            ></a-button>
                        </a-col>
                    </a-row>
                </a-tab-pane>
            </a-tabs>

            <a-form-item>
                <a-checkbox v-model="formLogin.rememberMe">自动登录</a-checkbox>
                <a
                        class="forge-password"
                        style="float: right;"
                        @click="routerLink('/member/forgot')"
                >忘记密码
                </a>
            </a-form-item>

            <a-form-item style="margin-top:24px">
                <a-button
                        size="large"
                        type="primary"
                        htmlType="submit"
                        class="login-button"
                        :loading="loginBtn"
                        @click.stop.prevent="handleSubmit"
                        :disabled="loginBtn"
                >登录
                </a-button>
            </a-form-item>

            <div class="user-login-other">
                <router-link class="register" :to="{ name: 'register' }">注册账户</router-link>
            </div>
        </a-form>
    </div>
</template>

<script>
    import md5 from 'md5'
    import {mapActions} from 'vuex'
    import {mapState} from 'vuex'
    import {Login, getCaptcha} from '@/api/user'
    import {info} from '@/api/system';
    import config from "@/config/config";
    import {appendMenuRoutes, checkResponse, getFirstAvailableRoute, timeFix} from '@/assets/js/utils'
    import {getStore} from '@/assets/js/storage'
    import {_checkLogin} from "../../api/user";
    import {notice} from "../../assets/js/notice";

    export default {
        components: {},
        data() {
            return {
                customActiveKey: 'tab1',
                loginBtn: false,
                oauthLoading: false,
                // login type: 0 email, 1 account, 2 telephone
                loginType: 0,
                requiredTwoStepCaptcha: false,
                stepCaptchaVisible: false,
                form: this.$form.createForm(this),
                state: {
                    time: 60,
                    smsSendBtn: false
                },
                formLogin: {
                    account: '',
                    password: '',
                    captcha: '',
                    mobile: '',
                    rememberMe: true
                }
            }
        },
        computed: {
            ...mapState({
                system: state => state.system,
            })
        },
        mounted() {
            if (this.$route.query.logged) {
                this.oauthLoading = true;
                this.checkLogin();
            }
            if (this.$route.query.message) {
                notice({title: this.$route.query.message}, 'notice');
                // notice(this.$route.query.message);
            }
        },
        methods: {
            ...mapActions(['Login', 'Logout']),
            // handler
            handleUsernameOrEmail(rule, value, callback) {
                const regex = /^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+((\.[a-zA-Z0-9_-]{2,3}){1,2})$/;
                if (regex.test(value)) {
                    this.loginType = 0
                } else {
                    this.loginType = 1
                }
                callback()
            },
            handleTabClick(key) {
                this.customActiveKey = key
                // this.form.resetFields()
            },
            handleSubmit() {
                const app = this;
                let flag = false;

                let loginParams = {
                    remember_me: app.formLogin.rememberMe
                };

                // 使用账户密码登录
                if (app.customActiveKey === 'tab1') {
                    app.form.validateFields(['account', 'password'], {force: true}, (err, values) => {
                        if (!err) {
                            flag = true;
                            loginParams[!app.loginType ? 'account' : 'account'] = values.account;
                            loginParams.password = md5(values.password)
                        }
                    })
                    // 使用手机号登录
                } else {
                    app.form.validateFields(['mobile', 'captcha'], {force: true}, (err, values) => {
                        if (!err) {
                            flag = true;
                            loginParams = Object.assign(loginParams, values)
                        }
                    })
                }

                if (!flag) return;

                app.loginBtn = true;
                loginParams.clientid = getStore('client_id');
                Login(loginParams).then(res => {
                    if (checkResponse(res)) {
                        loginParams.token = res.token;
                        this.dealDataBeforeLogin(res);
                    }
                    this.loginBtn = false
                }).catch(() => {
                    this.loginBtn = false
                });
            },
            getCaptcha(e) {
                e.preventDefault();
                const app = this;

                this.form.validateFields(['mobile'], {force: true}, (err, values) => {
                    if (!err) {
                        this.state.smsSendBtn = true;

                        const interval = window.setInterval(() => {
                            if (app.state.time-- <= 0) {
                                app.state.time = 60;
                                app.state.smsSendBtn = false;
                                window.clearInterval(interval)
                            }
                        }, 1000);

                        const hide = this.$message.loading('验证码发送中..', 0);
                        getCaptcha(values.mobile)
                            .then(res => {
                                this.$message.destroy();

                                if (!checkResponse(res)) {
                                    return false;
                                }
                                let tips = '验证码获取成功';
                                if (res.data) {
                                    tips += '，您的验证码为：' + res.data;
                                }
                                this.$notification['success']({
                                    message: '提示',
                                    description: tips,
                                    duration: 8,
                                    placement: 'bottomLeft'
                                });
                            })
                            .catch(err => {
                                // setTimeout(hide, 1);
                                clearInterval(interval);
                                app.state.time = 60;
                                app.state.smsSendBtn = false;
                                this.requestFailed(err)
                            })
                    }
                })
            },
            loginSuccess(res, org) {
                const menu = getStore('menu', true);
                if (menu) {
                    let routes = this.$router.options.routes;
                    appendMenuRoutes(routes[0].children, menu);
                    this.$router.addRoutes(routes);
                }
                this.loginBtn = false;
                const orgCode = org && org.code ? org.code : '';
                // 检查是否有 URL 参数指定 redirect，否则根据权限智能选择
                let redirect = this.$route.query.redirect;
                if (redirect == config.HOME_PAGE) {
                    redirect = null; // 等同于无指定，走智能路由
                }
                if (!redirect) {
                    const permissionNodes = this.$store.state.permissionNodes || getStore('permissionNodes', true) || [];
                    redirect = getFirstAvailableRoute(permissionNodes, org);
                }
                // 如果权限路由返回登录页（极端情况），则回退到首页
                if (redirect === '/member/login') {
                    redirect = config.HOME_PAGE + (orgCode ? '/' + orgCode : '');
                }
                this.$router.replace({
                    path: redirect
                }).catch(() => {});
                this.$notification.success({
                    message: '欢迎',
                    description: `${res.data.member.name}，${timeFix()}，欢迎回来`,
                });
                this.oauthLoading = false;
            },
            checkLogin() {
                _checkLogin().then(res => {
                    this.dealDataBeforeLogin(res);
                });
            },
            async dealDataBeforeLogin(res) {
                let app = this;
                if (res.data) {
                    const obj = {
                        userInfo: res.data.member,
                        tokenList: res.data.tokenList
                    };
                    let currentOrganization = getStore('currentOrganization', true);
                    const organizationList = res.data.organizationList || [];
                    await app.$store.dispatch('SET_LOGGED', obj);
                    await app.$store.dispatch('setOrganizationList', organizationList);
                    if (!currentOrganization && organizationList.length > 0) {
                        currentOrganization = organizationList[0];
                    } else if (currentOrganization && organizationList.length > 0) {
                        const has = organizationList.findIndex(item => item.id == currentOrganization.id);
                        if (has === -1) {
                            currentOrganization = organizationList[0];
                        }
                    }
                    await app.$store.dispatch('setCurrentOrganization', currentOrganization);
                    await app.$store.dispatch('GET_MENU').then(async () => {
                        // 获取权限节点（菜单API可能已返回，这里作为兜底）
                        await app.$store.dispatch('FETCH_PERMISSION_NODES');
                        app.loginSuccess(res, currentOrganization);
                    });
                } else {
                    app.oauthLoading = false;
                    app.$store.dispatch('SET_LOGOUT');
                    // app.$router.replace('/login?redirect=' + this.$router.currentRoute.fullPath);
                }
            },
            requestFailed(err) {
                this.$notification['error']({
                    message: '错误',
                    description: ((err.response || {}).data || {}).message || '请求出现错误，请稍后再试',
                    duration: 4
                });
                this.loginBtn = false
            }
        }
    }
</script>

<style lang="less">
    .user-layout-login {
        label {
            font-size: 14px;
        }

        .ant-tabs-bar {
            border-bottom: none;
        }

        .ant-tabs-tab {
            font-size: 15px;
            font-weight: 500;
        }

        .ant-input-lg {
            border-radius: 8px;
            height: 44px;
            transition: all 0.3s ease;

            &:focus, &:hover {
                border-color: #1890ff;
                box-shadow: 0 2px 8px rgba(24, 144, 255, 0.15);
            }
        }

        .ant-input-affix-wrapper .ant-input:not(:first-child) {
            padding-left: 36px;
        }

        .getCaptcha {
            display: block;
            width: 100%;
            height: 44px;
            border-radius: 8px;
            font-weight: 500;
        }

        .forge-password {
            font-size: 14px;
            color: #1890ff;
            transition: color 0.3s;

            &:hover {
                color: #40a9ff;
            }
        }

        button.login-button {
            padding: 0 15px;
            font-size: 16px;
            height: 46px;
            width: 100%;
            border-radius: 8px;
            font-weight: 600;
            letter-spacing: 1px;
            background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
            border: none;
            box-shadow: 0 4px 12px rgba(24, 144, 255, 0.35);
            transition: all 0.3s ease;

            &:hover {
                transform: translateY(-1px);
                box-shadow: 0 6px 16px rgba(24, 144, 255, 0.45);
                background: linear-gradient(135deg, #40a9ff 0%, #1890ff 100%);
            }

            &:active {
                transform: translateY(0);
                box-shadow: 0 2px 8px rgba(24, 144, 255, 0.3);
            }
        }

        .user-login-other {
            text-align: left;
            margin-top: 24px;
            line-height: 22px;
            padding-top: 16px;
            border-top: 1px solid #f0f0f0;

            .item-icon {
                font-size: 24px;
                color: rgba(0, 0, 0, 0.2);
                margin-left: 16px;
                vertical-align: middle;
                cursor: pointer;
                transition: all 0.3s;
                padding: 4px;
                border-radius: 50%;

                &:hover {
                    color: #1890ff;
                    background: rgba(24, 144, 255, 0.08);
                }
            }

            .register {
                float: right;
                font-weight: 500;
                color: #1890ff;
                transition: color 0.3s;

                &:hover {
                    color: #40a9ff;
                }
            }
        }
    }
</style>
