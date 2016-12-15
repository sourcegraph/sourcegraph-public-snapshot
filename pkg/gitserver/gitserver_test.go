package gitserver

import (
	"bytes"
	"context"
	"testing"

	"github.com/neelance/chanrpc/chanrpcutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestExec(t *testing.T) {
	type execTest struct {
		reply          *execReply
		expectedErr    error
		expectedStdout []byte
		expectedStderr []byte
	}

	tests := []*execTest{
		{
			reply:       &execReply{RepoNotFound: true},
			expectedErr: vcs.RepoNotExistError{},
		},
		{
			reply:          &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply:          &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			expectedErr: errRPCFailed,
		},
		{
			reply:       &execReply{CloneInProgress: true},
			expectedErr: vcs.RepoNotExistError{CloneInProgress: true},
		},
	}

	for _, test := range tests {
		server := make(chan *request)
		client := &Client{servers: [](chan<- *request){server}}

		go func(test *execTest) {
			req := <-server
			chanrpcutil.Drain(req.Exec.Stdin)
			if test.reply != nil {
				req.Exec.ReplyChan <- test.reply
			}
			close(req.Exec.ReplyChan)
		}(test)

		stdout, stderr, err := client.Command("git", "test").DividedOutput(context.Background())
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

func emptyProcessResult() <-chan *processResult {
	processResultChan := make(chan *processResult, 1)
	processResultChan <- &processResult{}
	return processResultChan
}
