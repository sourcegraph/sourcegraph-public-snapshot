package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

var (
	sshInitDir  = os.Getenv("SSH_INIT_DIR")
	sshDir      = os.Getenv("SSH_DIR")
	currentUser string
)

func init() {
	if u, err := user.Current(); err == nil {
		currentUser = u.Username
	} else {
		currentUser = "root"
	}
}

func initializeSSH() error {
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
