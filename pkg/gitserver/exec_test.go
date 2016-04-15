package gitserver

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	"github.com/neelance/chanrpc/chanrpcutil"

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
			reply1:      &execReply{RepoNotFound: true},
			reply2:      &execReply{RepoNotFound: true},
			expectedErr: vcs.RepoNotExistError{},
		},
		{
			reply1:         &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			reply2:         &execReply{RepoNotFound: true},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:         &execReply{RepoNotFound: true},
			reply2:         &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:         &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply2:         &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedStdout: []byte("out"),
			expectedStderr: []byte("err"),
		},
		{
			reply1:      &execReply{RepoNotFound: true},
			expectedErr: errRPCFailed,
		},
		{
			reply2:      &execReply{RepoNotFound: true},
			expectedErr: errRPCFailed,
		},
		{
			expectedErr: errRPCFailed,
		},

		{
			reply1:      &execReply{CloneInProgress: true},
			reply2:      &execReply{RepoNotFound: true},
			expectedErr: vcs.RepoNotExistError{CloneInProgress: true},
		},
		{
			reply1:      &execReply{RepoNotFound: true},
			reply2:      &execReply{CloneInProgress: true},
			expectedErr: vcs.RepoNotExistError{CloneInProgress: true},
		},
	}

	for _, test := range tests {
		server1 := make(chan *request)
		server2 := make(chan *request)
		servers = [](chan<- *request){server1, server2}

		go func(test *execTest) {
			req1 := <-server1
			chanrpcutil.Drain(req1.Exec.Stdin)
			if test.reply1 != nil {
				req1.Exec.ReplyChan <- test.reply1
			}
			close(req1.Exec.ReplyChan)

			req2 := <-server2
			chanrpcutil.Drain(req2.Exec.Stdin)
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

func emptyProcessResult() <-chan *processResult {
	processResultChan := make(chan *processResult, 1)
	processResultChan <- &processResult{}
	return processResultChan
}

// Test creating repository, ensure clone in progress is reported correctly.
func TestCreate(t *testing.T) {
	type createTest struct {
		gitRemote1  *execReply   // Reply to Command("git", "remote") from server 1.
		gitRemote2  *execReply   // Reply to Command("git", "remote") from server 2.
		reply       *createReply // Reply to Create from server selected using hash of repo name.
		expectedErr error
	}

	tests := []*createTest{
		{
			gitRemote1:  &execReply{RepoNotFound: true},
			gitRemote2:  &execReply{Stdout: chanrpcutil.ToChunks([]byte("out")), Stderr: chanrpcutil.ToChunks([]byte("err")), ProcessResult: emptyProcessResult()},
			expectedErr: vcs.ErrRepoExist,
		},
		{
			gitRemote1:  &execReply{RepoNotFound: true},
			gitRemote2:  &execReply{CloneInProgress: true},
			expectedErr: vcs.RepoNotExistError{CloneInProgress: true},
		},
		{
			gitRemote1:  &execReply{RepoNotFound: true},
			gitRemote2:  &execReply{RepoNotFound: true},
			expectedErr: errors.New("gitserver: create failed"),
		},
		{
			gitRemote1:  &execReply{RepoNotFound: true},
			gitRemote2:  &execReply{RepoNotFound: true},
			reply:       &createReply{Error: "something specific went wrong"},
			expectedErr: errors.New("something specific went wrong"),
		},
		{
			gitRemote1:  &execReply{RepoNotFound: true},
			gitRemote2:  &execReply{RepoNotFound: true},
			reply:       &createReply{CloneInProgress: true},
			expectedErr: vcs.RepoNotExistError{CloneInProgress: true},
		},
		{
			gitRemote1:  &execReply{RepoNotFound: true},
			gitRemote2:  &execReply{RepoNotFound: true},
			reply:       &createReply{RepoExist: true},
			expectedErr: vcs.ErrRepoExist,
		},
		{
			gitRemote1: &execReply{RepoNotFound: true},
			gitRemote2: &execReply{RepoNotFound: true},
			reply:      &createReply{},
		},
	}

	for _, test := range tests {
		testServers := [](chan *request){
			make(chan *request),
			make(chan *request),
		}
		servers = [](chan<- *request){testServers[0], testServers[1]}

		const repo = "test/repo"

		// Keep in sync with hashing algorithm in create.
		sum := md5.Sum([]byte(repo))
		serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(servers))

		go func(test *createTest) {
			req1 := <-testServers[0]
			chanrpcutil.Drain(req1.Exec.Stdin)
			if test.gitRemote2 != nil {
				req1.Exec.ReplyChan <- test.gitRemote1
			}
			close(req1.Exec.ReplyChan)

			req2 := <-testServers[1]
			chanrpcutil.Drain(req2.Exec.Stdin)
			if test.gitRemote2 != nil {
				req2.Exec.ReplyChan <- test.gitRemote2
			}
			close(req2.Exec.ReplyChan)

			req3 := <-testServers[serverIndex] // create will pick this server.
			if test.reply != nil {
				req3.Create.ReplyChan <- test.reply
			}
			close(req3.Create.ReplyChan)
		}(test)

		err := Clone(repo, "test/remote", nil)
		if !reflect.DeepEqual(test.expectedErr, err) {
			t.Errorf("expected error %#v, got %#v", test.expectedErr, err)
		}
	}
}
