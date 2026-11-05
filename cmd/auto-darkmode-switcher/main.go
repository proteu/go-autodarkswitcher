package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auto-darkmode-switcher/pkg/brightness"
)

func modeName(dark bool) string {
	if dark {
		return "Dark (暗黑)"
	}
	return "Light (浅色)"
}

func main() {
	threshold := flag.Float64("v", 0.8, "亮度切换阈值 (0.0 - 1.0)，低于此值切入暗黑模式")
	interval := flag.Int("i", 5, "轮询检测间隔时间（秒）")
	debounceSec := flag.Int("d", 3, "防抖观察时长（秒），状态稳定持续超过此时间才真正执行外观切换")
	flag.Parse()

	if *threshold < 0.0 || *threshold > 1.0 {
		fmt.Println("错误: 阈值必须在 0.0 到 1.0 之间")
		os.Exit(1)
	}

	fmt.Printf("=== macOS 自动外观切换守护程序已启动 ===\n")
	fmt.Printf("配置: 切换阈值 = %.2f | 检测周期 = %ds | 防抖观察 = %ds\n", *threshold, *interval, *debounceSec)

	// 捕获系统退出信号实现优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	normalInterval := time.Duration(*interval) * time.Second
	debounceDuration := time.Duration(*debounceSec) * time.Second

	var pendingState *bool
	var currentState *bool
	var pendingSince time.Time

	if isDark, err := brightness.IsDarkMode(); err == nil {
		currentState = &isDark
		fmt.Printf("[初始化] 当前系统外观状态 -> %s\n", modeName(*currentState))
	}

	// 动态睡眠控制：初始以正常间隔运行
	nextSleep := normalInterval

	for {
		select {
		case <-sigChan:
			fmt.Println("\n收到退出信号，程序正在停止...")
			return

		case <-time.After(nextSleep):
			currBrightness, err := brightness.GetCurrentBrightness()
			if err != nil {
				fmt.Printf("[ERROR] %v\n", err)
				nextSleep = normalInterval
				continue
			}

			shouldBeDark := currBrightness < *threshold

			// 1. 如果当前计算的状态与系统当前稳定状态一致，取消任何挂起的防抖任务
			if currentState != nil && *currentState == shouldBeDark {
				if pendingState != nil {
					fmt.Printf("[%s] 状态已恢复正常，取消防抖等待。\n", time.Now().Format("15:04:05"))
					pendingState = nil
				}
				nextSleep = normalInterval
				continue
			}

			// 2. 如果检测到新的状态改变需求
			now := time.Now()
			if pendingState == nil || *pendingState != shouldBeDark {
				// 第一次触发状态变更
				targetState := shouldBeDark
				pendingState = &targetState
				pendingSince = now

				fmt.Printf("[%s] 检测到外观拟切换 -> %s (亮度: %.4f)，开始 %ds 精准防抖...\n",
					pendingSince.Format("15:04:05"), modeName(targetState), currBrightness, *debounceSec)

				// 关键点：为了保证精准度，立刻将下一次唤醒时间设为“剩余的防抖时间”
				nextSleep = debounceDuration
				continue
			}
			// 3. 已经在防抖观察期中，检查是否达到防抖时长
			elapsed := now.Sub(pendingSince)
			if elapsed >= debounceDuration {
				err := brightness.SetDarkMode(*pendingState)
				if err != nil {
					fmt.Printf("[ERROR] 外观切换失败: %v\n", err)
				} else {
					fmt.Printf("[%s] 防抖确认结束 (持续超过 %ds) | 系统外观已成功切换 -> %s\n",
						now.Format("15:04:05"), *debounceSec, modeName(*pendingState))

					curr := *pendingState
					currentState = &curr
				}
				// 切换完成后清空挂起状态
				pendingState = nil
				nextSleep = normalInterval
			} else {
				// 还在防抖观察期内，继续等待
				remaining := debounceDuration - elapsed
				fmt.Printf("[%s] 观察期内 (已持续 %.1f秒)，将在 %.1f秒后再次确认...\n",
					now.Format("15:04:05"), elapsed.Seconds(), remaining.Seconds())
				nextSleep = remaining
			}
		}
	}
	/*
	   *

	   	for {
	   		currBrightness, err := brightness.GetCurrentBrightness()
	   		if err != nil {
	   			fmt.Printf("[ERROR] %v\n", err)
	   			time.Sleep(time.Duration(*interval) * time.Second)
	   			continue
	   		}

	   		shouldBeDark := currBrightness < *threshold

	   		// 仅在状态改变时执行系统切换，避免高频调用 osascript
	   		if lastState == nil || *lastState != shouldBeDark {
	   			err = brightness.SetDarkMode(shouldBeDark)
	   			if err != nil {
	   				fmt.Printf("[ERROR] %v\n", err)
	   			} else {
	   				modeName := "Light (浅色)"
	   				if shouldBeDark {
	   					modeName = "Dark (暗黑)"
	   				}
	   				fmt.Printf("[%s] 当前检测亮度: %.4f | 系统外观 -> %s\n",
	   					time.Now().Format("15:04:05"), currBrightness, modeName)

	   				state := shouldBeDark
	   				lastState = &state
	   			}
	   		}

	   		time.Sleep(time.Duration(*interval) * time.Second)
	   	}

	   *
	*/
}
