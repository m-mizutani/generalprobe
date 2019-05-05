package generalprobe

import (
	"fmt"
	"time"
)

// PauseScene is a scene of just sleeping
type PauseScene struct {
	interval int
	baseScene
}

// Pause creats a scene of sleeping.
func (x *Generalprobe) Pause(interval int) *PauseScene {
	scene := PauseScene{
		interval: interval,
	}
	return &scene
}

// String returns text explanation of the scene
func (x *PauseScene) String() string {
	return fmt.Sprintf("Pausing %d seconds", x.interval)
}

func (x *PauseScene) play() error {
	time.Sleep(time.Second * time.Duration(x.interval))
	return nil
}
