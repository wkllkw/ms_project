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
                        'primary-color': '#3a82f8',
                        'link-color': '#3a82f8',
                        'border-radius-base': '6px',
                        'border-radius-sm': '4px',
                        'font-size-base': '14px',
                        'font-family': 'Inter, -apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", sans-serif',
                        'heading-color': 'rgba(0, 0, 0, 0.88)',
                        'text-color': 'rgba(0, 0, 0, 0.75)',
                        'text-color-secondary': 'rgba(0, 0, 0, 0.45)',
                        'disabled-color': 'rgba(0, 0, 0, 0.25)',
                        'border-color-base': '#e0e0e0',
                        'border-color-split': '#f0f0f0',
                        'background-color-base': '#f5f5f5',
                        'box-shadow-base': '0 1px 4px rgba(0, 0, 0, 0.06)',
                        'btn-default-border': '#e0e0e0',
                        'btn-default-color': 'rgba(0, 0, 0, 0.75)',
                        'btn-border-radius-base': '6px',
                        'card-radius': '8px',
                        'card-shadow': '0 1px 4px rgba(0, 0, 0, 0.04)',
                        'modal-border-radius': '10px',
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
