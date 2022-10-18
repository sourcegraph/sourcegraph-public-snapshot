package main

import (
	"context"
	"net/url"
	"os"

	"github.com/sourcegraph/run"
)

type blankRepo struct {
	teardown func()
	path     string
	login    string
	password string
}

func newBlankRepo(login string, password string) (*blankRepo, error) {
	path, err := os.MkdirTemp(os.TempDir(), "sourcegraph-blank-repo")
	if err != nil {
		return nil, err
	}
	return &blankRepo{
		login:    login,
		password: password,
		path:     path,
		teardown: func() {
			_ = os.RemoveAll(path)
		},
	}, nil
}

func (r *blankRepo) inRepo(f func() error) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	err = os.Chdir(r.path)
	if err != nil {
		return err
	}

	return f()
}

func (r *blankRepo) init(ctx context.Context) error {
	return r.inRepo(func() error {
		err := run.Bash(ctx, "git init").Run().Stream(os.Stdout)
		if err != nil {
			return err
		}
		err = os.WriteFile("README.md", []byte("blank repo"), 0755)
		if err != nil {
			return err
		}
		err = run.Bash(ctx, "git add README.md").Run().Stream(os.Stdout)
		if err != nil {
			return err
		}
		err = run.Bash(ctx, "git commit -m \"initial commit\"").Run().Wait()
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *blankRepo) addRemote(ctx context.Context, name string, gitURL string) error {
	return r.inRepo(func() error {
		u, err := url.Parse(gitURL)
		if err != nil {
			return err
		}
		u.User = url.UserPassword(r.login, r.password)
		u.Scheme = "https"
		return run.Bash(ctx, "git remote add", name, u.String()).Run().Wait()
	})
}

func (r *blankRepo) pushRemote(ctx context.Context, name string) error {
	return r.inRepo(func() error {
		return run.Bash(ctx, "git push", name).Run().Wait()
	})
}
