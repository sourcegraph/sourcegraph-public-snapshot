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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type createRequest struct {
	Repo         string
	MirrorRemote string
	Opt          *vcs.RemoteOpts
	ReplyChan    chan<- *createReply
}

type createReply struct {
	RepoExist       bool   // If true, create returned with noop because repo exists.
	CloneInProgress bool   // If true, create returned with noop because clone is in progress.
	Error           string // If non-empty, an error happened.
}

var (
	// cloning tracks repositories (key is '/'-separated path) that are
	// in the process of being cloned.
	cloningMu sync.Mutex
	cloning   = make(map[string]struct{})
)

func handleCreateRequest(req *createRequest) {
	start := time.Now()
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { defer observeCreate(start, status) }()

	dir := path.Join(ReposDir, req.Repo)
	cloningMu.Lock()
	if _, ok := cloning[dir]; ok {
		cloningMu.Unlock()
		req.ReplyChan <- &createReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if repoExists(dir) {
		cloningMu.Unlock()
		req.ReplyChan <- &createReply{RepoExist: true}
		status = "repo-exists"
		return
	}

	// We'll take this repo and start cloning it.
	// Mark it as being cloned so no one else starts to.
	cloning[dir] = struct{}{}
	cloningMu.Unlock()

	defer func() {
		cloningMu.Lock()
		delete(cloning, dir)
		cloningMu.Unlock()
	}()

	if req.MirrorRemote != "" {
		cmd := exec.Command("git", "clone", "--mirror", req.MirrorRemote, dir)

		var outputBuf bytes.Buffer
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
		if err := runWithRemoteOpts(cmd, req.Opt); err != nil {
			req.ReplyChan <- &createReply{Error: fmt.Sprintf("cloning repository %s failed with output:\n%s", req.Repo, outputBuf.String())}
			status = "clone-fail"
			return
		}
		req.ReplyChan <- &createReply{}
		status = "clone-success"
		return
	}

	cmd := exec.Command("git", "init", "--bare", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		req.ReplyChan <- &createReply{Error: fmt.Sprintf("initializing repository %s failed with output:\n%s", req.Repo, string(out))}
		status = "init-fail"
		return
	}
	status = "init-success"
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
//
// A nil error is returned if the new repository was created successfully, but vcs.ErrRepoExist
// is returned if there was already an existing repository in its place, causing create to be noop.
// If the repository is in process of being cloned, vcs.RepoNotExistError{CloneInProgress: true} is returned.
func create(repo string, mirrorRemote string, opt *vcs.RemoteOpts) error {
	// We check if repo already exists by executing `git remote`. It may seem redundant since the
	// create request also checks that, but the purpose is to first do a broadcast and check if _any_
	// server already has the repo available.
	cmd := Command("git", "remote")
	cmd.Repo = repo
	err := cmd.Run()
	if err == nil {
		return vcs.ErrRepoExist
	}
	if !vcs.IsRepoNotExist(err) {
		// The only acceptable error is repo doesn't exist, if it's something else, there's a problem. Return the error.
		return err
	}
	if repoNotExistError := err.(vcs.RepoNotExistError); repoNotExistError.CloneInProgress {
		// If some server is already cloning this repository, report it and don't try to create another.
		return repoNotExistError
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
	if reply.CloneInProgress {
		return vcs.RepoNotExistError{CloneInProgress: true}
	}
	if reply.RepoExist {
		return vcs.ErrRepoExist
	}
	// Repo did not exist and was successfully created.
	return nil
}

type removeRequest struct {
	Repo      string
	ReplyChan chan<- *removeReply
}

type removeReply struct {
	RepoNotFound    bool   // If true, remove returned with noop because repo is not found.
	CloneInProgress bool   // If true, remove returned with noop because clone is in progress.
	Error           string // If non-empty, an error happened.
}

func (r *removeReply) repoFound() bool { return !r.RepoNotFound }

func handleRemoveRequest(req *removeRequest) {
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { defer observeRemove(status) }()

	dir := path.Join(ReposDir, req.Repo)
	cloningMu.Lock()
	_, cloneInProgress := cloning[dir]
	cloningMu.Unlock()
	if cloneInProgress {
		req.ReplyChan <- &removeReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		req.ReplyChan <- &removeReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}

	cmd := exec.Command("git", "remote")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		req.ReplyChan <- &removeReply{Error: fmt.Sprintf("not a repository: %s", req.Repo)}
		status = "not-a-repository"
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		req.ReplyChan <- &removeReply{Error: err.Error()}
		status = "failed"
		return
	}
	req.ReplyChan <- &removeReply{}
	status = "success"
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
	if reply.CloneInProgress {
		return vcs.RepoNotExistError{CloneInProgress: true}
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

// remove should be pretty much instant, so we just track counts
var removeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "remove_total",
	Help:      "Total calls to gitserver.Remove",
}, []string{"status"})
var createDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "create_duration_seconds",
	Help:      "gitserver.Init and gitserver.Clone latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"status"})

func init() {
	prometheus.MustRegister(removeCounter)
	prometheus.MustRegister(createDuration)
}

func observeRemove(status string) {
	removeCounter.WithLabelValues(status).Inc()
}

func observeCreate(start time.Time, status string) {
	createDuration.WithLabelValues(status).Observe(time.Since(start).Seconds())
}
