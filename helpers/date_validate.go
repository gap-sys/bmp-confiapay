package helpers

import (
	"time"
)

// Valida datas

func ValidateDateLayout(format string, date string) bool {

	_, err := time.Parse(format, date)
	return err == nil

}

func DateBefore(date string, before string, format string) bool {
	d, err := time.Parse(format, date)
	if err != nil {
		return false
	}
	b, err := time.Parse(format, before)
	if err != nil {
		return false
	}

	return d.Before(b) || d.Equal(b)

}

func DateAfter(date string, after string, format string) bool {
	d, err := time.Parse(format, date)
	if err != nil {

		return false
	}
	a, err := time.Parse(format, after)
	if err != nil {

		return false
	}

	return d.After(a) || d.Equal(a)

}

func DateGap(date1, date2, format, orientation string, interval int64) bool {

	d1, err := time.Parse(format, date1)
	if err != nil {
		return false
	}
	d2, err := time.Parse(format, date2)
	if err != nil {
		return false
	}

	switch orientation {
	case "after":
		diff := d1.Sub(d2)
		daysDiff := diff.Hours() / 24

		return int64(daysDiff) <= interval

	case "before":
		diff := d2.Sub(d1)
		daysDiff := diff.Hours() / 24
		return int64(daysDiff) <= interval

	default:
		return false

	}

}
