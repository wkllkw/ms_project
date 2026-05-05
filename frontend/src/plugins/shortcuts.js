// 全局快捷键注册表
// 使用方式：this.$shortcuts.register('ctrl+k', handler)
//          this.$shortcuts.unregister('ctrl+k')

const ShortcutManager = {
    _shortcuts: {},
    _enabled: true,

    install(Vue) {
        Vue.prototype.$shortcuts = this;
        document.addEventListener('keydown', this._handleKeydown.bind(this));
    },

    /**
     * 注册快捷键
     * @param {string} key - 快捷键组合，如 'ctrl+k', 'g+h', 't'
     * @param {Function} handler - 处理函数
     * @param {Object} options - { prevent: true, description: '' }
     */
    register(key, handler, options = {}) {
        const normalized = this._normalize(key);
        this._shortcuts[normalized] = {
            handler,
            prevent: options.prevent !== false,
            description: options.description || '',
            key: normalized,
        };
    },

    /**
     * 取消注册
     */
    unregister(key) {
        const normalized = this._normalize(key);
        delete this._shortcuts[normalized];
    },

    /**
     * 临时禁用所有快捷键（如模态框打开时）
     */
    disable() {
        this._enabled = false;
    },

    /**
     * 启用快捷键
     */
    enable() {
        this._enabled = true;
    },

    /**
     * 获取所有已注册的快捷键
     */
    getAll() {
        return Object.values(this._shortcuts);
    },

    _handleKeydown(e) {
        if (!this._enabled) return;
        // 忽略输入框中的快捷键（除了含 ctrl/meta 的）
        const target = e.target;
        const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;
        if (isInput && !e.ctrlKey && !e.metaKey) return;

        const parts = [];
        if (e.ctrlKey || e.metaKey) parts.push('ctrl');
        if (e.altKey) parts.push('alt');
        if (e.shiftKey) parts.push('shift');

        const key = e.key.toLowerCase();
        if (!['control', 'alt', 'shift', 'meta'].includes(key)) {
            parts.push(key);
        }

        const combo = parts.join('+');
        const shortcut = this._shortcuts[combo];
        if (shortcut) {
            if (shortcut.prevent) e.preventDefault();
            shortcut.handler(e);
        }
    },

    _normalize(key) {
        return key.toLowerCase().split('+').map(k => k.trim()).sort().join('+');
    },
};

export default ShortcutManager;
