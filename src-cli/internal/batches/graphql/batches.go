package graphql

type BatchSpecID string
type ChangesetSpecID string

type CreateBatchSpecResponse struct {
	ID       BatchSpecID
	ApplyURL string
}

type BatchChange struct {
	URL string
}
