package main

type Payload struct {
	Action      string             `json:"action"`
	PullRequest PullRequestPayload `json:"pull_request"`
	Repository  RepositoryPayload  `json:"repository"`
}

type PullRequestPayload struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`

	Merged   bool        `json:"merged"`
	MergedBy UserPayload `json:"merged_by"`

	URL string `json:"html_url"`
}

type UserPayload struct {
	Login string `json:"login"`
	URL   string `json:"html_url"`
}

type RepositoryPayload struct {
	FullName string `json:"full_name"`
	URL      string `json:"html_url"`
}
