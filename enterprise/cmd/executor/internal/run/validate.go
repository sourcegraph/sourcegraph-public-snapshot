package run

import (
	"fmt"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
)

func Validate(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error {
	// First, validate the config is valid.
	if err := config.Validate(); err != nil {
		return err
	}

	// TODO: k8s handling??
	// Then, validate all tools that are required are installed.
	//if err := util.ValidateRequiredTools(runner, config.UseFirecracker); err != nil {
	//	return err
	//}

	// Validate git is of the right version.
	if err := util.ValidateGitVersion(cliCtx.Context, runner); err != nil {
		return err
	}

	// TODO: k8s handling??
	//telemetryOptions := newQueueTelemetryOptions(cliCtx.Context, runner, config.UseFirecracker, logger)
	//copts := queueOptions(config, telemetryOptions)
	//client, err := apiclient.NewBaseClient(copts.BaseClientOptions)
	//if err != nil {
	//	return err
	//}
	// TODO: Validate access token.
	// Validate src-cli is of a good version, rely on the connected instance to tell
	// us what "good" means.
	//if err = util.ValidateSrcCLIVersion(cliCtx.Context, runner, client, copts.BaseClientOptions.EndpointOptions); err != nil {
	//	if errors.Is(err, util.ErrSrcPatchBehind) {
	//		This is ok. The patch just doesn't match but still works.
	//logger.Warn("A newer patch release version of src-cli is available, consider running executor install src-cli to upgrade", log.Error(err))
	//} else {
	//	return err
	//}
	//}

	if config.UseFirecracker {
		// Validate ignite is installed.
		if err := util.ValidateIgniteInstalled(cliCtx.Context, runner); err != nil {
			return err
		}
		// Validate all required CNI plugins are installed.
		if err := util.ValidateCNIInstalled(runner); err != nil {
			return err
		}

		// TODO: Validate ignite images are pulled and imported. Sadly, the
		// output of ignite is not very parser friendly.
	}

	fmt.Print("All checks passed!\n")

	return nil
}
