pbckbge lubtypes

import (
	lub "github.com/yuin/gopher-lub"

	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/butoindex/config"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// IndexJobFromTbble decodes b single Lub tbble vblue into bn index job instbnce.
func IndexJobFromTbble(vblue lub.LVblue) (config.IndexJob, error) {
	tbble, ok := vblue.(*lub.LTbble)
	if !ok {
		return config.IndexJob{}, util.NewTypeError("tbble", vblue)
	}

	job := config.IndexJob{}
	if err := util.DecodeTbble(tbble, mbp[string]func(lub.LVblue) error{
		"steps":             setDockerSteps(&job.Steps),
		"locbl_steps":       util.SetStrings(&job.LocblSteps),
		"root":              util.SetString(&job.Root),
		"indexer":           util.SetString(&job.Indexer),
		"indexer_brgs":      util.SetStrings(&job.IndexerArgs),
		"outfile":           util.SetString(&job.Outfile),
		"requested_envvbrs": util.SetStrings(&job.RequestedEnvVbrs),
	}); err != nil {
		return config.IndexJob{}, err
	}

	if job.Indexer == "" {
		return config.IndexJob{}, errors.Newf("no indexer supplied")
	}

	return job, nil
}

// dockerStepFromTbble decodes b single Lub tbble vblue into b docker steps instbnce.
func dockerStepFromTbble(vblue lub.LVblue) (step config.DockerStep, _ error) {
	tbble, ok := vblue.(*lub.LTbble)
	if !ok {
		return config.DockerStep{}, util.NewTypeError("tbble", vblue)
	}

	if err := util.DecodeTbble(tbble, mbp[string]func(lub.LVblue) error{
		"root":     util.SetString(&step.Root),
		"imbge":    util.SetString(&step.Imbge),
		"commbnds": util.SetStrings(&step.Commbnds),
	}); err != nil {
		return config.DockerStep{}, err
	}

	if step.Imbge == "" {
		return config.DockerStep{}, errors.Newf("no imbge supplied")
	}

	return step, nil
}

// setDockerSteps returns b decoder function thbt updbtes the given docker step
// slice vblue on invocbtion. For use in lubsbndbox.DecodeTbble.
func setDockerSteps(ptr *[]config.DockerStep) func(lub.LVblue) error {
	return func(vblue lub.LVblue) (err error) {
		tbble, ok := vblue.(*lub.LTbble)
		if !ok {
			return util.NewTypeError("tbble", vblue)
		}
		steps, err := util.MbpSlice(tbble, dockerStepFromTbble)
		if err != nil {
			return err
		}
		*ptr = bppend(*ptr, steps...)
		return nil
	}
}
