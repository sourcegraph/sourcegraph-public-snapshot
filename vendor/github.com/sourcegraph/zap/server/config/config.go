package config

import (
	"log"
	"os"
	"strings"
	"sync"

	gitconfig "github.com/src-d/go-git/plumbing/format/config"
)

// File is a configuration file.
// Reads and writes are synchronized on a mutex.
type File struct {
	path string
	sync.RWMutex
}

func (f *File) String() string {
	return f.path
}

// NewFile returns a new configuration file.
func NewFile(path string) *File {
	return &File{path: path}
}

// read reads the contents of the file as a Git config.
// The caller is responsible for acquiring a read or write lock on the file.
func (f *File) read() (*gitconfig.Config, error) {
	var gc gitconfig.Config
	file, err := os.Open(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			// It is ok if the file doesn't exist.
			// Return an empty configuration.
			return &gc, nil
		}
		return nil, err
	}
	defer file.Close()
	err = gitconfig.NewDecoder(file).Decode(&gc)
	return &gc, err
}

// write writes the contents of the file as a Git config.
// The caller is responsible for acquiring a write lock on the file.
func (f *File) write(gc *gitconfig.Config) error {
	file, err := os.Create(f.path)
	if err != nil {
		return err
	}
	defer file.Close()
	return gitconfig.NewEncoder(file).Encode(gc)
}

// ServerConfig represents the Zap server configuration file.
type ServerConfig struct {
	AuthToken  string
	Workspaces map[string]struct{}
}

func newServerConfig() *ServerConfig {
	return &ServerConfig{
		Workspaces: map[string]struct{}{},
	}
}

// ReadServerConfig parses the contents of file
// as a Zap server configuration file.
func ReadServerConfig(file *File) (*ServerConfig, error) {
	if file == nil {
		return newServerConfig(), nil
	}
	file.RLock()
	defer file.RUnlock()
	return readServerConfig(file)
}

func readServerConfig(file *File) (*ServerConfig, error) {
	gc, err := file.read()
	if err != nil {
		return nil, err
	}
	config := &ServerConfig{
		AuthToken:  gc.Section("auth").Option("token"),
		Workspaces: map[string]struct{}{},
	}

	// TODO(slimsag): After 5/21/2017 remove this.
	if strings.HasPrefix(config.AuthToken, "sg-session=") {
		log.Fatal("error: auth token is invalid, please run 'zap auth' to fix")
	}

	for _, opt := range gc.Section("workspaces").Options {
		if opt.IsKey("workspace") {
			config.Workspaces[opt.Value] = struct{}{}
		}
	}
	return config, nil
}

// WriteServerConfig writes the edited Zap server
// configuration to the file.
func WriteServerConfig(file *File, edit func(config *ServerConfig)) error {
	// Acquire the write lock before we read/edit/write the file.
	file.Lock()
	defer file.Unlock()
	config, err := readServerConfig(file)
	if err != nil {
		return err
	}
	edit(config)
	var gc gitconfig.Config
	workspaces := gc.Section("workspaces")
	for workspace := range config.Workspaces {
		workspaces.AddOption("workspace", workspace)
	}
	gc.Section("auth").SetOption("token", config.AuthToken)
	return file.write(&gc)
}

// RepoConfig contains Zap configuration for a specific repo.
type RepoConfig struct {
	// The URL of the Zap upstream.
	// (e.g. https://ws.sourcegraph.com/.api/zap)
	UpstreamURL string

	// The name of the repo on the upstream
	// (e.g. github.com/foo/bar)
	UpstreamRepo string
}

// Empty returns true if the configuration is empty.
func (rc *RepoConfig) Empty() bool {
	return rc.UpstreamRepo == "" && rc.UpstreamURL == ""
}

// ReadRepoConfig returns the Zap repo configuration from file.
func ReadRepoConfig(file *File) (*RepoConfig, error) {
	file.RLock()
	defer file.RUnlock()
	_, gc, err := readRepoConfig(file)
	return gc, err
}

func readRepoConfig(file *File) (*gitconfig.Config, *RepoConfig, error) {
	gc, err := file.read()
	if err != nil {
		return nil, nil, err
	}
	remoteOrigin := gc.Section("remote").Subsection("origin")
	return gc, &RepoConfig{
		UpstreamURL:  remoteOrigin.Option("zapurl"),
		UpstreamRepo: remoteOrigin.Option("zaprepo"),
	}, nil
}

// WriteRepoConfig writes the edited Zap repo configuration to file.
func WriteRepoConfig(file *File, edit func(config *RepoConfig)) error {
	// Acquire the write lock before we read/edit/write the file.
	file.Lock()
	defer file.Unlock()
	gc, config, err := readRepoConfig(file)
	if err != nil {
		return err
	}
	edit(config)
	remoteOrigin := gc.Section("remote").Subsection("origin")
	remoteOrigin.SetOption("zapurl", config.UpstreamURL)
	remoteOrigin.SetOption("zaprepo", config.UpstreamRepo)
	return file.write(gc)
}
