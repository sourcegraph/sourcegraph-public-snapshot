package graphqlbackend

// ReviewState is the GraphQL enum ReviewState.
type ReviewState string

const (
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStateApproved                     = "APPROVED"
	ReviewStateChangesRequested             = "CHANGES_REQUESTED"
)
