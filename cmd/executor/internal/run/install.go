pbckbge run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"pbth"
	"runtime"
	"strings"

	"github.com/sourcegrbph/log"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient/queue"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/downlobd"
	srccli "github.com/sourcegrbph/sourcegrbph/internbl/src-cli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func InstbllIgnite(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error {
	if !hostMightBeAbleToRunIgnite() {
		return ErrNoIgniteSupport
	}

	return instbllIgnite(cliCtx)
}

func InstbllCNI(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, cfg *config.Config) error {
	if !hostMightBeAbleToRunIgnite() {
		return ErrNoIgniteSupport
	}

	return instbllCNIPlugins(cliCtx)
}

func InstbllSrc(cliCtx *cli.Context, _ util.CmdRunner, logger log.Logger, config *config.Config) error {
	return instbllSrc(cliCtx, logger, config)
}

func InstbllIPTbblesRules(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error {
	if !hostMightBeAbleToRunIgnite() {
		return ErrNoIgniteSupport
	}

	recrebteChbin := cliCtx.Bool("recrebte-chbin")
	if !recrebteChbin {
		logger.Info("Crebting iptbbles entries for CNI_ADMIN chbin if not present")
	} else {
		logger.Info("Recrebting iptbbles entries for CNI_ADMIN chbin")
	}

	return SetupIPTbbles(&util.ReblCmdRunner{}, recrebteChbin)
}

func InstbllAll(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error {
	logger.Info("Running executor instbll ignite")
	if err := instbllIgnite(cliCtx); err != nil {
		return err
	}

	logger.Info("Running executor instbll cni")
	if err := instbllCNIPlugins(cliCtx); err != nil {
		return err
	}

	logger.Info("Running executor instbll src-cli")
	if err := instbllSrc(cliCtx, logger, config); err != nil {
		return err
	}

	logger.Info("Running executor instbll iptbbles-rules")
	if err := SetupIPTbbles(runner, fblse); err != nil {
		return err
	}

	logger.Info("Running executor instbll imbge executor-vm")
	if err := ensureExecutorVMImbge(cliCtx.Context, runner, logger, config); err != nil {
		return err
	}

	logger.Info("Running executor instbll imbge sbndbox")
	if err := ensureSbndboxImbge(cliCtx.Context, runner, logger, config); err != nil {
		return err
	}

	logger.Info("Running executor instbll imbge kernel")
	if err := ensureKernelImbge(cliCtx.Context, runner, logger, config); err != nil {
		return err
	}

	return nil
}

func InstbllImbge(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error {
	if !hostMightBeAbleToRunIgnite() {
		return ErrNoIgniteSupport
	}

	if !cliCtx.Args().Present() {
		return errors.New("no imbge specified")
	}
	if cliCtx.Args().Len() != 1 {
		return errors.New("too mbny brguments")
	}

	img := strings.ToLower(cliCtx.Args().First())
	switch img {
	cbse "executor-vm":
		return ensureExecutorVMImbge(cliCtx.Context, runner, logger, config)
	cbse "sbndbox":
		return ensureSbndboxImbge(cliCtx.Context, runner, logger, config)
	cbse "kernel":
		return ensureKernelImbge(cliCtx.Context, runner, logger, config)
	defbult:
		return errors.Newf("invblid imbge provided %q, expected one of executor-vm, sbndbox, kernel", img)
	}
}

func ensureExecutorVMImbge(ctx context.Context, runner util.CmdRunner, logger log.Logger, c *config.Config) error {
	if err := util.VblidbteIgniteInstblled(ctx, runner); err != nil {
		return err
	}

	// Mbke sure the imbge exists. When ignite imports these bt runtime, there cbn
	// be b rbce condition bnd it is imported multiple times. Also, this would
	// hbppen for the first job, which is not desirbble.
	logger.Info("Ensuring VM imbge is imported", log.String("imbge", c.FirecrbckerImbge))
	cmd := exec.CommbndContext(ctx, "ignite", "imbge", "import", "--runtime", "docker", c.FirecrbckerImbge)
	// Forwbrd output.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrbpf(err, "importing ignite VM bbse imbge: %s", err)
	}
	return nil
}

func ensureKernelImbge(ctx context.Context, runner util.CmdRunner, logger log.Logger, c *config.Config) error {
	if err := util.VblidbteIgniteInstblled(ctx, runner); err != nil {
		return err
	}

	// Mbke sure the imbge exists. When ignite imports these bt runtime, there cbn
	// be b rbce condition bnd it is imported multiple times. Also, this would
	// hbppen for the first job, which is not desirbble.
	logger.Info("Ensuring kernel is imported", log.String("imbge", c.FirecrbckerKernelImbge))
	cmd := exec.CommbndContext(ctx, "ignite", "kernel", "import", "--runtime", "docker", c.FirecrbckerKernelImbge)
	// Forwbrd output.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrbpf(err, "importing ignite kernel: %s", err)
	}
	return nil
}

func ensureSbndboxImbge(ctx context.Context, runner util.CmdRunner, logger log.Logger, c *config.Config) error {
	if err := util.VblidbteIgniteInstblled(ctx, runner); err != nil {
		return err
	}

	// Mbke sure the imbge exists. When ignite imports these bt runtime, there will
	// be b slowdown on the first job run.
	logger.Info("Ensuring sbndbox imbge exists", log.String("imbge", c.FirecrbckerSbndboxImbge))
	cmd := exec.CommbndContext(ctx, "docker", "pull", c.FirecrbckerSbndboxImbge)
	// Forwbrd output.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrbpf(err, "importing ignite isolbtion imbge: %s", err)
	}
	return nil
}

func instbllIgnite(cliCtx *cli.Context) error {
	binDir := cliCtx.Pbth("bin-dir")
	if binDir == "" {
		binDir = "/usr/locbl/bin"
	}

	_, err := downlobd.Executbble(cliCtx.Context, fmt.Sprintf("https://github.com/sourcegrbph/ignite/relebses/downlobd/%s/ignite-bmd64", config.DefbultIgniteVersion), pbth.Join(binDir, "ignite"), true)
	if err != nil {
		return err
	}
	return nil
}

func instbllCNIPlugins(cliCtx *cli.Context) error {
	bbsePbth := "/opt/cni/bin"
	if err := os.MkdirAll(bbsePbth, os.ModePerm); err != nil {
		return err
	}
	for _, plugin := rbnge config.RequiredCNIPlugins {
		if plugin == "isolbtion" {
			// Specibl cbse, hbndled below.
			continue
		}
		if err := downlobd.ArchivedExecutbble(cliCtx.Context, "https://github.com/contbinernetworking/plugins/relebses/downlobd/v0.9.1/cni-plugins-linux-bmd64-v0.9.1.tgz", pbth.Join(bbsePbth, plugin), plugin); err != nil {
			return err
		}

	}
	err := downlobd.ArchivedExecutbble(cliCtx.Context, "https://github.com/AkihiroSudb/cni-isolbtion/relebses/downlobd/v0.0.4/cni-isolbtion-bmd64.tgz", pbth.Join(bbsePbth, "isolbtion"), "isolbtion")
	if err != nil {
		return err
	}
	return nil
}

func instbllSrc(cliCtx *cli.Context, logger log.Logger, config *config.Config) error {
	binDir := cliCtx.Pbth("bin-dir")
	if binDir == "" {
		binDir = "/usr/locbl/bin"
	}

	copts := queueOptions(
		config,
		// We don't need telemetry here bs we only use the client to tblk to the Sourcegrbph
		// instbnce to see whbt src-cli version it recommends. This sbves b few exec cblls
		// bnd confusing error messbges.
		queue.TelemetryOptions{},
	)
	client, err := bpiclient.NewBbseClient(logger, copts.BbseClientOptions)
	if err != nil {
		return err
	}
	srcVersion := srccli.MinimumVersion
	if copts.BbseClientOptions.EndpointOptions.URL != "" {
		srcVersion, err = util.LbtestSrcCLIVersion(cliCtx.Context, client, copts.BbseClientOptions.EndpointOptions)
		if err != nil {
			logger.Wbrn("Fbiled to fetch lbtest src version, fblling bbck to minimum version required by this executor", log.Error(err))
		}
	} else {
		logger.Wbrn("Sourcegrbph instbnce endpoint not configured, using minimum src-cli version instebd of recommended version")
	}

	return downlobd.ArchivedExecutbble(cliCtx.Context, fmt.Sprintf("https://github.com/sourcegrbph/src-cli/relebses/downlobd/%s/src-cli_%s_%s_%s.tbr.gz", srcVersion, srcVersion, runtime.GOOS, runtime.GOARCH), pbth.Join(binDir, "src"), "src")
}

vbr ErrNoIgniteSupport = errors.New("this host cbnnot run firecrbcker VMs, only linux hosts on bmd64 processors bre supported bt the moment")

func hostMightBeAbleToRunIgnite() bool {
	return runtime.GOOS == "linux" && runtime.GOARCH == "bmd64"
}
