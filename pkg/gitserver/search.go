package gitserver

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type searchRequest struct {
	Repo      string
	Commit    vcs.CommitID
	Opt       vcs.SearchOptions
	ReplyChan chan<- *searchReply
}

type searchReply struct {
	RepoNotFound    bool // If true, search returned with noop because repo is not found.
	CloneInProgress bool // If true, search returned with noop because clone is in progress.
	Results         []*vcs.SearchResult
	Error           string // If non-empty, an error happened.
}

func (r *searchReply) repoFound() bool { return !r.RepoNotFound }

func handleSearchRequest(req *searchRequest) {
	start := time.Now()
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { defer observeSearch(req, start, status) }()

	dir := path.Join(ReposDir, req.Repo)
	cloningMu.Lock()
	_, cloneInProgress := cloning[dir]
	cloningMu.Unlock()
	if cloneInProgress {
		req.ReplyChan <- &searchReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		req.ReplyChan <- &searchReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}

	var queryType string
	switch req.Opt.QueryType {
	case vcs.FixedQuery:
		queryType = "--fixed-strings"
	default:
		req.ReplyChan <- &searchReply{Error: fmt.Sprintf("unrecognized QueryType: %q", req.Opt.QueryType)}
		status = "error"
		return
	}

	cmd := exec.Command("git", "grep", "--null", "--line-number", "-I", "--no-color", "--context", strconv.Itoa(int(req.Opt.ContextLines)), queryType, "-e", req.Opt.Query, string(req.Commit))
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		req.ReplyChan <- &searchReply{Error: err.Error()}
		status = "error"
		return
	}
	defer out.Close()
	if err := cmd.Start(); err != nil {
		req.ReplyChan <- &searchReply{Error: err.Error()}
		status = "error"
		return
	}

	var results []*vcs.SearchResult
	errc := make(chan error)
	go func() {
		rd := bufio.NewReader(out)
		var r *vcs.SearchResult
		addResult := func(rr *vcs.SearchResult) bool {
			if rr != nil {
				if req.Opt.Offset == 0 {
					results = append(results, rr)
				} else {
					req.Opt.Offset--
				}
				r = nil
			}
			// Return true if no more need to be added.
			return len(results) == int(req.Opt.N)
		}
		for {
			line, err := rd.ReadBytes('\n')
			if err == io.EOF {
				// git-grep output ends with a newline, so if we hit EOF, there's nothing left to
				// read
				break
			} else if err != nil {
				errc <- err
				return
			}
			// line is guaranteed to be '\n' terminated according to the contract of ReadBytes
			line = line[0 : len(line)-1]

			if bytes.Equal(line, []byte("--")) {
				// Match separator.
				if addResult(r) {
					break
				}
			} else {
				// Match line looks like: "HEAD:filename\x00lineno\x00matchline\n".
				fileEnd := bytes.Index(line, []byte{'\x00'})
				file := string(line[len(req.Commit)+1 : fileEnd])
				lineNoStart, lineNoEnd := fileEnd+1, fileEnd+1+bytes.Index(line[fileEnd+1:], []byte{'\x00'})
				lineNo, err := strconv.Atoi(string(line[lineNoStart:lineNoEnd]))
				if err != nil {
					panic("bad line number on line: " + string(line) + ": " + err.Error())
				}
				if r == nil || r.File != file {
					if r != nil {
						if addResult(r) {
							break
						}
					}
					r = &vcs.SearchResult{File: file, StartLine: uint32(lineNo)}
				}
				r.EndLine = uint32(lineNo)
				if r.Match != nil {
					r.Match = append(r.Match, '\n')
				}
				r.Match = append(r.Match, line[lineNoEnd+1:]...)
			}
		}
		addResult(r)

		if err := cmd.Process.Kill(); err != nil {
			if runtime.GOOS != "windows" {
				errc <- err
				return
			}
		}
		if err := cmd.Wait(); err != nil {
			if c := exitStatus(err); c != -1 && c != 1 {
				// -1 exit code = killed (by cmd.Process.Kill() call
				// above), 1 exit code means grep had no match (but we
				// don't translate that to a Go error)
				errc <- fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
				return
			}
		}
		errc <- nil
	}()

	err = <-errc
	cmd.Process.Kill()
	if err != nil {
		req.ReplyChan <- &searchReply{Error: err.Error()}
		status = "error"
		return
	}

	req.ReplyChan <- &searchReply{
		Results: results,
	}
	status = "success"
}

func exitStatus(err error) int {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		return 0
	}
	return 0
}

func Search(repo string, commit vcs.CommitID, opt vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	genReply, err := broadcastCall(func() (*request, func() (genericReply, bool)) {
		replyChan := make(chan *searchReply, 1)
		return &request{Search: &searchRequest{Repo: repo, Commit: commit, Opt: opt, ReplyChan: replyChan}},
			func() (genericReply, bool) { reply, ok := <-replyChan; return reply, ok }
	})
	if err != nil {
		return nil, err
	}

	reply := genReply.(*searchReply)
	if reply.CloneInProgress {
		return nil, vcs.RepoNotExistError{CloneInProgress: true}
	}
	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}
	return reply.Results, nil
}

var searchDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "search_duration_seconds",
	Help:      "gitserver.Search latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"query_type", "repo", "status"})

func init() {
	prometheus.MustRegister(searchDuration)
}

func observeSearch(req *searchRequest, start time.Time, status string) {
	repo := repotrackutil.GetTrackedRepo(req.Repo)
	searchDuration.WithLabelValues(req.Opt.QueryType, repo, status).Observe(time.Since(start).Seconds())
}
