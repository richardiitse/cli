## 文档组合模板

以下是常见文档类型的 XML 骨架，创建时可参考组合使用。

### 会议纪要

```xml
<title>XX 会议纪要 - YYYY/MM/DD</title>

<callout emoji="📝" background-color="light-blue" border-color="blue">
  <p>核心结论：[一句话摘要]</p>
</callout>

<h2>议题一：[主题]</h2>
<p>[讨论内容]</p>

<h2>决议事项</h2>
<table>
  <thead><tr><th background-color="light-gray">事项</th><th background-color="light-gray">负责人</th><th background-color="light-gray">截止日期</th></tr></thead>
  <tbody><tr><td>[事项]</td><td><cite type="user" user-id="[姓名]"></cite></td><td>[日期]</td></tr></tbody>
</table>

<h2>待办</h2>
<checkbox done="false">[待办 1]</checkbox>
<checkbox done="false">[待办 2]</checkbox>
```

### 项目方案

```xml
<title>[项目名称] 方案</title>

<callout emoji="💡" background-color="light-green" border-color="green">
  <p>推荐方案：[结论]。理由：[一句话]</p>
</callout>

<h1>背景</h1>
<p>[问题描述]</p>

<h1>方案对比</h1>
<grid>
  <column width-ratio="0.5">
    <h3>方案 A</h3>
    <ul><li>优点：[...]</li><li>缺点：[...]</li></ul>
  </column>
  <column width-ratio="0.5">
    <h3>方案 B</h3>
    <ul><li>优点：[...]</li><li>缺点：[...]</li></ul>
  </column>
</grid>

<h1>评估矩阵</h1>
<table>
  <thead><tr><th background-color="light-gray">维度</th><th background-color="light-gray">方案 A</th><th background-color="light-gray">方案 B</th></tr></thead>
  <tbody><tr><td>性能</td><td>[评分]</td><td>[评分]</td></tr></tbody>
</table>

<h1>下一步</h1>
<checkbox done="false">[行动项 1]</checkbox>
```

### 技术文档

```xml
<title>[功能/模块名称] 技术文档</title>

<callout emoji="💡" background-color="light-blue" border-color="blue">
  <p>TL;DR：[一句话概述核心设计]</p>
</callout>

<h1>设计概要</h1>
<p>[设计说明]</p>

<whiteboard type="mermaid">
graph LR
  A[客户端] --> B[网关]
  B --> C[服务]
  C --> D[数据库]
</whiteboard>

<h1>接口定义</h1>
<table>
  <thead><tr><th background-color="light-gray">参数</th><th background-color="light-gray">类型</th><th background-color="light-gray">必填</th><th background-color="light-gray">说明</th></tr></thead>
  <tbody><tr><td>user_id</td><td>string</td><td>是</td><td>用户 ID</td></tr></tbody>
</table>

<h1>核心实现</h1>
<pre lang="go" caption="核心逻辑"><code>[代码]</code></pre>
```


## 富文本更新示例

以下示例展示如何在更新操作中使用丰富的 XML 元素。

### 将普通段落替换为高亮提示框

```bash
lark-cli docs +update --doc "<doc_id>" --command block_replace \
  --block-id "blkcnXXXX" \
  --content '<callout emoji="⚠️" background-color="light-red" border-color="red"><p>此接口已废弃，请迁移至 v2。</p></callout>'
```

### 在标题后插入表格

```bash
lark-cli docs +update --doc "<doc_id>" --command block_insert \
  --block-id "blkcnXXXX" \
  --content '<table><colgroup><col span="3" width="120"/></colgroup><thead><tr><th background-color="light-gray">指标</th><th background-color="light-gray">当前</th><th background-color="light-gray">目标</th></tr></thead><tbody><tr><td>延迟</td><td>200ms</td><td><span text-color="green">100ms</span></td></tr></tbody></table>'
```

### 用 str_replace 将普通文本升级为 @人 + 公式

```bash
# 将 "由张三负责" 替换为 @人引用
lark-cli docs +update --doc "<doc_id>" --command str_replace \
  --pattern "由张三负责" --content '由 <cite type="user" user-id="id"></cite> 负责'

# 将文本公式替换为 LaTeX 渲染
lark-cli docs +update --doc "<doc_id>" --command str_replace \
  --pattern "E = mc2" --content '<latex>E = mc^2</latex>'
```
