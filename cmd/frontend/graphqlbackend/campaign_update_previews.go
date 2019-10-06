package graphqlbackend

import "context"

// CampaignUpdatePreview is the interface for the GraphQL type CampaignUpdatePreview.
type CampaignUpdatePreview interface {
	OldName() *string
	NewName() *string
	OldStartDate() *DateTime
	NewStartDate() *DateTime
	OldDueDate() *DateTime
	NewDueDate() *DateTime
	Threads(context.Context) (*[]ThreadUpdatePreview, error)
	RepositoryComparisons(context.Context) (*[]*RepositoryComparisonUpdatePreview, error)
}

// RepositoryComparisonUpdatePreview implements the RepositoryComparisonUpdatePreview GraphQL interface.
type RepositoryComparisonUpdatePreview struct {
	Repository_ *RepositoryResolver
	Old_        RepositoryComparison
	New_        RepositoryComparison
}

func (v *RepositoryComparisonUpdatePreview) Repository() *RepositoryResolver { return v.Repository_ }
func (v *RepositoryComparisonUpdatePreview) Old() RepositoryComparison       { return v.Old_ }
func (v *RepositoryComparisonUpdatePreview) New() RepositoryComparison       { return v.New_ }
