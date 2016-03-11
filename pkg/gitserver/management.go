package gitserver

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
	return call("Git.Init", &InitArgs{Repo: repo}, &struct{}{})
}

type RemoveArgs struct {
	Repo string
}

func (g *Git) Remove(args *RemoveArgs, reply *struct{}) error {
	dir := path.Join(ReposDir, args.Repo)

	cmd := exec.Command("git", "remote")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a repository: %s", args.Repo)
	}

	return os.RemoveAll(dir)
}

func Remove(repo string) error {
	return call("Git.Remove", &RemoveArgs{Repo: repo}, &struct{}{})
}
