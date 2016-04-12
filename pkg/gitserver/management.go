package gitserver

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mutexmap"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type createRequest struct {
	Repo         string
	MirrorRemote string
	Opt          *vcs.RemoteOpts
	ReplyChan    chan<- *createReply
}

type createReply struct {
	RepoExist bool
	Error     string
}

var createMu = mutexmap.New()

func handleCreateRequest(req *createRequest) {
	defer recoverAndLog()
	defer close(req.ReplyChan)

	dir := path.Join(ReposDir, req.Repo)
	createMu.Lock(dir)
	defer createMu.Unlock(dir)
	if repoExists(dir) {
		req.ReplyChan <- &createReply{RepoExist: true}
		return
	}

	if req.MirrorRemote != "" {
		cmd := exec.Command("git", "clone", "--mirror", req.MirrorRemote, dir)

		var outputBuf bytes.Buffer
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
		if err := runWithRemoteOpts(cmd, req.Opt); err != nil {
			req.ReplyChan <- &createReply{Error: fmt.Sprintf("cloning repository %s failed with output:\n%s", req.Repo, outputBuf.String())}
			return
		}
		req.ReplyChan <- &createReply{}
		return
	}

	cmd := exec.Command("git", "init", "--bare", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		req.ReplyChan <- &createReply{Error: fmt.Sprintf("initializing repository %s failed with output:\n%s", req.Repo, string(out))}
		return
	}
	req.ReplyChan <- &createReply{}
}

func Init(repo string) error {
	return create(repo, "", nil)
}

func Clone(repo string, remote string, opt *vcs.RemoteOpts) error {
	if remote == "" {
		return errors.New("empty remote")
	}
	return create(repo, remote, opt)
}

// create creates a new repository in the gitserver cluster by initializing an empty repository
// if mirrorRemote is empty or clones the given remote otherwise, using opt for authentication.
// The gitserver is selected pseudo-randomly.
func create(repo string, mirrorRemote string, opt *vcs.RemoteOpts) error {
	cmd := Command("git", "remote")
	cmd.Repo = repo
	err := cmd.Run()
	if err == nil {
		return vcs.ErrRepoExist
	}
	if err != vcs.ErrRepoNotExist {
		return err
	}

	// This hash is used to avoid concurrent init on two servers, it does not need to be stable over long timespans.
	h := fnv.New32a()
	if _, err := h.Write([]byte(repo)); err != nil {
		return err
	}

	sum := md5.Sum([]byte(repo))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(servers))

	replyChan := make(chan *createReply, 1)
	servers[serverIndex] <- &request{Create: &createRequest{
		Repo:         repo,
		MirrorRemote: mirrorRemote,
		Opt:          opt,
		ReplyChan:    replyChan,
	}}

	reply, ok := <-replyChan
	if !ok {
		return errors.New("gitserver: create failed")
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	if reply.RepoExist {
		return vcs.ErrRepoExist
	}
	return nil
}

type removeRequest struct {
	Repo      string
	ReplyChan chan<- *removeReply
}

type removeReply struct {
	RepoNotExist bool
	Error        string
}

func (r *removeReply) repoNotExist() bool {
	return r.RepoNotExist
}

func handleRemoveRequest(req *removeRequest) {
	defer recoverAndLog()
	defer close(req.ReplyChan)

	dir := path.Join(ReposDir, req.Repo)
	if !repoExists(dir) {
		req.ReplyChan <- &removeReply{RepoNotExist: true}
		return
	}

	cmd := exec.Command("git", "remote")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		req.ReplyChan <- &removeReply{Error: fmt.Sprintf("not a repository: %s", req.Repo)}
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		req.ReplyChan <- &removeReply{Error: err.Error()}
		return
	}
	req.ReplyChan <- &removeReply{}
}

func Remove(repo string) error {
	genReply, err := broadcastCall(func() (*request, func() (genericReply, bool)) {
		replyChan := make(chan *removeReply, 1)
		return &request{Remove: &removeRequest{Repo: repo, ReplyChan: replyChan}},
			func() (genericReply, bool) { reply, ok := <-replyChan; return reply, ok }
	})
	if err != nil {
		return err
	}

	reply := genReply.(*removeReply)
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}
