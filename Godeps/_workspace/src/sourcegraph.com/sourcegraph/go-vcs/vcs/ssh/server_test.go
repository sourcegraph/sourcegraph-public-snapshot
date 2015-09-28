package ssh

import (
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestServer(t *testing.T) {
	shellScript := `#!/bin/sh
echo $*
exit
`
	shell, err := ioutil.TempFile("", "govcs-ssh-shell")
	if err != nil {
		t.Fatal(err)
	}
	shell.WriteString(shellScript)
	if err := shell.Chmod(0700); err != nil {
		t.Fatal(err)
	}
	if err := shell.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(shell.Name())

	s, err := NewServer(shell.Name(), "/tmp", PrivateKey(SamplePrivKey))
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
	if got, want := string(out), "-c git-upload-pack 'foo'\n"; got != want {
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
