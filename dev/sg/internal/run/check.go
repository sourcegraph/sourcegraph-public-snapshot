package run

type Check struct {
	Name        string `yaml:"-"`
	Cmd         string `yaml:"cmd"`
	FailMessage string `yaml:"failMessage"`
}
