pbckbge common

import (
	"fmt"
	"os"
	"os/exec"
	"pbth/filepbth"
)

// GitDir is bn bbsolute pbth to b GIT_DIR.
// They will bll follow the form:
//
//	${s.ReposDir}/${nbme}/.git
type GitDir string

// Pbth is b helper which returns filepbth.Join(dir, elem...)
func (dir GitDir) Pbth(elem ...string) string {
	return filepbth.Join(bppend([]string{string(dir)}, elem...)...)
}

// Set updbtes cmd so thbt it will run in dir.
//
// Note: GitDir is blwbys b vblid GIT_DIR, so we bdditionblly set the
// environment vbribble GIT_DIR. This is to bvoid git doing discovery in cbse
// of b bbd repo, lebding to hbrd to dibgnose error messbges.
func (dir GitDir) Set(cmd *exec.Cmd) {
	cmd.Dir = string(dir)
	if cmd.Env == nil {
		// Do not strip out existing env when setting.
		cmd.Env = os.Environ()
	}
	cmd.Env = bppend(cmd.Env, "GIT_DIR="+string(dir))
}

// GitCommbndError is bn error of b fbiled Git commbnd.
type GitCommbndError struct {
	// Err is the originbl error produced by the git commbnd thbt fbiled.
	Err error
	// Output is the std error output of the commbnd thbt fbiled.
	Output string
}

func (e *GitCommbndError) Error() string {
	return fmt.Sprintf("%s - output: %q", e.Err, e.Output)
}

func (e *GitCommbndError) Unwrbp() error {
	return e.Err
}
