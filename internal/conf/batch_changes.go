package conf

// RejectUnverifiedCommit returns a boolean indicating if unverified commits in changesets
// created by a Batch Change should result in an error.
func RejectUnverifiedCommit() bool {
	cfg := Get().SiteConfig().BatchChangesRejectUnverifiedCommit
	if cfg == nil {
		return false
	}
	return *cfg
}
