package types

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type changesetEventUpdateMismatchError struct {
	field    string
	original any
	revised  any
}

func (e *changesetEventUpdateMismatchError) Error() string {
	return fmt.Sprintf("%s '%v' on the revised changeset event does not match %s '%v' on the original changeset event", e.field, e.revised, e.field, e.original)
}

// ChangesetEventKind defines the kind of a ChangesetEvent. This type is unexported
// so that users of ChangesetEvent can't instantiate it with a Kind being an arbitrary
// string.
type ChangesetEventKind string

// Valid ChangesetEvent kinds
const (
	ChangesetEventKindGitHubAssigned             ChangesetEventKind = "github:assigned"
	ChangesetEventKindGitHubClosed               ChangesetEventKind = "github:closed"
	ChangesetEventKindGitHubCommented            ChangesetEventKind = "github:commented"
	ChangesetEventKindGitHubRenamedTitle         ChangesetEventKind = "github:renamed"
	ChangesetEventKindGitHubMerged               ChangesetEventKind = "github:merged"
	ChangesetEventKindGitHubReviewed             ChangesetEventKind = "github:reviewed"
	ChangesetEventKindGitHubReopened             ChangesetEventKind = "github:reopened"
	ChangesetEventKindGitHubReviewDismissed      ChangesetEventKind = "github:review_dismissed"
	ChangesetEventKindGitHubReviewRequestRemoved ChangesetEventKind = "github:review_request_removed"
	ChangesetEventKindGitHubReviewRequested      ChangesetEventKind = "github:review_requested"
	ChangesetEventKindGitHubReviewCommented      ChangesetEventKind = "github:review_commented"
	ChangesetEventKindGitHubReadyForReview       ChangesetEventKind = "github:ready_for_review"
	ChangesetEventKindGitHubConvertToDraft       ChangesetEventKind = "github:convert_to_draft"
	ChangesetEventKindGitHubUnassigned           ChangesetEventKind = "github:unassigned"
	ChangesetEventKindGitHubCommit               ChangesetEventKind = "github:commit"
	ChangesetEventKindGitHubLabeled              ChangesetEventKind = "github:labeled"
	ChangesetEventKindGitHubUnlabeled            ChangesetEventKind = "github:unlabeled"
	ChangesetEventKindCommitStatus               ChangesetEventKind = "github:commit_status"
	ChangesetEventKindCheckSuite                 ChangesetEventKind = "github:check_suite"
	ChangesetEventKindCheckRun                   ChangesetEventKind = "github:check_run"

	ChangesetEventKindBitbucketServerApproved     ChangesetEventKind = "bitbucketserver:approved"
	ChangesetEventKindBitbucketServerUnapproved   ChangesetEventKind = "bitbucketserver:unapproved"
	ChangesetEventKindBitbucketServerDeclined     ChangesetEventKind = "bitbucketserver:declined"
	ChangesetEventKindBitbucketServerReviewed     ChangesetEventKind = "bitbucketserver:reviewed"
	ChangesetEventKindBitbucketServerOpened       ChangesetEventKind = "bitbucketserver:opened"
	ChangesetEventKindBitbucketServerReopened     ChangesetEventKind = "bitbucketserver:reopened"
	ChangesetEventKindBitbucketServerRescoped     ChangesetEventKind = "bitbucketserver:rescoped"
	ChangesetEventKindBitbucketServerUpdated      ChangesetEventKind = "bitbucketserver:updated"
	ChangesetEventKindBitbucketServerCommented    ChangesetEventKind = "bitbucketserver:commented"
	ChangesetEventKindBitbucketServerMerged       ChangesetEventKind = "bitbucketserver:merged"
	ChangesetEventKindBitbucketServerCommitStatus ChangesetEventKind = "bitbucketserver:commit_status"

	// BitbucketServer calls this an Unapprove event but we've called it Dismissed to more
	// clearly convey that it only occurs when a request for changes has been dismissed.
	ChangesetEventKindBitbucketServerDismissed ChangesetEventKind = "bitbucketserver:participant_status:unapproved"

	ChangesetEventKindGitLabApproved             ChangesetEventKind = "gitlab:approved"
	ChangesetEventKindGitLabClosed               ChangesetEventKind = "gitlab:closed"
	ChangesetEventKindGitLabMerged               ChangesetEventKind = "gitlab:merged"
	ChangesetEventKindGitLabPipeline             ChangesetEventKind = "gitlab:pipeline"
	ChangesetEventKindGitLabReopened             ChangesetEventKind = "gitlab:reopened"
	ChangesetEventKindGitLabUnapproved           ChangesetEventKind = "gitlab:unapproved"
	ChangesetEventKindGitLabMarkWorkInProgress   ChangesetEventKind = "gitlab:mark_wip"
	ChangesetEventKindGitLabUnmarkWorkInProgress ChangesetEventKind = "gitlab:unmark_wip"

	// These changeset events are created as the result of regular syncs with
	// Bitbucket Cloud.
	ChangesetEventKindBitbucketCloudApproved         ChangesetEventKind = "bitbucketcloud:approved"
	ChangesetEventKindBitbucketCloudChangesRequested ChangesetEventKind = "bitbucketcloud:changes_requested"
	ChangesetEventKindBitbucketCloudCommitStatus     ChangesetEventKind = "bitbucketcloud:commit_status"
	ChangesetEventKindBitbucketCloudReviewed         ChangesetEventKind = "bitbucketcloud:reviewed"

	// These changes events are created in response to webhooks received from
	// Bitbucket Cloud. The exact type that matches each event is included in a
	// comment after the constant.
	ChangesetEventKindBitbucketCloudPullRequestApproved              ChangesetEventKind = "bitbucketcloud:pullrequest:approved"                // PullRequestApprovalEvent
	ChangesetEventKindBitbucketCloudPullRequestChangesRequestCreated ChangesetEventKind = "bitbucketcloud:pullrequest:changes_request_created" // PullRequestChangesRequestCreatedEvent
	ChangesetEventKindBitbucketCloudPullRequestChangesRequestRemoved ChangesetEventKind = "bitbucketcloud:pullrequest:changes_request_removed" // PullRequestChangesRequestRemovedEvent
	ChangesetEventKindBitbucketCloudPullRequestCommentCreated        ChangesetEventKind = "bitbucketcloud:pullrequest:comment_created"         // PullRequestCommentCreatedEvent
	ChangesetEventKindBitbucketCloudPullRequestCommentDeleted        ChangesetEventKind = "bitbucketcloud:pullrequest:comment_deleted"         // PullRequestCommentDeletedEvent
	ChangesetEventKindBitbucketCloudPullRequestCommentUpdated        ChangesetEventKind = "bitbucketcloud:pullrequest:comment_updated"         // PullRequestCommentUpdatedEvent
	ChangesetEventKindBitbucketCloudPullRequestFulfilled             ChangesetEventKind = "bitbucketcloud:pullrequest:fulfilled"               // PullRequestFulfilledEvent
	ChangesetEventKindBitbucketCloudPullRequestRejected              ChangesetEventKind = "bitbucketcloud:pullrequest:rejected"                // PullRequestRejectedEvent
	ChangesetEventKindBitbucketCloudPullRequestUnapproved            ChangesetEventKind = "bitbucketcloud:pullrequest:unapproved"              // PullRequestUnapprovedEvent
	ChangesetEventKindBitbucketCloudPullRequestUpdated               ChangesetEventKind = "bitbucketcloud:pullrequest:updated"                 // PullRequestUpdatedEvent
	ChangesetEventKindBitbucketCloudRepoCommitStatusCreated          ChangesetEventKind = "bitbucketcloud:repo:commit_status_created"          // RepoCommitStatusCreatedEvent
	ChangesetEventKindBitbucketCloudRepoCommitStatusUpdated          ChangesetEventKind = "bitbucketcloud:repo:commit_status_updated"          // RepoCommitStatusUpdatedEvent

	ChangesetEventKindAzureDevOpsPullRequestMerged                  ChangesetEventKind = "azuredevops:pullrequest:merged"
	ChangesetEventKindAzureDevOpsPullRequestUpdated                 ChangesetEventKind = "azuredevops:pullrequest:updated"
	ChangesetEventKindAzureDevOpsPullRequestApproved                ChangesetEventKind = "azuredevops:pullrequest:approved"
	ChangesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions ChangesetEventKind = "azuredevops:pullrequest:approved_with_suggestions"
	ChangesetEventKindAzureDevOpsPullRequestReviewed                ChangesetEventKind = "azuredevops:pullrequest:reviewed"
	ChangesetEventKindAzureDevOpsPullRequestWaitingForAuthor        ChangesetEventKind = "azuredevops:pullrequest:waiting_for_author"
	ChangesetEventKindAzureDevOpsPullRequestRejected                ChangesetEventKind = "azuredevops:pullrequest:rejected"
	ChangesetEventKindAzureDevOpsPullRequestBuildSucceeded          ChangesetEventKind = "azuredevops:pullrequest:build_succeeded"
	ChangesetEventKindAzureDevOpsPullRequestBuildFailed             ChangesetEventKind = "azuredevops:pullrequest:build_failed"
	ChangesetEventKindAzureDevOpsPullRequestBuildError              ChangesetEventKind = "azuredevops:pullrequest:build_error"
	ChangesetEventKindAzureDevOpsPullRequestBuildPending            ChangesetEventKind = "azuredevops:pullrequest:build_pending"

	ChangesetEventKindGerritChangeApproved                ChangesetEventKind = "gerrit:change:approved"
	ChangesetEventKindGerritChangeApprovedWithSuggestions ChangesetEventKind = "gerrit:change:approved_with_suggestions"
	ChangesetEventKindGerritChangeReviewed                ChangesetEventKind = "gerrit:change:reviewed"
	ChangesetEventKindGerritChangeNeedsChanges            ChangesetEventKind = "gerrit:change:needs_changes"
	ChangesetEventKindGerritChangeRejected                ChangesetEventKind = "gerrit:change:rejected"
	ChangesetEventKindGerritChangeBuildSucceeded          ChangesetEventKind = "gerrit:change:build_succeeded"
	ChangesetEventKindGerritChangeBuildFailed             ChangesetEventKind = "gerrit:change:build_failed"
	ChangesetEventKindGerritChangeBuildPending            ChangesetEventKind = "gerrit:change:build_pending"

	ChangesetEventKindInvalid ChangesetEventKind = "invalid"
)

// A ChangesetEvent is an event that happened in the lifetime
// and context of a Changeset.
type ChangesetEvent struct {
	ID          int64
	ChangesetID int64
	Kind        ChangesetEventKind
	Key         string // Deduplication key
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    any
}

// Clone returns a clone of a ChangesetEvent.
func (e *ChangesetEvent) Clone() *ChangesetEvent {
	ee := *e
	return &ee
}

// ReviewAuthor returns the author of the review if the ChangesetEvent is related to a review.
// Returns an empty string if not a review event or the author has been deleted.
func (e *ChangesetEvent) ReviewAuthor() string {
	switch meta := e.Metadata.(type) {

	case *github.PullRequestReview:
		return meta.Author.Login
	case *github.ReviewDismissedEvent:
		return meta.Review.Author.Login

	case *bitbucketserver.Activity:
		return meta.User.Name
	case *bitbucketserver.ParticipantStatusEvent:
		return meta.User.Name

	case *gitlab.ReviewApprovedEvent:
		return meta.Author.Username
	case *gitlab.ReviewUnapprovedEvent:
		return meta.Author.Username

	// Bitbucket Cloud generally doesn't return the username in the objects that
	// we get when syncing or in webhooks, but since this just has to be unique
	// for each author and isn't surfaced in the UI, we can use the UUID.
	case *bitbucketcloud.Participant:
		return meta.User.UUID
	case *bitbucketcloud.PullRequestApprovedEvent:
		return meta.Approval.User.UUID
	case *bitbucketcloud.PullRequestUnapprovedEvent:
		return meta.Approval.User.UUID
	case *bitbucketcloud.PullRequestChangesRequestCreatedEvent:
		return meta.ChangesRequest.User.UUID
	case *bitbucketcloud.PullRequestChangesRequestRemovedEvent:
		return meta.ChangesRequest.User.UUID

	case *azuredevops.PullRequestApprovedEvent:
		if len(meta.PullRequest.Reviewers) == 0 {
			return meta.PullRequest.CreatedBy.UniqueName
		}
		return meta.PullRequest.Reviewers[len(meta.PullRequest.Reviewers)-1].UniqueName
	case *azuredevops.PullRequestApprovedWithSuggestionsEvent:
		if len(meta.PullRequest.Reviewers) == 0 {
			return meta.PullRequest.CreatedBy.UniqueName
		}
		return meta.PullRequest.Reviewers[len(meta.PullRequest.Reviewers)-1].UniqueName
	case *azuredevops.PullRequestWaitingForAuthorEvent:
		if len(meta.PullRequest.Reviewers) == 0 {
			return meta.PullRequest.CreatedBy.UniqueName
		}
		return meta.PullRequest.Reviewers[len(meta.PullRequest.Reviewers)-1].UniqueName
	case *azuredevops.PullRequestRejectedEvent:
		if len(meta.PullRequest.Reviewers) == 0 {
			return meta.PullRequest.CreatedBy.UniqueName
		}
		return meta.PullRequest.Reviewers[len(meta.PullRequest.Reviewers)-1].UniqueName
	case *azuredevops.PullRequestUpdatedEvent:
		return meta.PullRequest.CreatedBy.UniqueName
	default:
		return ""
	}
}

// ReviewState returns the review state of the ChangesetEvent if it is a review event.
func (e *ChangesetEvent) ReviewState() (ChangesetReviewState, error) {
	switch e.Kind {
	case ChangesetEventKindBitbucketServerApproved,
		ChangesetEventKindGitLabApproved,
		ChangesetEventKindBitbucketCloudApproved,
		ChangesetEventKindBitbucketCloudPullRequestApproved,
		ChangesetEventKindAzureDevOpsPullRequestApproved:
		return ChangesetReviewStateApproved, nil

	// BitbucketServer's "REVIEWED" activity is created when someone clicks
	// the "Needs work" button in the UI, which is why we map it to "Changes Requested"
	case ChangesetEventKindBitbucketServerReviewed,
		ChangesetEventKindBitbucketCloudChangesRequested,
		ChangesetEventKindBitbucketCloudPullRequestChangesRequestCreated,
		ChangesetEventKindAzureDevOpsPullRequestWaitingForAuthor,
		ChangesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions:
		return ChangesetReviewStateChangesRequested, nil

	case ChangesetEventKindGitHubReviewed:
		review, ok := e.Metadata.(*github.PullRequestReview)
		if !ok {
			return "", errors.New("ChangesetEvent metadata event not PullRequestReview")
		}

		s := ChangesetReviewState(strings.ToUpper(review.State))
		if !s.Valid() {
			// Ignore invalid states
			log15.Warn("invalid review state", "state", review.State)
			return ChangesetReviewStatePending, nil
		}
		return s, nil

	case ChangesetEventKindGitHubReviewDismissed,
		ChangesetEventKindBitbucketServerUnapproved,
		ChangesetEventKindBitbucketServerDismissed,
		ChangesetEventKindGitLabUnapproved,
		ChangesetEventKindBitbucketCloudPullRequestUnapproved,
		ChangesetEventKindBitbucketCloudPullRequestChangesRequestRemoved,
		ChangesetEventKindAzureDevOpsPullRequestRejected:
		return ChangesetReviewStateDismissed, nil
	default:
		return ChangesetReviewStatePending, nil
	}
}

// Type returns the ChangesetEventKind of the ChangesetEvent.
func (e *ChangesetEvent) Type() ChangesetEventKind {
	return e.Kind
}

// Changeset returns the changeset ID of the ChangesetEvent.
func (e *ChangesetEvent) Changeset() int64 {
	return e.ChangesetID
}

// Timestamp returns the time when the ChangesetEvent happened (or was updated)
// on the codehost, not when it was created in Sourcegraph's database.
func (e *ChangesetEvent) Timestamp() time.Time {
	var t time.Time

	switch ev := e.Metadata.(type) {
	case *github.AssignedEvent:
		t = ev.CreatedAt
	case *github.ClosedEvent:
		t = ev.CreatedAt
	case *github.IssueComment:
		t = ev.UpdatedAt
	case *github.RenamedTitleEvent:
		t = ev.CreatedAt
	case *github.MergedEvent:
		t = ev.CreatedAt
	case *github.PullRequestReview:
		t = ev.UpdatedAt
	case *github.PullRequestReviewComment:
		t = ev.UpdatedAt
	case *github.ReopenedEvent:
		t = ev.CreatedAt
	case *github.ReviewDismissedEvent:
		t = ev.CreatedAt
	case *github.ReviewRequestRemovedEvent:
		t = ev.CreatedAt
	case *github.ReviewRequestedEvent:
		t = ev.CreatedAt
	case *github.ReadyForReviewEvent:
		t = ev.CreatedAt
	case *github.ConvertToDraftEvent:
		t = ev.CreatedAt
	case *github.UnassignedEvent:
		t = ev.CreatedAt
	case *github.LabelEvent:
		t = ev.CreatedAt
	case *github.CommitStatus:
		t = ev.ReceivedAt
	case *github.CheckSuite:
		t = ev.ReceivedAt
	case *github.CheckRun:
		t = ev.ReceivedAt
	case *bitbucketserver.Activity:
		t = unixMilliToTime(int64(ev.CreatedDate))
	case *bitbucketserver.ParticipantStatusEvent:
		t = unixMilliToTime(int64(ev.CreatedDate))
	case *bitbucketserver.CommitStatus:
		t = unixMilliToTime(ev.Status.DateAdded)
	case *gitlab.ReviewApprovedEvent:
		t = ev.CreatedAt.Time
	case *gitlab.ReviewUnapprovedEvent:
		t = ev.CreatedAt.Time
	case *gitlab.MarkWorkInProgressEvent:
		t = ev.CreatedAt.Time
	case *gitlab.UnmarkWorkInProgressEvent:
		t = ev.CreatedAt.Time
	case *gitlab.MergeRequestClosedEvent:
		t = ev.CreatedAt.Time
	case *gitlab.MergeRequestReopenedEvent:
		t = ev.CreatedAt.Time
	case *gitlab.MergeRequestMergedEvent:
		t = ev.CreatedAt.Time
	case *gitlabwebhooks.PipelineEvent:
		// These events do not inherently have timestamps from GitLab, so we
		// fall back to the event record we created when we received the
		// webhook.
		t = e.CreatedAt
	case *bitbucketcloud.Participant:
		t = ev.ParticipatedOn
	case *bitbucketcloud.PullRequestStatus:
		t = ev.CreatedOn
	case *bitbucketcloud.PullRequestApprovedEvent:
		t = ev.Approval.Date
	case *bitbucketcloud.PullRequestChangesRequestCreatedEvent:
		t = ev.ChangesRequest.Date
	case *bitbucketcloud.PullRequestChangesRequestRemovedEvent:
		t = ev.ChangesRequest.Date
	case *bitbucketcloud.PullRequestCommentCreatedEvent:
		t = ev.Comment.CreatedOn
	case *bitbucketcloud.PullRequestCommentDeletedEvent:
		t = ev.Comment.UpdatedOn
	case *bitbucketcloud.PullRequestCommentUpdatedEvent:
		t = ev.Comment.UpdatedOn
	case *bitbucketcloud.PullRequestFulfilledEvent:
		t = ev.PullRequest.UpdatedOn
	case *bitbucketcloud.PullRequestRejectedEvent:
		t = ev.PullRequest.UpdatedOn
	case *bitbucketcloud.PullRequestUnapprovedEvent:
		t = ev.Approval.Date
	case *bitbucketcloud.PullRequestUpdatedEvent:
		t = ev.PullRequest.UpdatedOn
	case *bitbucketcloud.RepoCommitStatusCreatedEvent:
		t = ev.CommitStatus.CreatedOn
	case *bitbucketcloud.RepoCommitStatusUpdatedEvent:
		t = ev.CommitStatus.UpdatedOn
	case *azuredevops.PullRequestUpdatedEvent:
		t = ev.CreatedDate
	case *azuredevops.PullRequestApprovedEvent:
		t = ev.CreatedDate
	case *azuredevops.PullRequestApprovedWithSuggestionsEvent:
		t = ev.CreatedDate
	case *azuredevops.PullRequestWaitingForAuthorEvent:
		t = ev.CreatedDate
	case *azuredevops.PullRequestRejectedEvent:
		t = ev.CreatedDate
	case *azuredevops.PullRequestMergedEvent:
		t = ev.CreatedDate
	}

	return t
}

// Update updates the metadata of e with new metadata in o.
func (e *ChangesetEvent) Update(o *ChangesetEvent) error {
	if e.ChangesetID != o.ChangesetID {
		return &changesetEventUpdateMismatchError{
			field:    "ChangesetID",
			original: e.ChangesetID,
			revised:  o.ChangesetID,
		}
	}
	if e.Kind != o.Kind {
		return &changesetEventUpdateMismatchError{
			field:    "Kind",
			original: e.Kind,
			revised:  o.Kind,
		}
	}
	if e.Key != o.Key {
		return &changesetEventUpdateMismatchError{
			field:    "Key",
			original: e.Key,
			revised:  o.Key,
		}
	}

	switch e := e.Metadata.(type) {
	case *github.LabelEvent:
		o := o.Metadata.(*github.LabelEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		if e.Label == (github.Label{}) {
			e.Label = o.Label
		}

	case *github.AssignedEvent:
		o := o.Metadata.(*github.AssignedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.Assignee == (github.Actor{}) {
			e.Assignee = o.Assignee
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.ClosedEvent:
		o := o.Metadata.(*github.ClosedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.IssueComment:
		o := o.Metadata.(*github.IssueComment)

		if e.DatabaseID == 0 {
			e.DatabaseID = o.DatabaseID
		}

		if e.Author == (github.Actor{}) {
			e.Author = o.Author
		}

		if e.Editor == nil {
			e.Editor = o.Editor
		}

		if o.AuthorAssociation != "" && e.AuthorAssociation != o.AuthorAssociation {
			e.AuthorAssociation = o.AuthorAssociation
		}

		if o.Body != "" && e.Body != o.Body {
			e.Body = o.Body
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		if e.UpdatedAt.Before(o.UpdatedAt) {
			e.UpdatedAt = o.UpdatedAt
		}

		if o.IncludesCreatedEdit {
			e.IncludesCreatedEdit = true
		}

	case *github.RenamedTitleEvent:
		o := o.Metadata.(*github.RenamedTitleEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.PreviousTitle != "" && e.PreviousTitle != o.PreviousTitle {
			e.PreviousTitle = o.PreviousTitle
		}

		if o.CurrentTitle != "" && e.CurrentTitle != o.CurrentTitle {
			e.CurrentTitle = o.CurrentTitle
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.MergedEvent:
		o := o.Metadata.(*github.MergedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.MergeRefName != "" && e.MergeRefName != o.MergeRefName {
			e.MergeRefName = o.MergeRefName
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		updateGitHubCommit(&e.Commit, &o.Commit)

	case *github.PullRequestReview:
		o := o.Metadata.(*github.PullRequestReview)

		updateGitHubPullRequestReview(e, o)

	case *github.PullRequestReviewComment:
		o := o.Metadata.(*github.PullRequestReviewComment)

		if e.DatabaseID == 0 {
			e.DatabaseID = o.DatabaseID
		}

		if e.Author == (github.Actor{}) {
			e.Author = o.Author
		}

		if o.AuthorAssociation != "" && e.AuthorAssociation != o.AuthorAssociation {
			e.AuthorAssociation = o.AuthorAssociation
		}

		if e.Editor == (github.Actor{}) {
			e.Editor = o.Editor
		}

		if o.Body != "" && e.Body != o.Body {
			e.Body = o.Body
		}

		if o.State != "" && e.State != o.State {
			e.State = o.State
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		if e.UpdatedAt.Before(o.UpdatedAt) {
			e.UpdatedAt = o.UpdatedAt
		}

		if e, o := e.Commit, o.Commit; !reflect.DeepEqual(e, o) {
			updateGitHubCommit(&e, &o)
		}

		if o.IncludesCreatedEdit {
			e.IncludesCreatedEdit = true
		}

	case *github.ReopenedEvent:
		o := o.Metadata.(*github.ReopenedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
	case *github.ReviewDismissedEvent:
		o := o.Metadata.(*github.ReviewDismissedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.DismissalMessage != "" && e.DismissalMessage != o.DismissalMessage {
			e.DismissalMessage = o.DismissalMessage
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		updateGitHubPullRequestReview(&e.Review, &o.Review)

	case *github.ReviewRequestRemovedEvent:
		o := o.Metadata.(*github.ReviewRequestRemovedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.RequestedReviewer == (github.Actor{}) {
			e.RequestedReviewer = o.RequestedReviewer
		}

		if e.RequestedTeam == (github.Team{}) {
			e.RequestedTeam = o.RequestedTeam
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.ReviewRequestedEvent:
		o := o.Metadata.(*github.ReviewRequestedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.RequestedReviewer == (github.Actor{}) {
			e.RequestedReviewer = o.RequestedReviewer
		}

		if e.RequestedTeam == (github.Team{}) {
			e.RequestedTeam = o.RequestedTeam
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.ReadyForReviewEvent:
		o := o.Metadata.(*github.ReadyForReviewEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.ConvertToDraftEvent:
		o := o.Metadata.(*github.ConvertToDraftEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.UnassignedEvent:
		o := o.Metadata.(*github.UnassignedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.Assignee == (github.Actor{}) {
			e.Assignee = o.Assignee
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
	case *bitbucketserver.Activity:
		o := o.Metadata.(*bitbucketserver.Activity)

		if e.CreatedDate == 0 {
			e.CreatedDate = o.CreatedDate
		}

		if e.User == (bitbucketserver.User{}) {
			e.User = o.User
		}

		if e.Action == "" {
			e.Action = o.Action
		}

		if e.CommentAction == "" {
			e.CommentAction = o.CommentAction
		}

		if e.Comment == nil && o.Comment != nil {
			e.Comment = o.Comment
		}

		if len(e.AddedReviewers) == 0 {
			e.AddedReviewers = o.AddedReviewers
		}

		if len(e.RemovedReviewers) == 0 {
			e.RemovedReviewers = o.RemovedReviewers
		}

		if e.Commit == nil && o.Commit != nil {
			e.Commit = o.Commit
		}

	case *bitbucketserver.ParticipantStatusEvent:
		o := o.Metadata.(*bitbucketserver.ParticipantStatusEvent)

		if e.CreatedDate == 0 {
			e.CreatedDate = o.CreatedDate
		}

		if e.Action == "" {
			e.Action = o.Action
		}

		if e.User == (bitbucketserver.User{}) {
			e.User = o.User
		}

	case *bitbucketserver.CommitStatus:
		o := o.Metadata.(*bitbucketserver.CommitStatus)
		// We always get the full event, so safe to replace it
		*e = *o

	case *github.CheckRun:
		o := o.Metadata.(*github.CheckRun)
		if e.Status == "" {
			e.Status = o.Status
		}
		if e.Conclusion == "" {
			e.Conclusion = o.Conclusion
		}

	case *github.CheckSuite:
		o := o.Metadata.(*github.CheckSuite)
		if e.Status == "" {
			e.Status = o.Status
		}
		if e.Conclusion == "" {
			e.Conclusion = o.Conclusion
		}
		e.CheckRuns = o.CheckRuns

	case *gitlab.ReviewApprovedEvent:
		o := o.Metadata.(*gitlab.ReviewApprovedEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if !e.System {
			e.System = o.System
		}
		if e.Body == "" {
			e.Body = o.Body
		}
		if e.Author.ID == 0 {
			e.Author = o.Author
		}

	case *gitlab.ReviewUnapprovedEvent:
		o := o.Metadata.(*gitlab.ReviewUnapprovedEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if !e.System {
			e.System = o.System
		}
		if e.Body == "" {
			e.Body = o.Body
		}
		if e.Author.ID == 0 {
			e.Author = o.Author
		}

	case *gitlab.MarkWorkInProgressEvent:
		o := o.Metadata.(*gitlab.MarkWorkInProgressEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if !e.System {
			e.System = o.System
		}
		if e.Body == "" {
			e.Body = o.Body
		}
		if e.Author.ID == 0 {
			e.Author = o.Author
		}

	case *gitlab.UnmarkWorkInProgressEvent:
		o := o.Metadata.(*gitlab.UnmarkWorkInProgressEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if !e.System {
			e.System = o.System
		}
		if e.Body == "" {
			e.Body = o.Body
		}
		if e.Author.ID == 0 {
			e.Author = o.Author
		}

	case *gitlab.MergeRequestClosedEvent:
		o := o.Metadata.(*gitlab.MergeRequestClosedEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if e.ResourceID == 0 {
			e.ResourceID = o.ResourceID
		}
		if e.User.ID == 0 {
			e.User = o.User
		}

	case *gitlab.MergeRequestReopenedEvent:
		o := o.Metadata.(*gitlab.MergeRequestReopenedEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if e.ResourceID == 0 {
			e.ResourceID = o.ResourceID
		}
		if e.User.ID == 0 {
			e.User = o.User
		}

	case *gitlab.MergeRequestMergedEvent:
		o := o.Metadata.(*gitlab.MergeRequestMergedEvent)
		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
		if e.ResourceID == 0 {
			e.ResourceID = o.ResourceID
		}
		if e.User.ID == 0 {
			e.User = o.User
		}

	case *gitlabwebhooks.PipelineEvent:
		o := o.Metadata.(*gitlabwebhooks.PipelineEvent)
		// We always get the full event, so safe to replace it
		*e = *o

	case *bitbucketcloud.Participant:
		o := o.Metadata.(*bitbucketcloud.Participant)
		*e = *o
	case *bitbucketcloud.PullRequestStatus:
		o := o.Metadata.(*bitbucketcloud.PullRequestStatus)
		*e = *o
	case *bitbucketcloud.PullRequestApprovedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestApprovedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestChangesRequestCreatedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestChangesRequestCreatedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestChangesRequestRemovedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestChangesRequestRemovedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestCommentCreatedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestCommentCreatedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestCommentDeletedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestCommentDeletedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestCommentUpdatedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestCommentUpdatedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestFulfilledEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestFulfilledEvent)
		*e = *o
	case *bitbucketcloud.PullRequestRejectedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestRejectedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestUnapprovedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestUnapprovedEvent)
		*e = *o
	case *bitbucketcloud.PullRequestUpdatedEvent:
		o := o.Metadata.(*bitbucketcloud.PullRequestUpdatedEvent)
		*e = *o
	case *bitbucketcloud.RepoCommitStatusCreatedEvent:
		o := o.Metadata.(*bitbucketcloud.RepoCommitStatusCreatedEvent)
		*e = *o
	case *bitbucketcloud.RepoCommitStatusUpdatedEvent:
		o := o.Metadata.(*bitbucketcloud.RepoCommitStatusUpdatedEvent)
		*e = *o

	case *azuredevops.PullRequestUpdatedEvent:
		o := o.Metadata.(*azuredevops.PullRequestUpdatedEvent)
		*e = *o
	case *azuredevops.PullRequestMergedEvent:
		o := o.Metadata.(*azuredevops.PullRequestMergedEvent)
		*e = *o
	case *azuredevops.PullRequestApprovedEvent:
		o := o.Metadata.(*azuredevops.PullRequestApprovedEvent)
		*e = *o
	case *azuredevops.PullRequestApprovedWithSuggestionsEvent:
		o := o.Metadata.(*azuredevops.PullRequestApprovedWithSuggestionsEvent)
		*e = *o
	case *azuredevops.PullRequestWaitingForAuthorEvent:
		o := o.Metadata.(*azuredevops.PullRequestWaitingForAuthorEvent)
		*e = *o
	case *azuredevops.PullRequestRejectedEvent:
		o := o.Metadata.(*azuredevops.PullRequestRejectedEvent)
		*e = *o
	default:
		return errors.Errorf("unknown changeset event metadata %T", e)
	}

	return nil
}

////////////////////////////////////////////////////
// Helpers for updating changesets from metadata. //
////////////////////////////////////////////////////

func updateGitHubPullRequestReview(e, o *github.PullRequestReview) {
	if e.DatabaseID == 0 {
		e.DatabaseID = o.DatabaseID
	}

	if e.Author == (github.Actor{}) {
		e.Author = o.Author
	}

	if o.AuthorAssociation != "" && e.AuthorAssociation != o.AuthorAssociation {
		e.AuthorAssociation = o.AuthorAssociation
	}

	if o.Body != "" && e.Body != o.Body {
		e.Body = o.Body
	}

	if o.State != "" && e.State != o.State {
		e.State = o.State
	}

	if o.URL != "" && e.URL != o.URL {
		e.URL = o.URL
	}

	if e.CreatedAt.IsZero() {
		e.CreatedAt = o.CreatedAt
	}

	if e.UpdatedAt.Before(o.UpdatedAt) {
		e.UpdatedAt = o.UpdatedAt
	}

	if e, o := e.Commit, o.Commit; !reflect.DeepEqual(e, o) {
		updateGitHubCommit(&e, &o)
	}

	if o.IncludesCreatedEdit {
		e.IncludesCreatedEdit = true
	}
}

func updateGitHubCommit(e, o *github.Commit) {
	if o.OID != "" && e.OID != o.OID {
		e.OID = o.OID
	}

	if o.Message != "" && e.Message != o.Message {
		e.Message = o.Message
	}

	if o.MessageHeadline != "" && e.MessageHeadline != o.MessageHeadline {
		e.MessageHeadline = o.MessageHeadline
	}

	if o.URL != "" && e.URL != o.URL {
		e.URL = o.URL
	}

	if e.Committer != (github.GitActor{}) && e.Committer != o.Committer {
		e.Committer = o.Committer
	}

	if e.CommittedDate.IsZero() {
		e.CommittedDate = o.CommittedDate
	}

	if e.PushedDate.IsZero() {
		e.PushedDate = o.PushedDate
	}
}
