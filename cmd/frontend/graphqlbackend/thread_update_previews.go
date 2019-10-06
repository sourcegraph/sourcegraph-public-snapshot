package graphqlbackend

type ThreadUpdateOperation string

const (
	ThreadUpdateOperationCreation ThreadUpdateOperation = "CREATION"
	ThreadUpdateOperationUpdate                         = "UPDATE"
	ThreadUpdateOperationDeletion                       = "DELETION"
)

// ThreadUpdatePreview is the interface for the GraphQL type ThreadUpdatePreview.
type ThreadUpdatePreview interface {
	OldThread() Thread
	NewThread() ThreadPreview
	Operation() ThreadUpdateOperation
	OldTitle() *string
	NewTitle() *string
}
