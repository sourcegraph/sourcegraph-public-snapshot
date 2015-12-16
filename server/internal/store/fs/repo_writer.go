package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"strings"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

var (
	changesetsRefAuthor = vcs.Signature{
		Name:  "Sourcegraph",
		Email: "noreply@sourcegraph.com",
	}
	changesetsRefCommitter = changesetsRefAuthor
)

// changesetsRepoStage manages the staging area of a repository's ref and allows committing
// into it. Once initialized, it is possible to consequently stage & commit files.
// When the operation is completed, it is recommended that the Free() method be
// called.
type changesetsRepoStage struct {
	// stagingDir is a temp dir containing the repo's clone. In this
	// dir we stage and commit the changes, which is then `git push`'d
	// to the original repo. Using a temp staging repo dir lets us
	// avoid changing the index/work tree and concurrency conflicts
	// when editing the original repo (which are possibly destructive
	// actions).
	stagingDir string

	// repoDir is the dir containing the original repo.
	repoDir string

	refName string

	gitPassHelper    string
	gitPassHelperDir string
}

// changesetsNewRepoStage creates a new changesetsRepoStage to stage & commit into the
// repository located at the given repoPath at the ref specified by
// refName. It creates a staging repo in a temp dir to create the git
// index, commit it, and push to the original repo. This lets it avoid
// concurrency conflicts.
//
// When done, you MUST call the changesetsRepoStage's Free to remove the temp
// dir it creates.
func changesetsNewRepoStage(repoPath, refName string, password string) (rs *changesetsRepoStage, err error) {
	if err := checkGitArgSafety(repoPath); err != nil {
		return nil, err
	}

	rs = &changesetsRepoStage{
		repoDir: repoPath,
		refName: refName,
	}

	rs.stagingDir, err = ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer func() {
		// On error, clean up the abandoned staging dir.
		if err != nil && rs != nil {
			os.RemoveAll(rs.stagingDir)
		}
	}()

	if password != "" {
		rs.gitPassHelper, rs.gitPassHelperDir, err = makeGitPassHelper(password)
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.Command("git", "init", "--quiet")
	cmd.Dir = rs.stagingDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("exec %v: %s (output follows)\n\n%s", cmd.Args, err, out)
	}

	cmd = exec.Command("git", "pull", repoPath, refName)
	cmd.Dir = rs.stagingDir
	cmd.Env = rs.getEnviron()
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("exec %v: %s (output follows)\n\n%s", cmd.Args, err, out)
	}

	return rs, nil
}

// commit commits the staged files into the specified ref. It also
// pushes from the staging repo to the original repo, so that the
// commit is available to future readers.
func (rs *changesetsRepoStage) commit(author, committer vcs.Signature, message string) error {
	// Create commit in staging repo.
	authorStr := fmt.Sprintf("%s <%s>", author.Name, author.Email)
	cmd := exec.Command(
		"git", "commit",
		"--message="+message,
		"--author="+authorStr,
		"--date="+author.Date.Time().Format(time.RFC3339),
	)
	cmd.Dir = rs.stagingDir
	cmd.Env = append(os.Environ(),
		"GIT_COMMITTER_NAME="+committer.Name,
		"GIT_COMMITTER_EMAIL="+committer.Email,
		"GIT_COMMITTER_DATE="+committer.Date.Time().Format(time.RFC3339),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exec %v: %s (output follows)\n\n%s", cmd.Args, err, out)
	}

	// Push commit to original repo.
	//
	// TODO(sqs): Check for git 2.6(?) and use atomic pushes if
	// available.
	cmd = exec.Command("git", "push", rs.repoDir, "HEAD:"+rs.refName)
	cmd.Dir = rs.stagingDir
	cmd.Env = rs.getEnviron()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exec %v: %s (output follows)\n\n%s", cmd.Args, err, out)
	}

	return nil
}

// pull pulls the specified head branch into the current branch. The resulting
// changes will only be staged, so you must call changesetsRepoStage.commit if you want
// to commit merged changes.
func (rs *changesetsRepoStage) pull(head string, squash bool) error {
	if err := checkGitArgSafety(head); err != nil {
		return err
	}

	// Pull the head branch of the changeset into the base.
	args := []string{"pull", "--no-commit"}
	if squash {
		args = append(args, "--squash")
	} else {
		args = append(args, "--no-ff")
	}
	args = append(args, rs.repoDir, head)
	cmd := exec.Command("git", args...)
	cmd.Dir = rs.stagingDir
	env := rs.getEnviron()
	// Git requires you to configure a name and email to use "git pull", even if
	// you aren't committing anything.
	cmd.Env = append(env,
		"GIT_COMMITTER_NAME="+changesetsRefCommitter.Name,
		"GIT_COMMITTER_EMAIL="+changesetsRefCommitter.Email,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return execError(cmd.Args, err, out)
	}

	return nil
}

func (rs *changesetsRepoStage) getEnviron() []string {
	env := environ(os.Environ())

	if rs.gitPassHelper != "" {
		env.Unset("GIT_TERMINAL_PROMPT")
		env = append(env, "GIT_ASKPASS="+rs.gitPassHelper)
	}

	return env
}

// free frees up the resources used by the allocated repository and index.
func (rs *changesetsRepoStage) free() error {
	os.RemoveAll(rs.gitPassHelperDir)
	return os.RemoveAll(rs.stagingDir)
}

// checkGitArgSafety returns a non-nil error if a user-supplied arg beigins
// with a "-", which could cause it to be interpreted as a git command line
// argument.
func checkGitArgSafety(arg string) error {
	if strings.HasPrefix(arg, "-") {
		return fmt.Errorf("invalid git arg \"%s\" (begins with '-')", arg)
	}
	return nil
}

func execError(args []string, err error, out []byte) error {
	return fmt.Errorf("exec %v: %s (output follows)\n\n%s", args, err, out)
}

// TODO(renfred) the following methods (makeGitPassHelper, scriptFile,
// writeFileWithPermissions, environ.Unset), are copied from
// go-vcs/vcs/gitcmd/repo.go to aid in securely providing a password for git
// operations. Eventually when we remove support for Changeset persistence
// using refs, we can refactor changesetsRepoStage to use go-vcs and remove these
// methods.
func makeGitPassHelper(pass string) (passHelper string, tempDir string, err error) {
	tmpFile, dir, err := scriptFile("repo-stage-gitcmd-ask")
	if err != nil {
		return tmpFile, dir, err
	}

	passPath := filepath.Join(dir, "password")
	err = writeFileWithPermissions(passPath, []byte(pass), 0600)
	if err != nil {
		return tmpFile, dir, err
	}

	var script string

	if runtime.GOOS == "windows" {
		script = "@echo off\ntype " + passPath + "\n"
	} else {
		script = "#!/bin/sh\ncat '" + passPath + "'\n"
	}

	err = writeFileWithPermissions(tmpFile, []byte(script), 0500)
	return tmpFile, dir, err
}

func scriptFile(prefix string) (filePath string, rootPath string, err error) {
	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".bat"
	}

	tempDir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(tempDir, prefix+suffix), tempDir, nil
}

func writeFileWithPermissions(file string, content []byte, perm os.FileMode) error {
	err := ioutil.WriteFile(file, content, perm)
	if err != nil {
		return err
	}
	// ioutil.WriteFile applies permissions only for files that weren't exist
	return os.Chmod(file, perm)
}

type environ []string

func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			break
		}
	}
}
