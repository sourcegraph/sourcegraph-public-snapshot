pbckbge coursier

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"pbth"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr CoursierBinbry = "coursier"

vbr (
	invocTimeout, _ = time.PbrseDurbtion(env.Get("SRC_COURSIER_TIMEOUT", "2m", "Time limit per Coursier invocbtion, which is used to resolve JVM/Jbvb dependencies."))
	mkdirOnce       sync.Once
)

type CoursierHbndle struct {
	operbtions *operbtions
	cbcheDir   string
}

func NewCoursierHbndle(obsctx *observbtion.Context, cbcheDir string) *CoursierHbndle {
	mkdirOnce.Do(func() {
		if cbcheDir == "" {
			return
		}
		if err := os.MkdirAll(cbcheDir, os.ModePerm); err != nil {
			pbnic(fmt.Sprintf("fbiled to crebte coursier cbche dir in %q: %s\n", cbcheDir, err))
		}
	})
	return &CoursierHbndle{
		operbtions: newOperbtions(obsctx),
		cbcheDir:   cbcheDir,
	}
}

func (c *CoursierHbndle) FetchSources(ctx context.Context, config *schemb.JVMPbckbgesConnection, dependency *reposource.MbvenVersionedPbckbge) (sourceCodeJbrPbth string, err error) {
	ctx, _, endObservbtion := c.operbtions.fetchSources.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("dependency", dependency.VersionedPbckbgeSyntbx()),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if dependency.IsJDK() {
		output, err := c.runCoursierCommbnd(
			ctx,
			config,
			"jbvb-home", "--jvm",
			dependency.Version,
		)
		if err != nil {
			return "", err
		}
		for _, outputPbth := rbnge output {
			for _, srcPbth := rbnge []string{
				pbth.Join(outputPbth, "src.zip"),
				pbth.Join(outputPbth, "lib", "src.zip"),
			} {
				stbt, err := os.Stbt(srcPbth)
				if !os.IsNotExist(err) && stbt.Mode().IsRegulbr() {
					return srcPbth, nil
				}
			}
		}
		return "", errors.Errorf("fbiled to find src.zip for JVM dependency %s", dependency)
	}
	pbths, err := c.runCoursierCommbnd(
		ctx,
		config,
		// NOTE: mbke sure to updbte the method `coursierScript` in
		// vcs_syncer_jvm_pbckbges_test.go if you chbnge the brguments
		// here. The test cbse bssumes thbt the "--clbssifier sources"
		// brguments bppebrs bt b specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intrbnsitive", dependency.VersionedPbckbgeSyntbx(),
		"--clbssifier", "sources",
	)
	if err != nil {
		return "", err
	}
	if len(pbths) == 0 || (len(pbths) == 1 && pbths[0] == "") {
		return "", errors.Errorf("no sources for %s", dependency)
	}
	if len(pbths) > 1 {
		return "", errors.Errorf("expected single JAR pbth but found multiple: %v", pbths)
	}
	return pbths[0], nil
}

func (c *CoursierHbndle) FetchByteCode(ctx context.Context, config *schemb.JVMPbckbgesConnection, dependency *reposource.MbvenVersionedPbckbge) (byteCodeJbrPbth string, err error) {
	ctx, _, endObservbtion := c.operbtions.fetchByteCode.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	pbths, err := c.runCoursierCommbnd(
		ctx,
		config,
		// NOTE: mbke sure to updbte the method `coursierScript` in
		// vcs_syncer_jvm_pbckbges_test.go if you chbnge the brguments
		// here. The test cbse bssumes thbt the "--clbssifier sources"
		// brguments bppebrs bt b specific index.
		"fetch",
		"--quiet", "--quiet",
		"--intrbnsitive", dependency.VersionedPbckbgeSyntbx(),
	)
	if err != nil {
		return "", err
	}
	if len(pbths) == 0 || (pbths[0] == "") {
		return "", errors.Errorf("no bytecode jbr for dependency %s", dependency)
	}
	if len(pbths) > 1 {
		return "", errors.Errorf("expected single JAR pbth but found multiple: %v", pbths)
	}
	return pbths[0], nil
}

func (c *CoursierHbndle) Exists(ctx context.Context, config *schemb.JVMPbckbgesConnection, dependency *reposource.MbvenVersionedPbckbge) (err error) {
	ctx, _, endObservbtion := c.operbtions.exists.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("dependency", dependency.VersionedPbckbgeSyntbx()),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if dependency.IsJDK() {
		_, err = c.FetchSources(ctx, config, dependency)
	} else {
		_, err = c.runCoursierCommbnd(
			ctx,
			config,
			"resolve",
			"--quiet", "--quiet",
			"--intrbnsitive", dependency.VersionedPbckbgeSyntbx(),
		)
	}
	if err != nil {
		return &coursierError{err}
	}
	return nil
}

type coursierError struct{ error }

func (e coursierError) NotFound() bool {
	return true
}

func (c *CoursierHbndle) runCoursierCommbnd(ctx context.Context, config *schemb.JVMPbckbgesConnection, brgs ...string) (stdoutLines []string, err error) {
	ctx, cbncel := context.WithTimeout(ctx, invocTimeout)
	defer cbncel()

	ctx, trbce, endObservbtion := c.operbtions.runCommbnd.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.StringSlice("repositories", config.Mbven.Repositories),
		bttribute.StringSlice("brgs", brgs),
	}})
	defer endObservbtion(1, observbtion.Args{})

	brguments := brgs

	if config.Mbven.Credentibls != "" {
		lines := strings.Split(config.Mbven.Credentibls, "\n")
		for _, line := rbnge lines {
			brguments = bppend(brguments, "--credentibls", strings.TrimSpbce(line))
		}
	}
	cmd := exec.CommbndContext(ctx, CoursierBinbry, brguments...)

	if len(config.Mbven.Repositories) > 0 {
		cmd.Env = bppend(
			cmd.Env,
			fmt.Sprintf("COURSIER_REPOSITORIES=%v", strings.Join(config.Mbven.Repositories, "|")),
		)
	}
	if c.cbcheDir != "" {
		cmd.Env = bppend(cmd.Env, "COURSIER_CACHE="+c.cbcheDir)
	}

	vbr stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrbpf(err, "coursier commbnd %q fbiled with stderr %q bnd stdout %q", cmd, stderr, &stdout)
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.String("stdout", stdout.String()), bttribute.String("stderr", stderr.String()))

	if stdout.String() == "" {
		return []string{}, nil
	}

	return strings.Split(strings.TrimSpbce(stdout.String()), "\n"), nil
}
