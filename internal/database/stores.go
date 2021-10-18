package database

// Global reference to database stores using the global dbconn.Global connection handle.
// Deprecated: Use store constructors instead.
var (
	GlobalRepos = &RepoStore{}
	GlobalUsers = &UserStore{}
)
