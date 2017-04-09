package gitserver

import (
	"bytes"
	"context"
	"testing"

	"github.com/neelance/chanrpc/chanrpcutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestExec(t *testing.T) {
	type execTest struct {
		context        context.Context
		repo           *sourcegraph.Repo
		reply          *protocol.ExecReply
		expectedErr    error
		expectedStdout []byte
		expectedStderr []byte
		expectedPass   string
	}

	tests := []*execTest{
		{
			context:     context.Background(),
			repo:        &sourcegraph.Repo{URI: "github.com/gorilla/mux"},
			reply:       &protocol.ExecReply{RepoNotFound: true},
			expectedErr: vcs.RepoNotExistError{},
		},
		{
			context:        actor.WithActor(context.Background(), &actor.Actor{GitHubToken: "token"}),
			repo:           &sourcegraph.Repo{URI: "github.com/gorilla/mux"},
			reply:          &protocol.ExecReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
			expectedPass:   "", // no pass, repo not private
		},
		{
			context:        actor.WithActor(context.Background(), &actor.Actor{GitHubToken: "token"}),
			repo:           &sourcegraph.Repo{URI: "github.com/gorilla/mux", Private: true},
			reply:          &protocol.ExecReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
			expectedPass:   "token",
		},
		{
			context:     context.Background(),
			repo:        &sourcegraph.Repo{URI: "github.com/gorilla/mux"},
			expectedErr: errRPCFailed,
		},
		{
			context:     context.Background(),
			repo:        &sourcegraph.Repo{URI: "github.com/gorilla/mux"},
			reply:       &protocol.ExecReply{CloneInProgress: true},
			expectedErr: vcs.RepoNotExistError{CloneInProgress: true},
		},
	}

	for _, test := range tests {
		server := make(chan *protocol.Request)
		client := &Client{servers: [](chan<- *protocol.Request){server}}

		go func(test *execTest) {
			req := <-server
			pass := ""
			if req.Exec.Opt.HTTPS != nil {
				pass = req.Exec.Opt.HTTPS.Pass
			}
			if pass != test.expectedPass {
				t.Errorf("expected pass %#v, got %#v", test.expectedPass, pass)
			}
			chanrpcutil.Drain(req.Exec.Stdin)
			if test.reply != nil {
				req.Exec.ReplyChan <- test.reply
			}
			close(req.Exec.ReplyChan)
		}(test)

		cmd := client.Command("git", "test")
		cmd.Repo = test.repo
		stdout, stderr, err := cmd.DividedOutput(test.context)
		if err != test.expectedErr {
			t.Errorf("expected error %#v, got %#v", test.expectedErr, err)
		}
		if !bytes.Equal(stdout, test.expectedStdout) {
			t.Errorf("expected stdout %#v, got %#v", test.expectedStdout, stdout)
		}
		if !bytes.Equal(stderr, test.expectedStderr) {
			t.Errorf("expected stdout %#v, got %#v", test.expectedStderr, stderr)
		}
	}
}

func emptyProcessResult() <-chan *protocol.ProcessResult {
	processResultChan := make(chan *protocol.ProcessResult, 1)
	processResultChan <- &protocol.ProcessResult{}
	return processResultChan
}
