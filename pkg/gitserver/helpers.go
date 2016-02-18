package gitserver

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"src.sourcegraph.com/sourcegraph/pkg/vcs/util"
)

// InsecureSkipCheckVerifySSH controls whether the client verifies the
// SSH server's certificate or host key. If InsecureSkipCheckVerifySSH
// is true, the program is susceptible to a man-in-the-middle
// attack. This should only be used for testing.
var InsecureSkipCheckVerifySSH bool

// makeGitSSHWrapper writes a GIT_SSH wrapper that runs ssh with the
// private key. You should remove the sshWrapper, sshWrapperDir and
// the keyFile after using them.
func makeGitSSHWrapper(privKey []byte) (sshWrapper, sshWrapperDir, keyFile string, err error) {
	var otherOpt string
	if InsecureSkipCheckVerifySSH {
		otherOpt = "-o StrictHostKeyChecking=no"
	}

	kf, err := ioutil.TempFile("", "go-vcs-gitcmd-key")
	if err != nil {
		return "", "", "", err
	}
	keyFile = kf.Name()
	err = util.WriteFileWithPermissions(keyFile, privKey, 0600)
	if err != nil {
		return "", "", keyFile, err
	}

	tmpFile, tmpFileDir, err := gitSSHWrapper(keyFile, otherOpt)
	return tmpFile, tmpFileDir, keyFile, err
}

// environ is a slice of strings representing the environment, in the form "key=value".
type environ []string

// Unset a single environment variable.
func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			break
		}
	}
}

// Makes system-dependent SSH wrapper
func gitSSHWrapper(keyFile string, otherOpt string) (sshWrapperFile string, tempDir string, err error) {
	// TODO(sqs): encrypt and store the key in the env so that
	// attackers can't decrypt if they have disk access after our
	// process dies

	var script string

	if runtime.GOOS == "windows" {
		script = `
	@echo off
	ssh -o ControlMaster=no -o ControlPath=none ` + otherOpt + ` -i ` + filepath.ToSlash(keyFile) + ` "%@"
`
	} else {
		script = `
	#!/bin/sh
	exec /usr/bin/ssh -o ControlMaster=no -o ControlPath=none ` + otherOpt + ` -i ` + filepath.ToSlash(keyFile) + ` "$@"
`
	}

	sshWrapperName, tempDir, err := util.ScriptFile("go-vcs-gitcmd")
	if err != nil {
		return sshWrapperName, tempDir, err
	}

	err = util.WriteFileWithPermissions(sshWrapperName, []byte(script), 0500)
	return sshWrapperName, tempDir, err
}

// makeGitPassHelper writes a GIT_ASKPASS helper that supplies password over stdout.
// You should remove the passHelper (and tempDir if any) after using it.
func makeGitPassHelper(pass string) (passHelper string, tempDir string, err error) {
	tmpFile, dir, err := util.ScriptFile("go-vcs-gitcmd-ask")
	if err != nil {
		return tmpFile, dir, err
	}

	passPath := filepath.Join(dir, "password")
	err = util.WriteFileWithPermissions(passPath, []byte(pass), 0600)
	if err != nil {
		return tmpFile, dir, err
	}

	var script string

	// We assume passPath can be escaped with a simple wrapping of single
	// quotes. The path is not user controlled so this assumption should
	// not be violated.
	if runtime.GOOS == "windows" {
		script = "@echo off\ntype " + passPath + "\n"
	} else {
		script = "#!/bin/sh\ncat '" + passPath + "'\n"
	}

	err = util.WriteFileWithPermissions(tmpFile, []byte(script), 0500)
	return tmpFile, dir, err
}
