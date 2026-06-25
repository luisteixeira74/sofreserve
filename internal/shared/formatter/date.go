package formatter

import "time"

func FormatDateTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("02/01/2006 15:04")
}