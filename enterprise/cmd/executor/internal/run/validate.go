package run

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RunValidate(cliCtx *cli.Context, logger log.Logger, config *config.Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	// TODO: Validate access token.
	// TODO: Validate git version is >= 2.26
	client := apiclient.New(apiWorkerOptions(config, apiclient.TelemetryOptions{}).ClientOptions, nil, &observation.TestContext)
	if err := validateSrcCLIVersion(cliCtx.Context, logger, client); err != nil {
		return err
	}

	if err := validateCNIInstalled(); err != nil {
		return err
	}

	if err := validateToolsRequired(config.UseFirecracker); err != nil {
		return err
	}

	return nil
}

// validateSrcCLIVersion queries the latest recommended version of src-cli and makes sure it
// matches what is installed. If not, a warning message recommending to use a different
// version is logged.
func validateSrcCLIVersion(ctx context.Context, logger log.Logger, client *apiclient.Client) error {
	latestVersion, err := client.LatestSrcCLIVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve latest compatible src-cli version")
	}
	cmd := exec.CommandContext(ctx, "src", "version", "-client-only")
	out, err := cmd.Output()
	if err != nil {
		logger.Error("failed to get src-cli version, is it installed?", log.Error(err))
	}
	actualVersion := string(out)
	actualVersion = strings.TrimSpace(actualVersion)
	actualVersion = strings.TrimPrefix(actualVersion, "Current version: ")
	if version.IsDev(actualVersion) || actualVersion == "dev" {
		return nil
	}
	actual, err := semver.NewVersion(actualVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse src-cli version")
	}
	latest, err := semver.NewVersion(latestVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse latest src-cli version")
	}
	if actual.LessThan(latest) {
		return errors.Newf("installed src-cli is not the latest recommended version, consider upgrading actual=%s, latest=%s", actual.String(), latest.String())
	} else if actual.Major() != latest.Major() || actual.Minor() != latest.Minor() {
		return errors.Newf("installed src-cli is not the latest recommended version, consider switching actual=%s, recommended=%s", actual.String(), latest.String())
	}

	return nil
}

func validateToolsRequired(useFirecracker bool) error {
	notFoundTools := []string{}
	rt := append([]string{}, config.RequiredCLITools...)
	if useFirecracker {
		rt = append(rt, config.RequiredCLIToolsFirecracker...)
	}
	for _, tool := range rt {
		if found, err := existsPath(tool); err != nil {
			return err
		} else if !found {
			notFoundTools = append(notFoundTools, tool)
		}
	}

	if len(notFoundTools) > 0 {
		var errs error
		for _, tool := range notFoundTools {
			errs = errors.Append(errs, errors.Newf("%s not found in PATH, is it installed?", tool))
		}
		return errs
	}

	return nil
}

func existsPath(name string) (bool, error) {
	if _, err := exec.LookPath(name); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func validateIgniteInstalled() error {
	if _, err := exec.LookPath("ignite"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.Newf(`Ignite not found in PATH. Is it installed correctly?

Try running executor install ignite, or:
  $ curl -sfLo ignite https://github.com/sourcegraph/ignite/releases/download/%s/ignite-amd64
  $ chmod +x ignite
  $ mv ignite /usr/local/bin`, config.DefaultIgniteVersion)
		} else {
			return errors.Wrap(err, "failed to lookup ignite")
		}
	}

	return nil
}

func validateCNIInstalled() error {
	var errs error
	missingPlugins := []string{}
	missingIsolationPlugin := false
	if stat, err := os.Stat(config.CNIBinDir); err != nil {
		if os.IsNotExist(err) {
			errs = errors.Append(errs, errors.Newf("Cannot find directory %s. Are the CNI plugins for firecracker installed correctly?", config.CNIBinDir))
			missingPlugins = append([]string{}, config.RequiredCNIPlugins...)
			missingIsolationPlugin = true
		} else {
			errs = errors.Append(errs, errors.Wrap(err, "Checking for CNI_BIN_DIR"))
		}
	} else {
		if !stat.IsDir() {
			errs = errors.Append(errs, errors.Newf("%s expected to be a directory, but is a file", config.CNIBinDir))
		}
		for _, plugin := range config.RequiredCNIPlugins {
			pluginPath := path.Join(config.CNIBinDir, plugin)
			if stat, err := os.Stat(pluginPath); err != nil {
				if os.IsNotExist(err) {
					missingPlugins = append(missingPlugins, plugin)
					if plugin == "isolation" {
						missingIsolationPlugin = true
					}
				} else {
					errs = errors.Append(errs, errors.Wrapf(err, "Checking for existence of CNI plugin %q", plugin))
				}
			} else {
				if stat.IsDir() {
					errs = errors.Append(errs, errors.Newf("Expected %s to be a file, but is a directory", pluginPath))
				}
			}
		}
	}
	if len(missingPlugins) != 0 {
		hint := `To install the CNI plugins used by ignite run the following:
  $ mkdir -p /opt/cni/bin
  $ curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xz -C /opt/cni/bin`
		if missingIsolationPlugin {
			hint += `
To install the isolation plugin used by ignite run the following:
  $ curl -sSL https://github.com/AkihiroSuda/cni-isolation/releases/download/v0.0.4/cni-isolation-amd64.tgz | tar -xz -C /opt/cni/bin`
		}
		errs = errors.Append(errs, errors.Newf("Cannot find CNI plugins %v, are the CNI plugins for firecracker installed correctly?\n%s", missingPlugins, hint))
	}

	return errs
}
