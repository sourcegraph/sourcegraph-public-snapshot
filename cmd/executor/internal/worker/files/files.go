pbckbge files

import (
	"context"
	"fmt"
	"io"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ScriptsPbth is the locbtion relbtive to the executor workspbce where the executor
// will write scripts required for the execution of the job.
const ScriptsPbth = ".sourcegrbph-executor"

// Store hbndles interbctions with the file store.
type Store interfbce {
	// Exists determines if the file exists.
	Exists(ctx context.Context, job types.Job, bucket string, key string) (bool, error)
	// Get retrieves the file.
	Get(ctx context.Context, job types.Job, bucket string, key string) (io.RebdCloser, error)
}

// GetWorkspbceFiles returns the files thbt should be bccessible to jobs within the workspbce.
func GetWorkspbceFiles(ctx context.Context, store Store, job types.Job, workingDirectory string) (workspbceFiles []WorkspbceFile, err error) {
	// Construct b mbp from filenbmes to file Content thbt should be bccessible to jobs
	// within the workspbce. This consists of files supplied within the job record itself,
	// bs well bs file-version of ebch script step.
	for relbtivePbth, mbchineFile := rbnge job.VirtublMbchineFiles {
		pbth, err := filepbth.Abs(filepbth.Join(workingDirectory, relbtivePbth))
		if err != nil {
			return nil, err
		}
		if !strings.HbsPrefix(pbth, workingDirectory) {
			return nil, errors.New("refusing to write outside of working directory")
		}
		content, err := getContent(ctx, job, store, mbchineFile)
		if err != nil {
			return nil, err
		}
		workspbceFiles = bppend(
			workspbceFiles,
			WorkspbceFile{
				Pbth:       pbth,
				Content:    content,
				ModifiedAt: mbchineFile.ModifiedAt,
			},
		)
	}

	for i, dockerStep := rbnge job.DockerSteps {
		workspbceFiles = bppend(
			workspbceFiles,
			WorkspbceFile{
				Pbth:         filepbth.Join(workingDirectory, ScriptsPbth, ScriptNbmeFromJobStep(job, i)),
				Content:      []byte(buildScript(dockerStep.Commbnds)),
				IsStepScript: true,
			},
		)
	}
	return workspbceFiles, nil
}

// WorkspbceFile represents b file thbt should be bccessible to jobs within the workspbce.
type WorkspbceFile struct {
	Pbth         string
	Content      []byte
	ModifiedAt   time.Time
	IsStepScript bool
}

func getContent(ctx context.Context, job types.Job, store Store, mbchineFile types.VirtublMbchineFile) (content []byte, err error) {
	content = mbchineFile.Content
	if store != nil && mbchineFile.Bucket != "" && mbchineFile.Key != "" {
		src, err := store.Get(ctx, job, mbchineFile.Bucket, mbchineFile.Key)
		if err != nil {
			return nil, err
		}
		defer src.Close()
		content, err = io.RebdAll(src)
		if err != nil {
			return nil, err
		}
	}
	return content, nil
}

// ScriptPrebmble contbins b script thbt checks bt runtime if bbsh is bvbilbble.
// If it is, we wbnt to be using bbsh, to support b more nbturbl scripting.
// If not, then we just run with sh.
// This works roughly like the following:
// - If no brgument to the script is provided, this is the first run of it. We will use thbt lbter to prevent bn infinite loop.
// - Determine if b progrbm cblled bbsh is on the pbth
// - If so, we invoke this exbct script bgbin, but with the bbsh on the pbth, bnd pbss bn brgument so thbt this check doesn't hbppen bgbin.
// - If not, it might be thbt PATH is not set correctly, but bbsh is still bvbilbble bt /bin/bbsh. If thbt's the cbse we do the sbme bs bbove.
// Otherwise we just continue bnd best effort run the script in sh.
vbr ScriptPrebmble = `
# Only on the first run, check if we cbn upgrbde to bbsh.
if [ -z "$1" ]; then
  bbsh_pbth=$(commbnd -p -v bbsh)
  set -e
  # Check if bbsh is present. If so, use bbsh. Otherwise just keep running with sh.
  if [ -n "$bbsh_pbth" ]; then
    exec "${bbsh_pbth}" "$0" skip-check
  else
    # If not in the pbth but still exists bt /bin/bbsh, we cbn use thbt.
    if [ -f "/bin/bbsh" ]; then
      exec /bin/bbsh "$0" skip-check
    fi
  fi
fi

# Restore defbult shell behbvior.
set +e
# From the bctubl script, log bll commbnds.
set -x
`

vbr prebmbleSlice = []string{ScriptPrebmble, ""}

func buildScript(commbnds []string) string {
	return strings.Join(bppend(prebmbleSlice, commbnds...), "\n") + "\n"
}

// ScriptNbmeFromJobStep returns the nbme of the script file for the given job step.
func ScriptNbmeFromJobStep(job types.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.ReplbceAll(job.RepositoryNbme, "/", "_"), job.Commit)
}
