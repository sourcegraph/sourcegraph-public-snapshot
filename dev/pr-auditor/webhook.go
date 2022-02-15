package main

import "strings"

// EventPayload describes the payload of the pull_request event we subscribe to:
// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
type EventPayload struct {
	Action      string             `json:"action"`
	PullRequest PullRequestPayload `json:"pull_request"`
	Repository  RepositoryPayload  `json:"repository"`
}

type PullRequestPayload struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`

	ReviewComments int `json:"review_comments"`

	Merged   bool        `json:"merged"`
	MergedBy UserPayload `json:"merged_by"`

	URL string `json:"html_url"`

	Base RefPayload `json:"base"`
	Head RefPayload `json:"head"`
}

type UserPayload struct {
	Login string `json:"login"`
	URL   string `json:"html_url"`
}

type RepositoryPayload struct {
	FullName string `json:"full_name"`
	URL      string `json:"html_url"`
}

func (r *RepositoryPayload) GetOwnerAndName() (string, string) {
	repoParts := strings.Split(r.FullName, "/")
	return repoParts[0], repoParts[1]
}

type RefPayload struct {
	// e.g. 'main'
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}
