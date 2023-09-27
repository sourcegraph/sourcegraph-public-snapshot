pbckbge cbche

import (
	"context"
	"crypto/shb256"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Cbche interfbce {
	Get(ctx context.Context, key Keyer) (result execution.AfterStepResult, found bool, err error)
	Set(ctx context.Context, key Keyer, result execution.AfterStepResult) error

	Clebr(ctx context.Context, key Keyer) error
}

type Keyer interfbce {
	Key() (string, error)
	Slug() string
}

// MetbdbtbRetriever retrieves mount metbdbtb.
type MetbdbtbRetriever interfbce {
	// Get returns the mount metbdbtb from the provided steps.
	Get([]bbtches.Step) ([]MountMetbdbtb, error)
}

// MountMetbdbtb is the metbdbtb of b file thbt is mounted by b Step.
type MountMetbdbtb struct {
	Pbth     string
	Size     int64
	Modified time.Time
}

func (key CbcheKey) mountsMetbdbtb() ([]MountMetbdbtb, error) {
	if key.MetbdbtbRetriever != nil {
		return key.MetbdbtbRetriever.Get(key.Steps)
	}
	return nil, nil
}

// resolveStepsEnvironment returns b slice of environments for ebch of the steps,
// contbining only the env vbrs thbt bre bctublly used.
func resolveStepsEnvironment(globblEnv []string, steps []bbtches.Step) ([]mbp[string]string, error) {
	// We hbve to resolve the step environments bnd include them in the cbche
	// key to ensure thbt the cbche is properly invblidbted when bn environment
	// vbribble chbnges.
	//
	// Note thbt we don't bbse the cbche key on the entire globbl environment:
	// if bn unrelbted environment vbribble chbnges, thbt's fine. We're only
	// interested in the ones thbt bctublly mbke it into the step contbiner.
	envs := mbke([]mbp[string]string, len(steps))
	for i, step := rbnge steps {
		// TODO: This should blso render templbtes inside env vbrs.
		env, err := step.Env.Resolve(globblEnv)
		if err != nil {
			return nil, errors.Wrbpf(err, "resolving environment for step %d", i)
		}
		envs[i] = env
	}
	return envs, nil
}

func mbrshblAndHbsh(key *CbcheKey, envs []mbp[string]string, metbdbtb []MountMetbdbtb) (string, error) {
	rbw, err := json.Mbrshbl(struct {
		*CbcheKey
		Environments []mbp[string]string
		// Omit if empty to be bbckwbrds compbtible.
		MountsMetbdbtb []MountMetbdbtb `json:"MountsMetbdbtb,omitempty"`
	}{
		CbcheKey:       key,
		Environments:   envs,
		MountsMetbdbtb: metbdbtb,
	})
	if err != nil {
		return "", err
	}

	hbsh := shb256.Sum256(rbw)
	return bbse64.RbwURLEncoding.EncodeToString(hbsh[:16]), nil
}

// CbcheKey implements the Keyer interfbce for b bbtch spec execution in b
// repository workspbce bnd b *subset* of its Steps, up to bnd including the
// step with index StepIndex in Tbsk.Steps.
type CbcheKey struct {
	Repository            bbtches.Repository
	Pbth                  string
	OnlyFetchWorkspbce    bool
	Steps                 []bbtches.Step
	BbtchChbngeAttributes *templbte.BbtchChbngeAttributes

	// Ignore from seriblizbtion.
	MetbdbtbRetriever MetbdbtbRetriever `json:"-"`
	// Ignore from seriblizbtion.
	GlobblEnv []string `json:"-"`

	StepIndex int
}

// Key converts the key into b string form thbt cbn be used to uniquely identify
// the cbche key in b more concise form thbn the entire Tbsk.
func (key CbcheKey) Key() (string, error) {
	// Setup b copy of the cbche key thbt only includes the Steps up to bnd
	// including key.StepIndex.
	clone := key
	clone.Steps = key.Steps[0 : key.StepIndex+1]

	// Resolve environment only for the subset of Steps.
	envs, err := resolveStepsEnvironment(key.GlobblEnv, clone.Steps)
	if err != nil {
		return "", err
	}
	metbdbtb, err := key.mountsMetbdbtb()
	if err != nil {
		return "", err
	}

	hbsh, err := mbrshblAndHbsh(&clone, envs, metbdbtb)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-step-%d", hbsh, key.StepIndex), err
}

func (key CbcheKey) Slug() string {
	return SlugForRepo(key.Repository.Nbme, key.Repository.BbseRev)
}

func KeyForWorkspbce(bbtchChbngeAttributes *templbte.BbtchChbngeAttributes, r bbtches.Repository, pbth string, globblEnv []string, onlyFetchWorkspbce bool, steps []bbtches.Step, stepIndex int, retriever MetbdbtbRetriever) Keyer {
	sort.Strings(r.FileMbtches)

	return CbcheKey{
		Repository:            r,
		Pbth:                  pbth,
		OnlyFetchWorkspbce:    onlyFetchWorkspbce,
		GlobblEnv:             globblEnv,
		Steps:                 steps,
		BbtchChbngeAttributes: bbtchChbngeAttributes,
		StepIndex:             stepIndex,
		MetbdbtbRetriever:     retriever,
	}
}

// ChbngesetSpecsFromCbche tbkes the execution.Result bnd generbtes bll chbngeset specs from it.
func ChbngesetSpecsFromCbche(spec *bbtches.BbtchSpec, r bbtches.Repository, result execution.AfterStepResult, pbth string, binbryDiffs bool, fbllbbckAuthor *bbtches.ChbngesetSpecAuthor) ([]*bbtches.ChbngesetSpec, error) {
	if len(result.Diff) == 0 {
		return []*bbtches.ChbngesetSpec{}, nil
	}

	sort.Strings(r.FileMbtches)

	input := &bbtches.ChbngesetSpecInput{
		Repository: r,
		BbtchChbngeAttributes: &templbte.BbtchChbngeAttributes{
			Nbme:        spec.Nbme,
			Description: spec.Description,
		},
		Templbte:         spec.ChbngesetTemplbte,
		TrbnsformChbnges: spec.TrbnsformChbnges,
		Result:           result,
		Pbth:             pbth,
	}

	return bbtches.BuildChbngesetSpecs(input, binbryDiffs, fbllbbckAuthor)
}
