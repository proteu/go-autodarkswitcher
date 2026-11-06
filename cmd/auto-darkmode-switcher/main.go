package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"auto-darkmode-switcher/pkg/daemon"
)

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	threshold := flag.Float64("v", 0.7, "亮度切换阈值 (0.0 - 1.0)，低于此值切入暗黑模式")
	offset := flag.Float64("o", 0.035, "滞后偏移量，用于防止临界值反复横跳")
	interval := flag.Int("i", 5, "轮询检测间隔时间（秒）")
	debounceSec := flag.Int("d", 3, "防抖观察时长（秒），状态稳定持续超过此时间才真正执行外观切换")
	logLevelStr := flag.String("l", "info", "日志级别 (debug, info, warn, error)")
	flag.Parse()

	if *threshold < 0.0 || *threshold > 1.0 {
		fmt.Println("错误: 阈值必须在 0.0 到 1.0 之间")
		os.Exit(1)
	}
	if *offset < 0.0 || *offset > 0.1 {
		fmt.Println("错误: 偏移量必须在 0.0 到 0.1 之间")
		os.Exit(1)
	}

	// 初始化全局结构化日志
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(*logLevelStr),
	}))
	slog.SetDefault(logger)

	slog.Info("macOS 自动外观切换守护程序已启动",
		"threshold", *threshold,
		"offset", *offset,
		"interval_sec", *interval,
		"debounce_sec", *debounceSec,
		"log_level", strings.ToUpper(*logLevelStr),
	)

	opts := daemon.Options{
		Threshold:   *threshold,
		Offset:      *offset,
		Interval:    time.Duration(*interval) * time.Second,
		DebounceSec: time.Duration(*debounceSec) * time.Second,
	}

	if err := daemon.Run(opts); err != nil {
		slog.Error("守护程序异常退出", "error", err)
		os.Exit(1)
	}
}
