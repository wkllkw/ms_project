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
                <!-- 装饰元素（隐藏） -->
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

        // ========== 左侧品牌展示区 - 暖灰底 + 微装饰 ==========
        .brand-side {
            flex: 1;
            background: #f7f5f2;
            display: flex;
            align-items: center;
            justify-content: center;
            position: relative;
            overflow: hidden;
            min-height: 100vh;

            // 微妙的渐变装饰
            &::before {
                content: '';
                position: absolute;
                top: -50%;
                right: -20%;
                width: 70%;
                height: 100%;
                background: radial-gradient(ellipse, rgba(58, 130, 248, 0.04) 0%, transparent 70%);
                pointer-events: none;
            }

            &::after {
                content: '';
                position: absolute;
                bottom: -30%;
                left: -10%;
                width: 50%;
                height: 60%;
                background: radial-gradient(ellipse, rgba(82, 196, 26, 0.03) 0%, transparent 70%);
                pointer-events: none;
            }

            .brand-content {
                position: relative;
                z-index: 2;
                padding: 60px 48px;
                max-width: 440px;
                animation: fadeInUp 0.5s ease-out;
            }

            .brand-logo {
                margin-bottom: 24px;

                .brand-logo-img {
                    width: 56px;
                    height: 56px;
                    opacity: 0.85;
                }
            }

            .brand-title {
                font-size: 36px;
                font-weight: 700;
                color: rgba(0, 0, 0, 0.85);
                margin-bottom: 12px;
                letter-spacing: -0.5px;
            }

            .brand-desc {
                font-size: 15px;
                color: rgba(0, 0, 0, 0.45);
                margin-bottom: 48px;
                line-height: 1.6;
            }

            .brand-features {
                .feature-item {
                    display: flex;
                    align-items: flex-start;
                    margin-bottom: 24px;
                    transition: transform 0.2s ease;

                    &:hover {
                        transform: translateX(4px);

                        .feature-icon {
                            background: rgba(58, 130, 248, 0.08);
                            color: #3a82f8;
                        }
                    }

                    .feature-icon {
                        font-size: 24px;
                        color: rgba(0, 0, 0, 0.25);
                        margin-right: 14px;
                        margin-top: 2px;
                        background: rgba(0, 0, 0, 0.04);
                        padding: 8px;
                        border-radius: 8px;
                        transition: background 0.2s ease, color 0.2s ease;
                    }

                    .feature-text {
                        h4 {
                            color: rgba(0, 0, 0, 0.75);
                            font-size: 15px;
                            font-weight: 600;
                            margin-bottom: 4px;
                        }
                        p {
                            color: rgba(0, 0, 0, 0.4);
                            font-size: 13px;
                            margin: 0;
                        }
                    }
                }
            }

            // 装饰圆圈 - 恢复但极简化
            .brand-decoration {
                display: block;
                position: absolute;
                inset: 0;
                pointer-events: none;
                z-index: 1;

                .circle {
                    position: absolute;
                    border-radius: 50%;
                    border: 1px solid rgba(0, 0, 0, 0.04);
                }

                .circle-1 {
                    width: 300px;
                    height: 300px;
                    top: -80px;
                    right: -60px;
                    animation: float1 8s ease-in-out infinite;
                }

                .circle-2 {
                    width: 200px;
                    height: 200px;
                    bottom: -40px;
                    left: -30px;
                    animation: float2 10s ease-in-out infinite;
                }

                .circle-3 {
                    width: 120px;
                    height: 120px;
                    top: 40%;
                    right: 10%;
                    animation: float3 12s ease-in-out infinite;
                }
            }
        }

        @keyframes float1 {
            0%, 100% { transform: translate(0, 0); }
            50% { transform: translate(-8px, 12px); }
        }

        @keyframes float2 {
            0%, 100% { transform: translate(0, 0); }
            50% { transform: translate(10px, -8px); }
        }

        @keyframes float3 {
            0%, 100% { transform: translate(0, 0); }
            50% { transform: translate(-6px, -10px); }
        }

        @keyframes fadeInUp {
            from { opacity: 0; transform: translateY(16px); }
            to { opacity: 1; transform: translateY(0); }
        }

        // ========== 右侧表单区 ==========
        .form-side {
            flex: 0 0 480px;
            display: flex;
            align-items: center;
            justify-content: center;
            background: #fff;
            min-height: 100vh;
            position: relative;
            z-index: 2;

            .form-wrapper {
                width: 100%;
                max-width: 380px;
                padding: 40px 48px;
                animation: fadeInUp 0.4s ease-out 0.1s both;
            }

            a {
                text-decoration: none;
            }

            .top {
                text-align: center;
                margin-bottom: 28px;

                .header {
                    height: 40px;
                    line-height: 40px;

                    .logo {
                        height: 36px;
                        vertical-align: top;
                        margin-right: 10px;
                        border-style: none;
                    }

                    .title {
                        font-size: 24px;
                        color: rgba(0, 0, 0, 0.85);
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
                margin-top: 40px;

                .copyright {
                    color: rgba(0, 0, 0, 0.25);
                    font-size: 12px;
                }
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
                    font-size: 24px;
                }

                .brand-desc {
                    margin-bottom: 0;
                }
            }

            .form-side {
                flex: 1;
                min-height: auto;

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
                    font-size: 20px;
                }
            }
        }
    }
</style>
