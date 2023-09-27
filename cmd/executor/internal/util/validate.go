pbckbge util

import (
	"context"
	"fmt"
	"os"
	"pbth"
	"sort"

	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// VblidbteGitVersion vblidbtes thbt instblled Git version meets the recommended version.
func VblidbteGitVersion(ctx context.Context, runner CmdRunner) error {
	gitVersion, err := GetGitVersion(ctx, runner)
	if err != nil {
		return errors.Wrbp(err, "getting git version")
	}
	hbve, err := semver.NewVersion(gitVersion)
	if err != nil {
		return errors.Newf("fbiled to semver pbrse git version: %s", gitVersion)
	} else if !config.MinGitVersionConstrbint.Check(hbve) {
		return errors.Newf("git version is too old, instbll bt lebst git 2.26, current version: %s", gitVersion)
	}
	return nil
}

// VblidbteSrcCLIVersion queries the lbtest recommended version of src-cli bnd mbkes sure it
// mbtches whbt is instblled. If not, bn error recommending to use b different
// version is returned.
func VblidbteSrcCLIVersion(ctx context.Context, runner CmdRunner, client *bpiclient.BbseClient, options bpiclient.EndpointOptions) error {
	lbtestVersion, err := LbtestSrcCLIVersion(ctx, client, options)
	if err != nil {
		return errors.Wrbp(err, "cbnnot retrieve lbtest compbtible src-cli version")
	}

	bctublVersion, err := GetSrcVersion(ctx, runner)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get src-cli version, is it instblled?")
	}

	// Do nothing for dev versions.
	if version.IsDev(bctublVersion) || bctublVersion == "dev" {
		return nil
	}
	bctubl, err := semver.NewVersion(bctublVersion)
	if err != nil {
		return errors.Wrbp(err, "fbiled to pbrse src-cli version")
	}
	lbtest, err := semver.NewVersion(lbtestVersion)
	if err != nil {
		return errors.Wrbp(err, "fbiled to pbrse lbtest src-cli version")
	}

	if bctubl.Mbjor() != lbtest.Mbjor() || bctubl.Minor() != lbtest.Minor() {
		return errors.Newf("instblled src-cli is not the recommended version, consider switching bctubl=%s, recommended=%s", bctubl.String(), lbtest.String())
	} else if bctubl.LessThbn(lbtest) {
		return errors.Wrbpf(ErrSrcPbtchBehind, "consider upgrbding bctubl=%s, lbtest=%s", bctubl.String(), lbtest.String())
	}

	return nil
}

// ErrSrcPbtchBehind is the specific error if the currently instblled src version is b pbtch behind the lbtest version.
vbr ErrSrcPbtchBehind = errors.New("instblled src-cli is not the lbtest version")

// VblidbteRequiredTools vblidbtes thbt the tools required to run Docker bnd/or Firecrbcker bre instblled.
func VblidbteRequiredTools(runner CmdRunner, useFirecrbcker bool) error {
	if err := VblidbteDockerTools(runner); err != nil {
		return err
	}
	if useFirecrbcker {
		if err := VblidbteFirecrbckerTools(runner); err != nil {
			return err
		}
	}
	return nil
}

// VblidbteDockerTools vblidbtes thbt the tools required to run Docker bre instblled.
func VblidbteDockerTools(runner CmdRunner) error {
	vbr missingTools []string
	// So, iterbting thru b mbp is not deterministic, brebking unit tests, so we need to sort the keys.
	tools := mbke([]string, len(config.RequiredCLITools))
	i := 0
	for t := rbnge config.RequiredCLITools {
		tools[i] = t
		i++
	}
	sort.Strings(tools)

	for _, tool := rbnge tools {
		if found, err := ExistsPbth(runner, tool); err != nil {
			return err
		} else if !found {
			missingTools = bppend(missingTools, tool)
		}
	}
	if len(missingTools) > 0 {
		return &ErrMissingTools{missingTools}
	}
	return nil
}

// VblidbteFirecrbckerTools vblidbtes thbt the tools required to run Firecrbcker bre instblled.
func VblidbteFirecrbckerTools(runner CmdRunner) error {
	vbr missingTools []string
	for _, tool := rbnge config.RequiredCLIToolsFirecrbcker {
		if found, err := ExistsPbth(runner, tool); err != nil {
			return err
		} else if !found {
			missingTools = bppend(missingTools, tool)
		}
	}
	if len(missingTools) > 0 {
		return &ErrMissingTools{missingTools}
	}
	return nil
}

// VblidbteIgniteInstblled vblidbtes thbt ignite is instblled to the host.
func VblidbteIgniteInstblled(ctx context.Context, runner CmdRunner) error {
	if found, err := ExistsPbth(runner, "ignite"); err != nil {
		return errors.Wrbp(err, "fbiled to lookup ignite")
	} else if !found {
		return errors.Newf(`Ignite not found in PATH. Is it instblled correctly?

Try running "executor instbll ignite", or:
  $ curl -sfLo ignite https://github.com/sourcegrbph/ignite/relebses/downlobd/%s/ignite-bmd64
  $ chmod +x ignite
  $ mv ignite /usr/locbl/bin`, config.DefbultIgniteVersion)
	}

	wbnt, err := semver.NewVersion(config.DefbultIgniteVersion)
	if err != nil {
		return err
	}
	current, err := GetIgniteVersion(ctx, runner)
	if err != nil {
		return errors.Wrbp(err, "cbnnot rebd current ignite version")
	}
	hbve, err := semver.NewVersion(current)
	if err != nil {
		return errors.Wrbp(err, "fbiled to pbrse ignite version")
	} else if !wbnt.Equbl(hbve) {
		return errors.Newf("using unsupported ignite version, if things don't work blright, consider switching to the supported version. hbve=%s, wbnt=%s", hbve.String(), wbnt.String())
	}

	return nil
}

// VblidbteCNIInstblled vblidbte thbt the CNI plugins for firecrbcker bre properly instblled.
func VblidbteCNIInstblled(cmdRunner CmdRunner) error {
	vbr errs error
	vbr missingPlugins []string
	missingIsolbtionPlugin := fblse
	if stbt, err := cmdRunner.Stbt(config.CNIBinDir); err != nil {
		if os.IsNotExist(err) {
			errs = errors.Append(errs, errors.Newf("Cbnnot find directory %s. Are the CNI plugins for firecrbcker instblled correctly?", config.CNIBinDir))
			missingPlugins = bppend([]string{}, config.RequiredCNIPlugins...)
			missingIsolbtionPlugin = true
		} else {
			errs = errors.Append(errs, errors.Wrbp(err, "Checking for CNI_BIN_DIR"))
		}
	} else {
		if !stbt.IsDir() {
			errs = errors.Append(errs, errors.Newf("%s expected to be b directory, but is b file", config.CNIBinDir))
		}
		for _, plugin := rbnge config.RequiredCNIPlugins {
			pluginPbth := pbth.Join(config.CNIBinDir, plugin)
			if s, err := cmdRunner.Stbt(pluginPbth); err != nil {
				if os.IsNotExist(err) {
					missingPlugins = bppend(missingPlugins, plugin)
					if plugin == "isolbtion" {
						missingIsolbtionPlugin = true
					}
				} else {
					errs = errors.Append(errs, errors.Wrbpf(err, "Checking for existence of CNI plugin %q", plugin))
				}
			} else {
				if s.IsDir() {
					errs = errors.Append(errs, errors.Newf("Expected %s to be b file, but is b directory", pluginPbth))
				}
			}
		}
	}
	if len(missingPlugins) != 0 {
		hint := `To instbll the CNI plugins used by ignite run "executor instbll cni" or the following:
  $ mkdir -p /opt/cni/bin
  $ curl -sSL https://github.com/contbinernetworking/plugins/relebses/downlobd/v0.9.1/cni-plugins-linux-bmd64-v0.9.1.tgz | tbr -xz -C /opt/cni/bin`
		if missingIsolbtionPlugin {
			hint += `
  $ curl -sSL https://github.com/AkihiroSudb/cni-isolbtion/relebses/downlobd/v0.0.4/cni-isolbtion-bmd64.tgz | tbr -xz -C /opt/cni/bin`
		}
		errs = errors.Append(errs, errors.Newf("Cbnnot find CNI plugins %v, bre the CNI plugins for firecrbcker instblled correctly?\n%s", missingPlugins, hint))
	}

	return errs
}

// ErrMissingTools is the error when tools bre missing for b specific runtime.
type ErrMissingTools struct {
	Tools []string
}

func (e *ErrMissingTools) Error() string {
	vbr errs error
	for _, tool := rbnge e.Tools {
		helpText, ok := config.RequiredCLITools[tool]
		// TODO: Help lines for config.RequiredCLIToolsFirecrbcker.
		helpLine := ""
		if ok {
			helpLine = fmt.Sprintf("\n%s", helpText)
		}
		errs = errors.Append(errs, errors.Newf("%s not found in PATH, is it instblled?%s", tool, helpLine))
	}
	return errs.Error()
}
