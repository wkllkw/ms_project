<template>
    <div class="userLayout">
        <div class="container">
            <!-- 左侧品牌展示区 -->
            <div class="brand-side">
                <div class="brand-content">
                    <div class="brand-logo">
                        <img src="../../assets/image/common/logo.png" class="brand-logo-img" alt="logo">
                    </div>
                    <h1 class="brand-title">FlowHub</h1>
                    <p class="brand-desc">{{ desc }}</p>
                    <div class="brand-features">
                        <div class="feature-item">
                            <a-icon type="project" class="feature-icon" />
                            <div class="feature-text">
                                <h4>项目管理</h4>
                                <p>高效的任务看板与甘特图</p>
                            </div>
                        </div>
                        <div class="feature-item">
                            <a-icon type="team" class="feature-icon" />
                            <div class="feature-text">
                                <h4>团队协作</h4>
                                <p>实时沟通，无缝协同</p>
                            </div>
                        </div>
                        <div class="feature-item">
                            <a-icon type="bar-chart" class="feature-icon" />
                            <div class="feature-text">
                                <h4>数据分析</h4>
                                <p>多维度可视化数据报表</p>
                            </div>
                        </div>
                    </div>
                </div>
                <!-- 动态装饰元素 -->
                <div class="brand-decoration">
                    <div class="circle circle-1"></div>
                    <div class="circle circle-2"></div>
                    <div class="circle circle-3"></div>
                </div>
            </div>

            <!-- 右侧表单区 -->
            <div class="form-side">
                <div class="form-wrapper">
                    <div class="top">
                        <div class="header">
                            <a href="/">
                                <img src="../../assets/image/common/logo.png" class="logo" alt="logo">
                                <span class="title">FlowHub</span>
                            </a>
                        </div>
                    </div>

                    <router-view/>
                    <slot></slot>

                    <div class="footer">
                        <div class="copyright">
                            Copyright &copy; 2022 FlowHub
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>

    export default {
        name: 'UserLayout',
        props: {
            desc: {
                type: [String],
                default() {
                    return '欢迎使用 FlowHub 项目管理系统';
                }
            }
        }
    }
</script>

<style lang="less">
    .userLayout {
        height: 100%;
        overflow: auto;

        .container {
            width: 100%;
            min-height: 100vh;
            display: flex;
            position: relative;
        }

        // ========== 左侧品牌展示区 ==========
        .brand-side {
            flex: 1;
            background: linear-gradient(135deg, #1890ff 0%, #096dd9 40%, #0050b3 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            position: relative;
            overflow: hidden;
            min-height: 100vh;

            .brand-content {
                position: relative;
                z-index: 2;
                padding: 60px 48px;
                max-width: 480px;
            }

            .brand-logo {
                margin-bottom: 24px;

                .brand-logo-img {
                    width: 72px;
                    height: 72px;
                    filter: brightness(0) invert(1);
                    opacity: 0.95;
                }
            }

            .brand-title {
                font-size: 42px;
                font-weight: 700;
                color: #fff;
                margin-bottom: 12px;
                letter-spacing: -0.5px;
            }

            .brand-desc {
                font-size: 16px;
                color: rgba(255, 255, 255, 0.75);
                margin-bottom: 48px;
                line-height: 1.6;
            }

            .brand-features {
                .feature-item {
                    display: flex;
                    align-items: flex-start;
                    margin-bottom: 28px;
                    animation: fadeInUp 0.6s ease-out both;

                    &:nth-child(1) { animation-delay: 0.2s; }
                    &:nth-child(2) { animation-delay: 0.4s; }
                    &:nth-child(3) { animation-delay: 0.6s; }

                    .feature-icon {
                        font-size: 28px;
                        color: rgba(255, 255, 255, 0.9);
                        margin-right: 16px;
                        margin-top: 2px;
                        background: rgba(255, 255, 255, 0.15);
                        padding: 10px;
                        border-radius: 12px;
                    }

                    .feature-text {
                        h4 {
                            color: #fff;
                            font-size: 16px;
                            font-weight: 600;
                            margin-bottom: 4px;
                        }
                        p {
                            color: rgba(255, 255, 255, 0.65);
                            font-size: 13px;
                            margin: 0;
                        }
                    }
                }
            }

            // 装饰动画圆
            .brand-decoration {
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                pointer-events: none;
                z-index: 1;

                .circle {
                    position: absolute;
                    border-radius: 50%;
                    background: rgba(255, 255, 255, 0.06);

                    &.circle-1 {
                        width: 400px;
                        height: 400px;
                        top: -120px;
                        right: -100px;
                        animation: floatAnimation 8s ease-in-out infinite;
                    }

                    &.circle-2 {
                        width: 250px;
                        height: 250px;
                        bottom: -60px;
                        left: -40px;
                        animation: floatAnimation 6s ease-in-out infinite reverse;
                    }

                    &.circle-3 {
                        width: 150px;
                        height: 150px;
                        top: 50%;
                        right: 10%;
                        animation: floatAnimation 10s ease-in-out infinite 2s;
                    }
                }
            }
        }

        // ========== 右侧表单区 ==========
        .form-side {
            flex: 0 0 520px;
            display: flex;
            align-items: center;
            justify-content: center;
            background: #fff;
            min-height: 100vh;
            box-shadow: -8px 0 24px rgba(0, 0, 0, 0.06);
            position: relative;
            z-index: 2;

            .form-wrapper {
                width: 100%;
                max-width: 400px;
                padding: 40px 48px;
            }

            a {
                text-decoration: none;
            }

            .top {
                text-align: center;
                margin-bottom: 32px;

                .header {
                    height: 44px;
                    line-height: 44px;

                    .logo {
                        height: 40px;
                        vertical-align: top;
                        margin-right: 12px;
                        border-style: none;
                    }

                    .title {
                        font-size: 28px;
                        color: rgba(0, 0, 0, .85);
                        font-weight: 700;
                        position: relative;
                        top: 2px;
                    }
                }
            }

            .main {
                min-width: 260px;
                width: 100%;
                margin: 0 auto;
            }

            .footer {
                text-align: center;
                margin-top: 48px;

                .copyright {
                    color: rgba(0, 0, 0, 0.35);
                    font-size: 13px;
                }
            }
        }

        // ========== 动画 ==========
        @keyframes fadeInUp {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        @keyframes floatAnimation {
            0%, 100% {
                transform: translateY(0) scale(1);
            }
            50% {
                transform: translateY(-20px) scale(1.05);
            }
        }

        // ========== 响应式适配 ==========
        @media (max-width: 992px) {
            .container {
                flex-direction: column;
            }

            .brand-side {
                min-height: auto;
                padding: 40px 20px;

                .brand-content {
                    padding: 20px;
                    text-align: center;
                    max-width: 100%;
                }

                .brand-features {
                    display: none;
                }

                .brand-title {
                    font-size: 28px;
                }

                .brand-desc {
                    margin-bottom: 0;
                }
            }

            .form-side {
                flex: 1;
                min-height: auto;
                box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.06);

                .form-wrapper {
                    padding: 32px 24px;
                }
            }
        }

        @media (max-width: 576px) {
            .form-side {
                flex: 0 0 auto;

                .form-wrapper {
                    padding: 24px 16px;
                    max-width: 100%;
                }

                .top .header .title {
                    font-size: 22px;
                }
            }
        }
    }
</style>
