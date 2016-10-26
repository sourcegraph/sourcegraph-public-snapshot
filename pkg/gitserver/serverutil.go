package gitserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

func (s *Server) runWithRemoteOpts(cmd *exec.Cmd, opt *vcs.RemoteOpts) error {
	if opt != nil && opt.SSH != nil {
		gitSSHWrapper, gitSSHWrapperDir, keyFile, err := s.makeGitSSHWrapper(opt.SSH.PrivateKey)
		defer func() {
			if keyFile != "" {
				if err := os.Remove(keyFile); err != nil {
					log.Fatalf("Error removing SSH key file %s: %s.", keyFile, err)
				}
			}
		}()
		if err != nil {
			return err
		}
		defer os.Remove(gitSSHWrapper)
		if gitSSHWrapperDir != "" {
			defer os.RemoveAll(gitSSHWrapperDir)
		}
		cmd.Env = append(cmd.Env, "GIT_SSH="+gitSSHWrapper)
	}

	if opt != nil && opt.HTTPS != nil {
		gitPassHelperDir, err := s.makeGitPassHelper(opt.HTTPS.User, opt.HTTPS.Pass)
		if err != nil {
			return err
		}
		if gitPassHelperDir != "" {
			defer os.RemoveAll(gitPassHelperDir)
		}
		cmd.Args = append(cmd.Args[:1], append([]string{"-c", "credential.helper=gitserver-helper"}, cmd.Args[1:]...)...)
		env := environ(os.Environ())
		env.Set("PATH", gitPassHelperDir+string(filepath.ListSeparator)+os.Getenv("PATH"))
		cmd.Env = env
	}

	return cmd.Run()
}

// makeGitSSHWrapper writes a GIT_SSH wrapper that runs ssh with the
// private key. You should remove the sshWrapper, sshWrapperDir and
// the keyFile after using them.
func (s *Server) makeGitSSHWrapper(privKey []byte) (sshWrapper, sshWrapperDir, keyFile string, err error) {
	var otherOpt string
	if s.InsecureSkipCheckVerifySSH {
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

	tmpFile, tmpFileDir, err := s.gitSSHWrapper(keyFile, otherOpt)
	return tmpFile, tmpFileDir, keyFile, err
}

// gitSSHWrapper makes system-dependent SSH wrapper.
func (*Server) gitSSHWrapper(keyFile string, otherOpt string) (sshWrapperFile string, tempDir string, err error) {
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

// makeGitPassHelper writes a git credential helper that supplies username and password over stdout.
// Its name is "git-credential-gitserver-helper" and it's located inside gitPassHelperDir.
// If err is nil, the caller is responsible for removing gitPassHelperDir after it's done using it.
func (*Server) makeGitPassHelper(user, pass string) (gitPassHelperDir string, err error) {
	tempDir, err := ioutil.TempDir("", "gitserver_")
	if err != nil {
		return "", err
	}

	// Write the credentials content to credentialsFile file.
	// This is done to avoid code injection attacks.
	// Usernames and passwords are untrusted arbitrary user data. It's hard to escape
	// strings in shell scripts, so we opt to `cat` this non-executable credentials file instead.
	credentialsFile := filepath.Join(tempDir, "credentials-content")
	{
		// Always provide username and password via git credential helper.
		// Do this even if some of the values are blank strings.
		// Otherwise, git will fallback to asking via other means.
		content := fmt.Sprintf("username=%s\npassword=%s\n", user, pass)

		err := util.WriteFileWithPermissions(credentialsFile, []byte(content), 0600)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	// Write the credential helper executable that uses credentialsFile.
	{
		// We assume credentialsFile can be escaped with a simple wrapping of single
		// quotes. The path is not user controlled so this assumption should
		// not be violated.
		content := fmt.Sprintf("#!/bin/sh\ncat '%s'\n", credentialsFile)

		path := filepath.Join(tempDir, "git-credential-gitserver-helper")
		err := util.WriteFileWithPermissions(path, []byte(content), 0500)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	return tempDir, nil
}

// repoExists checks if dir is a valid GIT_DIR.
func repoExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "HEAD"))
	return !os.IsNotExist(err)
}

func recoverAndLog() {
	if err := recover(); err != nil {
		log.Print(err)
	}
}

// environ is a slice of strings representing the environment, in the form "key=value".
type environ []string

// Set environment variable key to value.
func (e *environ) Set(key, value string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = key + "=" + value
			return
		}
	}
	// If we get here, it's because the key isn't already present, so add a new one.
	*e = append(*e, key+"="+value)
}

// Unset environment variable key.
func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			return
		}
	}
}
