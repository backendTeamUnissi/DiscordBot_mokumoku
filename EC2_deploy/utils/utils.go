package utils

import (
	"fmt"
	"time"
)

// FormatDuration 滞在時間を時分秒の形式にフォーマット
func FormatDuration(d time.Duration) string {
    hours := int(d.Hours())
    minutes := int(d.Minutes()) % 60
    seconds := int(d.Seconds()) % 60

    return fmt.Sprintf("%02d時間%02d分%02d秒", hours, minutes, seconds)
}
