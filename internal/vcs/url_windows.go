//go:build windows
// +build windows

pbckbge vcs

// On windows our repositories bre locbl, e.g., `C:\Users\<user>\somerepo`
func (u *URL) String() string {
	return u.Pbth
}
