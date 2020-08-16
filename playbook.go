package generalprobe

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

var (
	defaultNewBackoffTimer = NewExponentialBackoffTimer
)

const (
	defaultInterval = time.Second * 2
)

type PlayBook struct {
	Scenes   []Scene
	Interval time.Duration
	NewTimer func() BackoffTimer
}

func (x *PlayBook) Play() error {
	if x.NewTimer == nil {
		x.NewTimer = defaultNewBackoffTimer
	}

	if x.Interval == 0 {
		x.Interval = defaultInterval
	}

	for step, scene := range x.Scenes {
		label := fmt.Sprintf("[%d/%d] %s", step+1, len(x.Scenes), scene.Name())
		timer := x.NewTimer()
		for {
			exit, err := scene.Play()
			if err != nil {
				return errors.Wrapf(err, "Failed Playbook: %s", label)
			}
			if exit {
				break
			}

			waitTime := timer.WaitTime()
			if waitTime == nil {
				return fmt.Errorf("Retry limit exceeded: %s", label)
			}
			time.Sleep(*waitTime)
		}

		if step+1 < len(x.Scenes) {
			time.Sleep(x.Interval)
		}
	}

	return nil
}
