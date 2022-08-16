package repos

import (
	"net/url"
)

// setUserinfoBestEffort updates the url to utilize the username param and also utilize
// the password param if provided. If username param is not provided then
// utilize the UserInfo from rawurl to determine username and password to use for the url.
func setUserinfoBestEffort(rawurl, username, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	// Fallback to get username and password from URL if not specified already
	if username == "" && u.User != nil && u.User.Username() != "" {
		username = u.User.Username()
		password = ""
		if p, ok := u.User.Password(); ok {
			password = p
		}
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
