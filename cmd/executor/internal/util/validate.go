package util

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ValidateGitVersion validates that installed Git version meets the recommended version.
func ValidateGitVersion(ctx context.Context, runner CmdRunner) error {
	gitVersion, err := GetGitVersion(ctx, runner)
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

// ValidateSrcCLIVersion queries the latest recommended version of src-cli and makes sure it
// matches what is installed. If not, an error recommending to use a different
// version is returned.
func ValidateSrcCLIVersion(ctx context.Context, runner CmdRunner, client *apiclient.BaseClient) error {
	latestVersion, err := LatestSrcCLIVersion(ctx, client)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve latest compatible src-cli version")
	}

	actualVersion, err := GetSrcVersion(ctx, runner)
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

	if actual.Major() != latest.Major() || actual.Minor() != latest.Minor() {
		return errors.Newf("installed src-cli is not the recommended version, consider switching actual=%s, recommended=%s", actual.String(), latest.String())
	} else if actual.LessThan(latest) {
		return errors.Wrapf(ErrSrcPatchBehind, "consider upgrading actual=%s, latest=%s", actual.String(), latest.String())
	}

	return nil
}

// ErrSrcPatchBehind is the specific error if the currently installed src version is a patch behind the latest version.
var ErrSrcPatchBehind = errors.New("installed src-cli is not the latest version")

// ValidateRequiredTools validates that the tools required to run Docker and/or Firecracker are installed.
func ValidateRequiredTools(runner CmdRunner, useFirecracker bool) error {
	if err := ValidateDockerTools(runner); err != nil {
		return err
	}
	if useFirecracker {
		if err := ValidateFirecrackerTools(runner); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDockerTools validates that the tools required to run Docker are installed.
func ValidateDockerTools(runner CmdRunner) error {
	var missingTools []string
	// So, iterating thru a map is not deterministic, breaking unit tests, so we need to sort the keys.
	tools := make([]string, len(config.RequiredCLITools))
	i := 0
	for t := range config.RequiredCLITools {
		tools[i] = t
		i++
	}
	sort.Strings(tools)

	for _, tool := range tools {
		if found, err := ExistsPath(runner, tool); err != nil {
			return err
		} else if !found {
			missingTools = append(missingTools, tool)
		}
	}
	if len(missingTools) > 0 {
		return &ErrMissingTools{missingTools}
	}
	return nil
}

// ValidateFirecrackerTools validates that the tools required to run Firecracker are installed.
func ValidateFirecrackerTools(runner CmdRunner) error {
	var missingTools []string
	for _, tool := range config.RequiredCLIToolsFirecracker {
		if found, err := ExistsPath(runner, tool); err != nil {
			return err
		} else if !found {
			missingTools = append(missingTools, tool)
		}
	}
	if len(missingTools) > 0 {
		return &ErrMissingTools{missingTools}
	}
	return nil
}

// ValidateIgniteInstalled validates that ignite is installed to the host.
func ValidateIgniteInstalled(ctx context.Context, runner CmdRunner) error {
	if found, err := ExistsPath(runner, "ignite"); err != nil {
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
	current, err := GetIgniteVersion(ctx, runner)
	if err != nil {
		return errors.Wrap(err, "cannot read current ignite version")
	}
	have, err := semver.NewVersion(current)
	if err != nil {
		return errors.Wrap(err, "failed to parse ignite version")
	} else if !want.Equal(have) {
		return errors.Newf("using unsupported ignite version, if things don't work alright, consider switching to the supported version. have=%s, want=%s", have.String(), want.String())
	}

	return nil
}

// ValidateCNIInstalled validate that the CNI plugins for firecracker are properly installed.
func ValidateCNIInstalled(cmdRunner CmdRunner) error {
	var errs error
	var missingPlugins []string
	missingIsolationPlugin := false
	if stat, err := cmdRunner.Stat(config.CNIBinDir); err != nil {
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
			if s, err := cmdRunner.Stat(pluginPath); err != nil {
				if os.IsNotExist(err) {
					missingPlugins = append(missingPlugins, plugin)
					if plugin == "isolation" {
						missingIsolationPlugin = true
					}
				} else {
					errs = errors.Append(errs, errors.Wrapf(err, "Checking for existence of CNI plugin %q", plugin))
				}
			} else {
				if s.IsDir() {
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

// ErrMissingTools is the error when tools are missing for a specific runtime.
type ErrMissingTools struct {
	Tools []string
}

func (e *ErrMissingTools) Error() string {
	var errs error
	for _, tool := range e.Tools {
		helpText, ok := config.RequiredCLITools[tool]
		// TODO: Help lines for config.RequiredCLIToolsFirecracker.
		helpLine := ""
		if ok {
			helpLine = fmt.Sprintf("\n%s", helpText)
		}
		errs = errors.Append(errs, errors.Newf("%s not found in PATH, is it installed?%s", tool, helpLine))
	}
	return errs.Error()
}
