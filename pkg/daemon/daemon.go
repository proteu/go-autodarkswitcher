package daemon

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auto-darkmode-switcher/pkg/brightness"
)

// Options 守护程序配置选项
type Options struct {
	Threshold   float64
	Offset      float64
	Interval    time.Duration
	DebounceSec time.Duration
}

// ModeName 辅助函数：将 bool 状态转为直观的外观名称
func ModeName(dark bool) string {
	if dark {
		return "Dark (暗黑)"
	}
	return "Light (浅色)"
}

// EvaluateTargetState 基于滞后区间（Hysteresis）计算目标外观状态
func EvaluateTargetState(currentState bool, currBrightness, threshold, offset float64) bool {
	lowThreshold := threshold - offset
	highThreshold := threshold + offset

	if !currentState {
		return currBrightness < lowThreshold
	} else {
		return currBrightness <= highThreshold
	}
}

// Run 启动自动切换守护程序
func Run(opts Options) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 初始化当前系统外观状态
	currentState := false
	if isDark, err := brightness.IsDarkMode(); err == nil {
		currentState = isDark
		slog.Info("初始化系统外观状态成功", "current_mode", ModeName(currentState))
	} else {
		slog.Error("初始化获取系统外观状态失败", "error", err)
	}

	var pendingState *bool
	var pendingSince time.Time
	nextSleep := opts.Interval

	for {
		select {
		case <-sigChan:
			slog.Info("收到退出信号，程序正在停止...")
			return nil

		case <-time.After(nextSleep):
			// 1. 实时从系统同步外观，防止外部手动修改导致错乱
			if realIsDark, err := brightness.IsDarkMode(); err == nil {
				if currentState != realIsDark {
					slog.Info("检测到系统外观已被外部手动修改，正在实时校准状态",
						"current_state", ModeName(currentState),
						"real_state", ModeName(realIsDark),
					)
					currentState = realIsDark
					if pendingState != nil {
						pendingState = nil
					}
				}
			}

			// 2. 获取当前亮度
			currBrightness, err := brightness.GetCurrentBrightness()
			if err != nil {
				slog.Error("获取当前屏幕亮度失败", "error", err)
				nextSleep = opts.Interval
				continue
			}

			slog.Debug("采集到当前屏幕亮度", "brightness", currBrightness)

			// 3. 计算目标状态
			shouldBeDark := EvaluateTargetState(currentState, currBrightness, opts.Threshold, opts.Offset)

			// 4. 状态一致，取消挂起
			if currentState == shouldBeDark {
				if pendingState != nil {
					slog.Info("亮度回到安全区间，取消防抖等待")
					pendingState = nil
				}
				nextSleep = opts.Interval
				continue
			}

			// 5. 触发拟切换（开始防抖）
			now := time.Now()
			if pendingState == nil {
				pendingState = &shouldBeDark
				pendingSince = now

				slog.Info("检测到外观状态拟切换，进入防抖观察期",
					"target_mode", ModeName(shouldBeDark),
					"brightness", currBrightness,
					"debounce_sec", opts.DebounceSec.Seconds(),
				)

				nextSleep = opts.DebounceSec
				continue
			}

			// 6. 防抖确认期判定
			elapsed := now.Sub(pendingSince)
			if elapsed >= opts.DebounceSec {
				err := brightness.SetDarkMode(*pendingState)
				if err != nil {
					slog.Error("系统外观切换失败", "error", err, "target_mode", ModeName(*pendingState))
				} else {
					slog.Info("防抖确认结束，系统外观已成功切换",
						"duration_sec", elapsed.Seconds(),
						"new_mode", ModeName(*pendingState),
					)
					currentState = *pendingState
				}
				pendingState = nil
				nextSleep = opts.Interval
			} else {
				remaining := opts.DebounceSec - elapsed
				slog.Debug("处于防抖观察期内",
					"elapsed_sec", elapsed.Seconds(),
					"remaining_sec", remaining.Seconds(),
				)
				nextSleep = remaining
			}
		}
	}
}
