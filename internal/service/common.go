package service

import (
	"time"

	"necutya/faker/internal/domain/domain"
)

func convertDollarsToCents(dollars int) int {
	return dollars * 100
}

func getDurationToMidnight() time.Duration {
	midnight := domain.DayAfter(time.Now())
	return midnight.Sub(time.Now())
}
