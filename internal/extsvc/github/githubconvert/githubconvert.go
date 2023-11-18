// This package serves as a translation layer between go-github versions.
// The intended purpose is to convert between the go-github/v55 user, and
// the user returned by gologin, which is go-github/v48
package githubconvert

import (
	"bytes"
	"encoding/json"

	//nolint:depguard
	gh48 "github.com/google/go-github/v48/github"
	gh55 "github.com/google/go-github/v55/github"
)

func ConvertUserV48ToV55(userV48 *gh48.User) *gh55.User {
	b, _ := json.Marshal(userV48)
	u55 := gh55.User{}
	_ = json.NewDecoder(bytes.NewReader(b)).Decode(&u55)

	return &u55
}

func ConvertUserV55ToV48(userV55 *gh55.User) *gh48.User {
	b, _ := json.Marshal(userV55)
	u48 := gh48.User{}
	_ = json.NewDecoder(bytes.NewReader(b)).Decode(&u48)

	return &u48
}
