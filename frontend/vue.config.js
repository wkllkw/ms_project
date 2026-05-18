const path = require('path');

function resolve(dir) {
    return path.join(__dirname, dir)
}

module.exports = {
    publicPath: process.env.NODE_ENV === 'production' ? process.env.VUE_APP_BUILD_PATH : '/',
    productionSourceMap: false,
    css: {
        loaderOptions: {
            less: {
                lessOptions: {
                    math: 'always',
                    modifyVars: {
                        // ========== Modern Theme v2.0 ==========
                        'primary-color': '#4f8cff',
                        'info-color': '#4f8cff',
                        'link-color': '#4f8cff',
                        'success-color': '#22c55e',
                        'warning-color': '#f59e0b',
                        'error-color': '#ef4444',
                        'border-radius-base': '8px',
                        'border-radius-sm': '4px',
                        'font-size-base': '14px',
                        'font-family': 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif',
                        'heading-color': '#1a1d23',
                        'text-color': '#1a1d23',
                        'text-color-secondary': '#5f6773',
                        'disabled-color': 'rgba(0, 0, 0, 0.25)',
                        'border-color-base': '#e4e7eb',
                        'border-color-split': '#eef0f2',
                        'background-color-base': '#f5f7fb',
                        'box-shadow-base': '0 1px 3px rgba(0, 0, 0, 0.06)',
                        'btn-default-border': '#e4e7eb',
                        'btn-default-color': '#5f6773',
                        'btn-border-radius-base': '8px',
                        'btn-height-base': '34px',
                        'input-height-base': '36px',
                        'card-radius': '12px',
                        'card-shadow': '0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.03)',
                        'modal-border-radius': '12px',
                        'layout-body-background': '#f5f7fb',
                        'layout-header-background': 'rgba(255,255,255,0.82)',
                        'layout-sider-background': '#f9fafb',
                    },
                    javascriptEnabled: true
                }
            }
        }
    },
    devServer: {
        host: process.env.VUE_APP_DEV_HOST || '0.0.0.0',
        port: process.env.VUE_APP_DEV_PORT || '8045',
        hot: true,
        client: {
            overlay: {
                errors: true,
                warnings: false,
                runtimeErrors: (error) => {
                    const navigationErrorNames = [
                        'NavigationCancelled',
                        'NavigationDuplicated',
                        'NavigationRedirected',
                        'NavigationAborted',
                    ];
                    if (error && navigationErrorNames.some(name => error.name === name)) {
                        return false;
                    }
                    if (error && error.message && error.message.includes('Navigation cancelled')) {
                        return false;
                    }
                    return true;
                }
            }
        },
        proxy: { // 配置跨域
            '/api': {
                //要访问的跨域的api的域名
                target: process.env.VUE_APP_API_URL || 'http://127.0.0.1:18000',
                ws: true,
                changeOrigin: true,
                pathRewrite: {
                    '^/api': ''
                }
            },
            // 代理上传文件路径，使上传的图片/文件可访问
            '/uploads': {
                target: process.env.VUE_APP_API_URL || 'http://127.0.0.1:18000',
                ws: true,
                changeOrigin: true,
            }
        },
    },
    configureWebpack: config => {
        config.resolve = {
            extensions: ['.js', '.vue', '.json', ".css"],
            alias: {
                'vue$': 'vue/dist/vue.esm.js',
                '@': resolve('src'),
                'assets': resolve('src/assets'),
                'components': resolve('src/components')
            }
        }
        // 开发模式启用持久化缓存加速构建
        if (process.env.NODE_ENV === 'development') {
            config.cache = {
                type: 'filesystem',
                buildDependencies: {
                    config: [__filename]
                }
            }
        }
    },
    lintOnSave: false
};
