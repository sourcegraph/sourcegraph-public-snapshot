// Package sshgit provides functionality for a git server over SSH.
package sshgit

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/flynn/go-shlex"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/accesstoken"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/gitserver/pktline"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

const (
	// gitTransactionTimeout controls a hard deadline on the time to perform a single git transaction.
	gitTransactionTimeout = 30 * time.Minute
)

// Server is SSH git server.
type Server struct {
	listener net.Listener
	// ctx is the scoped context to be used for RPC requests.
	ctx context.Context
	key *idkey.IDKey

	config *ssh.ServerConfig
	// reposRoot is the path to the directory where repositories are stored.
	reposRoot string
}

// ListenAndStart listens on the TCP network address addr and starts the server.
func (s *Server) ListenAndStart(ctx context.Context, addr string, key *idkey.IDKey) error {
	if key == nil {
		return errors.New("ssh git server requires non-nil key")
	}
	privateSigner, err := ssh.NewSignerFromKey(key.Private())
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = l
	s.ctx = ctx
	s.key = key

	s.config = &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if c.User() != "git" {
				return nil, fmt.Errorf(`unsupported SSH user %q; use "git" for SSH git access`, c.User())
			}

			cl, err := sourcegraph.NewClientFromContext(s.ctx)
			if err != nil {
				return nil, err
			}
			userSpec, err := cl.UserKeys.LookupUser(s.ctx, &sourcegraph.SSHPublicKey{Key: key.Marshal()})
			if err != nil {
				return nil, err
			}
			user, err := cl.Users.Get(s.ctx, userSpec)
			if err != nil {
				return nil, err
			}

			// Generate access token for the authenticated user,
			// all authorization checks for git operations will be
			// done by the RPC server based on the token's scope.
			actor := auth.GetActorFromUser(user)
			actor.ClientID = s.key.ID
			tok, err := accesstoken.New(s.key, actor, map[string]string{"GrantType": "SSHPublicKey"}, 7*24*time.Hour)
			if err != nil {
				return nil, err
			}
			return &ssh.Permissions{
				CriticalOptions: map[string]string{
					uidKey:         fmt.Sprint(user.UID),
					userLoginKey:   user.Login,
					accessTokenKey: tok.AccessToken,
				},
			}, nil
		},
	}
	s.config.AddHostKey(privateSigner)

	s.reposRoot = filepath.Join(os.Getenv("SGPATH"), "repos")

	go s.run()
	return nil
}

func (s *Server) run() {
	for {
		tcpConn, err := s.listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %v\n", err)
			continue
		}
		tcpConn.SetDeadline(time.Now().Add(gitTransactionTimeout))
		go s.handleConn(tcpConn)
	}
}

func (s *Server) handleConn(tcpConn net.Conn) {
	sshConns.Inc()
	defer sshConns.Dec()

	defer tcpConn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, s.config)
	if err != nil {
		log.Printf("failed to ssh handshake: %v\n", err)
		return
	}
	go ssh.DiscardRequests(reqs)
	for ch := range chans {
		if t := ch.ChannelType(); t != "session" {
			ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %v", t))
			continue
		}
		go s.handleChannel(sshConn, ch)
	}
}

func (s *Server) handleChannel(sshConn *ssh.ServerConn, newChan ssh.NewChannel) {
	ch, reqs, err := newChan.Accept()
	if err != nil {
		return
	}
	defer ch.Close()
	for req := range reqs {
		switch req.Type {
		case "shell":
			if req.WantReply {
				req.Reply(true, nil)
			}
			fmt.Fprintf(ch, "Hi %q, you've successfully authenticated, but we don't provide shell access.\n", sshConn.Permissions.CriticalOptions[userLoginKey])
			ch.SendRequest("exit-status", false, ssh.Marshal(exitStatusMsg{0}))
			return
		case "exec":
			if req.WantReply {
				req.Reply(true, nil)
			}
			err := s.execGitCommand(sshConn, ch, req)
			if err != nil {
				log.Println(err)
			}
			// Respond back with the exit-status otherwise some
			// git clients may incorrectly report errors.
			_, err = ch.SendRequest("exit-status", false, ssh.Marshal(exitStatus(err)))
			if err != nil {
				log.Println("Error while sending exit-status:", err)
			}
			return
		case "env":
			if req.WantReply {
				req.Reply(true, nil)
			}
			// Do nothing.
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func (s *Server) execGitCommand(sshConn *ssh.ServerConn, ch ssh.Channel, req *ssh.Request) error {
	if len(req.Payload) < 4 {
		return fmt.Errorf("invalid git transport protocol payload (less than 4 bytes): %q", req.Payload)
	}
	command := string(req.Payload[4:]) // E.g., "git-upload-pack '/user/repo'".
	args, err := shlex.Split(command)  // E.g., []string{"git-upload-pack", "/user/repo"}.
	if err != nil || len(args) != 2 {
		return fmt.Errorf("command %q is not a valid git command", command)
	}

	// Validate operation.
	op := args[0] // E.g., "git-upload-pack".
	switch op {
	case "git-upload-pack", "git-receive-pack": // valid
	default:
		return fmt.Errorf("%q is not a supported git operation", op)
	}
	op = op[4:]

	repoURI := args[1] // E.g., "/user/repo".
	repoURI = path.Clean(repoURI)
	if path.IsAbs(repoURI) {
		repoURI = repoURI[1:] // Normalize "/user/repo" to "user/repo".
	}
	repoDir := filepath.Join(s.reposRoot, filepath.FromSlash(repoURI))
	if repoURI == "" || !strings.HasPrefix(repoDir, s.reposRoot) {
		fmt.Fprintf(ch.Stderr(), "Specified repo %q lies outside of root.\n\n", repoURI)
		return fmt.Errorf("specified repo %q lies outside of root", repoURI)
	}

	cl, err := sourcegraph.NewClientFromContext(s.ctx)
	if err != nil {
		return err
	}
	repo, err := cl.Repos.Get(s.ctx, &sourcegraph.RepoSpec{URI: repoURI})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			fmt.Fprintf(ch.Stderr(), "Specified repo %q does not exist.\n\n", repoURI)
			return fmt.Errorf("specified repo %q does not exist: %v", repoURI, err)
		}
		fmt.Fprintf(ch.Stderr(), "Error accessing repo %q: %v\n\n", repoURI, err)
		return fmt.Errorf("error accessing repo %q: %v", repoURI, err)
	}

	userCtx := sourcegraph.WithCredentials(s.ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: sshConn.Permissions.CriticalOptions[accessTokenKey], TokenType: "Bearer"}))

	userClient, err := sourcegraph.NewClientFromContext(userCtx)
	if err != nil {
		return err
	}
	userGitClient := gitpb.NewGitTransportClient(userClient.Conn)

	pkt, err := userGitClient.InfoRefs(userCtx, &gitpb.InfoRefsOp{
		Repo:    sourcegraph.RepoSpec{URI: repo.URI},
		Service: op,
	})
	if err != nil {
		return err
	}

	// Parse the info refs response, write back to channel omitting comments.
	scanner := bufio.NewScanner(bytes.NewReader(pkt.Data))
	scanner.Split(pktline.SplitFunc)
	var skipNextFlush bool
	for scanner.Scan() {
		if pktline.IsComment(scanner.Bytes()) {
			// NOTE: Over HTTP, clients seem to expect a flush
			// packet right after the HTTP-specific service
			// announcement packet. This is not documented in the
			// protocol specifications, but the SSH protocol doesn't
			// operate with it so it must be filtered out.
			skipNextFlush = true
			continue
		}
		if pktline.IsFlush(scanner.Bytes()) && skipNextFlush {
			skipNextFlush = false
			continue
		}
		_, err := ch.Write(scanner.Bytes())
		if err != nil {
			return err
		}
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	// For fetch operations buffer the input from the client during several
	// rounds of negotiations to be able to make stateless requests to the
	// RPC service. The response from the server should be sliced each time
	// to remove data already written to the client's channel.
	scanner = bufio.NewScanner(ch)
	scanner.Split(pktline.SplitFunc)
	buf := new(bytes.Buffer)
	written := 0
	done := false
	// According to git's spec: The packfile MUST NOT be sent if the only
	// command used is 'delete'.
	expectPACK := false
	for !done {
		for scanner.Scan() {
			pl := scanner.Bytes()
			_, err := buf.Write(pl)
			if err != nil {
				return err
			}
			if pktline.IsFlush(pl) {
				break
			}
			if op == "receive-pack" {
				if pktline.HasPrefix(pl, []byte("push-cert-end")) {
					break
				}
				if pktline.IsCreate(pl) || pktline.IsUpdate(pl) {
					expectPACK = true
				}
			} else {
				if pktline.HasPrefix(pl, []byte("done")) {
					done = true
					break
				}
			}
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}

		if op == "receive-pack" {
			done = true
			if expectPACK {
				// Read PACK from channel.
				_, err := io.Copy(buf, ch)
				if err != nil {
					return err
				}
			}
		} else if pktline.IsFlush(buf.Bytes()) {
			// From git protocol specs: "After reference and
			// capabilities discovery, the client can decide to
			// terminate the connection by sending a flush-pkt,
			// telling the server it can now gracefully terminate,
			// and disconnect, when it does not need any pack data.
			// This can happen with the ls-remote command, and also
			// can happen when the client already is up-to-date.
			// Otherwise, it enters the negotiation phase."
			return nil
		}
		zipped, err := compress(buf.Bytes())
		if err != nil {
			return err
		}
		if op == "receive-pack" {
			pkt, err = userGitClient.ReceivePack(userCtx, &gitpb.ReceivePackOp{
				Repo:            sourcegraph.RepoSpec{URI: repo.URI},
				ContentEncoding: "gzip",
				Data:            zipped,
			})
		} else {
			pkt, err = userGitClient.UploadPack(userCtx, &gitpb.UploadPackOp{
				Repo:            sourcegraph.RepoSpec{URI: repo.URI},
				ContentEncoding: "gzip",
				Data:            zipped,
			})
		}
		if err != nil {
			return err
		}
		_, err = ch.Write(pkt.Data[written:])
		if err != nil {
			return err
		}
		written = len(pkt.Data)
	}

	return nil
}

func compress(in []byte) ([]byte, error) {
	zipped := new(bytes.Buffer)
	zip := gzip.NewWriter(zipped)
	_, err := io.Copy(zip, bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	err = zip.Close()
	if err != nil {
		return nil, err
	}
	return zipped.Bytes(), nil
}

type exitStatusMsg struct {
	Status uint32
}

// exitStatus converts the error value from exec.Command.Wait() to an exitStatusMsg.
func exitStatus(err error) exitStatusMsg {
	switch err {
	case nil:
		return exitStatusMsg{0}
	default:
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return exitStatusMsg{uint32(status.ExitStatus())}
			}
		}
		return exitStatusMsg{1}
	}
}

const (
	uidKey         = "sourcegraph-uid"
	userLoginKey   = "sourcegraph-user-login"
	accessTokenKey = "sourcegraph-access-token"
)

func uidFromSSHConn(sshConn *ssh.ServerConn) int32 {
	uid, err := strconv.ParseInt(sshConn.Permissions.CriticalOptions[uidKey], 10, 32)
	if err != nil {
		panic(fmt.Errorf("strconv.ParseInt error shouldn't happen since we encode it ourselves, but it happened: %v", err))
	}
	return int32(uid)
}
