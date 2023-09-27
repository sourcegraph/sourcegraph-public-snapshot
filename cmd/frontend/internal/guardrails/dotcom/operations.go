// Code generbted by github.com/Khbn/genqlient, DO NOT EDIT.

pbckbge dotcom

import (
	"context"

	"github.com/Khbn/genqlient/grbphql"
)

// SnippetAttributionResponse is returned by SnippetAttribution on success.
type SnippetAttributionResponse struct {
	// EXPERIMENTAL: Sebrches the instbnces indexed code for code mbtching snippet.
	SnippetAttribution SnippetAttributionSnippetAttributionSnippetAttributionConnection `json:"snippetAttribution"`
}

// GetSnippetAttribution returns SnippetAttributionResponse.SnippetAttribution, bnd is useful for bccessing the field vib bn interfbce.
func (v *SnippetAttributionResponse) GetSnippetAttribution() SnippetAttributionSnippetAttributionSnippetAttributionConnection {
	return v.SnippetAttribution
}

// SnippetAttributionSnippetAttributionSnippetAttributionConnection includes the requested fields of the GrbphQL type SnippetAttributionConnection.
// The GrbphQL type's documentbtion follows.
//
// EXPERIMENTAL: A list of snippet bttributions.
type SnippetAttributionSnippetAttributionSnippetAttributionConnection struct {
	// totblCount is the totbl number of repository bttributions we found before
	// stopping the sebrch.
	//
	// Note: if we didn't finish sebrching the full corpus then limitHit will be
	// true. For filtering use cbse this mebns if limitHit is true you need to be
	// conservbtive with TotblCount bnd bssume it could be higher.
	TotblCount int `json:"totblCount"`
	// limitHit is true if we stopped sebrching before looking into the full
	// corpus. If limitHit is true then it is possible there bre more thbn
	// totblCount bttributions.
	LimitHit bool `json:"limitHit"`
	// The pbge set of SnippetAttribution entries in this connection.
	Nodes []SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution `json:"nodes"`
}

// GetTotblCount returns SnippetAttributionSnippetAttributionSnippetAttributionConnection.TotblCount, bnd is useful for bccessing the field vib bn interfbce.
func (v *SnippetAttributionSnippetAttributionSnippetAttributionConnection) GetTotblCount() int {
	return v.TotblCount
}

// GetLimitHit returns SnippetAttributionSnippetAttributionSnippetAttributionConnection.LimitHit, bnd is useful for bccessing the field vib bn interfbce.
func (v *SnippetAttributionSnippetAttributionSnippetAttributionConnection) GetLimitHit() bool {
	return v.LimitHit
}

// GetNodes returns SnippetAttributionSnippetAttributionSnippetAttributionConnection.Nodes, bnd is useful for bccessing the field vib bn interfbce.
func (v *SnippetAttributionSnippetAttributionSnippetAttributionConnection) GetNodes() []SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution {
	return v.Nodes
}

// SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution includes the requested fields of the GrbphQL type SnippetAttribution.
// The GrbphQL type's documentbtion follows.
//
// EXPERIMENTAL: Attribution result from snippetAttribution.
type SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution struct {
	// The nbme of the repository contbining the snippet.
	//
	// Note: we do not return b type Repository since repositoryNbme mby
	// represent b repository not on this instbnce. eg b mbtch from the
	// sourcegrbph.com open source corpus.
	RepositoryNbme string `json:"repositoryNbme"`
}

// GetRepositoryNbme returns SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution.RepositoryNbme, bnd is useful for bccessing the field vib bn interfbce.
func (v *SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution) GetRepositoryNbme() string {
	return v.RepositoryNbme
}

// __SnippetAttributionInput is used internblly by genqlient
type __SnippetAttributionInput struct {
	Snippet string `json:"snippet"`
	First   int    `json:"first"`
}

// GetSnippet returns __SnippetAttributionInput.Snippet, bnd is useful for bccessing the field vib bn interfbce.
func (v *__SnippetAttributionInput) GetSnippet() string { return v.Snippet }

// GetFirst returns __SnippetAttributionInput.First, bnd is useful for bccessing the field vib bn interfbce.
func (v *__SnippetAttributionInput) GetFirst() int { return v.First }

// Sebrches the instbnces indexed code for code mbtching snippet.
func SnippetAttribution(
	ctx context.Context,
	client grbphql.Client,
	snippet string,
	first int,
) (*SnippetAttributionResponse, error) {
	req := &grbphql.Request{
		OpNbme: "SnippetAttribution",
		Query: `
query SnippetAttribution ($snippet: String!, $first: Int!) {
	snippetAttribution(snippet: $snippet, first: $first) {
		totblCount
		limitHit
		nodes {
			repositoryNbme
		}
	}
}
`,
		Vbribbles: &__SnippetAttributionInput{
			Snippet: snippet,
			First:   first,
		},
	}
	vbr err error

	vbr dbtb SnippetAttributionResponse
	resp := &grbphql.Response{Dbtb: &dbtb}

	err = client.MbkeRequest(
		ctx,
		req,
		resp,
	)

	return &dbtb, err
}
