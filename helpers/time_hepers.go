package helpers

import "time"

func NextMidnight(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
}
