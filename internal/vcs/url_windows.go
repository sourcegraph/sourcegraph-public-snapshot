//go:build windows
// +build windows

package vcs

// On windows our repositories are local, e.g., `C:\Users\<user>\somerepo`
func (u *URL) String() string {
	return u.Path
}
