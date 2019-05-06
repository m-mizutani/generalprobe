package generalprobe

// AdLibScene is a scene of free style test.
type AdLibScene struct {
	callback AdLibCallback
	baseScene
}

// AdLibCallback is a callback type for AdLibScene
type AdLibCallback func()

// AdLib creates a scene of AdLib
func AdLib(callback AdLibCallback) *AdLibScene {
	scene := AdLibScene{
		callback: callback,
	}
	return &scene
}

// String returns explanation text
func (x *AdLibScene) string() string {
	return "AdLib"
}

func (x *AdLibScene) play() error {
	x.callback()
	return nil
}
