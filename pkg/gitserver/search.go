package gitserver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"syscall"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

type SearchArgs struct {
	Repo   string
	Commit vcs.CommitID
	Opt    vcs.SearchOptions
}

type SearchReply struct {
	RepoExists bool
	Results    []*vcs.SearchResult
}

func (g *Git) Search(args *SearchArgs, reply *SearchReply) error {
	dir := path.Join(ReposDir, args.Repo)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	reply.RepoExists = true

	var queryType string
	switch args.Opt.QueryType {
	case vcs.FixedQuery:
		queryType = "--fixed-strings"
	default:
		return fmt.Errorf("unrecognized QueryType: %q", args.Opt.QueryType)
	}

	cmd := exec.Command("git", "grep", "--null", "--line-number", "-I", "--no-color", "--context", strconv.Itoa(int(args.Opt.ContextLines)), queryType, "-e", args.Opt.Query, string(args.Commit))
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer out.Close()
	if err := cmd.Start(); err != nil {
		return err
	}

	errc := make(chan error)
	go func() {
		rd := bufio.NewReader(out)
		var r *vcs.SearchResult
		addResult := func(rr *vcs.SearchResult) bool {
			if rr != nil {
				if args.Opt.Offset == 0 {
					reply.Results = append(reply.Results, rr)
				} else {
					args.Opt.Offset--
				}
				r = nil
			}
			// Return true if no more need to be added.
			return len(reply.Results) == int(args.Opt.N)
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
				file := string(line[len(args.Commit)+1 : fileEnd])
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
		return err
	}

	return nil
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
	done := make(chan *rpc.Call, len(servers))
	for _, server := range servers {
		server <- &rpc.Call{
			ServiceMethod: "Git.Search",
			Args:          &SearchArgs{Repo: repo, Commit: commit, Opt: opt},
			Reply:         &SearchReply{},
			Done:          done,
		}
	}
	var rpcError error
	for range servers {
		call := <-done
		if call.Error != nil {
			rpcError = call.Error
			continue
		}
		reply := call.Reply.(*SearchReply)
		if reply.RepoExists {
			return reply.Results, nil
		}
	}
	if rpcError != nil {
		return nil, rpcError
	}
	return nil, vcs.ErrRepoNotExist
}
