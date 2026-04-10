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

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/output"
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
	_ = opts.Ctx // TODO: pass to WebSocket/event subscriber

	// TODO: 实现完整的 Bot 启动逻辑
	// 1. 加载配置
	// 2. 初始化 session manager
	// 3. 启动 event +subscribe
	// 4. 处理消息循环
	// 5. 支持 daemon 模式

	fmt.Fprintf(f.IOStreams.Out, "=== Claude Code Bot 启动中 ===\n")

	// 临时实现：提示功能尚未完成
	result := map[string]interface{}{
		"status": "not_implemented",
		"message": "Bot 功能正在开发中，敬请期待",
		"config":  opts.Config,
		"daemon":  opts.Daemon,
	}
	output.PrintJson(f.IOStreams.Out, result)

	// 设置信号监听（用于优雅退出）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)
}
