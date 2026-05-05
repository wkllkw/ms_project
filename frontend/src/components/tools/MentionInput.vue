<template>
    <div class="mention-input-wrapper">
        <a-popover
            trigger="click"
            placement="topLeft"
            :visible="showPopover"
            :getPopupContainer="getPopupContainer"
            overlayClassName="mention-popover"
        >
            <template slot="content">
                <div class="mention-dropdown">
                    <div class="mention-search">
                        <a-input
                            ref="mentionSearchInput"
                            v-model="mentionQuery"
                            size="small"
                            placeholder="搜索成员..."
                            @keydown.down.prevent="navigateMemberDown"
                            @keydown.up.prevent="navigateMemberUp"
                            @keydown.enter.prevent="selectNavigatedMember"
                            @keydown.esc.prevent="showPopover = false"
                        />
                    </div>
                    <div class="mention-list">
                        <div
                            v-for="(member, index) in filteredMembers"
                            :key="member.code"
                            class="mention-item"
                            :class="{ active: index === activeMemberIndex }"
                            @click="selectMember(member)"
                            @mouseenter="activeMemberIndex = index"
                        >
                            <a-avatar :size="24" :src="member.avatar" icon="user"/>
                            <span class="member-name">{{ member.name }}</span>
                            <span class="member-email muted" v-if="member.email">{{ member.email }}</span>
                        </div>
                        <div class="mention-empty" v-if="filteredMembers.length === 0">
                            未找到成员
                        </div>
                    </div>
                </div>
            </template>
            <a-textarea
                ref="textarea"
                v-model="innerValue"
                :rows="rows"
                :placeholder="placeholder"
                @focus="$emit('focus')"
                @blur="handleBlur"
                @input="handleInput"
                @keydown="handleKeydown"
            />
        </a-popover>
        <div class="mention-actions" v-if="showActions">
            <span class="mention-hint">
                <kbd>@</kbd> 提及成员 · <kbd>Ctrl+Enter</kbd> 发送
            </span>
            <a-button type="primary" size="small" @click="$emit('submit')">
                {{ submitText }}
            </a-button>
        </div>
    </div>
</template>

<script>
export default {
    name: 'MentionInput',
    props: {
        value: { type: String, default: '' },
        members: { type: Array, default: () => [] },
        rows: { type: Number, default: 2 },
        placeholder: { type: String, default: '@提及成员，Ctrl+Enter发送' },
        submitText: { type: String, default: '评论' },
        showActions: { type: Boolean, default: true },
    },
    data() {
        return {
            innerValue: this.value,
            showPopover: false,
            mentionQuery: '',
            activeMemberIndex: 0,
            cursorPosition: 0,
        };
    },
    computed: {
        filteredMembers() {
            if (!this.mentionQuery) return this.members.slice(0, 10);
            const q = this.mentionQuery.toLowerCase();
            return this.members.filter(m =>
                (m.name && m.name.toLowerCase().includes(q)) ||
                (m.email && m.email.toLowerCase().includes(q))
            ).slice(0, 10);
        },
    },
    watch: {
        value(val) {
            this.innerValue = val;
        },
    },
    methods: {
        getPopupContainer() {
            return this.$el;
        },
        handleInput() {
            this.$emit('input', this.innerValue);
            this.detectMention();
        },
        handleKeydown(e) {
            if (e.key === '@') {
                this.showPopover = true;
                this.mentionQuery = '';
                this.activeMemberIndex = 0;
                this.$nextTick(() => {
                    if (this.$refs.mentionSearchInput) {
                        this.$refs.mentionSearchInput.focus();
                    }
                });
            }
            if (e.ctrlKey && e.key === 'Enter') {
                e.preventDefault();
                this.$emit('submit');
            }
        },
        handleBlur() {
            this.$emit('blur');
            // 延迟关闭，允许点击下拉项
            setTimeout(() => {
                this.showPopover = false;
            }, 200);
        },
        detectMention() {
            const textarea = this.$refs.textarea;
            if (!textarea) return;
            const pos = textarea.$el ? textarea.$el.selectionStart : 0;
            const text = this.innerValue.substring(0, pos);
            const atIndex = text.lastIndexOf('@');
            if (atIndex >= 0 && (atIndex === 0 || text[atIndex - 1] === ' ' || text[atIndex - 1] === '\n')) {
                this.mentionQuery = text.substring(atIndex + 1);
                this.showPopover = true;
                this.cursorPosition = atIndex;
            } else {
                this.showPopover = false;
            }
        },
        selectMember(member) {
            const before = this.innerValue.substring(0, this.cursorPosition);
            const after = this.innerValue.substring(this.cursorPosition + this.mentionQuery.length + 1);
            this.innerValue = before + '@' + member.name + ' ' + after;
            this.$emit('input', this.innerValue);
            this.showPopover = false;
            this.mentionQuery = '';
            this.$nextTick(() => {
                const textarea = this.$refs.textarea;
                if (textarea && textarea.$el) {
                    textarea.$el.focus();
                }
            });
        },
        navigateMemberUp() {
            if (this.activeMemberIndex > 0) this.activeMemberIndex--;
        },
        navigateMemberDown() {
            if (this.activeMemberIndex < this.filteredMembers.length - 1) this.activeMemberIndex++;
        },
        selectNavigatedMember() {
            const member = this.filteredMembers[this.activeMemberIndex];
            if (member) this.selectMember(member);
        },
        focus() {
            const textarea = this.$refs.textarea;
            if (textarea && textarea.$el) textarea.$el.focus();
        },
    },
};
</script>

<style lang="less">
.mention-popover {
    .ant-popover-inner {
        border-radius: 8px;
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
    }
}

.mention-input-wrapper {
    position: relative;

    .mention-actions {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-top: 6px;

        .mention-hint {
            font-size: 12px;
            color: rgba(0, 0, 0, 0.35);

            kbd {
                display: inline-block;
                padding: 0 4px;
                font-size: 10px;
                font-family: inherit;
                border: 1px solid #d9d9d9;
                border-radius: 3px;
                background: #fafafa;
                color: rgba(0, 0, 0, 0.45);
            }
        }
    }
}

.mention-dropdown {
    width: 220px;

    .mention-search {
        padding: 4px 4px 8px;
    }

    .mention-list {
        max-height: 200px;
        overflow-y: auto;

        .mention-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 6px 8px;
            border-radius: 4px;
            cursor: pointer;
            transition: background-color 0.1s;

            &:hover, &.active {
                background: #f5f5f5;
            }

            .member-name {
                font-size: 13px;
                color: rgba(0, 0, 0, 0.85);
            }

            .member-email {
                font-size: 11px;
                margin-left: auto;
            }
        }

        .mention-empty {
            padding: 16px;
            text-align: center;
            color: rgba(0, 0, 0, 0.25);
            font-size: 13px;
        }
    }
}
</style>
