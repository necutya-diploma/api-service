package domain

import "time"

const (
	DateLayout  = "2006-02-01"
	HoursInDay  = 24
	DaysInMonth = 30
)

func BeginningOfTheDay(t time.Time) time.Time {
	return t.UTC().Truncate(HoursInDay * time.Hour)
}

func DayAfter(t time.Time) time.Time {
	return BeginningOfTheDay(t.Add(HoursInDay * time.Hour))
}

func HoursInMonth() int {
	return HoursInDay * DaysInMonth
}
