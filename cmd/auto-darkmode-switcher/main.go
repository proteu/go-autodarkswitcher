package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"auto-darkmode-switcher/pkg/brightness"
)

func main() {
	threshold := flag.Float64("v", 0.8, "亮度切换阈值 (0.0 - 1.0)，低于此值切入暗黑模式")
	interval := flag.Int("i", 5, "轮询检测间隔时间（秒）")
	flag.Parse()

	if *threshold < 0.0 || *threshold > 1.0 {
		fmt.Println("错误: 阈值必须在 0.0 到 1.0 之间")
		os.Exit(1)
	}

	fmt.Printf("=== macOS 自动外观切换守护程序已启动 ===\n")
	fmt.Printf("配置: 切换阈值 = %.2f | 检测周期 = %ds\n", *threshold, *interval)

	var lastState *bool

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
}
