package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"gopkg.in/gcfg.v1"
)

// repoGitConfig represents a git config file in a repository (e.g.,
// ".git/config", or simply "config" in a bare repo).
type repoGitConfig struct {
	Remote map[string]*struct {
		URL    string
		Mirror bool
	}

	Sourcegraph struct {
		Description string
		Language    string
		Private     bool
		CreatedAt   string
		UpdatedAt   string
		PushedAt    string
	}
}

func (s *Repos) setGitConfig(ctx context.Context, dir, name, value string) error {
	if err := checkGitArg(name); err != nil {
		return err
	}
	if err := checkGitArg(value); err != nil {
		return err
	}

	var (
		err error
		out []byte
	)
	for attempt := 0; attempt < 3; attempt++ {
		cmd := exec.Command("git", "config", name, value)
		cmd.Dir = dir
		out, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("set git config %q %q failed with %s\n%s", name, value, err, out)
		}
	}
	return err
}

func (s *Repos) getGitConfig(ctx context.Context, fs vfs.FileSystem, dir string) (*repoGitConfig, error) {
	// TODO(sqs): Eliminate repetition: this next section of code is
	// copied several times in this file.
	var configPath string
	if _, err := fs.Stat(filepath.Join(dir, ".git")); err == nil {
		configPath = filepath.Join(dir, ".git", "config") // non-bare repo
	} else if os.IsNotExist(err) {
		configPath = filepath.Join(dir, "config") // bare repo
	} else {
		return nil, err
	}

	data, err := vfs.ReadFile(fs, configPath)
	if err != nil {
		return nil, err
	}

	var conf repoGitConfig
	if err := gcfg.ReadStringInto(&conf, string(data)); err != nil {
		return nil, err
	}
	return &conf, nil
}
