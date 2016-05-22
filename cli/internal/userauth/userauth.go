package userauth

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"
)

// UserAuth holds user auth credentials keyed on API endpoint
// URL. It's typically saved in a file named by userAuthFile.
type UserAuth map[string]*userEndpointAuth

func (ua UserAuth) SetDefault(endpoint string) {
	for k, v := range ua {
		if k == endpoint {
			v.Default = true
		} else {
			v.Default = false
		}
	}
}

// GetDefault returns the user-endpoint auth entry that is marked as
// the default, if any exists.
func (ua UserAuth) GetDefault() (endpoint string, a *userEndpointAuth) {
	for k, v := range ua {
		if v.Default {
			return k, v
		}
	}
	return "", nil
}

// Write writes ua to the userAuthFile.
func (ua UserAuth) Write(path string) error {
	f, err := os.Create(userAuthFileName(path))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := os.Chmod(f.Name(), 0600); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ua, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// userEndpointAuth holds a user's authentication credentials for a
// sourcegraph endpoint.
type userEndpointAuth struct {
	AccessToken string

	// Default is whether this endpoint and access token should be
	// used as the defaults if none are specified.
	Default bool `json:",omitempty"`
}

// Read attempts to read a UserAuth struct from the userAuthFile.
// It is not considered an error if the userAuthFile doesn't exist; in that
// case, an empty UserAuth and a nil error is returned.
// Typically path is client.Credentials.Auth
func Read(path string) (UserAuth, error) {
	if path == "/dev/null" {
		return UserAuth{}, nil
	}
	f, err := os.Open(userAuthFileName(path))
	if err != nil {
		if os.IsNotExist(err) {
			return UserAuth{}, nil
		}
		return nil, err
	}
	var ua UserAuth
	if err := json.NewDecoder(f).Decode(&ua); err != nil {
		return nil, err
	}
	return ua, nil
}

// Resolves user auth file name platform-independent way
func userAuthFileName(ret string) string {
	if runtime.GOOS == "windows" {
		// on Windows there is no HOME
		ret = strings.Replace(ret, "$HOME", env.CurrentUserHomeDir(), -1)
	}
	return filepath.FromSlash(os.ExpandEnv(ret))
}

// SaveCredentials is a wrapper around loading up src-auth, adding a
// credential and writing.
func SaveCredentials(path string, endpointURL *url.URL, accessTok string, makeDefault bool) error {
	a, err := Read(path)
	if err != nil {
		return err
	}

	var updatedDefault, updatedCredentials bool
	ua, ok := a[endpointURL.String()]
	if ok {
		if ua.AccessToken != accessTok {
			updatedCredentials = true
			ua.AccessToken = accessTok
		}
	} else {
		updatedCredentials = true
		ua = &userEndpointAuth{AccessToken: accessTok}
		a[endpointURL.String()] = ua
	}
	if makeDefault && !ua.Default {
		updatedDefault = true
		a.SetDefault(endpointURL.String())
	}

	if err := a.Write(path); err != nil {
		return err
	}
	if updatedCredentials {
		log.Printf("# Credentials for %s saved to %s.", endpointURL, userAuthFileName(path))
	}
	if updatedDefault {
		log.Printf("# Default endpoint set to %s.", endpointURL)
	}
	return nil
}
