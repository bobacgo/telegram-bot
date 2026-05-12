package timex

import (
	"context"
	"time"
)

// Sleep 可中断的等待，返回true表示被停止，false表示正常等待结束
func Sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return false
	case <-ctx.Done():
		return true
	}
}
