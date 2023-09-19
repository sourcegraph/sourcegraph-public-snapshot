package tst

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/run"
)

type LocalRepo struct {
	name     string
	path     string
	login    string
	password string
}

func NewLocalRepo(name, login, password string) (*LocalRepo, error) {
	path, err := os.MkdirTemp(os.TempDir(), name)
	if err != nil {
		return nil, err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &LocalRepo{
		name:     name,
		login:    login,
		password: password,
		path:     path,
	}, nil
}

func (r *LocalRepo) Cleanup() {
	_ = os.RemoveAll(r.path)
}

func (r *LocalRepo) Init(ctx context.Context) error {
	err := run.Bash(ctx, "git init -b main").Dir(r.path).Run().Stream(os.Stdout)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(r.path, "README.md"), []byte("blank repo"), 0755)
	if err != nil {
		return err
	}
	err = run.Bash(ctx, "git add README.md").Dir(r.path).Run().Stream(os.Stdout)
	if err != nil {
		return err
	}
	err = run.Bash(ctx, "git commit -m \"initial commit\"").Dir(r.path).Run().Wait()
	if err != nil {
		return err
	}
	return nil
}

func (r *LocalRepo) AddRemote(ctx context.Context, gitURL string) error {
	u, err := url.Parse(gitURL)
	if err != nil {
		return err
	}
	u.User = url.UserPassword(r.login, r.password)
	u.Scheme = "https"
	return run.Bash(ctx, "git remote add", r.name, u.String()).Dir(r.path).Run().Wait()
}

func (r *LocalRepo) PushRemote(ctx context.Context, retry int) error {
	var err error
	for i := 0; i < retry; i++ {
		err = r.doPushRemote(ctx)
		if err == nil {
			break
		}
	}
	return err
}

func (r *LocalRepo) doPushRemote(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return run.Bash(ctx, "git push", r.name).Dir(r.path).Run().Wait()
}
