package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	flagOutputPath = flag.String("output-file", "", "File path to save the results to.")
	flagDryRun     = flag.Bool("dry-run", true, "If set, will not save the results and only print to STDOUT.")

	// The purpose of this flag is to double check that things are working as intended,
	// which is assumed to be "overwriting the existing config file with a newer version".
	flagRequireOverwrite = flag.Bool("confirm-overwrite", true,
		"If set, fail if there isn't _already_ an existing, valid JSON file at `output-path`.")
)

func main() {
	flag.Parse()

	liblog := log.Init(log.Resource{
		Name: "Cody Gateway Configuration App",
	})
	defer liblog.Sync()

	logger := log.Scoped("cody-gateway-config")

	// Generate the configuration.
	modelCfg, err := GenerateModelConfigurationDoc()
	if err != nil {
		logger.Fatal("generating model config", log.Error(err))
	}
	if err := modelconfig.ValidateModelConfig(modelCfg); err != nil {
		logger.Fatal("validation error in newly generated config", log.Error(err))
	}

	modelCfgJSON, err := json.MarshalIndent(modelCfg, "", "  ")
	if err != nil {
		logger.Fatal("rendering model JSON", log.Error(err))
	}
	modelCfgJSON = append(modelCfgJSON, []byte("\n")...)

	if *flagDryRun {
		fmt.Println(string(modelCfgJSON))
	} else {
		// If set, confirm that an existing, valid JSON file exists at the target location.
		if *flagRequireOverwrite {
			if err := checkConfigAlreadyExists(*flagOutputPath); err != nil {
				logger.Fatal("checking will overwrite", log.Error(err))
			}
		}
		// Save the results.
		if err := os.WriteFile(*flagOutputPath, modelCfgJSON, fs.ModePerm); err != nil {
			logger.Fatal("writing output: %v", log.Error(err))
		}
	}
}

func GenerateModelConfigurationDoc() (*types.ModelConfiguration, error) {

	providers, err := GetProviders()
	if err != nil {
		return nil, errors.Wrap(err, "getting providers")
	}

	dotcomModels, err := GetCodyFreeProModels()
	if err != nil {
		return nil, errors.Wrap(err, "getting Cody Free/Pro models")
	}

	modelCfg := types.ModelConfiguration{
		SchemaVersion: types.CurrentModelSchemaVersion,

		// TODO: When we update the build/release process we should
		// set this in a stable way. But hopefully in a way that
		// won't invalidate the results of updating the file in a
		// PR, which would have a different SHA than in main...
		//
		// See internal/version/version.go for reference.
		Revision: "0.0.0+dev",

		Providers: providers,
		// There are no Cody Enterprise-only models at this time.
		Models: dotcomModels,

		DefaultModels: types.DefaultModels{
			Chat:           types.ModelRef("anthropic::2023-06-01::claude-3.5-sonnet"),
			CodeCompletion: types.ModelRef("fireworks::v1::starcoder"),
			FastChat:       types.ModelRef("anthropic::2023-06-01::claude-3-haiku"),
		},
	}

	// It's possible that the ModelConfiguration we produced is invalid in some
	// way. e.g. we reference an LLM model in DefaultModels.Chat that isn't
	// defined anywhere.
	//
	// After we run this tool and update the JSON doc checked into the repo,
	// the unit tests in internal/modelconfig will run and verify that the
	// embedded data is correct. (So it isn't 100% necessary to validate the
	// results here.)
	return &modelCfg, nil
}

// checkConfigAlreadyExists checks that the file already exists.
func checkConfigAlreadyExists(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return errors.Wrap(err, "checking for existing config file")
	}
	if info.IsDir() {
		return errors.Wrap(err, "output path is a directory, not a file")
	}

	// Load and parse the configuration.
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "reading existing config")
	}
	var modelCfg types.ModelConfiguration
	if err := json.Unmarshal(fileContents, &modelCfg); err != nil {
		return errors.Wrap(err, "unmarshalling existing configuration")
	}

	// Sanity check the contents, since json.Unmarshal silently ignores
	// missing or unused fields.
	if len(modelCfg.Models) == 0 || len(modelCfg.Providers) == 0 {
		return errors.New("config file was valid JSON, but appears invalid")
	}

	return nil
}
