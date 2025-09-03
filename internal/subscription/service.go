package subscription

import "time"

func parseMonthYear(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}
