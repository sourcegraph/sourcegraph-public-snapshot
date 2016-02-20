package papertrail

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

var ErrNoTokenFound = errors.New("no Papertrail API token found in PAPERTRAIL_API_TOKEN or ~/.papertrail.yml")

// ReadToken attempts to read the Papertrail API token from 2 sources, the
// PAPERTRAIL_API_TOKEN environment variable and the ~/.papertrail.yml file, in that
// order. It returns the token if found in either source, and otherwise returns
// ErrNoTokenFound. If an unexpected error occurs (e.g., while reading the
// config file), it returns that error.
func ReadToken() (string, error) {
	t := os.Getenv("PAPERTRAIL_API_TOKEN")
	if t != "" {
		return t, nil
	}

	t, err := readTokenFromConfig()
	if err != nil {
		return "", err
	}
	if t != "" {
		return t, nil
	}

	return "", ErrNoTokenFound
}

func readTokenFromConfig() (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".papertrail.yml"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("token: ")) {
			return string(line[len("token: "):]), nil
		}
	}
	return "", nil
}
