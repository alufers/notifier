package notifier

import (
	"time"

	"github.com/spf13/viper"
)

func truncateText(s string, max int) string {
	if len(s) > max {
		r := 0
		for i := range s {
			r++
			if r > max {
				return s[:i]
			}
		}
	}
	return s
}

func formatDate(d time.Time) string {
	return d.Format(viper.GetString("general.date_format"))
}
