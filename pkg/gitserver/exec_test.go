package gitserver

import (
	"bytes"
	"errors"
	"net/rpc"
	"testing"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

func TestExec(t *testing.T) {
	type execTest struct {
		reply1         ExecReply
		reply2         ExecReply
		error1         error
		error2         error
		expectedErr    error
		expectedStdout []byte
		expectedStderr []byte
	}

	rpcError := errors.New("rpc error")

	tests := []*execTest{
		{
			reply1:      ExecReply{RepoExists: false},
			reply2:      ExecReply{RepoExists: false},
			expectedErr: vcs.ErrRepoNotExist,
		},
		{
			reply1:         ExecReply{RepoExists: true, Stdout: []byte("out"), Stderr: []byte("err")},
			reply2:         ExecReply{RepoExists: false},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:         ExecReply{RepoExists: false},
			reply2:         ExecReply{RepoExists: true, Stdout: []byte("out"), Stderr: []byte("err")},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:         ExecReply{RepoExists: true, Stdout: []byte("out"), Stderr: []byte("err")},
			error2:         rpcError,
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			error1:         rpcError,
			reply2:         ExecReply{RepoExists: true, Stdout: []byte("out"), Stderr: []byte("err")},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:      ExecReply{RepoExists: false},
			error2:      rpcError,
			expectedErr: rpcError,
		},
		{
			error1:      rpcError,
			reply2:      ExecReply{RepoExists: false},
			expectedErr: rpcError,
		},
		{
			error1:      rpcError,
			error2:      rpcError,
			expectedErr: rpcError,
		},
	}

	for _, test := range tests {
		server1 := make(chan *rpc.Call)
		server2 := make(chan *rpc.Call)
		servers = [](chan<- *rpc.Call){server1, server2}

		go func() {
			call1 := <-server1
			*call1.Reply.(*ExecReply) = test.reply1
			call1.Error = test.error1
			call1.Done <- call1

			call2 := <-server2
			*call2.Reply.(*ExecReply) = test.reply2
			call2.Error = test.error2
			call2.Done <- call2
		}()

		stdout, stderr, err := Command("git", "test").DividedOutput()
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
