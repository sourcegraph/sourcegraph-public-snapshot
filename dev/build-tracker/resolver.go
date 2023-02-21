package main

type FinalizedBuild struct {
	Passed []*Step
	Failed []*Step
	Fixed  []*Step
	State  StepState
}

func (f *FinalizedBuild) isFailed() bool {
	return f.State == Failed
}

func (f *FinalizedBuild) isFixed() bool {
	return f.State == Fixed
}

func (f *FinalizedBuild) isPassed() bool {
	return f.State == Passed
}

func FinaliseBuild(build *Build) *FinalizedBuild {
	r := FinalizedBuild{}
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

	r.State = Passed
	if len(r.Failed) > 0 {
		r.State = Failed
	}
	if len(r.Fixed) > 0 && len(r.Failed) == 0 {
		r.State = Fixed
	}

	return &r
}
