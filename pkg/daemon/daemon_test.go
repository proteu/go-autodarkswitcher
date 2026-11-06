package daemon

import (
	"testing"
)

func TestEvaluateTargetState(t *testing.T) {
	threshold := 0.8
	offset := 0.035
	// lowThreshold = 0.765, highThreshold = 0.835

	tests := []struct {
		name           string
		currentState   bool // false = Light, true = Dark
		currBrightness float64
		expected       bool
	}{
		// --- 场景 1：当前为浅色模式 (Light, currentState = false) ---
		{
			name:           "浅色下亮度正常(高于低阈值)，保持浅色",
			currentState:   false,
			currBrightness: 0.85,
			expected:       false,
		},
		{
			name:           "浅色下亮度刚好处在死区内(0.78)，应保持浅色",
			currentState:   false,
			currBrightness: 0.78,
			expected:       false,
		},
		{
			name:           "浅色下亮度跌破低阈值(0.75)，触发切入暗黑",
			currentState:   false,
			currBrightness: 0.75,
			expected:       true,
		},

		// --- 场景 2：当前为暗黑模式 (Dark, currentState = true) ---
		{
			name:           "暗黑下亮度很低(0.50)，保持暗黑",
			currentState:   true,
			currBrightness: 0.50,
			expected:       true,
		},
		{
			name:           "暗黑下亮度回升但仍处在死区内(0.79)，应继续保持暗黑",
			currentState:   true,
			currBrightness: 0.79,
			expected:       true,
		},
		{
			name:           "暗黑下亮度涨超高阈值(0.85)，触发切回浅色",
			currentState:   true,
			currBrightness: 0.85,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateTargetState(tt.currentState, tt.currBrightness, threshold, offset)
			if got != tt.expected {
				t.Errorf("EvaluateTargetState() = %v, 期望 %v (当前状态=%v, 亮度=%.3f)",
					got, tt.expected, tt.currentState, tt.currBrightness)
			}
		})
	}
}
