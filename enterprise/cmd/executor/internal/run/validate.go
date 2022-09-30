package run

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/Masterminds/semver"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RunValidate(cliCtx *cli.Context, logger log.Logger, config *config.Config) error {
	// First, validate the config is valid.
	if err := config.Validate(); err != nil {
		return err
	}

	// Then, validate all tools that are required are installed.
	if err := validateToolsRequired(config.UseFirecracker); err != nil {
		return err
	}

	// Validate git is of the right version.
	if err := validateGitVersion(cliCtx.Context); err != nil {
		return err
	}

	telemetryOptions := newTelemetryOptions(cliCtx.Context, config.UseFirecracker)
	copts := clientOptions(config, telemetryOptions)
	client := apiclient.NewBaseClient(copts.BaseClientOptions)
	// TODO: Validate access token.
	// Validate src-cli is of a good version, rely on the connected instance to tell
	// us what "good" means.
	if err := validateSrcCLIVersion(cliCtx.Context, logger, client, copts.EndpointOptions); err != nil {
		return err
	}

	if config.UseFirecracker {
		// Validate ignite is installed.
		if err := validateIgniteInstalled(cliCtx.Context); err != nil {
			return err
		}
		// Validate all required CNI plugins are installed.
		if err := validateCNIInstalled(); err != nil {
			return err
		}

		// TODO: Validate ignite images are pulled and imported. Sadly, the
		// output of ignite is not very parser friendly.
	}

	fmt.Print("All checks passed!\n")

	return nil
}

func validateGitVersion(ctx context.Context) error {
	gitVersion, err := getGitVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "getting git version")
	}
	have, err := semver.NewVersion(gitVersion)
	if err != nil {
		return errors.Newf("failed to semver parse git version: %s", gitVersion)
	} else if !config.MinGitVersionConstraint.Check(have) {
		return errors.Newf("git version is too old, install at least git 2.26, current version: %s", gitVersion)
	}
	return nil
}

// validateSrcCLIVersion queries the latest recommended version of src-cli and makes sure it
// matches what is installed. If not, a warning message recommending to use a different
// version is logged.
func validateSrcCLIVersion(ctx context.Context, logger log.Logger, client *apiclient.BaseClient, options apiclient.EndpointOptions) error {
	latestVersion, err := latestSrcCLIVersion(ctx, client, options)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve latest compatible src-cli version")
	}

	actualVersion, err := getSrcVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get src-cli version, is it installed?")
	}

	// Do nothing for dev versions.
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
	// If the installed version is too old:
	if actual.LessThan(latest) {
		return errors.Newf("installed src-cli is not the latest recommended version, consider upgrading actual=%s, latest=%s", actual.String(), latest.String())
		// If the installed version is too new:
	} else if actual.Major() != latest.Major() || actual.Minor() != latest.Minor() {
		return errors.Newf("installed src-cli is not the recommended version, consider switching actual=%s, recommended=%s", actual.String(), latest.String())
	}

	return nil
}

func latestSrcCLIVersion(ctx context.Context, client *apiclient.BaseClient, options apiclient.EndpointOptions) (_ string, err error) {
	req, err := client.MakeRequest(http.MethodGet, options.URL, ".api/src-cli/version", nil)
	if err != nil {
		return "", err
	}

	type versionPayload struct {
		Version string `json:"version"`
	}
	var v versionPayload
	if _, err := client.DoAndDecode(ctx, req, &v); err != nil {
		return "", err
	}

	return v.Version, nil
}

func validateToolsRequired(useFirecracker bool) error {
	notFoundTools := []string{}
	for tool := range config.RequiredCLITools {
		if found, err := existsPath(tool); err != nil {
			return err
		} else if !found {
			notFoundTools = append(notFoundTools, tool)
		}
	}
	for _, tool := range config.RequiredCLIToolsFirecracker {
		if found, err := existsPath(tool); err != nil {
			return err
		} else if !found {
			notFoundTools = append(notFoundTools, tool)
		}
	}

	if len(notFoundTools) > 0 {
		var errs error
		for _, tool := range notFoundTools {
			helptext, ok := config.RequiredCLITools[tool]
			helpLine := ""
			if ok {
				helpLine = fmt.Sprintf("\n%s", helptext)
			}
			errs = errors.Append(errs, errors.Newf("%s not found in PATH, is it installed?%s", tool, helpLine))
		}
		return errs
	}

	return nil
}

func validateIgniteInstalled(ctx context.Context) error {
	if found, err := existsPath("ignite"); err != nil {
		return errors.Wrap(err, "failed to lookup ignite")
	} else if !found {
		return errors.Newf(`Ignite not found in PATH. Is it installed correctly?

Try running "executor install ignite", or:
  $ curl -sfLo ignite https://github.com/sourcegraph/ignite/releases/download/%s/ignite-amd64
  $ chmod +x ignite
  $ mv ignite /usr/local/bin`, config.DefaultIgniteVersion)
	}

	want, err := semver.NewVersion(config.DefaultIgniteVersion)
	if err != nil {
		return err
	}
	current, err := getIgniteVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot read current ignite version")
	}
	have, err := semver.NewVersion(current)
	if err != nil {
		return errors.Wrap(err, "failed to parse ignite version")
	} else if !want.Equal(have) {
		return errors.Wrapf(err, "using unsupported ignite version, if things don't work alright, consider switching to the supported version. have=%s, want=%s", have.String(), want.String())
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
		hint := `To install the CNI plugins used by ignite run "executor install cni" or the following:
  $ mkdir -p /opt/cni/bin
  $ curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xz -C /opt/cni/bin`
		if missingIsolationPlugin {
			hint += `
  $ curl -sSL https://github.com/AkihiroSuda/cni-isolation/releases/download/v0.0.4/cni-isolation-amd64.tgz | tar -xz -C /opt/cni/bin`
		}
		errs = errors.Append(errs, errors.Newf("Cannot find CNI plugins %v, are the CNI plugins for firecracker installed correctly?\n%s", missingPlugins, hint))
	}

	return errs
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
