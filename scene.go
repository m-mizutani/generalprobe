package generalprobe

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
)

// Scene is a part of playbook of test.
type Scene interface {
	play() error
	setGeneralprobe(gp *Generalprobe)
	string() string
}

type baseScene struct {
	gp *Generalprobe
}

type pollingScene struct {
	baseScene
	limit    int
	interval int
}

// Limit sets maximum retry number of the scene. Basically if the limit
// exceeded, the test will be fail by returning error from generalprobe.Run()
func (x *pollingScene) Limit(limit int) Scene {
	x.limit = limit
	return x
}

// Interval sets seconds of query interval.
func (x *pollingScene) Interval(limit int) Scene {
	x.limit = limit
	return x
}

func (x *pollingScene) play() error {
	return errors.New("pollingScene without mixin can no be used")
}

// String returns explain text
func (x *pollingScene) string() string {
	return "pollingScene"
}

func (x *baseScene) setGeneralprobe(gp *Generalprobe) { x.gp = gp }
func (x *baseScene) region() string                   { return x.gp.awsRegion }
func (x *baseScene) awsSession() *session.Session     { return x.gp.awsSession }
func (x *baseScene) startTime() time.Time             { return x.gp.StartTime }
func (x *baseScene) lookupPhysicalID(logicalID string) string {
	return x.gp.LookupID(logicalID)
}
