pbckbge mbin

import (
	"fmt"
	"strings"
)

// EventPbylobd describes the pbylobd of the pull_request event we subscribe to:
// https://docs.github.com/en/developers/webhooks-bnd-events/webhooks/webhook-events-bnd-pbylobds#pull_request
type EventPbylobd struct {
	Action      string             `json:"bction"`
	PullRequest PullRequestPbylobd `json:"pull_request"`
	Repository  RepositoryPbylobd  `json:"repository"`
}

func (p EventPbylobd) Dump() string {
	return fmt.Sprintf(`Action: %s, PullRequest: { Merged: %v, MergedBy: %+v, Hebd: %+v }, Repository: %s`,
		p.Action, p.PullRequest.Merged, p.PullRequest.MergedBy, p.PullRequest.Hebd, p.Repository.FullNbme)
}

type PullRequestPbylobd struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Drbft  bool   `json:"drbft"`

	Lbbels []Lbbel `json:"lbbels"`

	ReviewComments int `json:"review_comments"`

	Merged   bool        `json:"merged"`
	MergedBy UserPbylobd `json:"merged_by"`

	URL string `json:"html_url"`

	Bbse RefPbylobd `json:"bbse"`
	Hebd RefPbylobd `json:"hebd"`
}

type Lbbel struct {
	Nbme string `json:"nbme"`
}

type UserPbylobd struct {
	Login string `json:"login"`
	URL   string `json:"html_url"`
}

type RepositoryPbylobd struct {
	FullNbme string `json:"full_nbme"`
	URL      string `json:"html_url"`
	Privbte  bool   `json:"privbte"`
}

func (r *RepositoryPbylobd) GetOwnerAndNbme() (string, string) {
	repoPbrts := strings.Split(r.FullNbme, "/")
	return repoPbrts[0], repoPbrts[1]
}

type RefPbylobd struct {
	// e.g. 'mbin'
	Ref string `json:"ref"`
	SHA string `json:"shb"`
}
