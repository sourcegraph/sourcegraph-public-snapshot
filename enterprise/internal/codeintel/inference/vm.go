package inference

import (
	"github.com/robertkrimen/otto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	ErrInvalidScript error = errors.New("invalid script")
)

func EvaluateInferenceRule(rule string) ([]config.IndexJob, error) {
	vm := otto.New()

	if err := prepareVM(vm); err != nil {
		return nil, err
	}
	// TODO: Precompile script so we don't need to parse it multiple times.

	result, err := vm.Run(rule)
	if err != nil {
		return nil, errors.Wrap(err, "running rule script")
	}

	keys := result.Object().Keys()
	var jobs = make([]config.IndexJob, 0, len(keys))
	for _, ruleKey := range keys {
		r, err := result.Object().Get(ruleKey)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't read object")
		}
		job := config.IndexJob{}
		job.Indexer, err = readString(r.Object().Get("indexer"))
		if err != nil {
			return nil, err
		}
		job.Outfile, err = readString(r.Object().Get("outfile"))
		if err != nil {
			return nil, err
		}
		job.Root, err = readString(r.Object().Get("root"))
		if err != nil {
			return nil, err
		}
		job.IndexerArgs, err = readStringSlice(r.Object().Get("indexer_args"))
		if err != nil {
			return nil, err
		}
		job.LocalSteps, err = readStringSlice(r.Object().Get("local_steps"))
		if err != nil {
			return nil, err
		}
		job.Steps, err = readDockerStepSlice(r.Object().Get("steps"))
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func readString(value otto.Value, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if value.IsUndefined() || value.IsNull() {
		return "", nil
	}
	if !value.IsString() {
		return "", errors.New("not a string")
	}

	sV, err := value.ToString()
	if err != nil {
		return "", err
	}

	return sV, nil
}

func readStringSlice(value otto.Value, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}

	if value.IsUndefined() || value.IsNull() {
		return []string{}, nil
	}
	if !value.IsObject() {
		return nil, errors.New("not an object")
	}

	keys := value.Object().Keys()
	values := make([]string, 0, len(keys))

	for _, k := range keys {
		v, err := readString(value.Object().Get(k))
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}

func readDockerStepSlice(value otto.Value, err error) ([]config.DockerStep, error) {
	if err != nil {
		return nil, err
	}

	if value.IsUndefined() || value.IsNull() {
		return []config.DockerStep{}, nil
	}
	if !value.IsObject() {
		return nil, errors.New("not an object")
	}

	keys := value.Object().Keys()
	values := make([]config.DockerStep, 0, len(keys))

	for _, k := range keys {
		v, err := readDockerStep(value.Object().Get(k))
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}

func readDockerStep(value otto.Value, err error) (config.DockerStep, error) {
	ds := config.DockerStep{}
	if err != nil {
		return ds, err
	}

	if !value.IsObject() {
		return ds, errors.New("not an object")
	}

	ds.Root, err = readString(value.Object().Get("root"))
	if err != nil {
		return ds, err
	}

	ds.Image, err = readString(value.Object().Get("image"))
	if err != nil {
		return ds, err
	}
	ds.Commands, err = readStringSlice(value.Object().Get("commands"))
	if err != nil {
		return ds, err
	}

	return ds, nil
}

func prepareVM(vm *otto.Otto) error {
	vm.Set("lsFiles", func(call otto.FunctionCall) otto.Value {
		// TODO: Use arg.
		_, err := readString(call.Argument(0), nil)
		if err != nil {
			errRes, err2 := vm.ToValue("invalid argument, want string")
			if err2 != nil {
				panic(err2)
			}
			return errRes
		}
		result, err := vm.ToValue([]string{"/erick/go.mod"})
		if err != nil {
			panic(err)
		}
		return result
	})

	return nil
}
