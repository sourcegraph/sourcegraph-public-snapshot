package spec

type BuildSpec struct {
	Image  string          `yaml:"image"`
	Source BuildSourceSpec `yaml:"source"`
}

func (s BuildSpec) Validate() []error {
	var errs []error
	// TODO: Add validation
	return errs
}

type BuildSourceSpec struct {
	Repo string `yaml:"repo"`
	Dir  string `yaml:"dir"`
}
