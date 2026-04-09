# 文档样式指南

创建或编辑文档内容时，必须遵循本指南，使用结构化 block 提升可读性和视觉层次。

## 一、元素选择指南

优先使用结构化 block 代替纯文本段落。

| 场景 | 推荐 block | 避免 |
|-|-|-|
| 核心结论 / 摘要 / 注意事项 | `<callout>` + emoji + 背景色 | 纯 `<p>` 加粗 |
| 方案对比 / 优劣势 / Before vs After | `<grid>` 2 列分栏 | 用文字描述"左边…右边…" |
| 3+ 属性的结构化数据 / 指标表 | `<table>` + 表头背景色 | 用列表罗列 |
| 流程图 / 架构图 / 时序图 | `<whiteboard type="mermaid">` | 纯文字描述流程 |
| 任务清单 / 检查项 | `<checkbox>` | `- [ ]` 文本或普通列表 |
| 章节分隔 / 视觉节奏 | `<hr/>` | 连续空行 |
| 代码片段 | `<pre lang="x" caption="说明">` | 行内 `<code>` 包大段代码 |
| 引用 / 公式 | `<blockquote>` + `<latex>` | 纯文字写公式 |
| 操作入口 / 跳转链接 | `<button>` / `<a type="url-preview">` | 裸 URL |

## 二、默认丰富度规则

1. **每个 h1/h2 章节至少包含 1 个非纯文本 block**（callout / grid / table / whiteboard / checkbox 等）
2. 连续超过 3 个 `<p>` 段落时，必须考虑：用 `<callout>` 提炼要点、用 `<table>` 结构化、或用 `<grid>` 做对比
3. 文档开头用 `<callout>` front-load 核心信息
4. 涉及对比内容时，**必须**使用 `<grid>` 分栏
5. 用 `<checkbox>` 呈现待办/检查项，而非普通列表
6. 用 `<hr/>` 分隔不同主题的章节，创造视觉节奏

## 三、颜色语义

善用颜色系统区分信息类型：

| 语义 | callout 背景色 | 文字 / 边框色 |
|-|-|-|
| 信息、说明 | `light-blue` | `blue` |
| 成功、完成 | `light-green` | `green` |
| 警告、错误 | `light-red` | `red` |
| 注意、待确认 | `light-yellow` | `yellow` |

- 表头统一使用 `background-color="light-gray"`
- 关键指标用 `<span text-color="green/red">` 突出正负变化

## 四、排版节奏

- 空行分隔：不同块类型之间用空行分隔
- 标题层级 ≤ 4 层
- 用 `<callout>` 在文档开头 front-load 最重要的结论或摘要
- 用 `<grid>` 做 2 列对比（方案 A vs B），最多 3 列以保证可读性
- 有 3+ 属性的结构化数据用 `<table>`；简单列表用 `<ul>`
- 涉及流程/架构/关系时，用 `<whiteboard type="mermaid">` 插入图表，而非纯文字描述
