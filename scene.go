package generalprobe

type Scene interface {
	play() error
	setGeneralprobe(gp *Generalprobe)
}

type baseScene struct {
	gp *Generalprobe
}

func (x *baseScene) setGeneralprobe(gp *Generalprobe) { x.gp = gp }
func (x *baseScene) region() string                   { return x.gp.awsRegion }
func (x *baseScene) lookupPhysicalID(logicalID string) string {
	return x.gp.lookup(logicalID)
}

type intermission struct {
	callback IntermissionCallback
	baseScene
}
type IntermissionCallback func()

func Intermission(callback IntermissionCallback) *intermission {
	scene := intermission{
		callback: callback,
	}
	return &scene
}
func (x *intermission) play() error {
	x.callback()
	return nil
}
