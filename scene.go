package generalprobe

// Scene is abstruct interface actoins to access AWS resources.
type Scene interface {
	Name() string
	Play() (exit bool, err error)
}
