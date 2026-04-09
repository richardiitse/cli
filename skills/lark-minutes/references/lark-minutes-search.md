# minutes +search

> **前置条件：** 先阅读 [`../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

搜索妙记列表，支持关键词、所有者、参与者以及时间范围等多条件过滤。所有者与参与者都支持传入多个 open\_id，也支持传入 `me` 表示当前用户。只读操作，不修改任何妙记数据。

本 skill 对应 shortcut：`lark-cli minutes +search`（调用 `POST /open-apis/minutes/v1/minutes/search`）。

## 典型触发表达

以下说法通常应优先使用 `minutes +search`：

- 我的妙记
- 我拥有的妙记
- 我参与的妙记
- 最近的妙记
- 某个关键词的妙记
- 某段时间内的妙记

## 命令

```bash
# 关键词搜索
lark-cli minutes +search --query "预算复盘"

# 查询某一天内的妙记（单日查询时，建议将 start 和 end 都填写为同一天）
lark-cli minutes +search --start 2026-03-10 --end 2026-03-10

# 按时间范围搜索
lark-cli minutes +search --start "2026-03-10T00:00+08:00" --end "2026-03-17T00:00+08:00"
lark-cli minutes +search --start 2026-03-10 --end 2026-03-17

# 关键词 + 时间范围
lark-cli minutes +search --query "预算复盘" --start "2026-03-10T00:00+08:00" --end "2026-03-17T00:00+08:00"
lark-cli minutes +search --query "预算复盘" --start "2026-03-10T00:00+08:00"
lark-cli minutes +search --query "预算复盘" --end "2026-03-17T00:00+08:00"

# 按参与者过滤（open_id，逗号分隔）
lark-cli minutes +search --participant-ids "ou_x,ou_y"

# 按所有者过滤（open_id，逗号分隔）
lark-cli minutes +search --owner-ids "ou_owner,ou_owner_2"

# 查询我参与的妙记
lark-cli minutes +search --participant-ids "me"

# 查询我拥有的妙记
lark-cli minutes +search --owner-ids "me"

# 多条件组合查询
lark-cli minutes +search --owner-ids "ou_owner" --participant-ids "ou_x" --start "2026-03-10T00:00+08:00"

# 分页查询
lark-cli minutes +search --query "预算复盘" --page-size 20
lark-cli minutes +search --query "预算复盘" --page-size 20 --page-token '<PAGE_TOKEN>'

# 输出为结构化 JSON
lark-cli minutes +search --query "预算复盘" --format json
```

## 参数

| 参数                        | 必填 | 说明                                   |
| ------------------------- | -- | ------------------------------------ |
| `--query <text>`          | 否  | 搜索关键词                                |
| `--owner-ids <ids>`       | 否  | 所有者 open\_id 列表，逗号分隔；支持传 `me` 表示当前用户 |
| `--participant-ids <ids>` | 否  | 参与者 open\_id 列表，逗号分隔；支持传 `me` 表示当前用户 |
| `--start <time>`          | 否  | 开始时间（ISO 8601 或仅日期）                  |
| `--end <time>`            | 否  | 结束时间（ISO 8601 或仅日期）                  |
| `--page-size <n>`         | 否  | 每页数量，默认 `15`，最大 `200`                |
| `--page-token <token>`    | 否  | 下一页分页 token                          |
| `--dry-run`               | 否  | 预览 API 调用，不执行                        |

## 核心约束

### 1. 至少提供一个过滤条件

所有参数均可选，但必须至少提供一个过滤条件：`--query`、`--owner-ids`、`--participant-ids`、`--start` 或 `--end`。

### 2. 仅支持 user 身份

该接口仅支持 `user` 身份，使用前需完成 `lark-cli auth login` 并具备 `minutes:minutes.search:read` 权限。

### 3. `me` 表示当前用户

在 `--owner-ids` 和 `--participant-ids` 中可使用 `me`，表示当前登录用户。该值会在本地解析为当前用户的 `open_id`，无需手动先查询自己的用户 ID。
若当前环境尚未完成用户登录，或 CLI 无法解析出当前用户的 `open_id`，则应先执行 `lark-cli auth login`，再重新执行搜索。

### 4. 长时间范围优先拆分查询

当搜索时间范围很长时，优先按月拆分为多次查询，再合并结果，避免单次返回结果过多或遗漏目标妙记。

### 5. 使用 page\_token 分页

该接口使用 `page_token` 分页。继续翻页时，需要把下一次请求的 `--page-token` 设置为当前结果返回的 `page_token`。
单次查询最多返回 `200` 条，但结果总数没有固定上限，应按 `has_more` 和 `page_token` 持续翻页直到拿全。

### 6. 日期型 `--end` 包含当天整天

当 `--end` 传入的是仅日期格式（如 `2026-03-10`）时，CLI 会将它解释为当天 `23:59:59`，而不是当天 `00:00:00`。
CLI 会先按输入的本地日历日语义解析，再标准化为 RFC3339 时间戳发给 API；在 dry-run 或排查请求体时，看到的 `Z` 结尾时间表示同一个绝对时间点的 UTC 表示，不改变“按当天整天查询”的语义。

这意味着：

- `--start 2026-03-10 --end 2026-03-10` 表示只查 `2026-03-10` 当天
- `--start 2026-03-10 --end 2026-03-11` 表示查询 `2026-03-10` 和 `2026-03-11` 两天

如果用户说“昨天的妙记”“今天的妙记”“某一天内的妙记”，应把 `--start` 和 `--end` 都设置为同一天，而不是把 `--end` 设成下一天。

## 时间格式

`--start` 和 `--end` 支持以下时间格式：

| 格式             | 示例                          | 说明      |
| -------------- | --------------------------- | ------- |
| ISO 8601（带时区）  | `2026-03-10T14:00:00+08:00` | 推荐      |
| ISO 8601（不带时区） | `2026-03-10T14:00:00`       | 按本地时区解析 |
| 仅日期            | `2026-03-10`                | 按天粒度解析；若用于 `--end`，表示当天 `23:59:59` |

## 输出结果

- 结构化输出包含 `items`、`total`、`has_more` 和 `page_token`。
- 表格输出默认展示 `token`、`display_info`、`description`、`app_link` 和 `avatar`。

## 请求结构

- Query 参数仅包含分页字段：`page_size`、`page_token`
- Request Body 包含：
  - `query`
  - `filter.owner_ids`
  - `filter.participant_ids`
  - `filter.create_time.start_time`
  - `filter.create_time.end_time`

## Pagination (`has_more` / `page_token`)

- 当结果中返回 `has_more=true` 时，说明还有更多页可继续获取。
- 继续翻页时，使用响应中的 `page_token` 作为下一次查询的 `--page-token`。
- 单次查询的 `page_size` 最大为 `200`。
- 结果总数没有固定上限；不要假设调大 `--page-size` 就能拿全结果，分页遍历时应以 `has_more` 和 `page_token` 为准。

```bash
# First page
lark-cli minutes +search --query "预算复盘" --page-size 20

# Next page
lark-cli minutes +search --query "预算复盘" --page-size 20 --page-token '<PAGE_TOKEN>'
```

## 搜索结果中的下一步

搜索结果中的 `token` 可直接作为 `minute_token` 用于继续查询妙记产物：
通常先用搜索结果中的 `token` 获取妙记基础信息，确认描述、链接等元数据是否命中目标；需要进一步查看内容时，再继续查询关联的纪要产物。

如果你已经确定目标妙记，优先直接复用搜索结果中的 `token`，避免重复搜索。

```bash
# 首先查询妙记元信息（标题、时长、封面） → 用本 skill
lark-cli minutes minutes get --params '{"minute_token": "obcn***************"}'

# 查妙记关联的纪要产物：逐字稿、总结、待办、章节等 → 用 lark-cli vc +notes
lark-cli vc +notes --minute-tokens obcnhijv43vq6bcsl5xasfb2
```

## 常见错误与排查

| 错误现象                   | 根本原因                              | 解决方案                        |
| ---------------------- | --------------------------------- | --------------------------- |
| 命令直接报错，要求提供过滤条件        | 没有传入 `--query`、时间范围或任何过滤 ID       | 至少补充一个过滤条件后重试               |
| 时间参数校验失败               | `--start` 或 `--end` 格式不合法         | 改用 ISO 8601 或 `YYYY-MM-DD`  |
| `owner-ids` 校验失败       | 传入的不是 open\_id，且也不是 `me`；或传了 `me` 但当前用户 open\_id 不可解析 | 改为 `ou_` 开头的用户 ID，或先完成 `auth login` 后再传 `me` |
| `participant-ids` 校验失败 | 传入的不是 open\_id，且也不是 `me`；或传了 `me` 但当前用户 open\_id 不可解析 | 改为 `ou_` 开头的用户 ID，或先完成 `auth login` 后再传 `me` |
| 权限不足                   | 未授权 `minutes:minutes.search:read` | 使用 `auth login` 完成授权        |

## 提示

- 当用户说“我的妙记”时，优先理解为 `--owner-ids me`。
- 当用户说“我参与的妙记”时，优先理解为 `--participant-ids me`。
- 建议使用 `--format json` 输出，便于解析 `has_more` 和 `page_token`。
- 当结果存在多页时，持续使用 `page_token` 继续遍历；单次最多 `200` 条，总结果数没有上限，不要只看第一页。
- 排查参数与请求结构时优先使用 `--dry-run`。
- 不要使用 `yesterday`、`today` 这类相对时间字面量；请先转换成明确日期，例如 `2026-03-10`。
- 当用户已经明确给出 `minute_token` 或妙记链接时，优先进入 `vc +notes --minute-tokens ...`，而不是再次搜索。

## 参考

- [lark-minutes](../SKILL.md) -- 妙记相关命令
- [lark-vc-notes](../../lark-vc/references/lark-vc-notes.md) -- 基于 `minute_token` 获取逐字稿、总结、待办、章节等产物
- [lark-shared](../../lark-shared/SKILL.md) -- 认证和全局参数
