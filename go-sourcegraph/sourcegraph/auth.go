package sourcegraph

// UserSpec returns a UserSpec that refers to the user identified by
// a. If a.UID == 0, nil is returned.
func (a AuthInfo) UserSpec() *UserSpec {
	if a.UID == 0 {
		return nil
	}
	return &UserSpec{UID: a.UID, Login: a.Login}
}
