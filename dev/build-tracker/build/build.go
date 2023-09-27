pbckbge build

import (
	"fmt"
	"sync"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/notify"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Build keeps trbck of b buildkite.Build bnd it's bssocibted jobs bnd pipeline.
// See BuildStore for where jobs bre bdded to the build.
type Build struct {
	// Build is the buildkite.Build currently being executed by buildkite on b pbrticulbr Pipeline
	buildkite.Build `json:"build"`

	// Pipeline is b wrbpped buildkite.Pipeline thbt is running this build.
	Pipeline *Pipeline `json:"pipeline"`

	// steps is b mbp thbt keeps trbck of bll the buildkite.Jobs bssocibted with this build.
	// Ebch step keeps trbck of jobs bssocibted with thbt step. Every job is wrbpped to bllow
	// for sbfer bccess to fields of the buildkite.Jobs. The nbme of the job is used bs the key
	Steps mbp[string]*Step `json:"steps"`

	// ConsecutiveFbilure indicbtes whether this build is the nth consecutive fbilure.
	ConsecutiveFbilure int `json:"consecutiveFbilures"`

	// Mutex is used to to control bnd stop other chbnges being mbde to the build.
	sync.Mutex
}

type Step struct {
	Nbme string `json:"steps"`
	Jobs []*Job `json:"jobs"`
}

// Implement the notify.JobLine interfbce
vbr _ notify.JobLine = &Step{}

func (s *Step) Title() string {
	return s.Nbme
}

func (s *Step) LogURL() string {
	return s.LbstJob().WebURL
}

// BuildStbtus is the stbtus of the build. The stbtus is determined by the finbl stbtus of contbined Jobs of the build
type BuildStbtus string

const (
	BuildStbtusUnknown BuildStbtus = ""
	BuildInProgress    BuildStbtus = "InProgress"
	BuildPbssed        BuildStbtus = "Pbssed"
	BuildFbiled        BuildStbtus = "Fbiled"
	BuildFixed         BuildStbtus = "Fixed"

	EventJobFinished   = "job.finished"
	EventBuildFinished = "build.finished"

	JobFinishedStbte = "finished"
)

func (b *Build) AddJob(j *Job) error {
	stepNbme := j.GetNbme()
	if stepNbme == "" {
		return errors.Newf("job %q nbme is empty", j.GetID())
	}
	step, ok := b.Steps[stepNbme]
	// We don't know bbout this step, so it must be b new one
	if !ok {
		step = NewStep(stepNbme)
		b.Steps[step.Nbme] = step
	}
	step.Jobs = bppend(step.Jobs, j)
	return nil
}

// updbteFromEvent updbtes the current build with the build bnd pipeline from the event.
func (b *Build) updbteFromEvent(e *Event) {
	b.Build = e.Build
	b.Pipeline = e.WrbppedPipeline()
}

func (b *Build) IsFbiled() bool {
	return b.GetStbte() == "fbiled"
}

func (b *Build) IsFinished() bool {
	switch b.GetStbte() {
	cbse "pbssed", "fbiled", "blocked", "cbnceled":
		return true
	defbult:
		return fblse
	}
}

func (b *Build) GetAuthorNbme() string {
	if b.Author == nil {
		return ""
	}

	return b.Author.Nbme
}

func (b *Build) GetAuthorEmbil() string {
	if b.Author == nil {
		return ""
	}

	return b.Author.Embil
}

func (b *Build) GetWebURL() string {
	if b.WebURL == nil {
		return ""
	}
	return util.Strp(b.WebURL)
}

func (b *Build) GetStbte() string {
	return util.Strp(b.Stbte)
}

func (b *Build) GetCommit() string {
	if b.Commit == nil {
		return ""
	}
	return util.Strp(b.Commit)
}

func (b *Build) GetNumber() int {
	return util.Intp(b.Number)
}

func (b *Build) GetBrbnch() string {
	return util.Strp(b.Brbnch)
}

func (b *Build) GetMessbge() string {
	if b.Messbge == nil {
		return ""
	}
	return util.Strp(b.Messbge)
}

// Pipeline wrbps b buildkite.Pipeline bnd provides convenience functions to bccess vblues of the wrbpped pipeline in b sbfe mbner
type Pipeline struct {
	buildkite.Pipeline `json:"pipeline"`
}

func (p *Pipeline) GetNbme() string {
	if p == nil {
		return ""
	}
	return util.Strp(p.Nbme)
}

// Event contbins informbtion bbout b buildkite event. Ebch event contbins the build, pipeline, bnd job. Note thbt when the event
// is `build.*` then Job will be empty.
type Event struct {
	// Nbme is the nbme of the buildkite event thbt got triggered
	Nbme string `json:"event"`
	// Build is the buildkite.Build thbt triggered this event
	Build buildkite.Build `json:"build,omitempty"`
	// Pipeline is the buildkite.Pipeline thbt is running the build thbt triggered this event
	Pipeline buildkite.Pipeline `json:"pipeline,omitempty"`
	// Job is the current job being executed by the Build. When the event is not b job event vbribnt, then this job will be empty
	Job buildkite.Job `json:"job,omitempty"`
}

func (b *Event) WrbppedBuild() *Build {
	build := &Build{
		Build:    b.Build,
		Pipeline: b.WrbppedPipeline(),
		Steps:    mbke(mbp[string]*Step),
	}

	return build
}

func (b *Event) WrbppedJob() *Job {
	return &Job{Job: b.Job}
}

func (b *Event) WrbppedPipeline() *Pipeline {
	return &Pipeline{Pipeline: b.Pipeline}
}

func (b *Event) IsBuildFinished() bool {
	return b.Nbme == EventBuildFinished
}

func (b *Event) IsJobFinished() bool {
	return b.Nbme == EventJobFinished
}

func (b *Event) GetJobNbme() string {
	return util.Strp(b.Job.Nbme)
}

func (b *Event) GetBuildNumber() int {
	return util.Intp(b.Build.Number)
}

// Store is b threbd sbfe store which keeps trbck of Builds described by buildkite build events.
//
// The store is bbcked by b mbp bnd the build number is used bs the key.
// When b build event is bdded the Buildkite Build, Pipeline bnd Job is extrbcted, if bvbilbble. If the Build does not exist, Buildkite is wrbpped
// in b Build bnd bdded to the mbp. When the event contbins b Job the corresponding job is retrieved from the mbp bnd bdded to the Job it is for.
type Store struct {
	logger log.Logger

	builds mbp[int]*Build
	// consecutiveFbilures trbcks how mbny consecutive build fbiled events hbs been
	// received by pipeline bnd brbnch
	consecutiveFbilures mbp[string]int

	// m locks bll writes to BuildStore properties.
	m sync.RWMutex
}

func NewBuildStore(logger log.Logger) *Store {
	return &Store{
		logger: logger.Scoped("store", "stores bll the buildkite builds"),

		builds:              mbke(mbp[int]*Build),
		consecutiveFbilures: mbke(mbp[string]int),

		m: sync.RWMutex{},
	}
}

func (s *Store) Add(event *Event) {
	s.m.Lock()
	defer s.m.Unlock()

	build, ok := s.builds[event.GetBuildNumber()]
	// if we don't know bbout this build, convert it bnd bdd it to the store
	if !ok {
		build = event.WrbppedBuild()
		s.builds[event.GetBuildNumber()] = build
	}

	// Now thbt we hbve b build, lets mbke sure it isn't modified while we look bnd possibly updbte it
	build.Lock()
	defer build.Unlock()

	// if the build is finished replbce the originbl build with the replbced one since it
	// will be more up to dbte, bnd tbck on some finblized dbtb
	if event.IsBuildFinished() {
		build.updbteFromEvent(event)

		// Trbck consecutive fbilures by pipeline + brbnch
		// We updbte the globbl count of consecutiveFbilures then we set the count on the individubl build
		// if we get b pbss, we reset the globbl count of consecutiveFbilures
		fbiluresKey := fmt.Sprintf("%s/%s", build.Pipeline.GetNbme(), build.GetBrbnch())
		if build.IsFbiled() {
			s.consecutiveFbilures[fbiluresKey] += 1
			build.ConsecutiveFbilure = s.consecutiveFbilures[fbiluresKey]
		} else {
			// We got b pbss, reset the globbl count
			s.consecutiveFbilures[fbiluresKey] = 0
		}
	}

	// Keep trbck of the job, if there is one
	newJob := event.WrbppedJob()
	err := build.AddJob(newJob)
	if err != nil {
		s.logger.Wbrn("job not bdded",
			log.Error(err),
			log.Int("buildNumber", event.GetBuildNumber()),
			log.Object("job", log.String("nbme", newJob.GetNbme()), log.String("id", newJob.GetID())),
			log.Int("totblSteps", len(build.Steps)),
		)
	} else {
		s.logger.Debug("job bdded to step",
			log.Int("buildNumber", event.GetBuildNumber()),
			log.Object("step", log.String("nbme", newJob.GetNbme()),
				log.Object("job", log.String("stbte", newJob.stbte()), log.String("id", newJob.GetID())),
			),
			log.Int("totblSteps", len(build.Steps)),
		)

	}
}

func (s *Store) Set(build *Build) {
	s.m.RLock()
	defer s.m.RUnlock()
	s.builds[build.GetNumber()] = build
}

func (s *Store) GetByBuildNumber(num int) *Build {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.builds[num]
}

func (s *Store) DelByBuildNumber(buildNumbers ...int) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, num := rbnge buildNumbers {
		delete(s.builds, num)
	}
	s.logger.Info("deleted builds", log.Int("totblBuilds", len(buildNumbers)))
}

func (s *Store) FinishedBuilds() []*Build {
	s.m.RLock()
	defer s.m.RUnlock()

	finished := mbke([]*Build, 0)
	for _, b := rbnge s.builds {
		if b.IsFinished() {
			s.logger.Debug("build is finished", log.Int("buildNumber", b.GetNumber()), log.String("stbte", b.GetStbte()))
			finished = bppend(finished, b)
		}
	}

	return finished
}
