package accesslevels

var (
	// For every new grpc endpoint, add an entry to this map to
	// specify its access level. Every endpoint must have an
	// explicit entry to avoid missing the appropriate access
	// check on new endpoints.
	grpcMethodAccessLevel = map[string]string{
		"Accounts.AcceptInvite":         "none",
		"Accounts.Create":               "none",
		"Accounts.Delete":               "admin",
		"Accounts.DeleteInvite":         "admin",
		"Accounts.Invite":               "admin",
		"Accounts.ListInvites":          "admin",
		"Accounts.RequestPasswordReset": "none",
		"Accounts.ResetPassword":        "none",
		"Accounts.Update":               "read",

		"Auth.GetAccessToken":       "none",
		"Auth.GetAuthorizationCode": "read",
		"Auth.GetExternalToken":     "read",
		"Auth.Identify":             "none",
		"Auth.SetExternalToken":     "read",

		"Builds.Create":         "read",
		"Builds.CreateTasks":    "write",
		"Builds.DequeueNext":    "write",
		"Builds.Get":            "read",
		"Builds.GetRepoBuild":   "read",
		"Builds.GetTaskLog":     "read",
		"Builds.List":           "read",
		"Builds.ListBuildTasks": "read",
		"Builds.Update":         "write",
		"Builds.UpdateTask":     "write",

		"Changesets.Create":         "write",
		"Changesets.CreateReview":   "read",
		"Changesets.Get":            "read",
		"Changesets.List":           "read",
		"Changesets.ListEvents":     "read",
		"Changesets.ListReviews":    "read",
		"Changesets.Merge":          "write",
		"Changesets.Update":         "write",
		"Changesets.UpdateAffected": "write",

		"Defs.Get":          "read",
		"Defs.List":         "read",
		"Defs.ListAuthors":  "read",
		"Defs.ListClients":  "read",
		"Defs.ListExamples": "read",
		"Defs.ListRefs":     "read",

		"Deltas.Get":                 "read",
		"Deltas.ListAffectedAuthors": "read",
		"Deltas.ListAffectedClients": "read",
		"Deltas.ListDefs":            "read",
		"Deltas.ListFiles":           "read",
		"Deltas.ListUnits":           "read",

		"GitTransport.InfoRefs":    "read",
		"GitTransport.ReceivePack": "write",
		"GitTransport.UploadPack":  "read",

		"GraphUplink.Push":       "read",
		"GraphUplink.PushEvents": "read",

		"Markdown.Render": "read",

		"Meta.Config": "read",
		"Meta.Status": "read",

		"MirroredRepoSSHKeys.Create": "write",
		"MirroredRepoSSHKeys.Delete": "write",
		"MirroredRepoSSHKeys.Get":    "read",

		"MirrorRepos.AddToWaitlist": "read",
		"MirrorRepos.GetUserData":   "read",
		"MirrorRepos.RefreshVCS":    "read",

		"MultiRepoImporter.Import": "read",
		"MultiRepoImporter.Index":  "read",

		"Notify.GenericEvent": "read",

		"Orgs.Get":         "read",
		"Orgs.List":        "read",
		"Orgs.ListMembers": "read",

		"People.Get": "read",

		"RegisteredClients.Create":              "read",
		"RegisteredClients.Delete":              "read",
		"RegisteredClients.Get":                 "read",
		"RegisteredClients.GetCurrent":          "read",
		"RegisteredClients.GetUserPermissions":  "read",
		"RegisteredClients.List":                "admin",
		"RegisteredClients.ListUserPermissions": "read",
		"RegisteredClients.SetUserPermissions":  "read",
		"RegisteredClients.Update":              "read",

		"RepoBadges.CountHits":    "read",
		"RepoBadges.ListBadges":   "read",
		"RepoBadges.ListCounters": "read",
		"RepoBadges.RecordHit":    "read",

		"Repos.ConfigureApp":                "admin",
		"Repos.Create":                      "write",
		"Repos.Delete":                      "write",
		"Repos.Get":                         "read",
		"Repos.GetCommit":                   "read",
		"Repos.GetConfig":                   "read",
		"Repos.GetInventory":                "read",
		"Repos.GetReadme":                   "read",
		"Repos.GetSrclibDataVersionForPath": "read",
		"Repos.List":                        "read",
		"Repos.ListBranches":                "read",
		"Repos.ListCommits":                 "read",
		"Repos.ListCommitters":              "read",
		"Repos.ListTags":                    "read",
		"Repos.Update":                      "write",

		"RepoStatuses.Create":      "write",
		"RepoStatuses.GetCombined": "read",

		"RepoTree.Get":    "read",
		"RepoTree.List":   "read",
		"RepoTree.Search": "read",

		"Search.SearchText":   "read",
		"Search.SearchTokens": "read",

		"Storage.Delete":         "write",
		"Storage.Exists":         "read",
		"Storage.Get":            "read",
		"Storage.List":           "read",
		"Storage.Put":            "write",
		"Storage.PutNoOverwrite": "write",

		"Units.Get":  "read",
		"Units.List": "read",

		"UserKeys.AddKey":        "read",
		"UserKeys.DeleteAllKeys": "read",
		"UserKeys.DeleteKey":     "read",
		"UserKeys.ListKeys":      "read",
		"UserKeys.LookupUser":    "read",

		"Users.Count":         "read",
		"Users.Get":           "read",
		"Users.GetWithEmail":  "read",
		"Users.List":          "write",
		"Users.ListEmails":    "read",
		"Users.ListTeammates": "read",
	}
)

// GetMethodAccessLevel returns the access level for the given gRPC method.
// The access level can be one of:
//     none: no access checks required
//     read: check user has read access to given method (for given repo)
//     write: check user has read write to given method (for given repo)
//     admin: check user has admin access to given method
//
// If the empty string is returned, the given method has no specified
// access level.
func GetMethodAccessLevel(method string) string {
	if level, ok := grpcMethodAccessLevel[method]; ok {
		return level
	} else {
		return ""
	}
}
