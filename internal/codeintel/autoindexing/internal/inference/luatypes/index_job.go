package luatypes

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// IndexJobFromTable decodes a single Lua table value into an index job instance.
func IndexJobFromTable(value lua.LValue) (config.AutoIndexJobSpec, error) {
	table, ok := value.(*lua.LTable)
	if !ok {
		return config.AutoIndexJobSpec{}, util.NewTypeError("table", value)
	}

	job := config.AutoIndexJobSpec{}
	if err := util.DecodeTable(table, map[string]func(lua.LValue) error{
		"steps":             setDockerSteps(&job.Steps),
		"local_steps":       util.SetStrings(&job.LocalSteps),
		"root":              util.SetString(&job.Root),
		"indexer":           util.SetString(&job.Indexer),
		"indexer_args":      util.SetStrings(&job.IndexerArgs),
		"outfile":           util.SetString(&job.Outfile),
		"requested_envvars": util.SetStrings(&job.RequestedEnvVars),
	}); err != nil {
		return config.AutoIndexJobSpec{}, err
	}

	if job.Indexer == "" {
		return config.AutoIndexJobSpec{}, errors.Newf("no indexer supplied")
	}

	return job, nil
}

// dockerStepFromTable decodes a single Lua table value into a docker steps instance.
func dockerStepFromTable(value lua.LValue) (step config.DockerStep, _ error) {
	table, ok := value.(*lua.LTable)
	if !ok {
		return config.DockerStep{}, util.NewTypeError("table", value)
	}

	if err := util.DecodeTable(table, map[string]func(lua.LValue) error{
		"root":     util.SetString(&step.Root),
		"image":    util.SetString(&step.Image),
		"commands": util.SetStrings(&step.Commands),
	}); err != nil {
		return config.DockerStep{}, err
	}

	if step.Image == "" {
		return config.DockerStep{}, errors.Newf("no image supplied")
	}

	return step, nil
}

// setDockerSteps returns a decoder function that updates the given docker step
// slice value on invocation. For use in luasandbox.DecodeTable.
func setDockerSteps(ptr *[]config.DockerStep) func(lua.LValue) error {
	return func(value lua.LValue) (err error) {
		table, ok := value.(*lua.LTable)
		if !ok {
			return util.NewTypeError("table", value)
		}
		steps, err := util.MapSlice(table, dockerStepFromTable)
		if err != nil {
			return err
		}
		*ptr = append(*ptr, steps...)
		return nil
	}
}
