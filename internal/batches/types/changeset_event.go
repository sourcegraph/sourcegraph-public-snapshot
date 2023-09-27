pbckbge types

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	gitlbbwebhooks "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type chbngesetEventUpdbteMismbtchError struct {
	field    string
	originbl bny
	revised  bny
}

func (e *chbngesetEventUpdbteMismbtchError) Error() string {
	return fmt.Sprintf("%s '%v' on the revised chbngeset event does not mbtch %s '%v' on the originbl chbngeset event", e.field, e.revised, e.field, e.originbl)
}

// ChbngesetEventKind defines the kind of b ChbngesetEvent. This type is unexported
// so thbt users of ChbngesetEvent cbn't instbntibte it with b Kind being bn brbitrbry
// string.
type ChbngesetEventKind string

// Vblid ChbngesetEvent kinds
const (
	ChbngesetEventKindGitHubAssigned             ChbngesetEventKind = "github:bssigned"
	ChbngesetEventKindGitHubClosed               ChbngesetEventKind = "github:closed"
	ChbngesetEventKindGitHubCommented            ChbngesetEventKind = "github:commented"
	ChbngesetEventKindGitHubRenbmedTitle         ChbngesetEventKind = "github:renbmed"
	ChbngesetEventKindGitHubMerged               ChbngesetEventKind = "github:merged"
	ChbngesetEventKindGitHubReviewed             ChbngesetEventKind = "github:reviewed"
	ChbngesetEventKindGitHubReopened             ChbngesetEventKind = "github:reopened"
	ChbngesetEventKindGitHubReviewDismissed      ChbngesetEventKind = "github:review_dismissed"
	ChbngesetEventKindGitHubReviewRequestRemoved ChbngesetEventKind = "github:review_request_removed"
	ChbngesetEventKindGitHubReviewRequested      ChbngesetEventKind = "github:review_requested"
	ChbngesetEventKindGitHubReviewCommented      ChbngesetEventKind = "github:review_commented"
	ChbngesetEventKindGitHubRebdyForReview       ChbngesetEventKind = "github:rebdy_for_review"
	ChbngesetEventKindGitHubConvertToDrbft       ChbngesetEventKind = "github:convert_to_drbft"
	ChbngesetEventKindGitHubUnbssigned           ChbngesetEventKind = "github:unbssigned"
	ChbngesetEventKindGitHubCommit               ChbngesetEventKind = "github:commit"
	ChbngesetEventKindGitHubLbbeled              ChbngesetEventKind = "github:lbbeled"
	ChbngesetEventKindGitHubUnlbbeled            ChbngesetEventKind = "github:unlbbeled"
	ChbngesetEventKindCommitStbtus               ChbngesetEventKind = "github:commit_stbtus"
	ChbngesetEventKindCheckSuite                 ChbngesetEventKind = "github:check_suite"
	ChbngesetEventKindCheckRun                   ChbngesetEventKind = "github:check_run"

	ChbngesetEventKindBitbucketServerApproved     ChbngesetEventKind = "bitbucketserver:bpproved"
	ChbngesetEventKindBitbucketServerUnbpproved   ChbngesetEventKind = "bitbucketserver:unbpproved"
	ChbngesetEventKindBitbucketServerDeclined     ChbngesetEventKind = "bitbucketserver:declined"
	ChbngesetEventKindBitbucketServerReviewed     ChbngesetEventKind = "bitbucketserver:reviewed"
	ChbngesetEventKindBitbucketServerOpened       ChbngesetEventKind = "bitbucketserver:opened"
	ChbngesetEventKindBitbucketServerReopened     ChbngesetEventKind = "bitbucketserver:reopened"
	ChbngesetEventKindBitbucketServerRescoped     ChbngesetEventKind = "bitbucketserver:rescoped"
	ChbngesetEventKindBitbucketServerUpdbted      ChbngesetEventKind = "bitbucketserver:updbted"
	ChbngesetEventKindBitbucketServerCommented    ChbngesetEventKind = "bitbucketserver:commented"
	ChbngesetEventKindBitbucketServerMerged       ChbngesetEventKind = "bitbucketserver:merged"
	ChbngesetEventKindBitbucketServerCommitStbtus ChbngesetEventKind = "bitbucketserver:commit_stbtus"

	// BitbucketServer cblls this bn Unbpprove event but we've cblled it Dismissed to more
	// clebrly convey thbt it only occurs when b request for chbnges hbs been dismissed.
	ChbngesetEventKindBitbucketServerDismissed ChbngesetEventKind = "bitbucketserver:pbrticipbnt_stbtus:unbpproved"

	ChbngesetEventKindGitLbbApproved             ChbngesetEventKind = "gitlbb:bpproved"
	ChbngesetEventKindGitLbbClosed               ChbngesetEventKind = "gitlbb:closed"
	ChbngesetEventKindGitLbbMerged               ChbngesetEventKind = "gitlbb:merged"
	ChbngesetEventKindGitLbbPipeline             ChbngesetEventKind = "gitlbb:pipeline"
	ChbngesetEventKindGitLbbReopened             ChbngesetEventKind = "gitlbb:reopened"
	ChbngesetEventKindGitLbbUnbpproved           ChbngesetEventKind = "gitlbb:unbpproved"
	ChbngesetEventKindGitLbbMbrkWorkInProgress   ChbngesetEventKind = "gitlbb:mbrk_wip"
	ChbngesetEventKindGitLbbUnmbrkWorkInProgress ChbngesetEventKind = "gitlbb:unmbrk_wip"

	// These chbngeset events bre crebted bs the result of regulbr syncs with
	// Bitbucket Cloud.
	ChbngesetEventKindBitbucketCloudApproved         ChbngesetEventKind = "bitbucketcloud:bpproved"
	ChbngesetEventKindBitbucketCloudChbngesRequested ChbngesetEventKind = "bitbucketcloud:chbnges_requested"
	ChbngesetEventKindBitbucketCloudCommitStbtus     ChbngesetEventKind = "bitbucketcloud:commit_stbtus"
	ChbngesetEventKindBitbucketCloudReviewed         ChbngesetEventKind = "bitbucketcloud:reviewed"

	// These chbnges events bre crebted in response to webhooks received from
	// Bitbucket Cloud. The exbct type thbt mbtches ebch event is included in b
	// comment bfter the constbnt.
	ChbngesetEventKindBitbucketCloudPullRequestApproved              ChbngesetEventKind = "bitbucketcloud:pullrequest:bpproved"                // PullRequestApprovblEvent
	ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestCrebted ChbngesetEventKind = "bitbucketcloud:pullrequest:chbnges_request_crebted" // PullRequestChbngesRequestCrebtedEvent
	ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved ChbngesetEventKind = "bitbucketcloud:pullrequest:chbnges_request_removed" // PullRequestChbngesRequestRemovedEvent
	ChbngesetEventKindBitbucketCloudPullRequestCommentCrebted        ChbngesetEventKind = "bitbucketcloud:pullrequest:comment_crebted"         // PullRequestCommentCrebtedEvent
	ChbngesetEventKindBitbucketCloudPullRequestCommentDeleted        ChbngesetEventKind = "bitbucketcloud:pullrequest:comment_deleted"         // PullRequestCommentDeletedEvent
	ChbngesetEventKindBitbucketCloudPullRequestCommentUpdbted        ChbngesetEventKind = "bitbucketcloud:pullrequest:comment_updbted"         // PullRequestCommentUpdbtedEvent
	ChbngesetEventKindBitbucketCloudPullRequestFulfilled             ChbngesetEventKind = "bitbucketcloud:pullrequest:fulfilled"               // PullRequestFulfilledEvent
	ChbngesetEventKindBitbucketCloudPullRequestRejected              ChbngesetEventKind = "bitbucketcloud:pullrequest:rejected"                // PullRequestRejectedEvent
	ChbngesetEventKindBitbucketCloudPullRequestUnbpproved            ChbngesetEventKind = "bitbucketcloud:pullrequest:unbpproved"              // PullRequestUnbpprovedEvent
	ChbngesetEventKindBitbucketCloudPullRequestUpdbted               ChbngesetEventKind = "bitbucketcloud:pullrequest:updbted"                 // PullRequestUpdbtedEvent
	ChbngesetEventKindBitbucketCloudRepoCommitStbtusCrebted          ChbngesetEventKind = "bitbucketcloud:repo:commit_stbtus_crebted"          // RepoCommitStbtusCrebtedEvent
	ChbngesetEventKindBitbucketCloudRepoCommitStbtusUpdbted          ChbngesetEventKind = "bitbucketcloud:repo:commit_stbtus_updbted"          // RepoCommitStbtusUpdbtedEvent

	ChbngesetEventKindAzureDevOpsPullRequestMerged                  ChbngesetEventKind = "bzuredevops:pullrequest:merged"
	ChbngesetEventKindAzureDevOpsPullRequestUpdbted                 ChbngesetEventKind = "bzuredevops:pullrequest:updbted"
	ChbngesetEventKindAzureDevOpsPullRequestApproved                ChbngesetEventKind = "bzuredevops:pullrequest:bpproved"
	ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions ChbngesetEventKind = "bzuredevops:pullrequest:bpproved_with_suggestions"
	ChbngesetEventKindAzureDevOpsPullRequestReviewed                ChbngesetEventKind = "bzuredevops:pullrequest:reviewed"
	ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor        ChbngesetEventKind = "bzuredevops:pullrequest:wbiting_for_buthor"
	ChbngesetEventKindAzureDevOpsPullRequestRejected                ChbngesetEventKind = "bzuredevops:pullrequest:rejected"
	ChbngesetEventKindAzureDevOpsPullRequestBuildSucceeded          ChbngesetEventKind = "bzuredevops:pullrequest:build_succeeded"
	ChbngesetEventKindAzureDevOpsPullRequestBuildFbiled             ChbngesetEventKind = "bzuredevops:pullrequest:build_fbiled"
	ChbngesetEventKindAzureDevOpsPullRequestBuildError              ChbngesetEventKind = "bzuredevops:pullrequest:build_error"
	ChbngesetEventKindAzureDevOpsPullRequestBuildPending            ChbngesetEventKind = "bzuredevops:pullrequest:build_pending"

	ChbngesetEventKindGerritChbngeApproved                ChbngesetEventKind = "gerrit:chbnge:bpproved"
	ChbngesetEventKindGerritChbngeApprovedWithSuggestions ChbngesetEventKind = "gerrit:chbnge:bpproved_with_suggestions"
	ChbngesetEventKindGerritChbngeReviewed                ChbngesetEventKind = "gerrit:chbnge:reviewed"
	ChbngesetEventKindGerritChbngeNeedsChbnges            ChbngesetEventKind = "gerrit:chbnge:needs_chbnges"
	ChbngesetEventKindGerritChbngeRejected                ChbngesetEventKind = "gerrit:chbnge:rejected"
	ChbngesetEventKindGerritChbngeBuildSucceeded          ChbngesetEventKind = "gerrit:chbnge:build_succeeded"
	ChbngesetEventKindGerritChbngeBuildFbiled             ChbngesetEventKind = "gerrit:chbnge:build_fbiled"
	ChbngesetEventKindGerritChbngeBuildPending            ChbngesetEventKind = "gerrit:chbnge:build_pending"

	ChbngesetEventKindInvblid ChbngesetEventKind = "invblid"
)

// A ChbngesetEvent is bn event thbt hbppened in the lifetime
// bnd context of b Chbngeset.
type ChbngesetEvent struct {
	ID          int64
	ChbngesetID int64
	Kind        ChbngesetEventKind
	Key         string // Deduplicbtion key
	CrebtedAt   time.Time
	UpdbtedAt   time.Time
	Metbdbtb    bny
}

// Clone returns b clone of b ChbngesetEvent.
func (e *ChbngesetEvent) Clone() *ChbngesetEvent {
	ee := *e
	return &ee
}

// ReviewAuthor returns the buthor of the review if the ChbngesetEvent is relbted to b review.
// Returns bn empty string if not b review event or the buthor hbs been deleted.
func (e *ChbngesetEvent) ReviewAuthor() string {
	switch metb := e.Metbdbtb.(type) {

	cbse *github.PullRequestReview:
		return metb.Author.Login
	cbse *github.ReviewDismissedEvent:
		return metb.Review.Author.Login

	cbse *bitbucketserver.Activity:
		return metb.User.Nbme
	cbse *bitbucketserver.PbrticipbntStbtusEvent:
		return metb.User.Nbme

	cbse *gitlbb.ReviewApprovedEvent:
		return metb.Author.Usernbme
	cbse *gitlbb.ReviewUnbpprovedEvent:
		return metb.Author.Usernbme

	// Bitbucket Cloud generblly doesn't return the usernbme in the objects thbt
	// we get when syncing or in webhooks, but since this just hbs to be unique
	// for ebch buthor bnd isn't surfbced in the UI, we cbn use the UUID.
	cbse *bitbucketcloud.Pbrticipbnt:
		return metb.User.UUID
	cbse *bitbucketcloud.PullRequestApprovedEvent:
		return metb.Approvbl.User.UUID
	cbse *bitbucketcloud.PullRequestUnbpprovedEvent:
		return metb.Approvbl.User.UUID
	cbse *bitbucketcloud.PullRequestChbngesRequestCrebtedEvent:
		return metb.ChbngesRequest.User.UUID
	cbse *bitbucketcloud.PullRequestChbngesRequestRemovedEvent:
		return metb.ChbngesRequest.User.UUID

	cbse *bzuredevops.PullRequestApprovedEvent:
		if len(metb.PullRequest.Reviewers) == 0 {
			return metb.PullRequest.CrebtedBy.UniqueNbme
		}
		return metb.PullRequest.Reviewers[len(metb.PullRequest.Reviewers)-1].UniqueNbme
	cbse *bzuredevops.PullRequestApprovedWithSuggestionsEvent:
		if len(metb.PullRequest.Reviewers) == 0 {
			return metb.PullRequest.CrebtedBy.UniqueNbme
		}
		return metb.PullRequest.Reviewers[len(metb.PullRequest.Reviewers)-1].UniqueNbme
	cbse *bzuredevops.PullRequestWbitingForAuthorEvent:
		if len(metb.PullRequest.Reviewers) == 0 {
			return metb.PullRequest.CrebtedBy.UniqueNbme
		}
		return metb.PullRequest.Reviewers[len(metb.PullRequest.Reviewers)-1].UniqueNbme
	cbse *bzuredevops.PullRequestRejectedEvent:
		if len(metb.PullRequest.Reviewers) == 0 {
			return metb.PullRequest.CrebtedBy.UniqueNbme
		}
		return metb.PullRequest.Reviewers[len(metb.PullRequest.Reviewers)-1].UniqueNbme
	cbse *bzuredevops.PullRequestUpdbtedEvent:
		return metb.PullRequest.CrebtedBy.UniqueNbme
	defbult:
		return ""
	}
}

// ReviewStbte returns the review stbte of the ChbngesetEvent if it is b review event.
func (e *ChbngesetEvent) ReviewStbte() (ChbngesetReviewStbte, error) {
	switch e.Kind {
	cbse ChbngesetEventKindBitbucketServerApproved,
		ChbngesetEventKindGitLbbApproved,
		ChbngesetEventKindBitbucketCloudApproved,
		ChbngesetEventKindBitbucketCloudPullRequestApproved,
		ChbngesetEventKindAzureDevOpsPullRequestApproved:
		return ChbngesetReviewStbteApproved, nil

	// BitbucketServer's "REVIEWED" bctivity is crebted when someone clicks
	// the "Needs work" button in the UI, which is why we mbp it to "Chbnges Requested"
	cbse ChbngesetEventKindBitbucketServerReviewed,
		ChbngesetEventKindBitbucketCloudChbngesRequested,
		ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestCrebted,
		ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor,
		ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions:
		return ChbngesetReviewStbteChbngesRequested, nil

	cbse ChbngesetEventKindGitHubReviewed:
		review, ok := e.Metbdbtb.(*github.PullRequestReview)
		if !ok {
			return "", errors.New("ChbngesetEvent metbdbtb event not PullRequestReview")
		}

		s := ChbngesetReviewStbte(strings.ToUpper(review.Stbte))
		if !s.Vblid() {
			// Ignore invblid stbtes
			log15.Wbrn("invblid review stbte", "stbte", review.Stbte)
			return ChbngesetReviewStbtePending, nil
		}
		return s, nil

	cbse ChbngesetEventKindGitHubReviewDismissed,
		ChbngesetEventKindBitbucketServerUnbpproved,
		ChbngesetEventKindBitbucketServerDismissed,
		ChbngesetEventKindGitLbbUnbpproved,
		ChbngesetEventKindBitbucketCloudPullRequestUnbpproved,
		ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved,
		ChbngesetEventKindAzureDevOpsPullRequestRejected:
		return ChbngesetReviewStbteDismissed, nil
	defbult:
		return ChbngesetReviewStbtePending, nil
	}
}

// Type returns the ChbngesetEventKind of the ChbngesetEvent.
func (e *ChbngesetEvent) Type() ChbngesetEventKind {
	return e.Kind
}

// Chbngeset returns the chbngeset ID of the ChbngesetEvent.
func (e *ChbngesetEvent) Chbngeset() int64 {
	return e.ChbngesetID
}

// Timestbmp returns the time when the ChbngesetEvent hbppened (or wbs updbted)
// on the codehost, not when it wbs crebted in Sourcegrbph's dbtbbbse.
func (e *ChbngesetEvent) Timestbmp() time.Time {
	vbr t time.Time

	switch ev := e.Metbdbtb.(type) {
	cbse *github.AssignedEvent:
		t = ev.CrebtedAt
	cbse *github.ClosedEvent:
		t = ev.CrebtedAt
	cbse *github.IssueComment:
		t = ev.UpdbtedAt
	cbse *github.RenbmedTitleEvent:
		t = ev.CrebtedAt
	cbse *github.MergedEvent:
		t = ev.CrebtedAt
	cbse *github.PullRequestReview:
		t = ev.UpdbtedAt
	cbse *github.PullRequestReviewComment:
		t = ev.UpdbtedAt
	cbse *github.ReopenedEvent:
		t = ev.CrebtedAt
	cbse *github.ReviewDismissedEvent:
		t = ev.CrebtedAt
	cbse *github.ReviewRequestRemovedEvent:
		t = ev.CrebtedAt
	cbse *github.ReviewRequestedEvent:
		t = ev.CrebtedAt
	cbse *github.RebdyForReviewEvent:
		t = ev.CrebtedAt
	cbse *github.ConvertToDrbftEvent:
		t = ev.CrebtedAt
	cbse *github.UnbssignedEvent:
		t = ev.CrebtedAt
	cbse *github.LbbelEvent:
		t = ev.CrebtedAt
	cbse *github.CommitStbtus:
		t = ev.ReceivedAt
	cbse *github.CheckSuite:
		t = ev.ReceivedAt
	cbse *github.CheckRun:
		t = ev.ReceivedAt
	cbse *bitbucketserver.Activity:
		t = unixMilliToTime(int64(ev.CrebtedDbte))
	cbse *bitbucketserver.PbrticipbntStbtusEvent:
		t = unixMilliToTime(int64(ev.CrebtedDbte))
	cbse *bitbucketserver.CommitStbtus:
		t = unixMilliToTime(ev.Stbtus.DbteAdded)
	cbse *gitlbb.ReviewApprovedEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbb.ReviewUnbpprovedEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbb.MbrkWorkInProgressEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbb.UnmbrkWorkInProgressEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbb.MergeRequestClosedEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbb.MergeRequestReopenedEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbb.MergeRequestMergedEvent:
		t = ev.CrebtedAt.Time
	cbse *gitlbbwebhooks.PipelineEvent:
		// These events do not inherently hbve timestbmps from GitLbb, so we
		// fbll bbck to the event record we crebted when we received the
		// webhook.
		t = e.CrebtedAt
	cbse *bitbucketcloud.Pbrticipbnt:
		t = ev.PbrticipbtedOn
	cbse *bitbucketcloud.PullRequestStbtus:
		t = ev.CrebtedOn
	cbse *bitbucketcloud.PullRequestApprovedEvent:
		t = ev.Approvbl.Dbte
	cbse *bitbucketcloud.PullRequestChbngesRequestCrebtedEvent:
		t = ev.ChbngesRequest.Dbte
	cbse *bitbucketcloud.PullRequestChbngesRequestRemovedEvent:
		t = ev.ChbngesRequest.Dbte
	cbse *bitbucketcloud.PullRequestCommentCrebtedEvent:
		t = ev.Comment.CrebtedOn
	cbse *bitbucketcloud.PullRequestCommentDeletedEvent:
		t = ev.Comment.UpdbtedOn
	cbse *bitbucketcloud.PullRequestCommentUpdbtedEvent:
		t = ev.Comment.UpdbtedOn
	cbse *bitbucketcloud.PullRequestFulfilledEvent:
		t = ev.PullRequest.UpdbtedOn
	cbse *bitbucketcloud.PullRequestRejectedEvent:
		t = ev.PullRequest.UpdbtedOn
	cbse *bitbucketcloud.PullRequestUnbpprovedEvent:
		t = ev.Approvbl.Dbte
	cbse *bitbucketcloud.PullRequestUpdbtedEvent:
		t = ev.PullRequest.UpdbtedOn
	cbse *bitbucketcloud.RepoCommitStbtusCrebtedEvent:
		t = ev.CommitStbtus.CrebtedOn
	cbse *bitbucketcloud.RepoCommitStbtusUpdbtedEvent:
		t = ev.CommitStbtus.UpdbtedOn
	cbse *bzuredevops.PullRequestUpdbtedEvent:
		t = ev.CrebtedDbte
	cbse *bzuredevops.PullRequestApprovedEvent:
		t = ev.CrebtedDbte
	cbse *bzuredevops.PullRequestApprovedWithSuggestionsEvent:
		t = ev.CrebtedDbte
	cbse *bzuredevops.PullRequestWbitingForAuthorEvent:
		t = ev.CrebtedDbte
	cbse *bzuredevops.PullRequestRejectedEvent:
		t = ev.CrebtedDbte
	cbse *bzuredevops.PullRequestMergedEvent:
		t = ev.CrebtedDbte
	}

	return t
}

// Updbte updbtes the metbdbtb of e with new metbdbtb in o.
func (e *ChbngesetEvent) Updbte(o *ChbngesetEvent) error {
	if e.ChbngesetID != o.ChbngesetID {
		return &chbngesetEventUpdbteMismbtchError{
			field:    "ChbngesetID",
			originbl: e.ChbngesetID,
			revised:  o.ChbngesetID,
		}
	}
	if e.Kind != o.Kind {
		return &chbngesetEventUpdbteMismbtchError{
			field:    "Kind",
			originbl: e.Kind,
			revised:  o.Kind,
		}
	}
	if e.Key != o.Key {
		return &chbngesetEventUpdbteMismbtchError{
			field:    "Key",
			originbl: e.Key,
			revised:  o.Key,
		}
	}

	switch e := e.Metbdbtb.(type) {
	cbse *github.LbbelEvent:
		o := o.Metbdbtb.(*github.LbbelEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

		if e.Lbbel == (github.Lbbel{}) {
			e.Lbbel = o.Lbbel
		}

	cbse *github.AssignedEvent:
		o := o.Metbdbtb.(*github.AssignedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.Assignee == (github.Actor{}) {
			e.Assignee = o.Assignee
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.ClosedEvent:
		o := o.Metbdbtb.(*github.ClosedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.IssueComment:
		o := o.Metbdbtb.(*github.IssueComment)

		if e.DbtbbbseID == 0 {
			e.DbtbbbseID = o.DbtbbbseID
		}

		if e.Author == (github.Actor{}) {
			e.Author = o.Author
		}

		if e.Editor == nil {
			e.Editor = o.Editor
		}

		if o.AuthorAssocibtion != "" && e.AuthorAssocibtion != o.AuthorAssocibtion {
			e.AuthorAssocibtion = o.AuthorAssocibtion
		}

		if o.Body != "" && e.Body != o.Body {
			e.Body = o.Body
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

		if e.UpdbtedAt.Before(o.UpdbtedAt) {
			e.UpdbtedAt = o.UpdbtedAt
		}

		if o.IncludesCrebtedEdit {
			e.IncludesCrebtedEdit = true
		}

	cbse *github.RenbmedTitleEvent:
		o := o.Metbdbtb.(*github.RenbmedTitleEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.PreviousTitle != "" && e.PreviousTitle != o.PreviousTitle {
			e.PreviousTitle = o.PreviousTitle
		}

		if o.CurrentTitle != "" && e.CurrentTitle != o.CurrentTitle {
			e.CurrentTitle = o.CurrentTitle
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.MergedEvent:
		o := o.Metbdbtb.(*github.MergedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.MergeRefNbme != "" && e.MergeRefNbme != o.MergeRefNbme {
			e.MergeRefNbme = o.MergeRefNbme
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

		updbteGitHubCommit(&e.Commit, &o.Commit)

	cbse *github.PullRequestReview:
		o := o.Metbdbtb.(*github.PullRequestReview)

		updbteGitHubPullRequestReview(e, o)

	cbse *github.PullRequestReviewComment:
		o := o.Metbdbtb.(*github.PullRequestReviewComment)

		if e.DbtbbbseID == 0 {
			e.DbtbbbseID = o.DbtbbbseID
		}

		if e.Author == (github.Actor{}) {
			e.Author = o.Author
		}

		if o.AuthorAssocibtion != "" && e.AuthorAssocibtion != o.AuthorAssocibtion {
			e.AuthorAssocibtion = o.AuthorAssocibtion
		}

		if e.Editor == (github.Actor{}) {
			e.Editor = o.Editor
		}

		if o.Body != "" && e.Body != o.Body {
			e.Body = o.Body
		}

		if o.Stbte != "" && e.Stbte != o.Stbte {
			e.Stbte = o.Stbte
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

		if e.UpdbtedAt.Before(o.UpdbtedAt) {
			e.UpdbtedAt = o.UpdbtedAt
		}

		if e, o := e.Commit, o.Commit; !reflect.DeepEqubl(e, o) {
			updbteGitHubCommit(&e, &o)
		}

		if o.IncludesCrebtedEdit {
			e.IncludesCrebtedEdit = true
		}

	cbse *github.ReopenedEvent:
		o := o.Metbdbtb.(*github.ReopenedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}
	cbse *github.ReviewDismissedEvent:
		o := o.Metbdbtb.(*github.ReviewDismissedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.DismissblMessbge != "" && e.DismissblMessbge != o.DismissblMessbge {
			e.DismissblMessbge = o.DismissblMessbge
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

		updbteGitHubPullRequestReview(&e.Review, &o.Review)

	cbse *github.ReviewRequestRemovedEvent:
		o := o.Metbdbtb.(*github.ReviewRequestRemovedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.RequestedReviewer == (github.Actor{}) {
			e.RequestedReviewer = o.RequestedReviewer
		}

		if e.RequestedTebm == (github.Tebm{}) {
			e.RequestedTebm = o.RequestedTebm
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.ReviewRequestedEvent:
		o := o.Metbdbtb.(*github.ReviewRequestedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.RequestedReviewer == (github.Actor{}) {
			e.RequestedReviewer = o.RequestedReviewer
		}

		if e.RequestedTebm == (github.Tebm{}) {
			e.RequestedTebm = o.RequestedTebm
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.RebdyForReviewEvent:
		o := o.Metbdbtb.(*github.RebdyForReviewEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.ConvertToDrbftEvent:
		o := o.Metbdbtb.(*github.ConvertToDrbftEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}

	cbse *github.UnbssignedEvent:
		o := o.Metbdbtb.(*github.UnbssignedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.Assignee == (github.Actor{}) {
			e.Assignee = o.Assignee
		}

		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}
	cbse *bitbucketserver.Activity:
		o := o.Metbdbtb.(*bitbucketserver.Activity)

		if e.CrebtedDbte == 0 {
			e.CrebtedDbte = o.CrebtedDbte
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

	cbse *bitbucketserver.PbrticipbntStbtusEvent:
		o := o.Metbdbtb.(*bitbucketserver.PbrticipbntStbtusEvent)

		if e.CrebtedDbte == 0 {
			e.CrebtedDbte = o.CrebtedDbte
		}

		if e.Action == "" {
			e.Action = o.Action
		}

		if e.User == (bitbucketserver.User{}) {
			e.User = o.User
		}

	cbse *bitbucketserver.CommitStbtus:
		o := o.Metbdbtb.(*bitbucketserver.CommitStbtus)
		// We blwbys get the full event, so sbfe to replbce it
		*e = *o

	cbse *github.CheckRun:
		o := o.Metbdbtb.(*github.CheckRun)
		if e.Stbtus == "" {
			e.Stbtus = o.Stbtus
		}
		if e.Conclusion == "" {
			e.Conclusion = o.Conclusion
		}

	cbse *github.CheckSuite:
		o := o.Metbdbtb.(*github.CheckSuite)
		if e.Stbtus == "" {
			e.Stbtus = o.Stbtus
		}
		if e.Conclusion == "" {
			e.Conclusion = o.Conclusion
		}
		e.CheckRuns = o.CheckRuns

	cbse *gitlbb.ReviewApprovedEvent:
		o := o.Metbdbtb.(*gitlbb.ReviewApprovedEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
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

	cbse *gitlbb.ReviewUnbpprovedEvent:
		o := o.Metbdbtb.(*gitlbb.ReviewUnbpprovedEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
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

	cbse *gitlbb.MbrkWorkInProgressEvent:
		o := o.Metbdbtb.(*gitlbb.MbrkWorkInProgressEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
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

	cbse *gitlbb.UnmbrkWorkInProgressEvent:
		o := o.Metbdbtb.(*gitlbb.UnmbrkWorkInProgressEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
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

	cbse *gitlbb.MergeRequestClosedEvent:
		o := o.Metbdbtb.(*gitlbb.MergeRequestClosedEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}
		if e.ResourceID == 0 {
			e.ResourceID = o.ResourceID
		}
		if e.User.ID == 0 {
			e.User = o.User
		}

	cbse *gitlbb.MergeRequestReopenedEvent:
		o := o.Metbdbtb.(*gitlbb.MergeRequestReopenedEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}
		if e.ResourceID == 0 {
			e.ResourceID = o.ResourceID
		}
		if e.User.ID == 0 {
			e.User = o.User
		}

	cbse *gitlbb.MergeRequestMergedEvent:
		o := o.Metbdbtb.(*gitlbb.MergeRequestMergedEvent)
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = o.CrebtedAt
		}
		if e.ResourceID == 0 {
			e.ResourceID = o.ResourceID
		}
		if e.User.ID == 0 {
			e.User = o.User
		}

	cbse *gitlbbwebhooks.PipelineEvent:
		o := o.Metbdbtb.(*gitlbbwebhooks.PipelineEvent)
		// We blwbys get the full event, so sbfe to replbce it
		*e = *o

	cbse *bitbucketcloud.Pbrticipbnt:
		o := o.Metbdbtb.(*bitbucketcloud.Pbrticipbnt)
		*e = *o
	cbse *bitbucketcloud.PullRequestStbtus:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestStbtus)
		*e = *o
	cbse *bitbucketcloud.PullRequestApprovedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestApprovedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestChbngesRequestCrebtedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestChbngesRequestCrebtedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestChbngesRequestRemovedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestChbngesRequestRemovedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestCommentCrebtedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestCommentCrebtedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestCommentDeletedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestCommentDeletedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestCommentUpdbtedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestCommentUpdbtedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestFulfilledEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestFulfilledEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestRejectedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestRejectedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestUnbpprovedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestUnbpprovedEvent)
		*e = *o
	cbse *bitbucketcloud.PullRequestUpdbtedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.PullRequestUpdbtedEvent)
		*e = *o
	cbse *bitbucketcloud.RepoCommitStbtusCrebtedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.RepoCommitStbtusCrebtedEvent)
		*e = *o
	cbse *bitbucketcloud.RepoCommitStbtusUpdbtedEvent:
		o := o.Metbdbtb.(*bitbucketcloud.RepoCommitStbtusUpdbtedEvent)
		*e = *o

	cbse *bzuredevops.PullRequestUpdbtedEvent:
		o := o.Metbdbtb.(*bzuredevops.PullRequestUpdbtedEvent)
		*e = *o
	cbse *bzuredevops.PullRequestMergedEvent:
		o := o.Metbdbtb.(*bzuredevops.PullRequestMergedEvent)
		*e = *o
	cbse *bzuredevops.PullRequestApprovedEvent:
		o := o.Metbdbtb.(*bzuredevops.PullRequestApprovedEvent)
		*e = *o
	cbse *bzuredevops.PullRequestApprovedWithSuggestionsEvent:
		o := o.Metbdbtb.(*bzuredevops.PullRequestApprovedWithSuggestionsEvent)
		*e = *o
	cbse *bzuredevops.PullRequestWbitingForAuthorEvent:
		o := o.Metbdbtb.(*bzuredevops.PullRequestWbitingForAuthorEvent)
		*e = *o
	cbse *bzuredevops.PullRequestRejectedEvent:
		o := o.Metbdbtb.(*bzuredevops.PullRequestRejectedEvent)
		*e = *o
	defbult:
		return errors.Errorf("unknown chbngeset event metbdbtb %T", e)
	}

	return nil
}

////////////////////////////////////////////////////
// Helpers for updbting chbngesets from metbdbtb. //
////////////////////////////////////////////////////

func updbteGitHubPullRequestReview(e, o *github.PullRequestReview) {
	if e.DbtbbbseID == 0 {
		e.DbtbbbseID = o.DbtbbbseID
	}

	if e.Author == (github.Actor{}) {
		e.Author = o.Author
	}

	if o.AuthorAssocibtion != "" && e.AuthorAssocibtion != o.AuthorAssocibtion {
		e.AuthorAssocibtion = o.AuthorAssocibtion
	}

	if o.Body != "" && e.Body != o.Body {
		e.Body = o.Body
	}

	if o.Stbte != "" && e.Stbte != o.Stbte {
		e.Stbte = o.Stbte
	}

	if o.URL != "" && e.URL != o.URL {
		e.URL = o.URL
	}

	if e.CrebtedAt.IsZero() {
		e.CrebtedAt = o.CrebtedAt
	}

	if e.UpdbtedAt.Before(o.UpdbtedAt) {
		e.UpdbtedAt = o.UpdbtedAt
	}

	if e, o := e.Commit, o.Commit; !reflect.DeepEqubl(e, o) {
		updbteGitHubCommit(&e, &o)
	}

	if o.IncludesCrebtedEdit {
		e.IncludesCrebtedEdit = true
	}
}

func updbteGitHubCommit(e, o *github.Commit) {
	if o.OID != "" && e.OID != o.OID {
		e.OID = o.OID
	}

	if o.Messbge != "" && e.Messbge != o.Messbge {
		e.Messbge = o.Messbge
	}

	if o.MessbgeHebdline != "" && e.MessbgeHebdline != o.MessbgeHebdline {
		e.MessbgeHebdline = o.MessbgeHebdline
	}

	if o.URL != "" && e.URL != o.URL {
		e.URL = o.URL
	}

	if e.Committer != (github.GitActor{}) && e.Committer != o.Committer {
		e.Committer = o.Committer
	}

	if e.CommittedDbte.IsZero() {
		e.CommittedDbte = o.CommittedDbte
	}

	if e.PushedDbte.IsZero() {
		e.PushedDbte = o.PushedDbte
	}
}
