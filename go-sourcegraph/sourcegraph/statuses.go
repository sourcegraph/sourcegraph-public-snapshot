package sourcegraph

// GetStatus returns the status with the given context label. If none
// exists, nil is returned.
func (cs *CombinedStatus) GetStatus(context string) *RepoStatus {
	for _, s := range cs.Statuses {
		if s.Context == context {
			return s
		}
	}
	return nil
}
