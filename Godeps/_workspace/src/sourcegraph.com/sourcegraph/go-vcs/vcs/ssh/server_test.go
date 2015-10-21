package ssh

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"

	"sourcegraph.com/sourcegraph/go-vcs/vcs/internal"
)

func TestServer(t *testing.T) {
	var shellScript string

	if runtime.GOOS == "windows" {
		shellScript = `@echo off
	set flag=%1
	set flag=%flag:"=%
	set args=%2
	set args=%args:"=%
	echo %flag% %args%
`
	} else {
		shellScript = `#!/bin/sh
echo $*
exit
`
	}

	shell, err := internal.ScriptFile("govcs-ssh-shell")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(shell)

	err = internal.WriteFileWithPermissions(shell, []byte(shellScript), 0700)
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewServer(shell, os.TempDir(), PrivateKey(SamplePrivKey))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	// Client
	cauth, err := clientAuth(SamplePrivKey)
	if err != nil {
		t.Fatal(err)
	}
	cconf := ssh.ClientConfig{User: "go-vcs"}
	cconf.Auth = append(cconf.Auth, cauth)
	sshc, err := ssh.Dial(s.l.Addr().Network(), s.l.Addr().String(), &cconf)
	if err != nil {
		t.Fatal(err)
	}
	defer sshc.Close()

	session, err := sshc.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	out, err := session.CombinedOutput("git-upload-pack 'foo'")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(string(out)), "-c git-upload-pack 'foo'"; got != want {
		t.Errorf("got ssh session output %q, want %q", got, want)
	}
}

func clientAuth(pemData []byte) (ssh.AuthMethod, error) {
	privKey, err := ssh.ParseRawPrivateKey(pemData)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
