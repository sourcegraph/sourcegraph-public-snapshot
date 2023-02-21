package main

type BuildResolution struct {
	Passed []*Step
	Failed []*Step
	Fixed  []*Step
	State  string
}

func FinaliseBuild(build *Build) *BuildResolution {
	r := BuildResolution{}
	for _, step := range build.Steps {
		state := step.FinalState()
		switch state {
		case Failed:
			r.Failed = append(r.Failed, step)
		case Fixed:
			r.Fixed = append(r.Fixed, step)
		case Passed:
			r.Passed = append(r.Passed, step)
		}
	}

	r.State = string(Passed)
	if len(r.Failed) > 0 {
		r.State = string(Failed)
	}
	if len(r.Fixed) > 0 && len(r.Failed) == 0 {
		r.State = string(Fixed)
	}

	return &r
}
