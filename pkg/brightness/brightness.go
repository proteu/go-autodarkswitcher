package brightness

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// 匹配如 "brightness 0.785156"
var brightnessRegex = regexp.MustCompile(`brightness\s+([0-9]+\.[0-9]+)`)

// GetCurrentBrightness 获取当前主屏幕亮度 (0.0 ~ 1.0)
func GetCurrentBrightness() (float64, error) {
	// 使用 -l 参数列出当前亮度状态
	cmd := exec.Command("brightness", "-l")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("执行 brightness -l 失败: %w, stderr: %s", err, stderr.String())
	}

	output := out.String()

	// 解析输出内容
	matches := brightnessRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		return 0, fmt.Errorf("无法从命令输出中解析亮度数值，原始输出:\n%s", output)
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
	if err != nil {
		return 0, fmt.Errorf("数值转换失败 [%s]: %w", matches[1], err)
	}

	return val, nil
}
