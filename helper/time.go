package helper

import (
	"time"
)

type (
	Time interface {
		Now() time.Time
		GetElapsedTime(finishTime time.Time) time.Duration
	}

	timeHelper struct{}
)

func (h *timeHelper) Now() time.Time {
	return time.Now()
}

func (h *timeHelper) GetElapsedTime(finishTime time.Time) time.Duration {
	now := time.Now()
	return now.Sub(finishTime)
}

var TimeHelper Time = &timeHelper{}
