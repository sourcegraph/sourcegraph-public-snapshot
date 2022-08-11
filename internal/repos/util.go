package repos

import (
	"net/url"
)

// setUserinfoBestEffort adds the username and password to rawurl. If anything
// fails, the original rawurl is returned.
func setUserinfoBestEffort(rawurl, username, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	if username == "" {
		return rawurl
	}

	if password != "" {
		u.User = url.UserPassword(username, password)
	} else {
		u.User = url.User(username)
	}

	return u.String()
}
