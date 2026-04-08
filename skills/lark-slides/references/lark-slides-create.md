
# slides +create（创建飞书幻灯片）

> **前置条件：** 先阅读 [`../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

创建一个新的空白飞书幻灯片演示文稿。

## 命令

```bash
# 创建空白 PPT
lark-cli slides +create --title "项目汇报"

# 以应用身份创建（自动授权当前用户）
lark-cli slides +create --title "项目汇报" --as bot
```

## 返回值

工具成功执行后，返回一个 JSON 对象，包含以下字段：

- **`xml_presentation_id`**（string）：演示文稿的唯一标识符，后续添加页面时需要此 ID
- **`title`**（string）：演示文稿标题
- **`revision_id`**（integer）：演示文稿版本号
- **`permission_grant`**（object，可选）：仅 `--as bot` 时返回，说明是否已自动为当前 CLI 用户授予可管理权限

> [!IMPORTANT]
> `slides +create` 只创建空白演示文稿。创建后需要使用 `xml_presentation.slide create` 逐页添加 slide 内容。

> [!IMPORTANT]
> 如果演示文稿是**以应用身份（bot）创建**的，如 `lark-cli slides +create --as bot`，CLI 会**尝试为当前 CLI 用户自动授予该演示文稿的 `full_access`（可管理权限）**。
>
> 以应用身份创建时，结果里会额外返回 `permission_grant` 字段，明确说明授权结果：
> - `status = granted`：当前 CLI 用户已获得该演示文稿的可管理权限
> - `status = skipped`：本地没有可用的当前用户 `open_id`，因此不会自动授权
> - `status = failed`：演示文稿已创建成功，但自动授权用户失败
>
> **不要擅自执行 owner 转移。** 如果用户需要把 owner 转给自己，必须单独确认。

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--title` | 否 | 演示文稿标题（不传则默认 "Untitled"） |

## 创建后续步骤

`slides +create` 返回的 `xml_presentation_id` 用于后续操作：

```bash
# 第 1 步：创建空白 PPT
PRES_ID=$(lark-cli slides +create --title "项目汇报" | jq -r '.data.xml_presentation_id')

# 第 2 步：添加页面（使用返回的 xml_presentation_id）
lark-cli slides xml_presentation.slide create \
  --params "{\"xml_presentation_id\":\"$PRES_ID\"}" \
  --data '{
    "slide": {
      "content": "<slide xmlns=\"http://www.larkoffice.com/sml/2.0\">...</slide>"
    }
  }'
```

## 常见错误

| 错误码 | 含义 | 解决方案 |
|--------|------|----------|
| 400 | 参数错误 | 检查参数格式是否正确 |
| 403 | 权限不足 | 检查是否拥有 `slides:presentation:create` scope |

## 相关命令

- [xml_presentations create](lark-slides-xml-presentations-create.md) — 原生创建命令
- [xml_presentation.slides create](lark-slides-xml-presentation-slides-create.md) — 添加幻灯片页面
- [xml_presentations get](lark-slides-xml-presentations-get.md) — 读取 PPT 内容
