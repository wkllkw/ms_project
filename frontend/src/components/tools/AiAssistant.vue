<template>
    <div class="ai-assistant">
        <!-- 浮动按钮 -->
        <div class="ai-fab" :class="{'is-open': isOpen}" @click="toggleChat" v-show="!isOpen">
            <a-icon type="robot" />
        </div>
        <!-- 聊天窗口 -->
        <div class="ai-chat-window" v-show="isOpen">
            <div class="ai-chat-header">
                <span class="ai-chat-title">
                    <a-icon type="robot" style="margin-right:6px" />
                    AI 项目助手
                </span>
                <a-icon type="close" class="ai-chat-close" @click="toggleChat" />
            </div>
            <div class="ai-chat-messages" ref="messagesContainer">
                <div v-if="messages.length === 0" class="ai-chat-welcome">
                    <a-icon type="robot" style="font-size:36px;color:#1890ff;margin-bottom:12px" />
                    <p style="font-size:15px;font-weight:500;color:#333;margin-bottom:4px">你好！我是 AI 项目助手</p>
                    <p style="color:#8c8c8c">我可以帮你管理项目和任务，试试下面的快捷指令：</p>
                    <div class="ai-quick-actions">
                        <a-button size="small" @click="quickAction('查看我的项目')">查看我的项目</a-button>
                        <a-button size="small" @click="quickAction('查看我的任务')">查看我的任务</a-button>
                        <a-button size="small" @click="quickAction('查看项目统计')">查看统计</a-button>
                        <a-button size="small" @click="quickAction('创建一个新项目')">创建项目</a-button>
                    </div>
                </div>
                <div v-for="(msg, idx) in messages" :key="'msg-'+idx"
                     class="ai-chat-message" :class="'ai-chat-message--' + msg.role">
                    <div class="ai-chat-avatar">
                        <a-icon :type="msg.role === 'user' ? 'user' : 'robot'" />
                    </div>
                    <div class="ai-chat-bubble">
                        <template v-if="msg.loading">
                            <a-spin size="small" /> <span class="ai-loading-text">AI 正在思考中...</span>
                        </template>
                        <template v-else>
                            <div class="ai-message-content" v-html="renderMarkdown(msg.content)"></div>
                            <!-- 重试按钮：仅错误消息显示 -->
                            <a-icon v-if="msg.error" type="reload"
                                    class="ai-retry-btn" title="重试"
                                    @click="retryMessage(idx)" />
                        </template>
                    </div>
                </div>
            </div>
            <div class="ai-chat-input">
                <a-input-search
                    v-model="inputText"
                    placeholder="输入消息，如：查看我的项目..."
                    enter-button="发送"
                    :loading="sending"
                    :disabled="sending"
                    @search="sendMessage"
                    @pressEnter.native="sendMessage"
                />
            </div>
        </div>
    </div>
</template>

<script>
import {message} from 'ant-design-vue';
import {chat} from '@/api/assistant';

export default {
    name: 'AiAssistant',
    data() {
        return {
            isOpen: false,
            inputText: '',
            sending: false,
            messages: [],
        };
    },
    methods: {
        toggleChat() {
            this.isOpen = !this.isOpen;
            if (this.isOpen) {
                this.$nextTick(() => {
                    this.scrollToBottom();
                });
            }
        },
        quickAction(text) {
            this.inputText = text;
            this.sendMessage();
        },
        async sendMessage() {
            const text = this.inputText.trim();
            if (!text || this.sending) return;

            this.messages.push({role: 'user', content: text});
            this.inputText = '';
            this.sending = true;

            // 添加 AI 回复占位
            const aiMsg = {role: 'assistant', content: '', loading: true, error: false};
            const aiIndex = this.messages.push(aiMsg) - 1;
            this.scrollToBottom();

            try {
                // 构建对话历史（排除正在加载的占位消息）
                const chatMessages = this.messages
                    .slice(0, aiIndex)
                    .filter(m => m.role && m.content)
                    .map(m => ({role: m.role, content: m.content}));

                console.log('[AI Assistant] 发送消息:', chatMessages);

                const res = await chat({messages: chatMessages});
                console.log('[AI Assistant] 收到响应:', res);

                if (res && res.code === 200 && res.data) {
                    const choices = res.data.choices;
                    if (choices && choices.length > 0 && choices[0].message) {
                        aiMsg.content = choices[0].message.content || '（空回复）';
                    } else {
                        aiMsg.content = 'AI 暂无回复';
                        aiMsg.error = true;
                    }
                } else {
                    aiMsg.content = (res && res.msg) || '请求失败';
                    aiMsg.error = true;
                }
            } catch (e) {
                console.error('[AI Assistant] 异常:', e);
                aiMsg.content = '网络错误，请稍后重试';
                aiMsg.error = true;
            } finally {
                aiMsg.loading = false;
                this.sending = false;
                this.$set(this.messages, aiIndex, aiMsg);
                this.scrollToBottom();
            }
        },
        retryMessage(idx) {
            if (this.sending || idx < 1) return;
            // 找到对应的用户消息并重新发送
            const userMsg = this.messages[idx - 1];
            if (!userMsg || userMsg.role !== 'user') return;

            // 删除旧的 AI 错误回复
            this.messages.splice(idx, 1);
            this.inputText = userMsg.content;
            this.sendMessage();
        },
        scrollToBottom() {
            this.$nextTick(() => {
                const container = this.$refs.messagesContainer;
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            });
        },
        renderMarkdown(content) {
            if (!content) return '';
            let html = content
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;');

            // 简单 Markdown 渲染
            // 粗体 **text**
            html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
            // 行内代码 `code`
            html = html.replace(/`([^`]+)`/g, '<code class="ai-inline-code">$1</code>');
            // 标题 ### text
            html = html.replace(/^### (.+)$/gm, '<div class="ai-md-h3">$1</div>');
            html = html.replace(/^## (.+)$/gm, '<div class="ai-md-h2">$1</div>');
            // 无序列表 - item
            html = html.replace(/^- (.+)$/gm, '<div class="ai-md-li">• $1</div>');
            // 有序列表 1. item
            html = html.replace(/^\d+\. (.+)$/gm, (match, p1, offset, str) => {
                return '<div class="ai-md-li">' + match.split('.')[0] + '. ' + p1 + '</div>';
            });
            // 换行
            html = html.replace(/\n/g, '<br>');
            // 清理多余 <br>
            html = html.replace(/<br><div class="ai-md-/g, '<div class="ai-md-');

            return html;
        },
    },
};
</script>

<style lang="less" scoped>
.ai-assistant {
    position: fixed;
    z-index: 9999;
    bottom: 24px;
    right: 24px;
}

.ai-fab {
    width: 52px;
    height: 52px;
    border-radius: 50%;
    background: linear-gradient(135deg, #1890ff, #36cfc9);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    box-shadow: 0 4px 16px rgba(24, 144, 255, 0.45);
    font-size: 24px;
    transition: transform 0.2s, box-shadow 0.2s;

    &:hover {
        transform: scale(1.1);
        box-shadow: 0 6px 20px rgba(24, 144, 255, 0.55);
    }

    &.is-open {
        opacity: 0;
        pointer-events: none;
    }
}

.ai-chat-window {
    width: 420px;
    height: 560px;
    background: #fff;
    border-radius: 16px;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.2);
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.ai-chat-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 18px;
    background: linear-gradient(135deg, #1890ff, #36cfc9);
    color: #fff;
    font-size: 15px;
    font-weight: 600;
}

.ai-chat-close {
    cursor: pointer;
    font-size: 16px;
    opacity: 0.8;
    transition: opacity 0.2s;

    &:hover {
        opacity: 1;
    }
}

.ai-chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    background: #f7f8fa;
}

.ai-chat-welcome {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #8c8c8c;
    font-size: 13px;
    text-align: center;
    padding: 20px;
}

.ai-quick-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 16px;
    justify-content: center;

    /deep/ .ant-btn {
        border-radius: 16px;
        font-size: 12px;
        border-color: #1890ff;
        color: #1890ff;

        &:hover {
            background: #e6f7ff;
            border-color: #40a9ff;
            color: #40a9ff;
        }
    }
}

.ai-chat-message {
    display: flex;
    margin-bottom: 14px;

    &--user {
        flex-direction: row-reverse;
    }
}

.ai-chat-avatar {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 15px;
    flex-shrink: 0;

    .ai-chat-message--user & {
        background: #e6f7ff;
        color: #1890ff;
        margin-left: 8px;
    }

    .ai-chat-message--assistant & {
        background: #f0f0f0;
        color: #666;
        margin-right: 8px;
    }
}

.ai-chat-bubble {
    max-width: 300px;
    padding: 10px 14px;
    border-radius: 12px;
    font-size: 13px;
    line-height: 1.7;
    word-break: break-word;
    position: relative;

    .ai-chat-message--user & {
        background: #1890ff;
        color: #fff;
        border-top-right-radius: 4px;
    }

    .ai-chat-message--assistant & {
        background: #fff;
        color: #333;
        border-top-left-radius: 4px;
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
    }
}

.ai-message-content {
    /deep/ .ai-md-h2 {
        font-size: 14px;
        font-weight: 600;
        margin: 8px 0 4px 0;
        color: #1890ff;
    }

    /deep/ .ai-md-h3 {
        font-size: 13px;
        font-weight: 600;
        margin: 6px 0 2px 0;
        color: #333;
    }

    /deep/ .ai-md-li {
        padding-left: 4px;
        line-height: 1.8;
    }

    /deep/ .ai-inline-code {
        background: #f0f0f0;
        padding: 1px 5px;
        border-radius: 3px;
        font-size: 12px;
        color: #d4380d;
    }
}

.ai-loading-text {
    color: #999;
    font-size: 12px;
    margin-left: 4px;
}

.ai-retry-btn {
    margin-left: 6px;
    cursor: pointer;
    color: #ff4d4f;
    font-size: 12px;
    vertical-align: middle;
    transition: transform 0.2s;

    &:hover {
        transform: rotate(180deg);
    }
}

.ai-chat-input {
    padding: 12px 16px;
    border-top: 1px solid #f0f0f0;
    background: #fff;

    /deep/ .ant-input-search-button {
        background: #1890ff;
        border-color: #1890ff;
    }
}
</style>
