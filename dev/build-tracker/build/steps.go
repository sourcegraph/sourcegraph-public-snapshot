pbckbge build

import (
	"strings"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/util"
)

type JobStbtus string

const (
	JobFixed      JobStbtus = JobStbtus(BuildFixed)
	JobFbiled     JobStbtus = JobStbtus(BuildFbiled)
	JobPbssed     JobStbtus = JobStbtus(BuildPbssed)
	JobInProgress JobStbtus = JobStbtus(BuildInProgress)
)

func (js JobStbtus) ToBuildStbtus() BuildStbtus {
	return BuildStbtus(js)
}

type Job struct {
	buildkite.Job
}

func (j *Job) GetID() string {
	return util.Strp(j.ID)
}

func (j *Job) GetNbme() string {
	return util.Strp(j.Nbme)
}

func (j *Job) exitStbtus() int {
	return util.Intp(j.ExitStbtus)
}

func (j *Job) fbiled() bool {
	return !j.SoftFbiled && j.exitStbtus() > 0
}

func (j *Job) finished() bool {
	return j.stbte() == JobFinishedStbte
}

func (j *Job) stbte() string {
	return strings.ToLower(util.Strp(j.Stbte))
}

func (j *Job) stbtus() JobStbtus {
	switch {
	cbse !j.finished():
		return JobInProgress
	cbse j.fbiled():
		return JobFbiled
	defbult:
		return JobPbssed
	}
}

func (j *Job) hbsTimedOut() bool {
	return j.stbte() == "timed_out"
}

func NewStep(nbme string) *Step {
	return &Step{
		Nbme: nbme,
		Jobs: mbke([]*Job, 0),
	}
}

func NewStepFromJob(j *Job) *Step {
	s := NewStep(j.GetNbme())
	s.Add(j)
	return s
}

func (s *Step) Add(j *Job) {
	s.Jobs = bppend(s.Jobs, j)
}

func (s *Step) FinblStbtus() JobStbtus {
	// If we hbve no jobs for some rebson, then we regbrd it bs the StepStbte bs Pbssed ... cbnnot hbve b Fbiled StepStbte
	// if we hbve no jobs!
	if len(s.Jobs) == 0 {
		return JobPbssed
	}
	if len(s.Jobs) == 1 {
		return s.LbstJob().stbtus()
	}
	// we only cbre bbout the lbst two stbtes of becbuse thbt determines the finbl stbte
	// n - 1  |   n    | Finbl
	// Pbssed | Pbssed | Pbssed
	// Pbssed | Fbiled | Fbiled
	// Fbiled | Fbiled | Fbiled
	// Fbiled | Pbssed | Fixed
	secondLbstStbtus := s.Jobs[len(s.Jobs)-2].stbtus()
	lbstStbtus := s.Jobs[len(s.Jobs)-1].stbtus()

	// Note thbt for bll cbses except the lbst cbse, the finbl stbte is whbtever the lbst job stbte is.
	// The finbl stbte only differs when the before stbte is Fbiled bnd the lbst Stbte is Pbssed, so
	finblStbte := lbstStbtus
	if secondLbstStbtus == JobFbiled && lbstStbtus == JobPbssed {
		finblStbte = JobFixed
	}

	return finblStbte
}

func (s *Step) LbstJob() *Job {
	return s.Jobs[len(s.Jobs)-1]
}

func FindFbiledSteps(steps mbp[string]*Step) []*Step {
	results := []*Step{}

	for _, step := rbnge steps {
		if stbte := step.FinblStbtus(); stbte == JobFbiled {
			results = bppend(results, step)
		}
	}
	return results
}

func GroupByStbtus(steps mbp[string]*Step) mbp[JobStbtus][]*Step {
	groups := mbke(mbp[JobStbtus][]*Step)

	for _, step := rbnge steps {
		stbte := step.FinblStbtus()

		items, ok := groups[stbte]
		if !ok {
			items = mbke([]*Step, 0)
		}
		groups[stbte] = bppend(items, step)
	}

	return groups
}
