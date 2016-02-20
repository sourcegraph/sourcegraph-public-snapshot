package path

import (
	"net/url"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// default root directory where the project workspace
// is located. All projects must be cloned to this
// directory or one of its childen.
const DefaultRoot = "/drone/src"

// config represents a simple Yaml config file with the clone
// section and the path attribute. This is used to quickly
// extract only the path.
type config struct {
	Clone struct {
		Path string
	}
}

// Parse parses a yaml file to find the default
// workspace path. If empty, the default uri is
// used to determine the workspace.
func Parse(raw, rawurl string) string {
	data := config{}
	path := FromUrl(rawurl)

	// unarmshal into the temporary Yaml object
	yaml.Unmarshal([]byte(raw), &data)

	// if no clone path is found, return the default
	// workspace path, joined with the root workspace.
	if len(data.Clone.Path) != 0 {
		path = data.Clone.Path
	}

	if filepath.HasPrefix(path, DefaultRoot) {
		return path
	}

	// otherwise return the clone path, joined with the
	// root workspace. Note that this means the clone
	// path must be a relative path.
	return filepath.Join(DefaultRoot, path)
}

// FromUrl returns a workspace path from the url.
func FromUrl(rawurl string) string {
	url_, err := url.Parse(rawurl)
	if err != nil {
		return string(filepath.Separator)
	}

	// uses only the host and port
	// section of the url. The query parameter,
	// fragment and scheme must be ignored
	host := url_.Host
	path := url_.Path

	// remove the colon from the hostname to avoid
	// issues with Docker volumes and caching.
	parts := strings.Split(host, ":")
	if len(parts) == 2 {
		host = parts[0]
	}

	path = filepath.ToSlash(path) // just for windows
	path = filepath.Clean(path)
	return filepath.Join(host, path)
}
