package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	sshInitDir  = env.Get("SSH_INIT_DIR", "", "the directory from which SSH credentials are copied (typically contains the user's SSH credentials")
	sshDir      = env.Get("SSH_DIR", "", "the directory to which SSH credentials are copied, typically `~/.ssh`")
	currentUser string
)

func init() {
	currentUser = os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("USERNAME")
	}
	if currentUser == "" {
		currentUser = "root"
	}
}

// InitializeSSH best-effort initializes the .ssh directory using the contents of another directory.
// This should only be used for the local installer use case (mounting the host's .ssh directory as
// the source directory for SSH credentials).
func InitializeSSH() error {
	if sshInitDir == "" || sshDir == "" {
		return nil
	}
	if currentUser == "" {
		return fmt.Errorf("could not determine current user")
	}
	if runtime.GOOS != "linux" {
		return fmt.Errorf("unsupported on non-Linux platforms")
	}
	if _, err := os.Stat(sshInitDir); os.IsNotExist(err) {
		return err
	}
	log.Printf("initializing SSH directory (%s -> %s)", sshInitDir, sshDir)
	if _, err := os.Stat(sshDir); err == nil {
		if err := exec.Command("mv", sshDir, fmt.Sprintf("%s.old", sshDir)).Run(); err != nil {
			return err
		}
	}
	if err := exec.Command("cp", "-r", sshInitDir, sshDir).Run(); err != nil {
		return err
	}
	if err := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", currentUser, currentUser), sshDir).Run(); err != nil {
		return err
	}
	log.Printf("initialized SSH directory")
	return nil
}
