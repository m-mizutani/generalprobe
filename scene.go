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
	return x.gp.LookupID(logicalID)
}

type adLib struct {
	callback AdLibCallback
	baseScene
}
type AdLibCallback func()

func (x *Generalprobe) AdLib(callback AdLibCallback) *adLib {
	scene := adLib{
		callback: callback,
	}
	return &scene
}
func (x *adLib) play() error {
	x.callback()
	return nil
}
