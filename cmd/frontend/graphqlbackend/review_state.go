package graphqlbackend

// ReviewState is the GraphQL enum ReviewState.
type ReviewState string

const (
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
)
