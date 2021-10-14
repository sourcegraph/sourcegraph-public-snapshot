package database

// Global reference to database stores using the global dbconn.Global connection handle.
// Deprecated: Use store constructors instead.
var (
	GlobalRepos                 = &repoStore{}
	GlobalUsers                 = &UserStore{}
	GlobalUserEmails            = &UserEmailsStore{}
	GlobalAuthz      AuthzStore = &authzStore{}
)
