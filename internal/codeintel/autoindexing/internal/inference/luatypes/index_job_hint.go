package luatypes

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// IndexJobHintFromTable decodes a single Lua table value into an index job hint instance.
func IndexJobHintFromTable(value lua.LValue) (config.IndexJobHint, error) {
	table, ok := value.(*lua.LTable)
	if !ok {
		return config.IndexJobHint{}, util.NewTypeError("table", value)
	}

	var confidence string
	jobHint := config.IndexJobHint{}
	if err := util.DecodeTable(table, map[string]func(lua.LValue) error{
		"root":       util.SetString(&jobHint.Root),
		"indexer":    util.SetString(&jobHint.Indexer),
		"confidence": util.SetString(&confidence),
	}); err != nil {
		return config.IndexJobHint{}, err
	}

	if jobHint.Indexer == "" {
		return config.IndexJobHint{}, errors.Newf("no indexer supplied")
	}

	switch confidence {
	case "LANGUAGE_SUPPORTED":
		jobHint.HintConfidence = config.HintConfidenceLanguageSupport
	case "PROJECT_STRUCTURE_SUPPORTED":
		jobHint.HintConfidence = config.HintConfidenceProjectStructureSupported

	case "":
		return config.IndexJobHint{}, errors.Newf("no confidence supplied")
	default:
		return config.IndexJobHint{}, errors.Newf("illegal confidence %q supplied", confidence)
	}

	return jobHint, nil
}
