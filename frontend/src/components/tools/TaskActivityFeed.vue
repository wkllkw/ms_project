<template>
    <div class="task-activity-feed">
        <div class="feed-filter">
            <a-radio-group v-model="filterType" size="small" button-style="solid">
                <a-radio-button value="all">全部</a-radio-button>
                <a-radio-button value="comment">评论</a-radio-button>
                <a-radio-button value="log">动态</a-radio-button>
            </a-radio-group>
        </div>
        <div class="feed-timeline" v-if="filteredList.length > 0">
            <div
                v-for="item in filteredList"
                :key="item.code"
                class="feed-item"
                :class="{ 'is-comment': item.is_comment, 'is-log': !item.is_comment }"
            >
                <!-- 活动类型指示线 -->
                <div class="feed-indicator">
                    <div class="indicator-dot" :class="getItemClass(item)"/>
                    <div class="indicator-line"/>
                </div>

                <div class="feed-content">
                    <!-- 头部：头像 + 名字 + 时间 -->
                    <div class="feed-header">
                        <a-avatar :size="28" :src="item.member ? item.member.avatar : ''" icon="user"/>
                        <span class="feed-author">{{ item.member ? item.member.name : '未知' }}</span>
                        <span class="feed-time">{{ formatTime(item.create_time) }}</span>
                        <a-tag v-if="item.is_comment" color="blue" style="margin-left: 6px; font-size: 11px;">评论</a-tag>
                    </div>

                    <!-- 动态内容 -->
                    <div class="feed-body" v-if="!item.is_comment">
                        <a-icon :type="item.icon || 'file-text'" class="feed-icon"/>
                        <span class="feed-remark" v-html="item.remark"/>
                        <div class="feed-detail img-preview-content" v-if="item.content" v-html="item.content"/>
                    </div>

                    <!-- 评论内容 -->
                    <div class="feed-body comment-body" v-if="item.is_comment">
                        <div class="comment-content img-preview-content" v-html="checkLink(item.content)"/>
                    </div>
                </div>
            </div>
        </div>
        <div class="feed-empty" v-else>
            <a-empty :image="simpleImage" description="暂无动态"/>
        </div>
        <div class="feed-more" v-if="hasMore">
            <a @click="$emit('load-more')">加载更多动态</a>
        </div>
    </div>
</template>

<script>
import { Empty } from 'ant-design-vue';
import { relativelyTime } from '@/assets/js/dateTime';

export default {
    name: 'TaskActivityFeed',
    props: {
        items: { type: Array, default: () => [] },
        hasMore: { type: Boolean, default: false },
    },
    data() {
        return {
            filterType: 'all',
            simpleImage: Empty.PRESENTED_IMAGE_SIMPLE,
        };
    },
    computed: {
        filteredList() {
            if (this.filterType === 'all') return this.items;
            if (this.filterType === 'comment') return this.items.filter(i => i.is_comment);
            return this.items.filter(i => !i.is_comment);
        },
    },
    methods: {
        formatTime(time) {
            return relativelyTime(time);
        },
        getItemClass(item) {
            if (item.is_comment) return 'dot-comment';
            const icon = item.icon || '';
            if (icon.includes('check') || icon.includes('done')) return 'dot-success';
            if (icon.includes('edit') || icon.includes('form')) return 'dot-edit';
            if (icon.includes('delete') || icon.includes('close')) return 'dot-danger';
            return 'dot-default';
        },
        checkLink(content) {
            if (!content) return '';
            return content.replace(/(https?:\/\/[^\s<]+)/g, '<a href="$1" target="_blank" style="color: #1890ff;">$1</a>');
        },
    },
};
</script>

<style lang="less">
.task-activity-feed {
    .feed-filter {
        margin-bottom: 12px;
    }

    .feed-timeline {
        position: relative;
    }

    .feed-item {
        display: flex;
        gap: 12px;
        position: relative;
        padding-bottom: 16px;

        &:last-child {
            .indicator-line {
                display: none;
            }
        }
    }

    .feed-indicator {
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 12px;
        flex-shrink: 0;
        padding-top: 6px;

        .indicator-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            flex-shrink: 0;

            &.dot-comment { background: #1890ff; }
            &.dot-success { background: #52c41a; }
            &.dot-edit { background: #faad14; }
            &.dot-danger { background: #f5222d; }
            &.dot-default { background: #d9d9d9; }
        }

        .indicator-line {
            width: 1px;
            flex: 1;
            background: #f0f0f0;
            margin-top: 4px;
        }
    }

    .feed-content {
        flex: 1;
        min-width: 0;
    }

    .feed-header {
        display: flex;
        align-items: center;
        gap: 6px;
        margin-bottom: 4px;

        .feed-author {
            font-size: 13px;
            font-weight: 500;
            color: rgba(0, 0, 0, 0.85);
        }

        .feed-time {
            font-size: 12px;
            color: rgba(0, 0, 0, 0.35);
        }
    }

    .feed-body {
        padding-left: 2px;
        font-size: 13px;
        color: rgba(0, 0, 0, 0.65);
        line-height: 1.6;

        .feed-icon {
            margin-right: 4px;
            color: rgba(0, 0, 0, 0.35);
        }

        .feed-remark {
            color: rgba(0, 0, 0, 0.65);
        }

        .feed-detail {
            margin-top: 6px;
            padding: 8px 10px;
            background: #fafafa;
            border-radius: 4px;
            font-size: 12px;
            border: 1px solid #f0f0f0;
        }

        &.comment-body {
            .comment-content {
                padding: 8px 10px;
                background: #f0f5ff;
                border-radius: 6px;
                border: 1px solid #d6e4ff;
                word-break: break-word;
            }
        }
    }

    .feed-more {
        text-align: center;
        padding: 12px 0;
    }

    .feed-empty {
        padding: 32px 0;
    }
}
</style>
