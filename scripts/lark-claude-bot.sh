#!/bin/bash
# lark-claude-bot.sh
# 飞书 Bot + Claude Code 多轮对话服务
#
# 依赖：lark-cli, claude, jq
# 用法：bash /tmp/lark-claude-bot.sh
# pm2：pm2 start /tmp/lark-claude-bot.sh --name lark-claude-bot --interpreter bash

set -uo pipefail

# ── 配置 ──────────────────────────────────────────────────────────────────────
SESSION_DIR="${LARK_BOT_SESSION_DIR:-/tmp/lark-bot-sessions}"  # session 存储目录
WORK_DIR="${LARK_BOT_WORK_DIR:-$HOME}"                         # Claude 工作目录
SYSTEM_PROMPT="${LARK_BOT_SYSTEM_PROMPT:-你是一个智能助手，请用中文简洁地回答问题。}"
MAX_SESSIONS="${LARK_BOT_MAX_SESSIONS:-100}"                   # 最大并发 session 数
# ─────────────────────────────────────────────────────────────────────────────

mkdir -p "$SESSION_DIR"

log() { echo "[$(date '+%H:%M:%S')] $*" >&2; }

# 获取或创建某个 chat_id 对应的 Claude session-id
get_session() {
    local chat_id="$1"
    local session_file="$SESSION_DIR/${chat_id//[^a-zA-Z0-9_-]/_}.session"
    if [[ -f "$session_file" ]]; then
        cat "$session_file"
    else
        # 清理超出上限的旧 session（按修改时间，删最旧的）
        local count
        count=$(find "$SESSION_DIR" -name "*.session" | wc -l)
        if (( count >= MAX_SESSIONS )); then
            find "$SESSION_DIR" -name "*.session" -print0 \
                | xargs -0 ls -t | tail -1 | xargs rm -f
        fi
        echo ""  # 空 = 新 session
    fi
}

save_session() {
    local chat_id="$1"
    local session_id="$2"
    local session_file="$SESSION_DIR/${chat_id//[^a-zA-Z0-9_-]/_}.session"
    echo "$session_id" > "$session_file"
    touch "$session_file"  # 更新修改时间
}

# 处理单条消息（后台并发执行）
handle_message() {
    local chat_id="$1"
    local content="$2"
    local sender_id="$3"

    log "收到消息 chat=$chat_id sender=$sender_id: $content"

    # 获取该 chat 的历史 session
    local session_id
    session_id=$(get_session "$chat_id")

    # 构建 claude 调用参数
    local claude_args=(-p "$content" --add-dir "$WORK_DIR" --dangerously-skip-permissions)
    if [[ -n "$session_id" ]]; then
        claude_args+=(--resume "$session_id")
    fi

    # 调用 Claude Code，捕获输出和新 session-id
    local answer new_session_id
    local claude_output
    claude_output=$(claude "${claude_args[@]}" --output-format json 2>/tmp/claude_err.log) || {
        # 如果是 session 过期问题，尝试不用 resume 重新发起
        if [[ -n "$session_id" ]]; then
            log "Session 过期或无效，重新开始会话 chat=$chat_id"
            claude_args=(-p "$content" --add-dir "$WORK_DIR" --dangerously-skip-permissions)
            claude_output=$(claude "${claude_args[@]}" --output-format json 2>/tmp/claude_err.log) || {
                log "Claude 调用失败 chat=$chat_id: $(cat /tmp/claude_err.log)"
                lark-cli im +messages-send --chat-id "$chat_id" \
                    --text "抱歉，处理出错了，请稍后重试。" --as bot >/dev/null 2>&1
                return
            }
        else
            log "Claude 调用失败 chat=$chat_id: $(cat /tmp/claude_err.log)"
            lark-cli im +messages-send --chat-id "$chat_id" \
                --text "抱歉，处理出错了，请稍后重试。" --as bot >/dev/null 2>&1
            return
        fi
    }

    answer=$(echo "$claude_output" | jq -r '.result // .message // empty' 2>/dev/null)
    new_session_id=$(echo "$claude_output" | jq -r '.session_id // empty' 2>/dev/null)

    # 保存新 session-id
    if [[ -n "$new_session_id" ]]; then
        save_session "$chat_id" "$new_session_id"
    fi

    [[ -z "$answer" ]] && answer="（无回复）"

    log "Claude 回复 chat=$chat_id: ${answer:0:80}..."

    # 发送回复到飞书
    lark-cli im +messages-send \
        --chat-id "$chat_id" \
        --text "$answer" \
        --as bot >/dev/null 2>&1

    log "已回复 chat=$chat_id"
}

log "=== lark-claude-bot 启动 ==="
log "工作目录: $WORK_DIR"
log "Session 目录: $SESSION_DIR"
log "监听事件: im.message.receive_v1"
log "等待飞书消息..."

# 主循环：订阅飞书事件，并发处理每条消息
lark-cli event +subscribe \
    --event-types im.message.receive_v1 \
    --compact --quiet \
| while IFS= read -r line; do
    content=$(echo "$line" | jq -r '.content // empty')
    chat_id=$(echo "$line" | jq -r '.chat_id // empty')
    sender_id=$(echo "$line" | jq -r '.sender_id // empty')

    [[ -z "$content" || -z "$chat_id" ]] && continue

    # 后台并发处理，重定向 stdin 避免与主循环冲突
    handle_message "$chat_id" "$content" "$sender_id" </dev/null &
done
