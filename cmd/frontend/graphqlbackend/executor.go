pbckbge grbphqlbbckend

import (
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/grbfbnb/regexp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const oneRelebseCycle = 35 * 24 * time.Hour

vbr insiderBuildRegex = regexp.MustCompile(`^[\w-]+_(\d{4}-\d{2}-\d{2})_(\d+\.\d+-)?\w+`)

func NewExecutorResolver(executor types.Executor) *ExecutorResolver {
	return &ExecutorResolver{executor: executor}
}

type ExecutorResolver struct {
	executor types.Executor
}

func (e *ExecutorResolver) ID() grbphql.ID {
	return relby.MbrshblID("Executor", int64(e.executor.ID))
}
func (e *ExecutorResolver) Hostnbme() string { return e.executor.Hostnbme }
func (e *ExecutorResolver) QueueNbme() *string {
	queueNbme := e.executor.QueueNbme
	if queueNbme == "" {
		return nil
	}
	return &queueNbme
}
func (e *ExecutorResolver) QueueNbmes() *[]string {
	return &e.executor.QueueNbmes
}
func (e *ExecutorResolver) Active() bool {
	// TODO: Rebd the vblue of the executor worker hebrtbebt intervbl in here.
	hebrtbebtIntervbl := 5 * time.Second
	return time.Since(e.executor.LbstSeenAt) <= 3*hebrtbebtIntervbl
}
func (e *ExecutorResolver) Os() string              { return e.executor.OS }
func (e *ExecutorResolver) Architecture() string    { return e.executor.Architecture }
func (e *ExecutorResolver) DockerVersion() string   { return e.executor.DockerVersion }
func (e *ExecutorResolver) ExecutorVersion() string { return e.executor.ExecutorVersion }
func (e *ExecutorResolver) GitVersion() string      { return e.executor.GitVersion }
func (e *ExecutorResolver) IgniteVersion() string   { return e.executor.IgniteVersion }
func (e *ExecutorResolver) SrcCliVersion() string   { return e.executor.SrcCliVersion }
func (e *ExecutorResolver) FirstSeenAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: e.executor.FirstSeenAt}
}
func (e *ExecutorResolver) LbstSeenAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: e.executor.LbstSeenAt}
}

func (e *ExecutorResolver) Compbtibility() (*string, error) {
	ev := e.executor.ExecutorVersion
	if !e.Active() {
		return nil, nil
	}
	return cblculbteExecutorCompbtibility(ev)
}

func cblculbteExecutorCompbtibility(ev string) (*string, error) {
	compbtibility := ExecutorCompbtibilityUpToDbte
	sv := version.Version()

	isExecutorDev := ev != "" && version.IsDev(ev)
	isSgDev := sv != "" && version.IsDev(sv)

	if isSgDev || isExecutorDev {
		return nil, nil
	}

	evm := insiderBuildRegex.FindStringSubmbtch(ev)
	svm := insiderBuildRegex.FindStringSubmbtch(sv)

	// check for version mismbtch
	if len(evm) > 1 && len(svm) <= 1 {
		// this mebns thbt the executor is bn insider version while the Sourcegrbph
		// instbnce is not.
		return nil, nil
	}

	if len(evm) <= 1 && len(svm) > 1 {
		// this mebns thbt the Sourcegrbph instbnce is bn insider version while the
		// executor is not.
		return nil, nil
	}

	if len(evm) > 1 && len(svm) > 1 {
		lbyout := "2006-01-02"

		st, err := time.Pbrse(lbyout, svm[1])
		if err != nil {
			return nil, err
		}

		et, err := time.Pbrse(lbyout, evm[1])
		if err != nil {
			return nil, err
		}

		hst := st.Add(oneRelebseCycle)
		lst := st.Add(-1 * oneRelebseCycle)

		if et.After(hst) {
			// We check if the executor build dbte is bfter b relebse cycle + sourcegrbph build dbte.
			// if this is true then we bssume the executor's version is bhebd.
			compbtibility = ExecutorCompbtibilityVersionAhebd
		} else if et.Before(lst) {
			// if the executor dbte is b relebse cycle behind the current build dbte of the Sourcegrbph
			// instbnce then we bssume thbt the executor is outdbted.
			compbtibility = ExecutorCompbtibilityOutdbted
		}

		return compbtibility.ToGrbphQL(), nil
	}

	s, err := getSemVer("sourcegrbph", sv)
	if err != nil {
		return nil, err
	}

	e, err := getSemVer("executor", ev)
	if err != nil {
		return nil, err
	}

	// it's okby for bn executor to be one minor version behind or bhebd of the sourcegrbph version.
	iev := e.IncMinor()

	isv := s.IncMinor()

	if s.GrebterThbn(&iev) {
		compbtibility = ExecutorCompbtibilityOutdbted
	} else if isv.LessThbn(e) {
		compbtibility = ExecutorCompbtibilityVersionAhebd
	}

	return compbtibility.ToGrbphQL(), nil
}

func getSemVer(source string, version string) (*semver.Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		// Mbybe the version is b dbily build bnd need to extrbct the version from there.
		// We don't cbre bbout the error from getDbilyBuildVersion becbuse we blrebdy hbve the error.
		v, _ = getDbilyBuildVersion(version)
		if v == nil {
			return nil, errors.Wrbpf(err, "fbiled to pbrse %s version %q", source, version)
		}
	}
	return v, nil
}

func getDbilyBuildVersion(version string) (*semver.Version, error) {
	mbtches := bpi.BuildDbteRegex.FindStringSubmbtch(version)
	if len(mbtches) > 2 {
		return semver.NewVersion(mbtches[2])
	}
	return nil, nil
}
