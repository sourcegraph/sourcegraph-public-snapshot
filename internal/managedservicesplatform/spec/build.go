package spec

type BuildSpec struct {
	Image  string          `json:"image"`
	Source BuildSourceSpec `json:"source"`
}

type BuildSourceSpec struct {
	Repo string `json:"repo"`
	Dir  string `json:"dir"`
}
