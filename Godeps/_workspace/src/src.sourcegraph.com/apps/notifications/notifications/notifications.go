package notifications

import (
	"html/template"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/tracker/issues"
)

type Service interface {
	InternalService
	ExternalService
}

type InternalService interface {
	List(ctx context.Context, opt interface{}) (Notifications, error)
	Count(ctx context.Context, opt interface{}) (uint64, error)
}

type ExternalService interface {
	Subscribe(ctx context.Context, appID string, repo issues.RepoSpec, threadID uint64, subscribers []issues.UserSpec) error

	MarkRead(ctx context.Context, appID string, repo issues.RepoSpec, threadID uint64) error

	Notify(ctx context.Context, appID string, repo issues.RepoSpec, threadID uint64, notification Notification) error
}

type CopierFrom interface {
	CopyFrom(src Service, repo issues.RepoSpec) error // TODO: Consider best place for RepoSpec?
}

type Notification struct {
	RepoSpec  issues.RepoSpec
	RepoURL   template.URL
	Title     string
	Icon      OcticonID
	UpdatedAt time.Time
	HTMLURL   template.URL // Address of notification target.
}

// Octicon ID. E.g., "issue-opened".
type OcticonID string

// Notifications implements sort.Interface.
type Notifications []Notification

func (s Notifications) Len() int           { return len(s) }
func (s Notifications) Less(i, j int) bool { return !s[i].UpdatedAt.Before(s[j].UpdatedAt) }
func (s Notifications) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
