package gitserver

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net/rpc"
	"os"
	"os/exec"
	"path"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

type InitArgs struct {
	Repo string
}

func (g *Git) Init(args *InitArgs, reply *struct{}) error {
	cmd := exec.Command("git", "init", "--bare", path.Join(ReposDir, args.Repo))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating repository %s failed with output:\n%s", args.Repo, string(out))
	}
	return nil
}

func Init(repo string) error {
	cmd := Command("git", "remote")
	cmd.Repo = repo
	err := cmd.Run()
	if err == nil {
		return errors.New("repository already exists")
	}
	if err != vcs.ErrRepoNotExist {
		return err
	}

	// this hash is used to avoid concurrent init on two servers, it does not need to be stable over long timespans
	h := fnv.New32a()
	if _, err := h.Write([]byte(repo)); err != nil {
		return err
	}
	serverIndex := int(h.Sum32()) % len(servers)

	done := make(chan *rpc.Call, 1)
	servers[serverIndex] <- &rpc.Call{
		ServiceMethod: "Git.Init",
		Args:          &InitArgs{Repo: repo},
		Reply:         &struct{}{},
		Done:          done,
	}
	return (<-done).Error
}

type RemoveArgs struct {
	Repo string
}

type RemoveReply struct {
	RepoExists bool
}

func (r *RemoveReply) repoExists() bool {
	return r.RepoExists
}

func (g *Git) Remove(args *RemoveArgs, reply *RemoveReply) error {
	dir := path.Join(ReposDir, args.Repo)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	reply.RepoExists = true

	cmd := exec.Command("git", "remote")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a repository: %s", args.Repo)
	}

	return os.RemoveAll(dir)
}

func Remove(repo string) error {
	_, err := broadcastCall(
		"Git.Remove",
		&RemoveArgs{Repo: repo},
		func() repoExistsReply { return new(RemoveReply) },
	)
	return err
}
