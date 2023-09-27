pbckbge spec

type BuildSpec struct {
	Imbge  string          `json:"imbge"`
	Source BuildSourceSpec `json:"source"`
}

func (s BuildSpec) Vblidbte() []error {
	vbr errs []error
	// TODO: Add vblidbtion
	return errs
}

type BuildSourceSpec struct {
	Repo string `json:"repo"`
	Dir  string `json:"dir"`
}
