package executor

import (
	"fmt"
	"time"
)

// formatDuration formats a duration with max 3 decimal places for cleaner display
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.3gμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.3gms", float64(d.Nanoseconds())/1000000)
	}
	return fmt.Sprintf("%.3gs", d.Seconds())
}
