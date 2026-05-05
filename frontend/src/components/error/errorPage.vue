<template>
    <div class="error-page">
        <div class="error-bg-decoration">
            <div class="bg-circle bg-circle-1"></div>
            <div class="bg-circle bg-circle-2"></div>
        </div>
        <div class="error-brand">
            <a href="/">
                <img src="../../assets/image/common/logo.png" class="error-logo" alt="logo">
                <span class="error-brand-title">FlowHub</span>
            </a>
        </div>
        <div class="exception">
            <div class="imgBlock">
                <div class="img-exception">
                    <slot name="img"></slot>
                </div>
            </div>
            <div class="content">
                <h1>{{code}}</h1>
                <div class="desc">{{desc}}</div>
                <div class="sub-desc">{{subDesc}}</div>
                <div class="actions">
                    <router-link :to="url">
                        <a-button type="primary" size="large" icon="home">{{urlText}}</a-button>
                    </router-link>
                    <a-button size="large" icon="rollback" class="action-back" @click="goBack">返回上一页</a-button>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
    import {getStore} from "../../assets/js/storage";

    export default {
        props: {
            code: {
                default: '500'
            },
            desc: {
                default: '抱歉，服务器出错了',
            },
            subDesc: {
                default: '请稍后重试或联系管理员',
            },
            urlText: {
                default: '返回首页'
            }
        },
        data() {
            return {
                url: this.getHomeUrl()
            }
        },
        methods: {
            getHomeUrl() {
                const currentOrganization = getStore('currentOrganization', true);
                const homePath = currentOrganization ? '/home/' + currentOrganization.code : '/home';
                // 检查用户是否有 home 权限，没有则找第一个有权限的路由
                const permissionNodes = getStore('permissionNodes', true) || [];
                if (Array.isArray(permissionNodes) && permissionNodes.length > 0 && !permissionNodes.includes('home')) {
                    // 根据权限节点映射到可访问的路由
                    const nodeRouteMap = {
                        'project.list': '/project/list/my',
                        'project.template': '/project/template',
                        'project.analysis': '/project/analysis',
                        'notify.notice': '/notify/notice',
                        'calendar': '/calendar',
                        'members.index': '/members',
                    };
                    for (const node of permissionNodes) {
                        if (nodeRouteMap[node]) {
                            return nodeRouteMap[node];
                        }
                    }
                }
                return homePath;
            },
            goBack() {
                this.$router.go(-1);
            }
        }
    }
</script>
<style lang="less">
    .error-page {
        background: linear-gradient(135deg, #f5f7fa 0%, #e4e9f0 100%);
        height: 100vh;
        position: relative;
        overflow: hidden;

        .error-bg-decoration {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: 0;

            .bg-circle {
                position: absolute;
                border-radius: 50%;
                opacity: 0.5;

                &.bg-circle-1 {
                    width: 500px;
                    height: 500px;
                    background: radial-gradient(circle, rgba(24, 144, 255, 0.08) 0%, transparent 70%);
                    top: -150px;
                    right: -100px;
                    animation: bgFloat 10s ease-in-out infinite;
                }

                &.bg-circle-2 {
                    width: 400px;
                    height: 400px;
                    background: radial-gradient(circle, rgba(114, 46, 209, 0.06) 0%, transparent 70%);
                    bottom: -100px;
                    left: -80px;
                    animation: bgFloat 8s ease-in-out infinite reverse;
                }
            }
        }

        .error-brand {
            position: absolute;
            top: 32px;
            left: 40px;
            z-index: 2;

            a {
                display: flex;
                align-items: center;
                text-decoration: none;
                transition: opacity 0.3s;

                &:hover {
                    opacity: 0.8;
                }
            }

            .error-logo {
                width: 36px;
                height: 36px;
                margin-right: 10px;
            }

            .error-brand-title {
                font-size: 20px;
                font-weight: 700;
                color: #1890ff;
            }
        }
    }

    .exception {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100vh;
        padding: 0 48px;
        max-width: 1100px;
        margin: 0 auto;
        position: relative;
        z-index: 1;
    }

    .exception .imgBlock {
        flex: 0 0 50%;
        padding-right: 48px;
        animation: fadeInLeft 0.8s ease-out both;
    }

    .exception .img-exception {
        width: 100%;
        max-width: 420px;
        float: right;

        img {
            width: 100%;
            filter: drop-shadow(0 8px 24px rgba(0, 0, 0, 0.08));
        }
    }

    .exception .content {
        flex: auto;
        animation: fadeInRight 0.8s ease-out both;
        animation-delay: 0.2s;
    }

    .exception .content h1 {
        font-size: 80px;
        font-weight: 800;
        line-height: 1;
        margin-bottom: 16px;
        background: linear-gradient(135deg, #1890ff, #722ed1);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .exception .content .desc {
        color: rgba(0, 0, 0, .75);
        font-size: 24px;
        font-weight: 500;
        line-height: 1.4;
        margin-bottom: 8px;
    }

    .exception .content .sub-desc {
        color: rgba(0, 0, 0, .4);
        font-size: 15px;
        line-height: 1.6;
        margin-bottom: 28px;
    }

    .exception .content .actions {
        display: flex;
        gap: 12px;

        .ant-btn-primary {
            height: 44px;
            padding: 0 28px;
            border-radius: 8px;
            font-size: 15px;
            font-weight: 600;
            background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
            border: none;
            box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
            transition: all 0.3s;

            &:hover {
                transform: translateY(-1px);
                box-shadow: 0 6px 16px rgba(24, 144, 255, 0.4);
            }
        }

        .action-back {
            height: 44px;
            padding: 0 28px;
            border-radius: 8px;
            font-size: 15px;
            font-weight: 500;
            transition: all 0.3s;

            &:hover {
                color: #1890ff;
                border-color: #1890ff;
            }
        }
    }

    @keyframes fadeInLeft {
        from { opacity: 0; transform: translateX(-30px); }
        to { opacity: 1; transform: translateX(0); }
    }

    @keyframes fadeInRight {
        from { opacity: 0; transform: translateX(30px); }
        to { opacity: 1; transform: translateX(0); }
    }

    @keyframes bgFloat {
        0%, 100% { transform: translateY(0) scale(1); }
        50% { transform: translateY(-25px) scale(1.05); }
    }

    @media (max-width: 768px) {
        .exception {
            flex-direction: column;
            padding: 80px 24px 24px;
            text-align: center;
        }

        .exception .imgBlock {
            flex: none;
            padding: 0;
            margin-bottom: 24px;
        }

        .exception .img-exception {
            float: none;
            margin: 0 auto;
            max-width: 250px;
        }

        .exception .content h1 {
            font-size: 56px;
        }

        .exception .content .actions {
            justify-content: center;
        }

        .error-brand {
            left: 50%;
            transform: translateX(-50%);
        }
    }
</style>
