// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package bot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larksuite/cli/cmd/cmdutil"
	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/bot"
	"github.com/spf13/cobra"
)

// BotStartOptions holds inputs for the bot start command.
type BotStartOptions struct {
	Factory  *cmdutil.Factory
	Ctx      context.Context
	Config   string
	Daemon   bool
}

// newCmdBotStart creates the bot start command.
func newCmdBotStart(opts *BotOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "启动 Claude Code Bot",
		Long:  "启动飞书 Bot，监听消息并路由给 Claude Code 处理",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Ctx = cmd.Context()
			return botStartRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Config, "config", "", "配置文件路径")
	cmd.Flags().BoolVar(&opts.Daemon, "daemon", false, "后台运行模式")

	return cmd
}

// botStartRun executes the bot start command.
func botStartRun(opts *BotStartOptions) error {
	f := opts.Factory
	ctx := opts.Ctx
	io := f.IOStreams

	fmt.Fprintf(io.Out, "=== Claude Code Bot 启动中 ===\n")

	// 1. Validate claude CLI is available
	fmt.Fprintf(io.Out, "验证 Claude Code CLI...\n")
	if err := bot.ValidateClaudeCLI(ctx); err != nil {
		return fmt.Errorf("claude CLI validation failed: %w", err)
	}
	fmt.Fprintf(io.Out, "✓ Claude Code CLI 已就绪\n")

	// 2. Initialize session manager
	fmt.Fprintf(io.Out, "初始化 Session 管理器...\n")
	sessionMgr, err := bot.NewSessionManager(bot.SessionManagerConfig{
		TTL: 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	fmt.Fprintf(io.Out, "✓ Session 管理器已初始化\n")

	// 3. Initialize Claude client
	claudeClient := bot.NewClaudeClient(bot.ClaudeClientConfig{
		WorkDir:         "/tmp/lark-claude-bot",
		Timeout:         5 * time.Minute,
		MaxRetries:      3,
		SkipPermissions: true,
	})

	// 4. Initialize bot handler
	fmt.Fprintf(io.Out, "初始化 Bot 处理器...\n")
	botHandler, err := bot.NewBotHandler(bot.BotHandlerConfig{
		ClaudeClient:   claudeClient,
		SessionManager: sessionMgr,
		WorkDir:        "/tmp/lark-claude-bot",
	})
	if err != nil {
		return fmt.Errorf("failed to create bot handler: %w", err)
	}
	fmt.Fprintf(io.Out, "✓ Bot 处理器已初始化\n")

	// 5. Start event subscription (TODO: integrate event +subscribe)
	fmt.Fprintf(io.Out, "\n⚠️  事件订阅集成待实现\n")
	fmt.Fprintf(io.Out, "当前状态：核心模块已完成，需要集成 event +subscribe\n")

	// Show status
	stats, _ := botHandler.GetStats(ctx)
	output.PrintJson(io.Out, stats)

	// 设置信号监听（用于优雅退出）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 如果不是 daemon 模式，等待信号
	if !opts.Daemon {
		fmt.Fprintf(io.Out, "\n按 Ctrl+C 停止 Bot\n")
		<-sigChan
		fmt.Fprintf(io.Out, "\n=== Bot 已停止 ===\n")
		return nil
	}

	// Daemon 模式：后台运行
	// TODO: 实现 daemon 模式（fork 进程、PID 文件等）
	return errors.New("daemon 模式尚未实现")
}
