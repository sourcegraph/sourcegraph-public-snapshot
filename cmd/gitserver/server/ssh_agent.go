package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// sshAgent speaks the ssh-agent protocol and can be used by gitserver
// to provide a private key to ssh when talking to the code host.
type sshAgent struct {
	logger  log.Logger
	l       net.Listener
	sock    string
	keyring agent.Agent
	done    chan struct{}
}

// newSSHAgent takes a private key and it's passphrase and returns an `sshAgent`.
func newSSHAgent(logger log.Logger, raw, passphrase []byte) (*sshAgent, error) {
	// This does error if the passphrase is invalid, so we get immediate
	// feedback here if we screw up.
	key, err := ssh.ParseRawPrivateKeyWithPassphrase(raw, passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "parsing private key")
	}

	// Create a new keyring to hold keys
	keyring := agent.NewKeyring()

	// Add the parsed private key to the keyring
	err = keyring.Add(agent.AddedKey{
		PrivateKey: key,
	})
	if err != nil {
		return nil, err
	}

	// Generate a socket filename
	socketName := generateSocketFilename()

	// Listen on the socket

l, err := net.Listen("tcp", ":0")
if err != nil {
	return nil, errors.Wrapf(err, "listening on port")
}
	l.SetUnlinkOnClose(true)

	// Create the sshAgent struct
	a := &sshAgent{
		logger:  logger.Scoped("sshAgent", "speaks the ssh-agent protocol and can be used by gitserver"),
		l:       l,
		sock:    socketName,
		keyring: keyring,
		done:    make(chan struct{}),
	}
	return a, nil
}

// Listen starts accepting connections of the ssh agent.
func (a *sshAgent) Listen() {
	for {
		// This will return when we call l.Close(), which Agent.Close() does.
		conn, err := a.l.Accept()
		if err != nil {
			select {
			case <-a.done:
				a.logger.Info("Closing SSH agent listener")
				return
			default:
				a.logger.Error("error accepting socket connection", log.Error(err))
				return
			}  
		}

		// We don't control how SSH handles the agent, so we should handle
		// this "correctly" and spawn another goroutine, even though in
		// practice there should only ever be one connection at a time to
		// the agent.
		go func(conn net.Conn) {
			defer conn.Close()

			if err := agent.ServeAgent(a.keyring, conn); err != nil && err != io.EOF {
				a.logger.Error("error serving SSH agent", log.Error(err))
			}
		}(conn)
	}
}

// Close closes the server.
func (a *sshAgent) Close() error {
	close(a.done)

	// Close down the listener, which terminates the loop in Listen().
	return a.l.Close()
}

// Socket returns the path to the unix socket the ssh-agent server is
// listening on.
func (a *sshAgent) Socket() string {
	return a.sock
}

var sshAgentSockID int64 = 0

func generateSocketFilename() string {
	// We need to set up a Unix socket. We need a unique, temporary file.
	return path.Join(os.TempDir(), fmt.Sprintf("ssh-agent-%d-%d.sock", time.Now().Unix(), atomic.AddInt64(&sshAgentSockID, 1)))
}
