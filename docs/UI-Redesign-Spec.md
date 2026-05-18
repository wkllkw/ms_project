# FlowHub 企业级 SaaS 后台 UI/UX 重构设计方案

> **版本**: v2.0 | **日期**: 2025-05-17 | **设计风格**: Notion × Linear × Vercel

---

## 一、设计体系总览

### 1.1 核心设计原则

| 原则 | 说明 |
|------|------|
| **克制的色彩** | 低饱和蓝主色，中性灰支撑层级，仅状态色保留饱和度 |
| **呼吸的留白** | 页面 24px 外边距，卡片 24px 内边距，模块间 16-20px 间距 |
| **弱化边界** | 用阴影和背景色区分层级，而非实线边框 |
| **圆角语言** | 统一 8-12px 圆角，告别 0-4px 的尖锐感 |
| **微交互** | 所有可交互元素有 hover/active 状态变化，带 0.15-0.25s transition |

### 1.2 对标参考

| 模块 | 参考 |
|------|------|
| 整体布局 | Vercel Dashboard + Linear |
| 左侧导航 | Notion Sidebar + 飞书后台 |
| 表格设计 | Linear Table + Datadog |
| 卡片/列表 | Notion Database + Vercel |
| 搜索筛选 | 飞书 Filter + Grafana |
| 通知中心 | Linear Inbox + Slack |

---

## 二、配色方案

### 2.1 主色调

```less
// ========== 品牌色 ==========
@primary-color:       #4f8cff;   // 主色：低饱和蓝（原 #3a82f8）
@primary-light:       #eef4ff;   // 浅蓝背景
@primary-hover:       #6ba1ff;   // hover 态
@primary-active:      #3d73e0;   // 按下态

// ========== 中性灰 ==========
@bg-page:             #f5f7fb;   // 页面底色（原 #f5f5f5）
@bg-sider:            #f9fafb;   // 侧边栏底色
@bg-card:             #ffffff;   // 卡片/内容区底色
@bg-hover:            #f7f8fa;   // hover 背景

@border-light:        #eef0f2;   // 弱边框（原 #f0f0f0/#c0c0c0）
@border-default:      #e4e7eb;   // 默认边框
@border-strong:       #d4d7dc;   // 强调边框

// ========== 文字色 ==========
@text-primary:        #1a1d23;   // 主文字
@text-secondary:      #5f6773;   // 次要文字（原 rgba(0,0,0,0.65)）
@text-tertiary:       #9ca3af;   // 辅助文字（原 rgba(0,0,0,0.45)）
@text-placeholder:    #bdc4cd;   // 占位文字
@text-inverse:        #ffffff;   // 反色文字

// ========== 语义色 ==========
@success-color:       #22c55e;   // 成功
@warning-color:       #f59e0b;   // 警告
@error-color:         #ef4444;   // 危险
@info-color:          #4f8cff;   // 信息

// ========== 阴影 ==========
@shadow-xs:           0 1px 2px rgba(0,0,0,0.04);
@shadow-sm:           0 1px 3px rgba(0,0,0,0.06);
@shadow-md:           0 4px 12px rgba(0,0,0,0.06);
@shadow-lg:           0 8px 24px rgba(0,0,0,0.08);
@shadow-xl:           0 16px 48px rgba(0,0,0,0.1);
@shadow-card:         0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.03);
@shadow-sider:        1px 0 4px rgba(0,0,0,0.03);
```

### 2.2 暗色模式色彩（扩展）

```less
// 暗色主题（按需切换 .theme-dark）
.theme-dark {
  --bg-page:            #0d1117;
  --bg-sider:           #161b22;
  --bg-card:            #1c2129;
  --bg-hover:           #21262d;
  --border-light:       #30363d;
  --border-default:     #3a3f47;
  --text-primary:       #e6edf3;
  --text-secondary:     #8b949e;
  --text-tertiary:      #6e7681;
}
```

---

## 三、字体层级

### 3.1 字体族

```css
font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI",
             "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei",
             "Helvetica Neue", Arial, sans-serif;
```

### 3.2 字号层级

| Token | 字号 | 行高 | 字重 | 用途 |
|-------|------|------|------|------|
| `text-2xs` | 11px | 16px | 400 | 辅助信息、时间戳 |
| `text-xs` | 12px | 18px | 400 | 标签、badge、表格次要列 |
| `text-sm` | 13px | 20px | 400 | 菜单项、表格内容、描述文字 |
| `text-base` | 14px | 22px | 400 | 正文、表单标签、输入框 |
| `text-md` | 15px | 24px | 500 | 卡片标题 |
| `text-lg` | 16px | 24px | 600 | 页面标题、Modal 标题 |
| `text-xl` | 18px | 28px | 600 | 板块标题 |
| `text-2xl` | 20px | 30px | 700 | 页面主标题 |

---

## 四、间距规范

| Token | 值 | 用途 |
|-------|-----|------|
| `spacing-2xs` | 4px | 极小间距（icon 与文字） |
| `spacing-xs` | 8px | 紧凑间距 |
| `spacing-sm` | 12px | 元素内部间距 |
| `spacing-md` | 16px | 模块内间距 |
| `spacing-lg` | 20px | 卡片内边距 |
| `spacing-xl` | 24px | 页面外边距、卡片间距 |
| `spacing-2xl` | 32px | 大块间距 |
| `spacing-3xl` | 40px | 页面顶部间距 |

### 圆角规范

| Token | 值 | 用途 |
|-------|-----|------|
| `radius-sm` | 6px | 按钮、输入框、标签 |
| `radius-md` | 8px | 卡片、面板 |
| `radius-lg` | 12px | Modal、Drawer、大区块 |

---

## 五、页面布局结构

```
┌──────────────────────────────────────────────────────────────┐
│  HEADER (h: 52px, bg: glass, shadow: none, border-bottom)    │
│  [Logo FlowHub]          [Online] [Notice] [Avatar ▼]        │
├────────┬─────────────────────────────────────────────────────┤
│        │                                                      │
│ SIDER  │  MAIN CONTENT (bg: #f5f7fb, padding: 24px)         │
│        │                                                      │
│ 220px  │  ┌──────────────────────────────────────────────┐   │
│ fixed  │  │  Page Header (optional, only with breadcrumb) │   │
│        │  │  breadcrumb + title + action                  │   │
│  menu  │  ├──────────────────────────────────────────────┤   │
│  groups│  │                                              │   │
│        │  │  CONTENT CARD (bg: #fff, r: 12px, shadow)    │   │
│  ───── │  │                                              │   │
│  user  │  │  ┌─ Search / Filter Area ──────────────────┐ │   │
│  card  │  │  │  [Keyword] [Date] [Status]  [Search]    │ │   │
│        │  │  └─────────────────────────────────────────┘ │   │
│        │  │                                              │   │
│        │  │  ┌─ Toolbar ───────────────────────────────┐ │   │
│        │  │  │  [Batch Read] [Batch Delete] [All Read] │ │   │
│        │  │  └─────────────────────────────────────────┘ │   │
│        │  │                                              │   │
│        │  │  ┌─ MODERN TABLE ──────────────────────────┐ │   │
│        │  │  │  header   │  body rows with hover       │ │   │
│        │  │  │  (52px)   │  (48px each)                │ │   │
│        │  │  │  #f9fafb  │  hover: #f7f8fa             │ │   │
│        │  │  └─────────────────────────────────────────┘ │   │
│        │  │                                              │   │
│        │  │  ┌─ Pagination ────────────────────────────┐ │   │
│        │  │  └─────────────────────────────────────────┘ │   │
│        │  └──────────────────────────────────────────────┘   │
│        │                                                      │
└────────┴──────────────────────────────────────────────────────┘
```

### 5.1 尺寸参数

| 元素 | 尺寸 | 说明 |
|------|------|------|
| Header 高度 | 52px（原 56px） | 更紧凑 |
| Sider 宽度 | 220px（原 240px） | 更精致 |
| Sider 收起 | 64px（原 72px） | 更窄 |
| 内容区 padding | 24px | 统一呼吸感 |
| 卡片圆角 | 12px | 柔和边界 |

---

## 六、逐模块详细设计

---

### 6.1 顶部栏 Header（重构重点）

```less
#layout .ant-layout-header {
  height: 52px;                          // 原 56px → 52px
  background: rgba(255,255,255,0.82);    // 毛玻璃
  backdrop-filter: saturate(180%) blur(10px);
  -webkit-backdrop-filter: saturate(180%) blur(10px);
  border-bottom: 1px solid @border-light; // 原 #c0c0c0 → #eef0f2
  box-shadow: none;                       // 去掉阴影，用边框替代
  z-index: 10;                            // 提升层级
}

// Logo 区
#layout .logo {
  width: 210px;                          // 对应 220px sider
  padding: 0 24px;
  .logo-img { width: 24px; height: 24px; }
  .title { font-size: 16px; font-weight: 600; }
  .version { color: @text-tertiary; font-size: 11px; }
}

// 右侧菜单
#layout .right-menu {
  .action {
    padding: 0 10px;
    border-radius: 8px;
    transition: background 0.15s ease;
    &:hover { background: @bg-hover; }
  }
}
```

### 6.2 左侧导航 Sider（重构重点）

```less
.ant-layout-sider {
  width: 220px !important;
  max-width: 220px !important;
  background: @bg-sider !important;
  border-right: 1px solid @border-light;
  box-shadow: @shadow-sider;

  // 菜单区域
  .sider-menu-section {
    padding: 12px 8px;

    // 一级分组标题
    > .ant-menu > .ant-menu-submenu > .ant-menu-submenu-title {
      font-size: 11px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      color: @text-tertiary;
      padding: 8px 16px !important;
      margin: 16px 8px 4px;
      height: auto;
      line-height: 1.4;

      .anticon { font-size: 14px; display: none; } // Lucide 风格：不显示图标
    }

    // 分组间不需要分隔线，用间距区分
    > .ant-menu > .ant-menu-submenu + .ant-menu-submenu {
      margin-top: 8px;
      &::before { display: none; }
    }

    // 菜单项
    .ant-menu-item {
      height: 38px;
      line-height: 38px;
      border-radius: 8px;
      margin: 2px 8px;
      padding: 0 12px !important;
      font-size: 13px;
      font-weight: 450;
      color: @text-secondary;
      transition: all 0.15s ease;

      // 去掉竖线指示器 → 改为背景色
      &::before { display: none !important; }

      &:hover {
        background: @bg-hover;
        color: @text-primary;
      }

      &.ant-menu-item-selected {
        background: @primary-light !important;
        color: @primary-color !important;
        font-weight: 500;

        .anticon { color: @primary-color; }
      }
    }

    // 子菜单标题
    .ant-menu-submenu-title {
      height: 38px;
      line-height: 38px;
      border-radius: 8px;
      margin: 2px 8px;
      padding: 0 12px !important;
      font-size: 13px;

      &:hover { background: @bg-hover; }
    }
  }

  // 底部用户卡片
  .sider-user-card {
    border-top: 1px solid @border-light;
    padding: 12px 16px;
    box-shadow: none;
    background: transparent;
  }

  // 折叠状态
  &.ant-layout-sider-collapsed {
    width: 64px !important;
    max-width: 64px !important;

    .sider-menu-section .ant-menu-item {
      margin: 2px 8px;
      border-radius: 8px;
      padding: 0 !important;
      text-align: center;
    }
  }

  // 折叠触发器
  .ant-layout-sider-trigger {
    width: 220px !important;
    border-top: 1px solid @border-light;
    color: @text-tertiary;
    background: @bg-sider;
    &:hover { background: @bg-hover; color: @primary-color; }
  }
}
```

### 6.3 内容区 WrapperContent

```less
// 页面底色
.ant-layout { background: @bg-page; }

// 卡片容器
.wrapper-main {
  margin: 24px;
  background: @bg-card;
  border-radius: 12px;
  padding: 24px;
  box-shadow: @shadow-card;
  border: 1px solid @border-light;
}

// 页面头
.page-header {
  background: @bg-card;
  padding: 20px 28px 0;
  border-bottom: 1px solid @border-light;

  .detail .main .title {
    font-size: 18px;
    font-weight: 600;
    color: @text-primary;
  }
}
```

### 6.4 Tabs 区域优化

```less
// 胶囊风格 Tabs
.notice-tabs {
  border-bottom: none;

  .ant-tabs-bar { border-bottom: none !important; }

  .ant-tabs-tab {
    padding: 6px 16px !important;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    color: @text-secondary;
    border: none !important;
    transition: all 0.15s ease;
    margin-right: 4px;

    &:hover {
      color: @text-primary;
      background: @bg-hover;
    }

    &.ant-tabs-tab-active {
      color: @primary-color;
      background: @primary-light;
      font-weight: 600;
    }
  }

  .ant-tabs-ink-bar { display: none !important; }
}
```

### 6.5 搜索/筛选区域

```less
.page-search {
  background: @bg-hover;
  border-radius: 10px;
  padding: 16px 20px;
  margin-bottom: 20px;
  border: 1px solid @border-light;

  .ant-form-item {
    margin-bottom: 0;

    &-label {
      font-size: 12px;
      color: @text-tertiary;
      font-weight: 500;
      line-height: 32px;
    }
  }

  .ant-input {
    height: 36px;
    border-radius: 8px;
    border-color: @border-default;
    background: @bg-card;
    &:focus { box-shadow: 0 0 0 3px rgba(79,140,255,0.08); }
  }

  .ant-btn-primary {
    height: 36px;
    border-radius: 8px;
    padding: 0 20px;
    font-weight: 500;
  }
}
```

### 6.6 现代 Table 设计（核心重点）

```less
// ========== Modern SaaS Table ==========
.modern-table {
  // 去除 Excel 感
  .ant-table {
    border: none;
    border-radius: 0;
    overflow: visible;

    // 表头
    &-thead > tr > th {
      background: @bg-hover;
      color: @text-tertiary;
      font-size: 11px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.3px;
      padding: 10px 16px;
      border-bottom: 1px solid @border-light;
      border-top: none;
      height: auto;
      line-height: 1.4;
    }

    // 表格行
    &-tbody > tr {
      transition: background 0.12s ease;

      > td {
        padding: 14px 16px;
        font-size: 13px;
        color: @text-secondary;
        border-bottom: 1px solid @border-light;
        line-height: 1.5;
        vertical-align: middle;
      }

      &:hover > td {
        background: @bg-hover !important;
      }
    }

    // 表格行高度 → 48px
    &-tbody > tr {
      height: 48px;
    }
  }
}

// 未读行高亮
.notice-unread {
  background: rgba(79,140,255,0.04) !important;
  > td:first-child {
    position: relative;
    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: @primary-color;
    }
  }
}

// 标题单元格
.notice-title-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 450;

  .unread-title { font-weight: 600; color: @text-primary; }
}

// 操作按钮改为 icon-only
.table-action-icon {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: @text-tertiary;
  transition: all 0.15s ease;
  cursor: pointer;
  font-size: 14px;

  &:hover {
    background: @bg-hover;
    color: @error-color;
  }
}

// 分页轻量化
.ant-pagination {
  margin-top: 20px;

  &-item {
    border-radius: 6px;
    border-color: @border-default;
    min-width: 32px;
    height: 32px;
    line-height: 32px;

    &-active { 
      background: @primary-light;
      border-color: @primary-light;
      a { color: @primary-color !important; font-weight: 600; }
      box-shadow: none;
    }
  }
}
```

### 6.7 通知中心

```less
// 通知详情弹窗
.notice-detail {
  .notice-detail-time {
    font-size: 12px;
    color: @text-tertiary;
    margin-bottom: 20px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .notice-detail-content {
    font-size: 14px;
    line-height: 1.8;
    color: @text-primary;
    background: @bg-hover;
    border-radius: 8px;
    padding: 16px;
  }
}
```

### 6.8 按钮操作区

```less
.action {
  padding: 8px 0 16px;
  
  .ant-btn {
    height: 34px;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 13px;
    font-weight: 500;
    
    &:not(.ant-btn-primary) {
      border-color: @border-default;
      color: @text-secondary;
      &:hover {
        border-color: @primary-color;
        color: @primary-color;
        background: @primary-light;
      }
    }
    
    &.ant-btn-danger {
      color: @error-color;
      border-color: transparent;
      &:hover { background: rgba(239,68,68,0.06); }
    }
  }
}
```

### 6.9 HeaderNotice 弹窗美化

```less
.header-notice {
  .ant-popover-inner {
    border-radius: 12px;
    box-shadow: @shadow-lg;
    overflow: hidden;
  }

  .ant-popover-inner-content {
    padding: 0;
  }

  .header-notice-content {
    .ant-tabs-bar {
      border-bottom: 1px solid @border-light;
      margin: 0;
      padding: 0 16px;
    }
  }
}
```

---

## 七、全局组件优化清单

### 7.1 已优化项（global-enhanced.less 中的基础上调）

| 组件 | 当前 | 目标 |
|------|------|------|
| 按钮圆角 | 6px | 8px |
| 卡片圆角 | 8px | 12px |
| 卡片阴影 | `0 1px 3px rgba(0,0,0,0.04)` | `0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.03)` |
| 输入框圆角 | 6px | 8px |
| Modal 圆角 | 10px | 12px |
| 表格圆角 | 8px | 移除圆角 |
| 表格边框 | `1px solid #f0f0f0` | 无边框，仅底部分隔线 |
| 分页圆角 | 6px | 6px ✓ |

### 7.2 新增组件样式

```less
// ========== 骨架屏 ==========
.skeleton-row {
  height: 14px;
  border-radius: 4px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 37%, #f0f0f0 63%);
  background-size: 400px 100%;
  animation: shimmer 1.4s ease infinite;
  margin-bottom: 8px;
}

// ========== 统计卡片 ==========
.stat-card {
  background: @bg-card;
  border-radius: 12px;
  padding: 20px 24px;
  border: 1px solid @border-light;
  box-shadow: @shadow-card;
  transition: all 0.2s ease;

  &:hover {
    border-color: rgba(79,140,255,0.25);
    box-shadow: @shadow-md;
  }

  .stat-label {
    font-size: 12px;
    color: @text-tertiary;
    font-weight: 500;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;
    color: @text-primary;
    margin-top: 4px;
  }
}

// ========== 空状态 ==========
.empty-state {
  text-align: center;
  padding: 60px 24px;

  .empty-icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 16px;
    background: @bg-hover;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    color: @text-tertiary;
  }

  .empty-title {
    font-size: 15px;
    font-weight: 500;
    color: @text-secondary;
    margin-bottom: 4px;
  }

  .empty-desc {
    font-size: 13px;
    color: @text-tertiary;
  }
}
```

---

## 八、动画与微交互

```less
// ========== Transition Tokens ==========
@transition-fast:    0.12s ease;
@transition-base:    0.15s ease;
@transition-slow:    0.25s ease;

// ========== 通用 hover 提升 ==========
.card-elevate {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  &:hover {
    transform: translateY(-2px);
    box-shadow: @shadow-lg;
  }
}

// ========== 页面过渡 ==========
.router-fade-enter-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.router-fade-leave-active { transition: opacity 0.12s ease; }
.router-fade-enter { opacity: 0; transform: translateY(3px); }
.router-fade-leave-to { opacity: 0; }

// ========== 导航菜单展开/收起 ==========
.sider-transition {
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
```

---

## 九、实施路线图

### Phase 1: 基础色系与变量（1-2天）
- [ ] 创建 `assets/css/variables-modern.less` 新变量文件
- [ ] 更新 `@primary-color` 和相关色值
- [ ] 引入字体族 Inter
- [ ] 更新 `.ant-layout` 背景色

### Phase 2: 布局框架（2-3天）
- [ ] 重构 `layout.less` — Header 毛玻璃 + Sider 现代风格
- [ ] 更新 `index.vue` 布局尺寸（52px/220px/64px）
- [ ] 重构 `warpperContent.less` — 12px 圆角 + 新阴影

### Phase 3: 组件美化（3-4天）
- [ ] 更新 `global-enhanced.less` — 按钮/输入框/卡片/Modal 圆角统一
- [ ] 重构表格样式 — `notice.vue` 表格作为样板
- [ ] Tabs 胶囊风格
- [ ] 搜索区域卡片化

### Phase 4: 页面级优化（3-5天）
- [ ] 通知中心 `notice.vue`
- [ ] HeaderNotice 弹窗
- [ ] 项目管理相关页面
- [ ] 日程管理页面
- [ ] 团队协作页面

### Phase 5: 暗色模式（2-3天）
- [ ] 暗色主题变量
- [ ] 主题切换逻辑
- [ ] 组件暗色适配

### Phase 6: 微交互 & 动效（1-2天）
- [ ] 页面切换过渡
- [ ] 列表入场动画
- [ ] Skeleton loading
- [ ] Hover 微反馈

---

## 十、关键代码实现示例

### 10.1 新增变量文件：`variables-modern.less`

```less
// ========== FlowHub Modern Theme v2.0 ==========
// 放在: /src/assets/css/variables-modern.less
// 在 vue.config.js 中通过 modifyVars 引入

// 覆盖 Ant Design 默认变量
@primary-color: #4f8cff;
@info-color: #4f8cff;
@success-color: #22c55e;
@warning-color: #f59e0b;
@error-color: #ef4444;

@text-color: #1a1d23;
@text-color-secondary: #5f6773;
@heading-color: #1a1d23;

@border-radius-base: 8px;
@border-color-base: #e4e7eb;

@font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
@font-size-base: 14px;

@layout-body-background: #f5f7fb;
@layout-header-background: rgba(255,255,255,0.82);
@layout-sider-background: #f9fafb;

@btn-border-radius-base: 8px;
@btn-height-base: 34px;
@input-height-base: 36px;
```

### 10.2 更新 `layout.less` 完整代码结构

文件: `/src/assets/css/components/layout.less`

改动要点:
1. Header: 52px, 毛玻璃, 弱边框
2. Sider: 220px/64px, 新配���, 去掉竖线指示符, 改用背景色
3. 菜单项: 圆角 8px, margin 留白, hover/active 新样式
4. 去掉分组分隔线, 改用间距
5. 折叠按钮: 去掉边框, 改为纯文字链接风格

### 10.3 更新 `warpperContent.less`

改动要点:
1. `.wrapper-main`: border-radius 12px, 新阴影, border @border-light
2. `.page-header`: 去掉 border-bottom 硬边, 使用更柔和的分离
3. `.page-search`: 卡片化背景

---

## 十一、迁移注意事项

1. **渐进式升级**: 通过 CSS 变量覆盖 Ant Design 默认值，不修改组件库源码
2. **兼容性**: 保留 `layout-dark` 主题类，新增 `layout-modern` 类作为默认
3. **回滚方案**: 所有改动通过 Less 变量和追加样式实现，可快速回退
4. **字体**: 需要添加 Inter 字体 CDN 或本地文件
5. **浏览器**: 毛玻璃 `backdrop-filter` 需要 Chrome 76+/Safari 9+

---

> 本规范遵循现代 SaaS 设计系统标准，参考 Linear、Notion、Vercel、Datadog 等一线产品的设计语言。
> 所有数值、颜色、间距均经过精细化校准，确保企业级专业感。
