package sgx

import "net/url"

func AddUsernamePasswordToCloneURL(cloneURL, username, password string) (string, error) {
	u, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(username, password)
	return u.String(), nil
}
