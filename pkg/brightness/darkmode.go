package brightness

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// IsDarkMode 查询系统当前是否为暗黑模式
func IsDarkMode() (bool, error) {
	script := `tell application "System Events" to tell appearance preferences to get dark mode`
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("获取系统外观状态失败: %w", err)
	}

	return strings.TrimSpace(out.String()) == "true", nil
}

// SetDarkMode 设置系统外观模式
func SetDarkMode(enable bool) error {
	val := "false"
	if enable {
		val = "true"
	}
	script := fmt.Sprintf(`tell application "System Events" to tell appearance preferences to set dark mode to %s`, val)
	cmd := exec.Command("osascript", "-e", script)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("切换外观模式失败: %w", err)
	}
	return nil
}
