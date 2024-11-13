package utils

import (
	"fmt"
	"time"
)

// FormatDuration 滞在時間を時分秒の形式にフォーマット
func FormatDuration(duration time.Duration) string {
	totalSeconds := int(duration.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d時間%02d分%02d秒", hours, minutes, seconds)
}
