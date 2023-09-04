package spec

type BuildSpec struct {
	Image  string          `json:"image"`
	Source BuildSourceSpec `json:"source"`
}

func (s BuildSpec) Validate() []error {
	var errs []error
	// TODO: Add validation
	return errs
}

type BuildSourceSpec struct {
	Repo string `json:"repo"`
	Dir  string `json:"dir"`
}
