package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/run"
)

type blankRepo struct {
	path     string
	login    string
	password string
	sync.Mutex
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

func (r *blankRepo) clone(ctx context.Context, num int) (*blankRepo, error) {
	folder := fmt.Sprintf("%s_%d", filepath.Base(r.path), num)
	err := run.Bash(ctx, "cp -R", r.path, folder).Run().Wait()
	if err != nil {
		return nil, err
	}
	other := *r
	other.path = filepath.Join(filepath.Dir(r.path), folder)
	return &other, nil
}

func (r *blankRepo) teardown() {
	_ = os.RemoveAll(r.path)
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
		r.Lock()
		defer r.Unlock()
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
		var err error
		for i := 0; i < 3; i++ {
			err = run.Bash(ctx, "git push", name).Run().Wait()
			if err != nil && strings.Contains(err.Error(), "timed out") {
				println("retrying", i)
				continue
			}
			break
		}
		return err
	})
}
