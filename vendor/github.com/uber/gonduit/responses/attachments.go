package responses

import "github.com/uber/gonduit/util"

// SearchAttachmentSubscribers is common attachment with subscribers information
// for *.search API methods.
type SearchAttachmentSubscribers struct {
	SubscriberPHIDs    []string `json:"subscriberPHIDs"`
	SubscriberCount    int      `json:"subscriberCount"`
	ViewerIsSubscribed bool     `json:"viewerIsSubscribed"`
}

// SearchAttachmentProjects is common attachment with projects information for
// *.search API methods.
type SearchAttachmentProjects struct {
	ProjectPHIDs []string `json:"projectPHIDs"`
}

// SearchAttachmentReviewers is attachment with revision reviewers information
// for differenial.revision.search API method.
type SearchAttachmentReviewers struct {
	Reviewers []AttachmentReviewer `json:"reviewers"`
}

// AttachmentReviewer is a single revision reviewer in reviewers list.
type AttachmentReviewer struct {
	ReviewerPHID    string `json:"reviewerPHID"`
	Status          string `json:"status"`
	IsBlocking      bool   `json:"isBlocking"`
	ActorPHID       string `json:"actorPHID"`
	IsCurrentAction bool   `json:"isCurrentAction"`
}

// SearchAttachmentMetrics is an attachment of repository metrics.
type SearchAttachmentMetrics struct {
	CommitCount int `json:"commitCount"`
}

type SearchAttachmentURIs struct {
	URIs []RepositoryURIItem `json:"uris"`
}

// SearchAttachmentCommits is an attachment of diff commits.
type SearchAttachmentCommits struct {
	Commits []AttachmentCommit `json:"commits"`
}

type AttachmentCommit struct {
	Identifier string                 `json:"identifier"`
	Tree       string                 `json:"tree"`
	Parents    []string               `json:"parents"`
	Author     AttachmentCommitAuthor `json:"author"`
	Message    string                 `json:"message"`
}

type AttachmentCommitAuthor struct {
	Name  string             `json:"name"`
	Email string             `json:"email"`
	Raw   string             `json:"raw"`
	Epoch util.UnixTimestamp `json:"epoch"`
}

type SearchAttachmentMembers struct {
	Members []AttachmentMember `json:"members"`
}

type AttachmentMember struct {
	PHID string `json:"phid"`
}

type SearchAttachmentWatchers struct {
	Watchers []AttachmentWatcher `json:"watchers"`
}

type AttachmentWatcher struct {
	PHID string `json:"phid"`
}

type SearchAttachmentAncestors struct {
	Ancestors []ProjectParent `json:"ancestors"`
}
