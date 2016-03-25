package gitserver

import (
	"bytes"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestExec(t *testing.T) {
	type execTest struct {
		reply1         *execReply
		reply2         *execReply
		expectedErr    error
		expectedStdout []byte
		expectedStderr []byte
	}

	tests := []*execTest{
		{
			reply1:      &execReply{RepoNotExist: true},
			reply2:      &execReply{RepoNotExist: true},
			expectedErr: vcs.ErrRepoNotExist,
		},
		{
			reply1:         &execReply{Stdout: []byte("out"), Stderr: []byte("err")},
			reply2:         &execReply{RepoNotExist: true},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:         &execReply{RepoNotExist: true},
			reply2:         &execReply{Stdout: []byte("out"), Stderr: []byte("err")},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:         &execReply{Stdout: []byte("out"), Stderr: []byte("err")},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply2:         &execReply{Stdout: []byte("out"), Stderr: []byte("err")},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:      &execReply{RepoNotExist: true},
			expectedErr: errRPCFailed,
		},
		{
			reply2:      &execReply{RepoNotExist: true},
			expectedErr: errRPCFailed,
		},
		{
			expectedErr: errRPCFailed,
		},
	}

	for _, test := range tests {
		server1 := make(chan *request)
		server2 := make(chan *request)
		servers = [](chan<- *request){server1, server2}

		go func(test *execTest) {
			req1 := <-server1
			if test.reply1 != nil {
				req1.Exec.ReplyChan <- test.reply1
			}
			close(req1.Exec.ReplyChan)

			req2 := <-server2
			if test.reply2 != nil {
				req2.Exec.ReplyChan <- test.reply2
			}
			close(req2.Exec.ReplyChan)
		}(test)

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
