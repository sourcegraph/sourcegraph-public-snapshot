package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/flynn/go-shlex"

	"golang.org/x/crypto/ssh"
)

// NewServer creates a new test SSH server that runs a shell
// command upon login (with the current directory set to dir). It can
// be used to test remote SSH communication.
func NewServer(shell, dir string, opt ...func(*Server) error) (*Server, error) {
	s := &Server{Shell: shell, Dir: dir}

	for _, opt := range opt {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Server is an SSH server.
type Server struct {
	Shell string
	Dir   string

	SSH ssh.ServerConfig

	GitURL string

	l *net.TCPListener

	closedMu sync.Mutex
	closed   bool // whether l is closed
}

// PrivateKey sets the server's private key and host key.
func PrivateKey(pemData []byte) func(*Server) error {
	return func(s *Server) error {
		privKey, err := ssh.ParseRawPrivateKey(pemData)
		if err != nil {
			return err
		}
		hostKey, err := ssh.NewSignerFromKey(privKey)
		if err != nil {
			return err
		}
		pubKey := hostKey.PublicKey()

		s.SSH.AddHostKey(hostKey)
		s.SSH.PublicKeyCallback = func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(key.Marshal(), pubKey.Marshal()) {
				return nil, nil
			}
			return nil, errors.New("public key rejected")
		}

		return nil
	}
}

// Verbose enables verbose logging.
func Verbose(s *Server) error {
	s.SSH.AuthLogCallback = func(conn ssh.ConnMetadata, method string, err error) {
		log.Printf("ssh: user %q, method %q: %v", conn.User(), method, err)
	}
	return nil
}

// Start starts the server in a goroutine. If the server was unable to
// start, an error is returned.
func (s *Server) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	s.l = l

	s.GitURL = fmt.Sprintf("ssh://go-vcs@%s", s.l.Addr().String())

	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				s.closedMu.Lock()
				if s.closed {
					s.closedMu.Unlock()
					return
				}
				s.closedMu.Unlock()
				log.Println(err)
				continue
			}
			go s.handleConn(conn)
		}
	}()

	return nil
}

// SSH server code adapted from gitreceived in Flynn
// (https://sourcegraph.com/flynn/flynn/.tree/gitreceived/gitreceived.go).

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, &s.SSH)
	if err != nil {
		log.Println("Failed to handshake:", err)
		return
	}

	go ssh.DiscardRequests(reqs)

	for ch := range chans {
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		go s.handleChannel(sshConn, ch)
	}
}

func (s *Server) handleChannel(conn *ssh.ServerConn, newChan ssh.NewChannel) {
	ch, reqs, err := newChan.Accept()
	if err != nil {
		log.Println("newChan.Accept failed:", err)
		return
	}
	defer ch.Close()
	for req := range reqs {
		switch req.Type {
		case "exec":
			fail := func(at string, err error) {
				log.Printf("%s failed: %s", at, err)
				ch.Stderr().Write([]byte("Internal error.\n"))
			}
			if req.WantReply {
				req.Reply(true, nil)
			}
			cmdline := string(req.Payload[4:])
			cmdargs, err := shlex.Split(cmdline)
			if err != nil || len(cmdargs) != 2 {
				ch.Stderr().Write([]byte("Invalid arguments.\n"))
				return
			}
			if cmdargs[0] != "git-upload-pack" {
				ch.Stderr().Write([]byte("Only `git fetch` is supported.\n"))
				return
			}
			cmdargs[1] = strings.TrimSuffix(strings.TrimPrefix(cmdargs[1], "/"), ".git")
			if strings.Contains(cmdargs[1], "..") {
				ch.Stderr().Write([]byte("Invalid repo.\n"))
				return
			}

			cmd := exec.Command(s.Shell, "-c", cmdargs[0]+" '"+cmdargs[1]+"'")
			cmd.Dir = s.Dir
			cmd.Env = append(os.Environ(),
				"RECEIVE_USER="+conn.User(),
				"RECEIVE_REPO="+cmdargs[1],
			)
			done, err := attachCmd(cmd, ch, ch.Stderr(), ch)
			if err != nil {
				fail("attachCmd", err)
				return
			}
			if err := cmd.Start(); err != nil {
				fail("cmd.Start", err)
				return
			}
			done.Wait()
			status, err := exitStatus(cmd.Wait())
			if err != nil {
				fail("exitStatus", err)
				return
			}
			if _, err := ch.SendRequest("exit-status", false, ssh.Marshal(&status)); err != nil {
				fail("sendExit", err)
			}
			return
		case "env":
			if req.WantReply {
				req.Reply(true, nil)
			}
		}
	}
}

func (s *Server) Close() error {
	s.closedMu.Lock()
	s.closed = true
	s.closedMu.Unlock()
	return s.l.Close()
}

func attachCmd(cmd *exec.Cmd, stdout, stderr io.Writer, stdin io.Reader) (*sync.WaitGroup, error) {
	var wg sync.WaitGroup
	wg.Add(2)

	stdinIn, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrOut, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		io.Copy(stdinIn, stdin)
		stdinIn.Close()
	}()
	go func() {
		io.Copy(stdout, stdoutOut)
		wg.Done()
	}()
	go func() {
		io.Copy(stderr, stderrOut)
		wg.Done()
	}()

	return &wg, nil
}

type exitStatusMsg struct {
	Status uint32
}

func exitStatus(err error) (exitStatusMsg, error) {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return exitStatusMsg{uint32(status.ExitStatus())}, nil
			}
		}
		return exitStatusMsg{0}, err
	}
	return exitStatusMsg{0}, nil
}

// SamplePrivKey is taken from go.crypto's testdata.
var SamplePrivKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA2eP2yEw47wYi7DCgNKDAswlR4rRSKpDSOGnYst1Tj2Lkt6BH
3HWMUCgSZfCG2JzE6WSZaOsKrSxvq6+n0qSOxm8CR8TlJcpIerFycR1IioAPOTaT
RtDVZnQEIDGH7TCEJHFZZ600IiIeQF2JOfiqe8FbDlw/Z8AqWTn5Vh5aIq32MosO
3BUkHku9lgTj1NyGc0al20nSCf/btdlH/oQQVtUJNwfYj/jx/dvfg70pD9IINR5r
S9dwgfNi5Uk1nwscY2pduA9rp1TB9uXQUyFvFjIMslrR0jcCIXXuiVg7Fc+Uttqr
YH9+qVPTSrWwb6Mm4VEwY2RVOBhzvZ9Sp/RoXQIDAQABAoIBAQC8vwfw3G5ZSASK
e1jcHgCvVrxzWObwboFcUvxffPA8fltIYfS+GamRahT970ywaaT91KI7y5d1CdA2
djQ3eUsgw9rC1uH1SXRdrEdJiydiqqoFUqxjpNWnKYrFZIKtyeA+PV5IPDaz9sAj
26La7/imuYkqOGjIdCN7JYhCvIoyDMx5iNlTGgBK/fmHrhnFj57fl87Y5ZIcFDl1
7oEq76TeuqT9wGpK0NfC9CfKq4XqhHLEVJLXUlKjBqqd93UTjvZljTjV9rezhsDs
h1MSK54PD1oMH7O1LeQSy0NpkwCHqkccBg+163FLUwre7VokdhqO5wPQP10NUmoR
19FaZ4O1AoGBAPYG1Dbc6xGoYB8vn8rmTFjQodadwIo0qx3a1sN6pQF2gxumerNM
2MSu7TzmaD6A3fHuFS4/0RNzvFOYtQ6OnpzpqjskBSiVmlwo+FIakuqjE+nGYhjI
dgfeYkTf1nhyfwwAmjNxfJBFukryWnXcM2zCjJa7rYH2hCEh5ZZKWs2DAoGBAOK5
JftPRFG45X9JfLcd4MNG5xFfCe6is/6dDlz9HWswm0mKO192mWh58npzXJ/kewX7
UXIhpqOqcKCum2YfbR2t2mw80oYDjZ4rgSdBiB4aQA4i1DGAUweUXFkSZZlWntp0
ucxRLLflCfjjtk7ozmVoWU+HgxsvpxNF/WgaXuyfAoGBAOhDtC8DS00FR5HJlTKp
TqR+ensxvOb9KBrsUdqEO6jw6H+/IJGLWA3/EutunjV71Yyj9w0NpGWX2tCVF0Fh
9W4vzs08iT4yVmLxLtXcTp0DTjZiWpQJFB0DnoRlSYW2miiLnQg5+J3/pgtBV5Nz
Sn1AAhf/oKNURpM8/BFxqt3fAoGBAIF+rqrzg1oJ+UrSdmFAt3fRr3jEh6+9ToFG
w0VpbLwkbw153p+P5d8+h7hY27aXkYzBFqvRfJRObTXZhPi3SmOBQRhBRR02OlT1
FDePvmczJxLr4bbETKgvnO9jCpSiXOj5coW4d4oxT5jQtvgrEHfrOdeq1r9YYF0p
xKsJJN6RAoGAY0dKncvCQaBnKEsdnYXGZTkiWkbP3f9lvHqcpsNOeMionSn+MNNV
eLSIhRZBZeDbd5/GJsmKcXPqD6St3nShObjJva8/zw63sBT9nng65RuU2oyCnSwu
NdnXpXB1bDYDbc3Gy33eJck6czWrHsOREM/d4L8xBfTSzr24I36qk5Y=
-----END RSA PRIVATE KEY-----`)
