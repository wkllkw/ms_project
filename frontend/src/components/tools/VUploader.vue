<template>
    <div class="v-uploader-wrapper" v-show="showUploader">
        <div class="v-uploader-mask" @click="closeUploader"></div>
        <div class="v-uploader">
            <a-card :title="uploaderTitle">
            <div class="actions" slot="extra">
                <a class="muted action-item" @click="showFiles = !showFiles">
                    <a-icon type="shrink" v-show="showFiles"/>
                    <a-icon type="arrows-alt" v-show="!showFiles"/>
                </a>
                <a class="muted action-item" @click="closeUploader">
                    <a-icon type="close"/>
                </a>
            </div>
            <uploader ref="uploader"
                      :options="options"
                      :autoStart="autoStart"
                      @files-added="fileAdded"
                      @files-submitted="filesSubmitted"
                      @file-progress="fileProgress"
                      @file-success="fileSuccess"
                      @file-error="fileError"
                      @file-complete="fileComplete"
                      @complete="complete"
                      class="uploader-workplace">
                <vue-scroll>
                    <!--<a-button @click="testSomeThing">测试</a-button>-->
                    <!--<uploader-unsupport></uploader-unsupport>-->
                    <!--<uploader-btn>select files</uploader-btn>-->
                    <!--<uploader-drop>
                        <p>Drop files here to upload or</p>
                        <uploader-btn>select files</uploader-btn>
                        <uploader-btn :attrs="attrs">select images</uploader-btn>
                        <uploader-btn :directory="true">select folder</uploader-btn>
                    </uploader-drop>-->
                    <uploader-list>
                        <template slot-scope="files">
                            <ul class="uploader-wrapper">
                                <uploader-file :key="file.id" :file="file" :list="true"
                                               v-for="file in files.fileList">
                                    <template slot-scope="file">
                                        <li class="uploader-item">
                                            <div class="item-content">
                                                <div class="item-info">
                                                    <div class="file-item">
                                                        <div class="file-icon">
                                                            <a-avatar icon="link" shape="square"
                                                                      :src="file.file.fileUrl"/>
                                                        </div>
                                                        <div class="file-info">
                                                            <div class="file-content">
                                                                <div class="file-title">
                                                                    <a-tooltip placement="top" :mouseEnterDelay="0.3"
                                                                               :title="file.file.name">
                                                                        {{file.file.name}}
                                                                    </a-tooltip>
                                                                </div>
                                                                <div class="file-extra">
                                                                    <span>{{file.file.projectName ? file.file.projectName : tempData.projectName}}</span>
                                                                    <span v-if="file.status == 'success'">({{file.formatedSize }})</span>
                                                                    <span v-else>({{(Number(file.uploadedSize) / (1024 * 1000)).toFixed(2) }}MB/{{file.formatedSize }})</span>
                                                                </div>
                                                            </div>
                                                            <div class="uploader-progress"
                                                                 v-show="file.status != 'success'">
                                                                <a-progress :strokeWidth="2" :showInfo="false"
                                                                            :percent="file.progress * 100"/>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                                <div class="item-status">
                                                    <a class="muted" @click="cancelUpload(file)">
                                                        <a-icon type="close" v-show="file.status != 'success'"/>
                                                    </a>
                                                    <a-icon class="text-success" type="check"
                                                            v-show="file.status == 'success'"/>
                                                </div>
                                            </div>
                                        </li>
                                    </template>
                                </uploader-file>
                            </ul>
                        </template>
                    </uploader-list>
                    <!--<uploader-files>
                    </uploader-files>-->
                </vue-scroll>
            </uploader>
            </a-card>
        </div>
    </div>
</template>

<script>
    import {checkResponse, getApiUrl, getAuthorization} from "../../assets/js/utils";
    import {mapState} from 'vuex'
    import {getStore} from "../../assets/js/storage";
    import {notice} from "../../assets/js/notice";
    import {uploadFiles} from "../../api/file";

    export default {
        name: "v-uploader",
        props: {
            code: {
                type: [String],
                default() {
                    return ''
                }
            },
        },
        data() {
            return {
                loading: false,
                singleMaxSize: 30,//单个文件最大大小，MB
                showFiles: false, //显示上传文件
                showUploader: false,//显示上传窗口
                progressTotal: 0, //上传中的文件数
                completeTotal: 0, //已完成的文件数
                failedTotal: 0, //上传失败的文件数
                lastErrorMsg: '', //最后一次上传失败的错误消息
                options: {
                    target: uploadFiles,
                    testChunks: false,
                    // 后端不支持分片上传，设置极大值禁用分片
                    chunkSize: 500 * 1024 * 1024,
                    simultaneousUploads: 1,
                    query: function () {
                        return getStore('tempData', true);//query暂时无法动态响应
                    },
                    headers: function () {
                        const auth = getAuthorization();
                        const organization = getStore('currentOrganization', true);
                        if (organization && organization.code) {
                            auth.organizationCode = organization.code;
                        }
                        return auth;
                    },
                },
                attrs: {
                    // 不限制文件格式，后端支持图片/文档/视频/音频/代码等所有类型
                },
                autoStart: true,
            }
        },
        computed: {
            ...mapState({
                uploader: state => state.common.uploader,
                tempData: state => state.common.tempData,
            }),
            uploaderTitle() {
                if (!this.progressTotal) {
                    return '上传完成';
                }
                let current = this.completeTotal + 1;
                if (current > this.progressTotal) {
                    current = this.progressTotal;
                }
                return `正在上传 ${current}/${this.progressTotal}`;
            }
        },
        watch: {
            code() {
                this.init();
            },
        },
        created() {
            this.init();
        },
        mounted() {
            this.$nextTick(() => {
                window.uploader = this.$refs.uploader.uploader;
                this.$store.dispatch('setUploader', window.uploader);
            })
        },
        methods: {
            init() {

            },
            closeUploader() {//关闭上传窗口
                this.showUploader = false;
                this.uploader.cancel();
            },
            fileAdded(files, fileList, event) {
                let ignored = false;
                let fileName = '';
                const singleMaxSize = this.singleMaxSize * 1024 * 1024;
                fileList.forEach((v, k) => {
                    if (v.size > singleMaxSize) {
                        ignored = true;
                        fileName = v.file.name;
                        return false;
                    }
                });
                files.ignored = ignored;
                if (ignored) {
                    this.$info({
                        title: '文件超过最大限制',
                        content: `上传文件「${fileName}」过大，请选择${this.singleMaxSize}MB以内的文件`,
                    });
                }
            },
            filesSubmitted(files) { //添加上传文件
                // this.$refs.uploader.uploader.opts.query = this.tempData;
                this.showUploader = true;
                this.showFiles = true;
                this.progressTotal += files.length;

            },
            fileProgress(rootFile, file, chunk) { //有文件上传中
                this.showUploader = true;
                this.showFiles = true;
            },
            fileSuccess(rootFile, file, message, chunk) { //一个文件上传成功
                let response;
                try {
                    response = JSON.parse(message);
                } catch (e) {
                    this.failedTotal++;
                    this.lastErrorMsg = '上传响应解析失败: ' + String(message).substring(0, 100);
                    notice({title: this.lastErrorMsg}, 'notice', 'error');
                    return false;
                }
                if (!checkResponse(response)) {
                    this.lastErrorMsg = response.msg || '上传失败';
                    notice({title: this.lastErrorMsg}, 'notice', 'error');
                    this.failedTotal++;
                    return false;
                }
                rootFile.projectName = response.data.projectName || response.data.title || '';
                rootFile.fileUrl = response.data.fileUrl || response.data.url;
                this.completeTotal++;
            },
            fileError(rootFile, file, message, chunk) { //一个文件上传失败
                this.progressTotal--;
                this.failedTotal++;
                let errorMsg = '网络请求失败';
                try {
                    const response = JSON.parse(message);
                    errorMsg = response.msg || '网络请求失败';
                } catch (e) {
                    // message 可能不是 JSON（如网络错误）
                    if (message) {
                        errorMsg = '网络错误: ' + String(message).substring(0, 80);
                    }
                }
                this.lastErrorMsg = errorMsg;
                file.cancel();
                notice({title: errorMsg}, 'notice', 'error');
            },
            fileComplete(rootFile) { //一个文件上传完成
            },
            complete() { //所有文件上传完成
                // 只有真正有文件上传成功时才提示成功并刷新列表
                if (this.completeTotal > 0) {
                    notice({title: '关联文件成功'}, 'notice', 'success');
                    // 通知文件页刷新
                    const tempData = getStore('tempData', true);
                    this.$root.$emit('fileUploadComplete', tempData);
                }
                if (this.failedTotal > 0) {
                    notice({
                        title: `${this.failedTotal}个文件上传失败`,
                        desc: this.lastErrorMsg || ''
                    }, 'notice', 'error', 8);
                }
                this.progressTotal = this.completeTotal = this.failedTotal = 0;
                this.lastErrorMsg = '';
                // 上传完成后立即隐藏上传面板，不延迟
                this.showFiles = false;
                this.showUploader = false;
            },
            cancelUpload(file) {
                this.progressTotal--;
                this.completeTotal--;
                file.file.cancel();
            },
            filterList(list) {
                return list;
                return list.reverse();
            },
            testSomeThing() {
                this.uploader.fileList[0].resume();
            },
        }
    }
</script>

<style lang="less">
    @import "~ant-design-vue/lib/style/themes/default";

    .v-uploader-wrapper {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        z-index: 1001;
        display: flex;
        align-items: center;
        justify-content: center;

        .v-uploader-mask {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.3);
        }

        .v-uploader {
            position: relative;
            z-index: 1;
            width: 500px;
            box-shadow: 0 7px 21px rgba(0, 0, 0, .15);
            border-radius: 6px;
            overflow: hidden;

            .ant-card {
                box-shadow: none;
                border-radius: 6px;

                .ant-card-head {
                    margin-bottom: 0;
                    border-bottom: 1px solid #e1e1e1;
                }

                .ant-card-head-title, .ant-card-extra {
                    padding: 12px 0;
                }

                .ant-card-body {
                    padding: 0;
                }
            }

            .actions {
                .action-item {
                    margin-left: 12px;
                    font-size: 16px;
                }
            }

            .uploader-workplace {
                height: 240px;
                background-color: #f7f7f7;
                padding: 12px 0 0 12px;

                .uploader-list {
                    padding-right: 12px;

                    .uploader-wrapper {
                        .uploader-file {
                            padding: 0;
                            height: auto;
                            line-height: 36px;
                            border-bottom: none;
                            border-radius: 4px;

                            .uploader-item {
                                width: 100%;
                                margin-bottom: 8px;
                                background: #eee;

                                .item-content {
                                    padding: 6px 12px;
                                    display: flex;
                                    justify-content: space-between;
                                    flex: 1;

                                    .item-info {
                                        flex: 1 1 auto;
                                        position: relative;
                                        overflow: hidden;
                                        text-overflow: ellipsis;
                                        white-space: nowrap;

                                        .file-item {
                                            display: flex;

                                            .file-icon {
                                                margin-right: 8px;
                                            }

                                            .file-info {
                                                width: 100%;
                                                min-width: 0;
                                                padding: 4px 0;
                                                line-height: 16px;
                                                display: flex;
                                                flex-direction: column;
                                                justify-content: center;

                                                .file-content {
                                                    width: 100%;
                                                    display: flex;
                                                    align-items: center;
                                                    justify-content: space-between;

                                                    .file-title {
                                                        max-width: 210px;
                                                        text-overflow: ellipsis;
                                                        overflow: hidden;
                                                        min-width: 0;
                                                    }

                                                    .file-extra {
                                                        display: flex;
                                                        align-items: center;
                                                        max-width: 200px;
                                                        margin-left: 16px;
                                                        color: gray;
                                                        font-size: 12px;

                                                        span {
                                                            margin-left: 3px;
                                                        }
                                                    }
                                                }
                                            }

                                            .uploader-progress {
                                                .ant-progress-outer {
                                                    margin: 0;
                                                }
                                            }
                                        }
                                    }

                                    .item-status {
                                        display: flex;
                                        justify-content: flex-end;
                                        align-items: center;
                                        width: 30px;
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
</style>
