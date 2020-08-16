package generalprobe

import (
	"math"
	"time"
)

type BackoffTimer interface {
	WaitTime() *time.Duration
}

type ExponentialBackoffTimer struct {
	retry int
}

func (x *ExponentialBackoffTimer) WaitTime() *time.Duration {
	x.retry++
	if x.retry > 10 {
		return nil
	}

	waitMillSec := time.Duration(math.Pow(2, float64(x.retry))*100) * time.Millisecond
	return &waitMillSec
}

func NewExponentialBackoffTimer() BackoffTimer {
	return &ExponentialBackoffTimer{}
}
