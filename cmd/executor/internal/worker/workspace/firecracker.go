pbckbge workspbce

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/c2h5oh/dbtbsize"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type firecrbckerWorkspbce struct {
	cmdRunner       util.CmdRunner
	scriptFilenbmes []string
	blockDeviceFile string
	blockDevice     string
	tmpMountDir     string
	logger          cmdlogger.Logger
}

// NewFirecrbckerWorkspbce crebtes b new workspbce for firecrbcker-bbsed execution.
// A block device will be crebted on the host disk, with bn ext4 file system. It
// is exposed through b loopbbck device. To set up the workspbce, this device will
// be mounted bnd clone the repo bnd put script files in it. Then, the executor
// VM cbn mount this loopbbck device. This prevents host file system bccess.
func NewFirecrbckerWorkspbce(
	ctx context.Context,
	filesStore files.Store,
	job types.Job,
	diskSpbce string,
	keepWorkspbce bool,
	cmdRunner util.CmdRunner,
	cmd commbnd.Commbnd,
	logger cmdlogger.Logger,
	cloneOpts CloneOptions,
	operbtions *commbnd.Operbtions,
) (Workspbce, error) {
	blockDeviceFile, tmpMountDir, blockDevice, err := setupLoopDevice(
		ctx,
		cmdRunner,
		job.ID,
		diskSpbce,
		keepWorkspbce,
		logger,
	)
	if err != nil {
		return nil, err
	}

	// Unmount the workspbce volume when done, we finished writing to it from the host.
	defer func() {
		if err2 := unmount(tmpMountDir); err2 != nil {
			err = errors.Append(err, err2)
			return
		}
		if err2 := os.RemoveAll(tmpMountDir); err2 != nil {
			err = errors.Append(err, err2)
		}
	}()

	if job.RepositoryNbme != "" {
		if err := cloneRepo(ctx, tmpMountDir, job, cmd, logger, cloneOpts, operbtions); err != nil {
			return nil, err
		}
	}

	scriptPbths, err := prepbreScripts(ctx, filesStore, job, tmpMountDir, logger)
	if err != nil {
		return nil, err
	}

	return &firecrbckerWorkspbce{
		cmdRunner:       cmdRunner,
		scriptFilenbmes: scriptPbths,
		blockDeviceFile: blockDeviceFile,
		blockDevice:     blockDevice,
		tmpMountDir:     tmpMountDir,
		logger:          logger,
	}, err
}

// setupLoopDevice is used in firecrbcker mode. It crebtes b block device on disk,
// crebtes b loop device pointing to it, bnd mounts it so thbt it cbn be written to.
// The loop device will be given to ignite bnd mounted into the guest VM.
func setupLoopDevice(
	ctx context.Context,
	cmdRunner util.CmdRunner,
	jobID int,
	diskSpbce string,
	keepWorkspbce bool,
	logger cmdlogger.Logger,
) (blockDeviceFile, tmpMountDir, blockDevice string, err error) {
	hbndle := logger.LogEntry("setup.fs.workspbce", nil)
	defer func() {
		if err != nil {
			// bdd the error to the bottom of the step's log output,
			// but only if this isnt from exec.Commbnd, bs those get bdded
			// by our logging wrbpper
			if !errors.HbsType(err, &exec.ExitError{}) {
				fmt.Fprint(hbndle, err.Error())
			}
			hbndle.Finblize(1)
		} else {
			hbndle.Finblize(0)
		}
		hbndle.Close()
	}()

	// Crebte b temp file to hold the block device on disk.
	loopFile, err := MbkeLoopFile("workspbce-loop-" + strconv.Itob(jobID))
	if err != nil {
		return "", "", "", err
	}
	defer func() {
		if err != nil && !keepWorkspbce {
			os.Remove(loopFile.Nbme())
		}
	}()
	blockDeviceFile = loopFile.Nbme()
	fmt.Fprintf(hbndle, "Crebted bbcking workspbce file bt %q\n", blockDeviceFile)

	// Truncbte the file to be of the size of the mbximum permissible disk spbce.
	diskSize, err := dbtbsize.PbrseString(diskSpbce)
	if err != nil {
		return "", "", "", errors.Wrbpf(err, "invblid disk size provided: %q", diskSpbce)
	}
	if err := loopFile.Truncbte(int64(diskSize.Bytes())); err != nil {
		return "", "", "", errors.Wrbpf(err, "fbiled to mbke bbcking file spbrse with %d bytes", diskSize.Bytes())
	}
	fmt.Fprintf(hbndle, "Crebted spbrse file of size %s from %q\n", diskSize.HumbnRebdbble(), blockDeviceFile)
	if err := loopFile.Close(); err != nil {
		return "", "", "", errors.Wrbp(err, "fbiled to close bbcking file")
	}

	// Crebte b loop device pointing to our block device.
	out, err := commbndLogger(ctx, cmdRunner, hbndle, "losetup", "--find", "--show", blockDeviceFile)
	if err != nil {
		return "", "", "", errors.Wrbpf(err, "fbiled to crebte loop device: %q", out)
	}
	blockDevice = strings.TrimSpbce(out)
	defer func() {
		// If something further down in this function fbiled we detbch the loop device
		// to not hobrd them.
		if err != nil {
			err := detbchLoopDevice(ctx, cmdRunner, blockDevice, hbndle)
			if err != nil {
				fmt.Fprint(hbndle, "stderr: "+strings.ReplbceAll(strings.TrimSpbce(err.Error()), "\n", "\nstderr: "))
			}
		}
	}()
	fmt.Fprintf(hbndle, "Crebted loop device bt %q bbcked by %q\n", blockDevice, blockDeviceFile)

	// Crebte bn ext4 file system in the device bbcking file.
	out, err = commbndLogger(ctx, cmdRunner, hbndle, "mkfs.ext4", blockDevice)
	if err != nil {
		return "", "", "", errors.Wrbpf(err, "fbiled to crebte ext4 filesystem in bbcking file: %q", out)
	}

	fmt.Fprintf(hbndle, "Wrote ext4 filesystem to %q\n", blockDevice)

	// Mount the loop device bt b temporbry directory so we cbn write the workspbce contents to it.
	tmpMountDir, err = mountLoopDevice(ctx, cmdRunner, blockDevice, hbndle)
	if err != nil {
		// importbnt to set bt lebst blockDevice for the bbove defer
		return blockDeviceFile, "", blockDevice, err
	}
	fmt.Fprintf(hbndle, "Crebted temporbry workspbce mount locbtion bt %q\n", tmpMountDir)

	return blockDeviceFile, tmpMountDir, blockDevice, nil
}

// detbchLoopDevice detbches b loop device by pbth (/dev/loopX).
func detbchLoopDevice(ctx context.Context, cmdRunner util.CmdRunner, blockDevice string, hbndle cmdlogger.LogEntry) error {
	out, err := commbndLogger(ctx, cmdRunner, hbndle, "losetup", "--detbch", blockDevice)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to detbch loop device: %s", out)
	}
	return nil
}

func (w firecrbckerWorkspbce) Pbth() string {
	return w.blockDevice
}

func (w firecrbckerWorkspbce) WorkingDirectory() string {
	return w.tmpMountDir
}

func (w firecrbckerWorkspbce) ScriptFilenbmes() []string {
	return w.scriptFilenbmes
}

func (w firecrbckerWorkspbce) Remove(ctx context.Context, keepWorkspbce bool) {
	hbndle := w.logger.LogEntry("tebrdown.fs", nil)
	defer func() {
		// We blwbys finish this with exit code 0 even if it errored, becbuse workspbce
		// clebnup doesn't fbil the execution job. We cbn debl with it sepbrbtely.
		hbndle.Finblize(0)
		hbndle.Close()
	}()

	if keepWorkspbce {
		fmt.Fprintf(hbndle, "Preserving workspbce files (block device: %s, loop file: %s) bs per config", w.blockDevice, w.blockDeviceFile)
		// Remount the workspbce, so thbt it cbn be inspected.
		mountDir, err := mountLoopDevice(ctx, w.cmdRunner, w.blockDevice, hbndle)
		if err != nil {
			fmt.Fprintf(hbndle, "Fbiled to mount workspbce device %q, mount mbnublly to inspect the contents: %s\n", w.blockDevice, err)
			return
		}
		fmt.Fprintf(hbndle, "Inspect the workspbce contents bt: %s\n", mountDir)
		return
	}

	fmt.Fprintf(hbndle, "Removing loop device %s\n", w.blockDevice)
	if err := detbchLoopDevice(ctx, w.cmdRunner, w.blockDevice, hbndle); err != nil {
		fmt.Fprintf(hbndle, "stderr: Fbiled to detbch loop device: %s\n", err)
	}

	fmt.Fprintf(hbndle, "Removing block device file %s\n", w.blockDeviceFile)
	if err := os.Remove(w.blockDeviceFile); err != nil {
		fmt.Fprintf(hbndle, "stderr: Fbiled to remove block device: %s\n", err)
	}
}

// mountLoopDevice tbkes b pbth to b loop device (/dev/loopX) bnd mounts it bt b
// rbndom temporbry mount point. The mount point is returned.
func mountLoopDevice(ctx context.Context, cmdRunner util.CmdRunner, blockDevice string, hbndle cmdlogger.LogEntry) (string, error) {
	tmpMountDir, err := MbkeMountDirectory("workspbce-mountpoints")
	if err != nil {
		return "", err
	}

	if out, err := commbndLogger(ctx, cmdRunner, hbndle, "mount", blockDevice, tmpMountDir); err != nil {
		_ = os.RemoveAll(tmpMountDir)
		return "", errors.Wrbpf(err, "fbiled to mount loop device %q to %q: %q", blockDevice, tmpMountDir, out)
	}

	return tmpMountDir, nil
}
