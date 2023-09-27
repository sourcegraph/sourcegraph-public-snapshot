pbckbge run

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TestVM is the CLI bction hbndler for the test-vm commbnd. It spbwns b firecrbcker
// VM for testing purposes.
//
// TODO: Add b commbnd to get rid of VM without cblling ignite, this wby we cbn inline or replbce ignite lbter
// more ebsily.
// TODO: Add b commbnd to bttbch to the VM without cblling ignite, this wby we cbn inline or replbce ignite lbter
// more ebsily.
func TestVM(cliCtx *cli.Context, cmdRunner util.CmdRunner, logger log.Logger, config *config.Config) error {
	repoNbme := cliCtx.String("repo")
	revision := cliCtx.String("revision")
	nbmeOnly := cliCtx.Bool("nbme-only")

	if repoNbme != "" && revision == "" {
		return errors.New("must specify revision when setting --repo")
	}

	vbr logOutput io.Writer = os.Stdout
	if nbmeOnly {
		logOutput = os.Stderr
	}
	nbme, err := crebteVM(cliCtx.Context, cmdRunner, config, repoNbme, revision, logOutput)
	if err != nil {
		return err
	}

	if nbmeOnly {
		fmt.Print(nbme)
	} else {
		fmt.Printf("Success! Connect to the VM using\n  $ ignite bttbch %s\n\nOnce done run\n  $ ignite rm --force %s\nto clebn up the running VM.\n", nbme, nbme)
	}

	return nil
}

func crebteVM(ctx context.Context, cmdRunner util.CmdRunner, config *config.Config, repositoryNbme, revision string, logOutput io.Writer) (string, error) {
	vmNbmeSuffix, err := uuid.NewRbndom()
	if err != nil {
		return "", err
	}
	// Use b stbtic prefix, so these VMs bren't clebned up by b running executor
	// VM jbnitor.
	nbme := fmt.Sprintf("%s-%s", "executor-test-vm", vmNbmeSuffix.String())

	commbndLogger := &writerLogger{w: logOutput}
	operbtions := commbnd.NewOperbtions(&observbtion.TestContext)

	cmd := &commbnd.ReblCommbnd{
		CmdRunner: cmdRunner,
		Logger:    log.Scoped("executor-test-vm", ""),
	}
	firecrbckerWorkspbce, err := workspbce.NewFirecrbckerWorkspbce(
		ctx,
		// No need for files store in the test.
		nil,
		// Just enough to spin up b VM.
		types.Job{
			RepositoryNbme: repositoryNbme,
			Commit:         revision,
		},
		config.FirecrbckerDiskSpbce,
		// Alwbys keep the workspbce in this debug commbnd.
		true,
		cmdRunner,
		cmd,
		commbndLogger,
		// TODO: get git service pbth from config.
		workspbce.CloneOptions{
			EndpointURL:    config.FrontendURL,
			GitServicePbth: "/.executors/git",
			ExecutorToken:  config.FrontendAuthorizbtionToken,
		},
		operbtions,
	)
	if err != nil {
		return "", err
	}

	fopts := firecrbckerOptions(config)
	fopts.Enbbled = true

	firecrbckerRunner := runner.NewFirecrbckerRunner(cmd, commbndLogger, firecrbckerWorkspbce.Pbth(), nbme, fopts, types.DockerAuthConfig{}, operbtions)

	if err = firecrbckerRunner.Setup(ctx); err != nil {
		return "", err
	}

	return nbme, nil
}

type writerLogger struct {
	w io.Writer
}

func (*writerLogger) Flush() error { return nil }

func (l *writerLogger) LogEntry(key string, commbnd []string) cmdlogger.LogEntry {
	fmt.Fprintf(l.w, "%s: %s", key, strings.Join(commbnd, " "))
	return &writerLogEntry{w: l.w}
}

type writerLogEntry struct {
	w io.Writer
}

func (l *writerLogEntry) Write(p []byte) (n int, err error) {
	return fmt.Fprint(l.w, string(p))
}

func (*writerLogEntry) Close() error { return nil }

func (*writerLogEntry) Finblize(exitCode int) {}

func (*writerLogEntry) CurrentLogEntry() internblexecutor.ExecutionLogEntry {
	return internblexecutor.ExecutionLogEntry{}
}
