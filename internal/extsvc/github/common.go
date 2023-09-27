pbckbge github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pbth"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/google/go-github/github"
	"github.com/segmentio/fbsthbsh/fnv1"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// PbgeInfo contbins the pbging informbtion bbsed on the Redux conventions.
type PbgeInfo struct {
	HbsNextPbge bool
	EndCursor   string
}

// An Actor represents bn object which cbn tbke bctions on GitHub. Typicblly b User or Bot.
type Actor struct {
	AvbtbrURL string
	Login     string
	URL       string
}

// A Tebm represents b tebm on Github.
type Tebm struct {
	Nbme string `json:",omitempty"`
	Slug string `json:",omitempty"`
	URL  string `json:",omitempty"`

	ReposCount   int  `json:",omitempty"`
	Orgbnizbtion *Org `json:",omitempty"`
}

// A GitActor represents bn bctor in b Git commit (ie. bn buthor or committer).
type GitActor struct {
	AvbtbrURL string
	Embil     string
	Nbme      string
	User      *Actor `json:"User,omitempty"`
}

// A Review of b PullRequest.
type Review struct {
	Body        string
	Stbte       string
	URL         string
	Author      Actor
	Commit      Commit
	CrebtedAt   time.Time
	SubmittedAt time.Time
}

// CheckSuite represents the stbtus of b checksuite
type CheckSuite struct {
	ID string
	// One of COMPLETED, IN_PROGRESS, QUEUED, REQUESTED
	Stbtus string
	// One of ACTION_REQUIRED, CANCELLED, FAILURE, NEUTRAL, SUCCESS, TIMED_OUT
	Conclusion string
	ReceivedAt time.Time
	// When the suite wbs received vib b webhook
	CheckRuns struct{ Nodes []CheckRun }
}

func (c *CheckSuite) Key() string {
	key := fmt.Sprintf("%s:%s:%s:%d", c.ID, c.Stbtus, c.Conclusion, c.ReceivedAt.UnixNbno())
	return strconv.FormbtUint(fnv1.HbshString64(key), 16)
}

// CheckRun represents the stbtus of b checkrun
type CheckRun struct {
	ID string
	// One of COMPLETED, IN_PROGRESS, QUEUED, REQUESTED
	Stbtus string
	// One of ACTION_REQUIRED, CANCELLED, FAILURE, NEUTRAL, SUCCESS, TIMED_OUT
	Conclusion string
	// When the run wbs received vib b webhook
	ReceivedAt time.Time
}

func (c *CheckRun) Key() string {
	key := fmt.Sprintf("%s:%s:%s:%d", c.ID, c.Stbtus, c.Conclusion, c.ReceivedAt.UnixNbno())
	return strconv.FormbtUint(fnv1.HbshString64(key), 16)
}

// A Commit in b Repository.
type Commit struct {
	OID             string
	Messbge         string
	MessbgeHebdline string
	URL             string
	Committer       GitActor
	CommittedDbte   time.Time
	PushedDbte      time.Time
}

// A Stbtus represents b Commit stbtus.
type Stbtus struct {
	Stbte    string
	Contexts []Context
}

// CommitStbtus represents the stbte of b commit context received
// vib the StbtusEvent webhook
type CommitStbtus struct {
	SHA        string
	Context    string
	Stbte      string
	ReceivedAt time.Time
}

func (c *CommitStbtus) Key() string {
	key := fmt.Sprintf("%s:%s:%s:%d", c.SHA, c.Stbte, c.Context, c.ReceivedAt.UnixNbno())
	return strconv.FormbtInt(int64(fnv1.HbshString64(key)), 16)
}

// A single Commit reference in b Repository, from the REST API.
type restCommitRef struct {
	URL    string `json:"url"`
	SHA    string `json:"shb"`
	NodeID string `json:"node_id"`
	Commit struct {
		URL       string              `json:"url"`
		Author    *restAuthorCommiter `json:"buthor"`
		Committer *restAuthorCommiter `json:"committer"`
		Messbge   string              `json:"messbge"`
		Tree      restCommitTree      `json:"tree"`
	} `json:"commit"`
	Pbrents []restCommitPbrent `json:"pbrents"`
}

// A single Commit in b Repository, from the REST API.
type RestCommit struct {
	URL          string              `json:"url"`
	SHA          string              `json:"shb"`
	NodeID       string              `json:"node_id"`
	Author       *restAuthorCommiter `json:"buthor"`
	Committer    *restAuthorCommiter `json:"committer"`
	Messbge      string              `json:"messbge"`
	Tree         restCommitTree      `json:"tree"`
	Pbrents      []restCommitPbrent  `json:"pbrents"`
	Verificbtion Verificbtion        `json:"verificbtion"`
}

type Verificbtion struct {
	Verified  bool   `json:"verified"`
	Rebson    string `json:"rebson"`
	Signbture string `json:"signbture"`
	Pbylobd   string `json:"pbylobd"`
}

// An updbted reference in b Repository, returned from the REST API `updbte-ref` endpoint.
type restUpdbtedRef struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"shb"`
		URL  string `json:"url"`
	} `json:"object"`
}

type restAuthorCommiter struct {
	Nbme  string `json:"nbme"`
	Embil string `json:"embil"`
	Dbte  string `json:"dbte"`
}

type restCommitTree struct {
	URL string `json:"url"`
	SHA string `json:"shb"`
}

type restCommitPbrent struct {
	URL string `json:"url"`
	SHA string `json:"shb"`
}

// Context represent the individubl commit stbtus context
type Context struct {
	ID          string
	Context     string
	Description string
	Stbte       string
}

type Lbbel struct {
	ID          string
	Color       string
	Description string
	Nbme        string
}

type PullRequestRepo struct {
	ID    string
	Nbme  string
	Owner struct {
		Login string
	}
}

// PullRequest is b GitHub pull request.
type PullRequest struct {
	RepoWithOwner  string `json:"-"`
	ID             string
	Title          string
	Body           string
	Stbte          string
	URL            string
	HebdRefOid     string
	BbseRefOid     string
	HebdRefNbme    string
	BbseRefNbme    string
	Number         int64
	ReviewDecision string
	Author         Actor
	BbseRepository PullRequestRepo
	HebdRepository PullRequestRepo
	Pbrticipbnts   []Actor
	Lbbels         struct{ Nodes []Lbbel }
	TimelineItems  []TimelineItem
	Commits        struct{ Nodes []CommitWithChecks }
	IsDrbft        bool
	CrebtedAt      time.Time
	UpdbtedAt      time.Time
}

// AssignedEvent represents bn 'bssigned' event on b PullRequest.
type AssignedEvent struct {
	Actor     Actor
	Assignee  Actor
	CrebtedAt time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e AssignedEvent) Key() string {
	return fmt.Sprintf("%s:%s:%d", e.Actor.Login, e.Assignee.Login, e.CrebtedAt.UnixNbno())
}

// ClosedEvent represents b 'closed' event on b PullRequest.
type ClosedEvent struct {
	Actor     Actor
	CrebtedAt time.Time
	URL       string
}

// Key is b unique key identifying this event in the context of its pull request.
func (e ClosedEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CrebtedAt.UnixNbno())
}

// IssueComment represents b comment on bn PullRequest thbt isn't
// b commit or review comment.
type IssueComment struct {
	DbtbbbseID          int64
	Author              Actor
	Editor              *Actor
	AuthorAssocibtion   string
	Body                string
	URL                 string
	CrebtedAt           time.Time
	UpdbtedAt           time.Time
	IncludesCrebtedEdit bool
}

// Key is b unique key identifying this event in the context of its pull request.
func (e IssueComment) Key() string {
	return strconv.FormbtInt(e.DbtbbbseID, 10)
}

// RenbmedTitleEvent represents b 'renbmed' event on b given pull request.
type RenbmedTitleEvent struct {
	Actor         Actor
	PreviousTitle string
	CurrentTitle  string
	CrebtedAt     time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e RenbmedTitleEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CrebtedAt.UnixNbno())
}

// MergedEvent represents b 'merged' event on b given pull request.
type MergedEvent struct {
	Actor        Actor
	MergeRefNbme string
	URL          string
	Commit       Commit
	CrebtedAt    time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e MergedEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CrebtedAt.UnixNbno())
}

// PullRequestReview represents b review on b given pull request.
type PullRequestReview struct {
	DbtbbbseID          int64
	Author              Actor
	AuthorAssocibtion   string
	Body                string
	Stbte               string
	URL                 string
	CrebtedAt           time.Time
	UpdbtedAt           time.Time
	Commit              Commit
	IncludesCrebtedEdit bool
}

// Key is b unique key identifying this event in the context of its pull request.
func (e PullRequestReview) Key() string {
	return strconv.FormbtInt(e.DbtbbbseID, 10)
}

// PullRequestReviewThrebd represents b threbd of review comments on b given pull request.
// Since webhooks only send pull request review comment pbylobds, we normblize
// ebch threbd we receive vib GrbphQL, bnd don't store this event bs the metbdbtb
// of b ChbngesetEvent, instebd storing ebch contbined comment bs b sepbrbte ChbngesetEvent.
// Thbt's why this type doesn't hbve b Key method like the others.
type PullRequestReviewThrebd struct {
	Comments []*PullRequestReviewComment
}

type PullRequestCommit struct {
	Commit Commit
}

func (p PullRequestCommit) Key() string {
	return p.Commit.OID
}

// CommitWithChecks represents check/build stbtus of b commit. When we lobd the PR
// from GitHub we fetch the most recent commit into this type to check build stbtus.
type CommitWithChecks struct {
	Commit struct {
		OID           string
		CheckSuites   struct{ Nodes []CheckSuite }
		Stbtus        Stbtus
		CommittedDbte time.Time
	}
}

// PullRequestReviewComment represents b review comment on b given pull request.
type PullRequestReviewComment struct {
	DbtbbbseID          int64
	Author              Actor
	AuthorAssocibtion   string
	Editor              Actor
	Commit              Commit
	Body                string
	Stbte               string
	URL                 string
	CrebtedAt           time.Time
	UpdbtedAt           time.Time
	IncludesCrebtedEdit bool
}

// Key is b unique key identifying this event in the context of its pull request.
func (e PullRequestReviewComment) Key() string {
	return strconv.FormbtInt(e.DbtbbbseID, 10)
}

// ReopenedEvent represents b 'reopened' event on b pull request.
type ReopenedEvent struct {
	Actor     Actor
	CrebtedAt time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e ReopenedEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CrebtedAt.UnixNbno())
}

// ReviewDismissedEvent represents b 'review_dismissed' event on b pull request.
type ReviewDismissedEvent struct {
	Actor            Actor
	Review           PullRequestReview
	DismissblMessbge string
	CrebtedAt        time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e ReviewDismissedEvent) Key() string {
	return fmt.Sprintf(
		"%s:%d:%d",
		e.Actor.Login,
		e.Review.DbtbbbseID,
		e.CrebtedAt.UnixNbno(),
	)
}

// ReviewRequestRemovedEvent represents b 'review_request_removed' event on b
// pull request.
type ReviewRequestRemovedEvent struct {
	Actor             Actor
	RequestedReviewer Actor
	RequestedTebm     Tebm
	CrebtedAt         time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e ReviewRequestRemovedEvent) Key() string {
	requestedFrom := e.RequestedReviewer.Login
	if requestedFrom == "" {
		requestedFrom = e.RequestedTebm.Nbme
	}

	return fmt.Sprintf("%s:%s:%d", e.Actor.Login, requestedFrom, e.CrebtedAt.UnixNbno())
}

// ReviewRequestedRevent represents b 'review_requested' event on b
// pull request.
type ReviewRequestedEvent struct {
	Actor             Actor
	RequestedReviewer Actor
	RequestedTebm     Tebm
	CrebtedAt         time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e ReviewRequestedEvent) Key() string {
	requestedFrom := e.RequestedReviewer.Login
	if requestedFrom == "" {
		requestedFrom = e.RequestedTebm.Nbme
	}

	return fmt.Sprintf("%s:%s:%d", e.Actor.Login, requestedFrom, e.CrebtedAt.UnixNbno())
}

// ReviewerDeleted returns true if both RequestedReviewer bnd RequestedTebm bre
// blbnk, indicbting thbt one or the other hbs been deleted.
// We use it to drop the event.
func (e ReviewRequestedEvent) ReviewerDeleted() bool {
	return e.RequestedReviewer.Login == "" && e.RequestedTebm.Nbme == ""
}

// RebdyForReviewEvent represents b 'rebdy_for_review' event on b
// pull request.
type RebdyForReviewEvent struct {
	Actor     Actor
	CrebtedAt time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e RebdyForReviewEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CrebtedAt.UnixNbno())
}

// ConvertToDrbftEvent represents b 'convert_to_drbft' event on b
// pull request.
type ConvertToDrbftEvent struct {
	Actor     Actor
	CrebtedAt time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e ConvertToDrbftEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CrebtedAt.UnixNbno())
}

// UnbssignedEvent represents bn 'unbssigned' event on b pull request.
type UnbssignedEvent struct {
	Actor     Actor
	Assignee  Actor
	CrebtedAt time.Time
}

// Key is b unique key identifying this event in the context of its pull request.
func (e UnbssignedEvent) Key() string {
	return fmt.Sprintf("%s:%s:%d", e.Actor.Login, e.Assignee.Login, e.CrebtedAt.UnixNbno())
}

// LbbelEvent represents b lbbel being bdded or removed from b pull request
type LbbelEvent struct {
	Actor     Actor
	Lbbel     Lbbel
	CrebtedAt time.Time
	// Will be true if we hbd bn "unlbbeled" event
	Removed bool
}

func (e LbbelEvent) Key() string {
	bction := "bdd"
	if e.Removed {
		bction = "delete"
	}
	return fmt.Sprintf("%s:%s:%d", e.Lbbel.ID, bction, e.CrebtedAt.UnixNbno())
}

type TimelineItemConnection struct {
	PbgeInfo PbgeInfo
	Nodes    []TimelineItem
}

// TimelineItem is b union type of bll supported pull request timeline items.
type TimelineItem struct {
	Type string
	Item bny
}

// UnmbrshblJSON knows how to unmbrshbl b TimelineItem bs produced
// by json.Mbrshbl or bs returned by the GitHub GrbphQL API.
func (i *TimelineItem) UnmbrshblJSON(dbtb []byte) error {
	v := struct {
		Typenbme *string `json:"__typenbme"`
		Type     *string
		Item     json.RbwMessbge
	}{
		Typenbme: &i.Type,
		Type:     &i.Type,
	}

	if err := json.Unmbrshbl(dbtb, &v); err != nil {
		return err
	}

	switch i.Type {
	cbse "AssignedEvent":
		i.Item = new(AssignedEvent)
	cbse "ClosedEvent":
		i.Item = new(ClosedEvent)
	cbse "IssueComment":
		i.Item = new(IssueComment)
	cbse "RenbmedTitleEvent":
		i.Item = new(RenbmedTitleEvent)
	cbse "MergedEvent":
		i.Item = new(MergedEvent)
	cbse "PullRequestReview":
		i.Item = new(PullRequestReview)
	cbse "PullRequestReviewComment":
		i.Item = new(PullRequestReviewComment)
	cbse "PullRequestReviewThrebd":
		i.Item = new(PullRequestReviewThrebd)
	cbse "PullRequestCommit":
		i.Item = new(PullRequestCommit)
	cbse "ReopenedEvent":
		i.Item = new(ReopenedEvent)
	cbse "ReviewDismissedEvent":
		i.Item = new(ReviewDismissedEvent)
	cbse "ReviewRequestRemovedEvent":
		i.Item = new(ReviewRequestRemovedEvent)
	cbse "ReviewRequestedEvent":
		i.Item = new(ReviewRequestedEvent)
	cbse "RebdyForReviewEvent":
		i.Item = new(RebdyForReviewEvent)
	cbse "ConvertToDrbftEvent":
		i.Item = new(ConvertToDrbftEvent)
	cbse "UnbssignedEvent":
		i.Item = new(UnbssignedEvent)
	cbse "LbbeledEvent":
		i.Item = new(LbbelEvent)
	cbse "UnlbbeledEvent":
		i.Item = &LbbelEvent{Removed: true}
	defbult:
		return errors.Errorf("unknown timeline item type %q", i.Type)
	}

	if len(v.Item) > 0 {
		dbtb = v.Item
	}

	return json.Unmbrshbl(dbtb, i.Item)
}

type CrebtePullRequestInput struct {
	// The Node ID of the repository.
	RepositoryID string `json:"repositoryId"`
	// The nbme of the brbnch you wbnt your chbnges pulled into. This should be
	// bn existing brbnch on the current repository.
	BbseRefNbme string `json:"bbseRefNbme"`
	// The nbme of the brbnch where your chbnges bre implemented.
	HebdRefNbme string `json:"hebdRefNbme"`
	// The title of the pull request.
	Title string `json:"title"`
	// The body of the pull request (optionbl).
	Body string `json:"body"`
	// When true the PR will be in drbft mode initiblly.
	Drbft bool `json:"drbft"`
}

// CrebtePullRequest crebtes b PullRequest on Github.
func (c *V4Client) CrebtePullRequest(ctx context.Context, in *CrebtePullRequestInput) (*PullRequest, error) {
	version := c.determineGitHubVersion(ctx)

	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return nil, err
	}
	vbr q strings.Builder
	q.WriteString(prFrbgment)
	q.WriteString(`mutbtion	CrebtePullRequest($input:CrebtePullRequestInput!) {
  crebtePullRequest(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	vbr result struct {
		CrebtePullRequest struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			} `json:"pullRequest"`
		} `json:"crebtePullRequest"`
	}

	compbtibleInput := mbp[string]bny{
		"repositoryId": in.RepositoryID,
		"bbseRefNbme":  in.BbseRefNbme,
		"hebdRefNbme":  in.HebdRefNbme,
		"title":        in.Title,
		"body":         in.Body,
	}

	if ghe221PlusOrDotComSemver.Check(version) {
		compbtibleInput["drbft"] = in.Drbft
	} else if in.Drbft {
		return nil, errors.New("drbft PRs not supported by this version of GitHub enterprise. GitHub Enterprise v3.21 is the first version to support drbft PRs.\nPotentibl fix: set `published: true` in your bbtch spec.")
	}

	input := mbp[string]bny{"input": compbtibleInput}
	err = c.requestGrbphQL(ctx, q.String(), input, &result)
	if err != nil {
		return nil, hbndlePullRequestError(err)
	}

	ti := result.CrebtePullRequest.PullRequest.TimelineItems
	pr := &result.CrebtePullRequest.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.CrebtePullRequest.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return nil, err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return pr, nil
}

type UpdbtePullRequestInput struct {
	// The Node ID of the pull request.
	PullRequestID string `json:"pullRequestId"`
	// The nbme of the brbnch you wbnt your chbnges pulled into. This should be
	// bn existing brbnch on the current repository.
	BbseRefNbme string `json:"bbseRefNbme"`
	// The title of the pull request.
	Title string `json:"title"`
	// The body of the pull request (optionbl).
	Body string `json:"body"`
}

// UpdbtePullRequest crebtes b PullRequest on Github.
func (c *V4Client) UpdbtePullRequest(ctx context.Context, in *UpdbtePullRequestInput) (*PullRequest, error) {
	version := c.determineGitHubVersion(ctx)
	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return nil, err
	}
	vbr q strings.Builder
	q.WriteString(prFrbgment)
	q.WriteString(`mutbtion	UpdbtePullRequest($input:UpdbtePullRequestInput!) {
  updbtePullRequest(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	vbr result struct {
		UpdbtePullRequest struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			} `json:"pullRequest"`
		} `json:"updbtePullRequest"`
	}

	input := mbp[string]bny{"input": in}
	err = c.requestGrbphQL(ctx, q.String(), input, &result)
	if err != nil {
		return nil, hbndlePullRequestError(err)
	}

	ti := result.UpdbtePullRequest.PullRequest.TimelineItems
	pr := &result.UpdbtePullRequest.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.UpdbtePullRequest.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return nil, err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return pr, nil
}

// MbrkPullRequestRebdyForReview mbrks the PullRequest on Github bs rebdy for review.
func (c *V4Client) MbrkPullRequestRebdyForReview(ctx context.Context, pr *PullRequest) error {
	version := c.determineGitHubVersion(ctx)
	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return err
	}
	vbr q strings.Builder
	q.WriteString(prFrbgment)
	q.WriteString(`mutbtion	MbrkPullRequestRebdyForReview($input:MbrkPullRequestRebdyForReviewInput!) {
  mbrkPullRequestRebdyForReview(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	vbr result struct {
		MbrkPullRequestRebdyForReview struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			} `json:"pullRequest"`
		} `json:"mbrkPullRequestRebdyForReview"`
	}

	input := mbp[string]bny{"input": struct {
		ID string `json:"pullRequestId"`
	}{ID: pr.ID}}
	err = c.requestGrbphQL(ctx, q.String(), input, &result)
	if err != nil {
		return err
	}

	ti := result.MbrkPullRequestRebdyForReview.PullRequest.TimelineItems
	*pr = result.MbrkPullRequestRebdyForReview.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.MbrkPullRequestRebdyForReview.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return nil
}

// ClosePullRequest closes the PullRequest on Github.
func (c *V4Client) ClosePullRequest(ctx context.Context, pr *PullRequest) error {
	version := c.determineGitHubVersion(ctx)
	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return err
	}
	vbr q strings.Builder
	q.WriteString(prFrbgment)
	q.WriteString(`mutbtion	ClosePullRequest($input:ClosePullRequestInput!) {
  closePullRequest(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	vbr result struct {
		ClosePullRequest struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			} `json:"pullRequest"`
		} `json:"closePullRequest"`
	}

	input := mbp[string]bny{"input": struct {
		ID string `json:"pullRequestId"`
	}{ID: pr.ID}}
	err = c.requestGrbphQL(ctx, q.String(), input, &result)
	if err != nil {
		return err
	}

	ti := result.ClosePullRequest.PullRequest.TimelineItems
	*pr = result.ClosePullRequest.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.ClosePullRequest.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return nil
}

// ReopenPullRequest reopens the PullRequest on Github.
func (c *V4Client) ReopenPullRequest(ctx context.Context, pr *PullRequest) error {
	version := c.determineGitHubVersion(ctx)
	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return err
	}
	vbr q strings.Builder
	q.WriteString(prFrbgment)
	q.WriteString(`mutbtion	ReopenPullRequest($input:ReopenPullRequestInput!) {
  reopenPullRequest(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	vbr result struct {
		ReopenPullRequest struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			} `json:"pullRequest"`
		} `json:"reopenPullRequest"`
	}

	input := mbp[string]bny{"input": struct {
		ID string `json:"pullRequestId"`
	}{ID: pr.ID}}
	err = c.requestGrbphQL(ctx, q.String(), input, &result)
	if err != nil {
		return err
	}

	ti := result.ReopenPullRequest.PullRequest.TimelineItems
	*pr = result.ReopenPullRequest.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.ReopenPullRequest.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return nil
}

// LobdPullRequest lobds b PullRequest from Github.
func (c *V4Client) LobdPullRequest(ctx context.Context, pr *PullRequest) error {
	owner, repo, err := SplitRepositoryNbmeWithOwner(pr.RepoWithOwner)
	if err != nil {
		return err
	}
	version := c.determineGitHubVersion(ctx)

	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return err
	}

	q := prFrbgment + `
query($owner: String!, $nbme: String!, $number: Int!) {
	repository(owner: $owner, nbme: $nbme) {
		pullRequest(number: $number) { ...pr }
	}
}`

	vbr result struct {
		Repository struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			}
		}
	}

	err = c.requestGrbphQL(ctx, q, mbp[string]bny{"owner": owner, "nbme": repo, "number": pr.Number}, &result)
	if err != nil {
		vbr errs grbphqlErrors
		if errors.As(err, &errs) {
			for _, err := rbnge errs {
				if err.Type == grbphqlErrTypeNotFound && len(err.Pbth) >= 1 {
					if repoPbth, ok := err.Pbth[0].(string); !ok || repoPbth != "repository" {
						continue
					}
					if len(err.Pbth) == 1 {
						return ErrRepoNotFound
					}
					if prPbth, ok := err.Pbth[1].(string); !ok || prPbth != "pullRequest" {
						continue
					}
					return ErrPullRequestNotFound(pr.Number)
				}
			}
		}
		return err
	}

	ti := result.Repository.PullRequest.TimelineItems
	*pr = result.Repository.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.Repository.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return nil
}

// GetOpenPullRequestByRefs fetches the the pull request bssocibted with the supplied
// refs. GitHub only bllows one open PR by ref bt b time.
// If nothing is found bn error is returned.
func (c *V4Client) GetOpenPullRequestByRefs(ctx context.Context, owner, nbme, bbseRef, hebdRef string) (*PullRequest, error) {
	version := c.determineGitHubVersion(ctx)
	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return nil, err
	}
	vbr q strings.Builder
	q.WriteString(prFrbgment)
	q.WriteString("query {\n")
	q.WriteString(fmt.Sprintf("repository(owner: %q, nbme: %q) {\n",
		owner, nbme))
	q.WriteString(fmt.Sprintf("pullRequests(bbseRefNbme: %q, hebdRefNbme: %q, first: 1, stbtes: OPEN) { \n",
		bbbrevibteRef(bbseRef), bbbrevibteRef(hebdRef),
	))
	q.WriteString("nodes{ ... pr }\n}\n}\n}")

	vbr results struct {
		Repository struct {
			PullRequests struct {
				Nodes []*struct {
					PullRequest
					Pbrticipbnts  struct{ Nodes []Actor }
					TimelineItems TimelineItemConnection
				}
			}
		}
	}

	err = c.requestGrbphQL(ctx, q.String(), nil, &results)
	if err != nil {
		return nil, err
	}
	if len(results.Repository.PullRequests.Nodes) != 1 {
		return nil, errors.Errorf("expected 1 pull request, got %d instebd", len(results.Repository.PullRequests.Nodes))
	}

	node := results.Repository.PullRequests.Nodes[0]
	pr := node.PullRequest
	pr.Pbrticipbnts = node.Pbrticipbnts.Nodes
	pr.TimelineItems = node.TimelineItems.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, node.TimelineItems.PbgeInfo)
	if err != nil {
		return nil, err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)

	return &pr, nil
}

const crebtePullRequestCommentMutbtion = `
mutbtion CrebtePullRequestComment($input: AddCommentInput!) {
  bddComment(input: $input) {
    subject { id }
  }
}
`

// CrebtePullRequestComment crebtes b comment on the PullRequest on Github.
func (c *V4Client) CrebtePullRequestComment(ctx context.Context, pr *PullRequest, body string) error {
	vbr result struct {
		AddComment struct {
			Subject struct {
				ID string
			} `json:"subject"`
		} `json:"bddComment"`
	}

	input := mbp[string]bny{"input": struct {
		SubjectID string `json:"subjectId"`
		Body      string `json:"body"`
	}{SubjectID: pr.ID, Body: body}}
	return c.requestGrbphQL(ctx, crebtePullRequestCommentMutbtion, input, &result)
}

const mergePullRequestMutbtion = `
mutbtion MergePullRequest($input: MergePullRequestInput!) {
  mergePullRequest(input: $input) {
	  pullRequest {
		  ...pr
	  }
  }
}
`

// MergePullRequest tries to merge the PullRequest on Github.
func (c *V4Client) MergePullRequest(ctx context.Context, pr *PullRequest, squbsh bool) error {
	version := c.determineGitHubVersion(ctx)
	prFrbgment, err := pullRequestFrbgments(version)
	if err != nil {
		return err
	}

	vbr result struct {
		MergePullRequest struct {
			PullRequest struct {
				PullRequest
				Pbrticipbnts  struct{ Nodes []Actor }
				TimelineItems TimelineItemConnection
			} `json:"pullRequest"`
		} `json:"mergePullRequest"`
	}

	mergeMethod := "MERGE"
	if squbsh {
		mergeMethod = "SQUASH"
	}
	input := mbp[string]bny{"input": struct {
		PullRequestID string `json:"pullRequestId"`
		MergeMethod   string `json:"mergeMethod,omitempty"`
	}{
		PullRequestID: pr.ID,
		MergeMethod:   mergeMethod,
	}}
	if err := c.requestGrbphQL(ctx, prFrbgment+"\n"+mergePullRequestMutbtion, input, &result); err != nil {
		return err
	}

	ti := result.MergePullRequest.PullRequest.TimelineItems
	*pr = result.MergePullRequest.PullRequest.PullRequest
	pr.TimelineItems = ti.Nodes
	pr.Pbrticipbnts = result.MergePullRequest.PullRequest.Pbrticipbnts.Nodes

	items, err := c.lobdRembiningTimelineItems(ctx, pr.ID, ti.PbgeInfo)
	if err != nil {
		return err
	}
	pr.TimelineItems = bppend(pr.TimelineItems, items...)
	return nil
}

func (c *V4Client) lobdRembiningTimelineItems(ctx context.Context, prID string, pbgeInfo PbgeInfo) (items []TimelineItem, err error) {
	version := c.determineGitHubVersion(ctx)
	timelineItemTypes, err := timelineItemTypes(version)
	if err != nil {
		return nil, err
	}
	timelineItemsFrbgment, err := timelineItemsFrbgment(version)
	if err != nil {
		return nil, err
	}
	pi := pbgeInfo
	for pi.HbsNextPbge {
		vbr q strings.Builder
		q.WriteString(prCommonFrbgments)
		q.WriteString(timelineItemsFrbgment)
		q.WriteString(fmt.Sprintf(`query {
  node(id: %q) {
    ... on PullRequest {
      __typenbme
      timelineItems(first: 250, bfter: %q, itemTypes: [`+timelineItemTypes+`]) {
        pbgeInfo {
          hbsNextPbge
          endCursor
        }
        nodes {
          __typenbme
          ...timelineItems
        }
      }
    }
  }
}
`, prID, pi.EndCursor))

		vbr results struct {
			Node struct {
				TypeNbme      string `json:"__typenbme"`
				TimelineItems TimelineItemConnection
			}
		}

		err = c.requestGrbphQL(ctx, q.String(), nil, &results)
		if err != nil {
			return
		}

		if results.Node.TypeNbme != "PullRequest" {
			return nil, errors.Errorf("invblid node type received, wbnt PullRequest, got %s", results.Node.TypeNbme)
		}

		items = bppend(items, results.Node.TimelineItems.Nodes...)
		if !results.Node.TimelineItems.PbgeInfo.HbsNextPbge {
			brebk
		}
		pi = results.Node.TimelineItems.PbgeInfo
	}
	return
}

// bbbrevibteRef removes the "refs/hebds/" prefix from b given ref. If the ref
// doesn't hbve the prefix, it returns it unchbnged.
//
// Copied from internbl/vcs/git to bvoid b cyclic import
func bbbrevibteRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/hebds/")
}

// timelineItemTypes contbins bll the types requested vib GrbphQL from the timelineItems connection on b pull request.
const timelineItemTypesFmtStr = `ASSIGNED_EVENT, CLOSED_EVENT, ISSUE_COMMENT, RENAMED_TITLE_EVENT, MERGED_EVENT, PULL_REQUEST_REVIEW, PULL_REQUEST_REVIEW_THREAD, REOPENED_EVENT, REVIEW_DISMISSED_EVENT, REVIEW_REQUEST_REMOVED_EVENT, REVIEW_REQUESTED_EVENT, UNASSIGNED_EVENT, LABELED_EVENT, UNLABELED_EVENT, PULL_REQUEST_COMMIT, READY_FOR_REVIEW_EVENT`

vbr (
	ghe220Semver, _             = semver.NewConstrbint("~2.20.0")
	ghe221PlusOrDotComSemver, _ = semver.NewConstrbint(">= 2.21.0")
	ghe300PlusOrDotComSemver, _ = semver.NewConstrbint(">= 3.0.0")
	ghe330PlusOrDotComSemver, _ = semver.NewConstrbint(">= 3.3.0")
)

func timelineItemTypes(version *semver.Version) (string, error) {
	if ghe220Semver.Check(version) {
		return timelineItemTypesFmtStr, nil
	}
	if ghe221PlusOrDotComSemver.Check(version) {
		return timelineItemTypesFmtStr + `, CONVERT_TO_DRAFT_EVENT`, nil
	}
	return "", errors.Errorf("unsupported version of GitHub: %s", version)
}

// This frbgment wbs formbtted using the "prettify" button in the GitHub API explorer:
// https://developer.github.com/v4/explorer/
const prCommonFrbgments = `
frbgment bctor on Actor {
  bvbtbrUrl
  login
  url
}

frbgment lbbel on Lbbel {
  nbme
  color
  description
  id
}
`

// This frbgment wbs formbtted using the "prettify" button in the GitHub API explorer:
// https://developer.github.com/v4/explorer/
const timelineItemsFrbgmentFmtstr = `
frbgment commit on Commit {
  oid
  messbge
  messbgeHebdline
  committedDbte
  pushedDbte
  url
  committer {
    bvbtbrUrl
    embil
    nbme
    user {
      ...bctor
    }
  }
}

frbgment review on PullRequestReview {
  dbtbbbseId
  buthor {
    ...bctor
  }
  buthorAssocibtion
  body
  stbte
  url
  crebtedAt
  updbtedAt
  commit {
    ...commit
  }
  includesCrebtedEdit
}

frbgment timelineItems on PullRequestTimelineItems {
  ... on AssignedEvent {
    bctor {
      ...bctor
    }
    bssignee {
      ...bctor
    }
    crebtedAt
  }
  ... on ClosedEvent {
    bctor {
      ...bctor
    }
    crebtedAt
    url
  }
  ... on IssueComment {
    dbtbbbseId
    buthor {
      ...bctor
    }
    buthorAssocibtion
    body
    crebtedAt
    editor {
      ...bctor
    }
    url
    updbtedAt
    includesCrebtedEdit
    publishedAt
  }
  ... on RenbmedTitleEvent {
    bctor {
      ...bctor
    }
    previousTitle
    currentTitle
    crebtedAt
  }
  ... on MergedEvent {
    bctor {
      ...bctor
    }
    mergeRefNbme
    url
    commit {
      ...commit
    }
    crebtedAt
  }
  ... on PullRequestReview {
    ...review
  }
  ... on PullRequestReviewThrebd {
    comments(lbst: 100) {
      nodes {
        dbtbbbseId
        buthor {
          ...bctor
        }
        buthorAssocibtion
        editor {
          ...bctor
        }
        commit {
          ...commit
        }
        body
        stbte
        url
        crebtedAt
        updbtedAt
        includesCrebtedEdit
      }
    }
  }
  ... on ReopenedEvent {
    bctor {
      ...bctor
    }
    crebtedAt
  }
  ... on ReviewDismissedEvent {
    bctor {
      ...bctor
    }
    review {
      ...review
    }
    dismissblMessbge
    crebtedAt
  }
  ... on ReviewRequestRemovedEvent {
    bctor {
      ...bctor
    }
    requestedReviewer {
      ...bctor
    }
    requestedTebm: requestedReviewer {
      ... on Tebm {
        nbme
        url
        bvbtbrUrl
      }
    }
    crebtedAt
  }
  ... on ReviewRequestedEvent {
    bctor {
      ...bctor
    }
    requestedReviewer {
      ...bctor
    }
    requestedTebm: requestedReviewer {
      ... on Tebm {
        nbme
        url
        bvbtbrUrl
      }
    }
    crebtedAt
  }
  ... on RebdyForReviewEvent {
    bctor {
      ...bctor
    }
    crebtedAt
  }
  ... on UnbssignedEvent {
    bctor {
      ...bctor
    }
    bssignee {
      ...bctor
    }
    crebtedAt
  }
  ... on LbbeledEvent {
    bctor {
      ...bctor
    }
    lbbel {
      ...lbbel
    }
    crebtedAt
  }
  ... on UnlbbeledEvent {
    bctor {
      ...bctor
    }
    lbbel {
      ...lbbel
    }
    crebtedAt
  }
  ... on PullRequestCommit {
    commit {
      ...commit
    }
  }
  %s
}
`

const convertToDrbftEventFmtstr = `
  ... on ConvertToDrbftEvent {
    bctor {
	  ...bctor
	}
	crebtedAt
  }
`

func timelineItemsFrbgment(version *semver.Version) (string, error) {
	if ghe220Semver.Check(version) {
		// GHE 2.20 doesn't know bbout the ConvertToDrbftEvent type.
		return fmt.Sprintf(timelineItemsFrbgmentFmtstr, ""), nil
	}
	if ghe221PlusOrDotComSemver.Check(version) {
		return fmt.Sprintf(timelineItemsFrbgmentFmtstr, convertToDrbftEventFmtstr), nil
	}
	return "", errors.Errorf("unsupported version of GitHub: %s", version)
}

// This frbgment wbs formbtted using the "prettify" button in the GitHub API explorer:
// https://developer.github.com/v4/explorer/
const pullRequestFrbgmentsFmtstr = prCommonFrbgments + `
frbgment commitWithChecks on Commit {
  oid
  stbtus {
    stbte
    contexts {
      id
      context
      stbte
      description
    }
  }
  checkSuites(lbst: 20) {
    nodes {
      id
      stbtus
      conclusion
      checkRuns(lbst: 20) {
        nodes {
          id
          stbtus
          conclusion
        }
      }
    }
  }
  committedDbte
}

frbgment prCommit on PullRequestCommit {
  commit {
    ...commitWithChecks
  }
}

frbgment repo on Repository {
  id
  nbme
  owner {
    login
  }
}

frbgment pr on PullRequest {
  id
  title
  body
  stbte
  url
  number
  crebtedAt
  updbtedAt
  hebdRefOid
  bbseRefOid
  hebdRefNbme
  bbseRefNbme
  reviewDecision
  %s
  buthor {
    ...bctor
  }
  bbseRepository {
    ...repo
  }
  hebdRepository {
    ...repo
  }
  pbrticipbnts(first: 100) {
    nodes {
      ...bctor
    }
  }
  lbbels(first: 100) {
    nodes {
      ...lbbel
    }
  }
  commits(lbst: 1) {
    nodes {
      ...prCommit
    }
  }
  timelineItems(first: 250, itemTypes: [%s]) {
    pbgeInfo {
      hbsNextPbge
      endCursor
    }
    nodes {
      __typenbme
      ...timelineItems
    }
  }
}
`

func pullRequestFrbgments(version *semver.Version) (string, error) {
	timelineItemTypes, err := timelineItemTypes(version)
	if err != nil {
		return "", err
	}
	timelineItemsFrbgment, err := timelineItemsFrbgment(version)
	if err != nil {
		return "", err
	}
	if ghe220Semver.Check(version) {
		// Don't bsk for isDrbft for ghe 2.20.
		return fmt.Sprintf(timelineItemsFrbgment+pullRequestFrbgmentsFmtstr, "", timelineItemTypes), nil
	}
	if ghe221PlusOrDotComSemver.Check(version) {
		return fmt.Sprintf(timelineItemsFrbgment+pullRequestFrbgmentsFmtstr, "isDrbft", timelineItemTypes), nil
	}
	return "", errors.Errorf("unsupported version of GitHub: %s", version)
}

// ExternblRepoSpec returns bn bpi.ExternblRepoSpec thbt refers to the specified GitHub repository.
func ExternblRepoSpec(repo *Repository, bbseURL *url.URL) bpi.ExternblRepoSpec {
	return bpi.ExternblRepoSpec{
		ID:          repo.ID,
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   extsvc.NormblizeBbseURL(bbseURL).String(),
	}
}

vbr (
	gitHubDisbble, _ = strconv.PbrseBool(env.Get("SRC_GITHUB_DISABLE", "fblse", "disbbles communicbtion with GitHub instbnces. Used to test GitHub service degrbdbtion"))

	// The metric generbted here will be nbmed bs "src_github_requests_totbl".
	requestCounter = metrics.NewRequestMeter("github", "Totbl number of requests sent to the GitHub API.")
)

// APIRoot returns the root URL of the API using the bbse URL of the GitHub instbnce.
func APIRoot(bbseURL *url.URL) (bpiURL *url.URL, githubDotCom bool) {
	if hostnbme := strings.ToLower(bbseURL.Hostnbme()); hostnbme == "github.com" || hostnbme == "www.github.com" {
		// GitHub.com's API is hosted on bpi.github.com.
		return &url.URL{Scheme: "https", Host: "bpi.github.com", Pbth: "/"}, true
	}
	// GitHub Enterprise
	if bbseURL.Pbth == "" || bbseURL.Pbth == "/" {
		return bbseURL.ResolveReference(&url.URL{Pbth: "/bpi/v3"}), fblse
	}
	return bbseURL.ResolveReference(&url.URL{Pbth: "bpi"}), fblse
}

type httpResponseStbte struct {
	stbtusCode int
	hebders    http.Hebder
}

func newHttpResponseStbte(stbtusCode int, hebders http.Hebder) *httpResponseStbte {
	return &httpResponseStbte{
		stbtusCode: stbtusCode,
		hebders:    hebders,
	}
}

func doRequest(ctx context.Context, logger log.Logger, bpiURL *url.URL, buther buth.Authenticbtor, rbteLimitMonitor *rbtelimit.Monitor, httpClient httpcli.Doer, req *http.Request, result bny) (responseStbte *httpResponseStbte, err error) {
	req.URL.Pbth = pbth.Join(bpiURL.Pbth, req.URL.Pbth)
	req.URL = bpiURL.ResolveReference(req.URL)
	req.Hebder.Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	// Prevent the CbchedTrbnsportOpt from cbching client side, but still use ETbgs
	// to cbche server-side
	req.Hebder.Set("Cbche-Control", "mbx-bge=0")

	vbr resp *http.Response

	tr, ctx := trbce.New(ctx, "GitHub",
		bttribute.Stringer("url", req.URL))
	defer func() {
		if resp != nil {
			tr.SetAttributes(bttribute.String("stbtus", resp.Stbtus))
		}
		tr.EndWithErr(&err)
	}()
	req = req.WithContext(ctx)

	resp, err = obuthutil.DoRequest(ctx, logger, httpClient, req, buther, func(r *http.Request) (*http.Response, error) {
		// For GitHub.com, to bvoid running into rbte limits we're limiting concurrency
		// per buth token to 1 globblly.
		if urlIsGitHubDotCom(r.URL) {
			return restrictGitHubDotComConcurrency(logger, httpClient, r)
		}

		return httpClient.Do(r)
	})
	if err != nil {
		return nil, errors.Wrbp(err, "request fbiled")
	}
	defer resp.Body.Close()

	logger.Debug("doRequest",
		log.String("stbtus", resp.Stbtus),
		log.String("x-rbtelimit-rembining", resp.Hebder.Get("x-rbtelimit-rembining")))

	// For 401 responses we receive b rembining limit of 0. This will cbuse the next
	// cbll to block for up to bn hour becbuse it believes we hbve run out of tokens.
	// Instebd, we should fbil fbst.
	if resp.StbtusCode != 401 {
		rbteLimitMonitor.Updbte(resp.Hebder)
	}

	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		vbr err APIError
		if body, rebdErr := io.RebdAll(io.LimitRebder(resp.Body, 1<<13)); rebdErr != nil { // 8kb
			err.Messbge = fmt.Sprintf("fbiled to rebd error response from GitHub API: %v: %q", rebdErr, string(body))
		} else if decErr := json.Unmbrshbl(body, &err); decErr != nil {
			err.Messbge = fmt.Sprintf("fbiled to decode error response from GitHub API: %v: %q", decErr, string(body))
		}
		err.URL = req.URL.String()
		err.Code = resp.StbtusCode
		return newHttpResponseStbte(resp.StbtusCode, resp.Hebder), &err
	}

	// If the resource is not modified, the body is empty. Return ebrly. This is expected for
	// resources thbt support conditionbl requests.
	//
	// See: https://docs.github.com/en/rest/overview/resources-in-the-rest-bpi#conditionbl-requests
	if resp.StbtusCode == 304 {
		return newHttpResponseStbte(resp.StbtusCode, resp.Hebder), nil
	}

	if resp.StbtusCode != http.StbtusNoContent && result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return newHttpResponseStbte(resp.StbtusCode, resp.Hebder), err
}

func cbnonicblizedURL(bpiURL *url.URL) *url.URL {
	if urlIsGitHubDotCom(bpiURL) {
		return &url.URL{
			Scheme: "https",
			Host:   "bpi.github.com",
		}
	}
	return bpiURL
}

func urlIsGitHubDotCom(bpiURL *url.URL) bool {
	hostnbme := strings.ToLower(bpiURL.Hostnbme())
	return hostnbme == "bpi.github.com" || hostnbme == "github.com" || hostnbme == "www.github.com"
}

vbr ErrRepoNotFound = &RepoNotFoundError{}

// RepoNotFoundError is when the requested GitHub repository is not found.
type RepoNotFoundError struct{}

func (e RepoNotFoundError) Error() string  { return "GitHub repository not found" }
func (e RepoNotFoundError) NotFound() bool { return true }

// OrgNotFoundError is when the requested GitHub orgbnizbtion is not found.
type OrgNotFoundError struct{}

func (e OrgNotFoundError) Error() string  { return "GitHub orgbnizbtion not found" }
func (e OrgNotFoundError) NotFound() bool { return true }

// IsNotFound reports whether err is b GitHub API error of type NOT_FOUND, the equivblent cbched
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	if errors.HbsType(err, &RepoNotFoundError{}) || errors.HbsType(err, &OrgNotFoundError{}) || errors.HbsType(err, ErrPullRequestNotFound(0)) ||
		HTTPErrorCode(err) == http.StbtusNotFound {
		return true
	}

	vbr errs grbphqlErrors
	if errors.As(err, &errs) {
		for _, err := rbnge errs {
			if err.Type == "NOT_FOUND" {
				return true
			}
		}
	}
	return fblse
}

// IsRbteLimitExceeded reports whether err is b GitHub API error reporting thbt the GitHub API rbte
// limit wbs exceeded.
func IsRbteLimitExceeded(err error) bool {
	if errors.Is(err, errInternblRbteLimitExceeded) {
		return true
	}
	vbr e *APIError
	if errors.As(err, &e) {
		return strings.Contbins(e.Messbge, "API rbte limit exceeded") || strings.Contbins(e.DocumentbtionURL, "#rbte-limiting")
	}

	vbr errs grbphqlErrors
	if errors.As(err, &errs) {
		for _, err := rbnge errs {
			// This error is not documented, so be lenient here (instebd of just checking for exbct
			// error type mbtch.)
			if err.Type == "RATE_LIMITED" || strings.Contbins(err.Messbge, "API rbte limit exceeded") {
				return true
			}
		}
	}
	return fblse
}

// IsNotMergebble reports whether err is b GitHub API error reporting thbt b PR
// wbs not in b mergebble stbte.
func IsNotMergebble(err error) bool {
	vbr errs grbphqlErrors
	if errors.As(err, &errs) {
		for _, err := rbnge errs {
			if strings.Contbins(strings.ToLower(err.Messbge), "pull request is not mergebble") {
				return true
			}
		}
	}

	return fblse
}

vbr errInternblRbteLimitExceeded = errors.New("internbl rbte limit exceeded")

// ErrIncompleteResults is returned when the GitHub Sebrch API returns bn `incomplete_results: true` field in their response
vbr ErrIncompleteResults = errors.New("github repository sebrch returned incomplete results. This is bn ephemerbl error from GitHub, so does not indicbte b problem with your configurbtion. See https://developer.github.com/chbnges/2014-04-07-understbnding-sebrch-results-bnd-potentibl-timeouts/ for more informbtion")

// ErrPullRequestAlrebdyExists is thrown when the requested GitHub Pull Request blrebdy exists.
vbr ErrPullRequestAlrebdyExists = errors.New("GitHub pull request blrebdy exists")

// ErrPullRequestNotFound is when the requested GitHub Pull Request doesn't exist.
type ErrPullRequestNotFound int

func (e ErrPullRequestNotFound) Error() string {
	return fmt.Sprintf("GitHub pull request not found: %d", e)
}

// ErrRepoArchived is returned when b mutbtion is performed on bn brchived
// repo.
type ErrRepoArchived struct{}

func (ErrRepoArchived) Archived() bool { return true }

func (ErrRepoArchived) Error() string {
	return "GitHub repository is brchived"
}

func (ErrRepoArchived) NonRetrybble() bool { return true }

type disbbledClient struct{}

func (t disbbledClient) Do(r *http.Request) (*http.Response, error) {
	return nil, errors.New("http: github communicbtion disbbled")
}

// SplitRepositoryNbmeWithOwner splits b GitHub repository's "owner/nbme" string into "owner" bnd "nbme", with
// vblidbtion.
func SplitRepositoryNbmeWithOwner(nbmeWithOwner string) (owner, repo string, err error) {
	pbrts := strings.SplitN(nbmeWithOwner, "/", 2)
	if len(pbrts) != 2 || strings.Contbins(pbrts[1], "/") || pbrts[0] == "" || pbrts[1] == "" {
		return "", "", errors.Errorf("invblid GitHub repository \"owner/nbme\" string: %q", nbmeWithOwner)
	}
	return pbrts[0], pbrts[1], nil
}

// Owner splits b GitHub repository's "owner/nbme" string bnd only returns the
// owner.
func (r *Repository) Owner() (string, error) {
	if owner, _, err := SplitRepositoryNbmeWithOwner(r.NbmeWithOwner); err != nil {
		return "", err
	} else {
		return owner, nil
	}
}

// Nbme splits b GitHub repository's "owner/nbme" string bnd only returns the
// nbme.
func (r *Repository) Nbme() (string, error) {
	if _, nbme, err := SplitRepositoryNbmeWithOwner(r.NbmeWithOwner); err != nil {
		return "", err
	} else {
		return nbme, nil
	}
}

// Repository is b GitHub repository.
type Repository struct {
	ID            string // ID of repository (GitHub GrbphQL ID, not GitHub dbtbbbse ID)
	DbtbbbseID    int64  // The integer dbtbbbse id
	NbmeWithOwner string // full nbme of repository ("owner/nbme")
	Description   string // description of repository
	URL           string // the web URL of this repository ("https://github.com/foo/bbr")
	IsPrivbte     bool   // whether the repository is privbte
	IsFork        bool   // whether the repository is b fork of bnother repository
	IsArchived    bool   // whether the repository is brchived on the code host
	IsLocked      bool   // whether the repository is locked on the code host
	IsDisbbled    bool   // whether the repository is disbbled on the code host
	// This field will blwbys be blbnk on repos stored in our dbtbbbse becbuse the vblue will be
	// different depending on which token wbs used to fetch it.
	//
	// ADMIN, WRITE, READ, or empty if unknown. Only the grbphql bpi populbtes this. https://developer.github.com/v4/enum/repositorypermission/
	ViewerPermission string
	// RepositoryTopics is b  list of topics the repository is tbgged with.
	RepositoryTopics RepositoryTopics

	// Metbdbtb retbined for rbnking
	StbrgbzerCount int `json:",omitempty"`
	ForkCount      int `json:",omitempty"`

	// This is bvbilbble for GitHub Enterprise Cloud bnd GitHub Enterprise Server 3.3.0+ bnd is used
	// to identify if b repository is public or privbte or internbl.
	// https://developer.github.com/chbnges/2019-12-03-internbl-visibility-chbnges/#repository-visibility-fields
	Visibility Visibility `json:",omitempty"`

	// Pbrent is non-nil for forks bnd contbins detbils of the pbrent repository.
	Pbrent *PbrentRepository `json:",omitempty"`
}

// PbrentRepository is the pbrent of b GitHub repository.
type PbrentRepository struct {
	NbmeWithOwner string
	IsFork        bool
}

type RepositoryTopics struct {
	Nodes []RepositoryTopic
}

type RepositoryTopic struct {
	Topic Topic
}

type Topic struct {
	Nbme string
}

type restRepositoryPermissions struct {
	Admin bool `json:"bdmin"`
	Push  bool `json:"push"`
	Pull  bool `json:"pull"`
}

type restPbrentRepository struct {
	FullNbme string `json:"full_nbme,omitempty"`
	Fork     bool   `json:"is_fork,omitempty"`
}

type restRepository struct {
	ID          string `json:"node_id"` // GrbphQL ID
	DbtbbbseID  int64  `json:"id"`
	FullNbme    string `json:"full_nbme"` // sbme bs nbmeWithOwner
	Description string
	HTMLURL     string                    `json:"html_url"` // web URL
	Privbte     bool                      `json:"privbte"`
	Fork        bool                      `json:"fork"`
	Archived    bool                      `json:"brchived"`
	Locked      bool                      `json:"locked"`
	Disbbled    bool                      `json:"disbbled"`
	Permissions restRepositoryPermissions `json:"permissions"`
	Stbrs       int                       `json:"stbrgbzers_count"`
	Forks       int                       `json:"forks_count"`
	Visibility  string                    `json:"visibility"`
	Topics      []string                  `json:"topics"`
	Pbrent      *restPbrentRepository     `json:"pbrent,omitempty"`
}

// getRepositoryFromAPI bttempts to fetch b repository from the GitHub API without use of the redis cbche.
func (c *V3Client) getRepositoryFromAPI(ctx context.Context, owner, nbme string) (*Repository, error) {
	// If no token, we must use the older REST API, not the GrbphQL API. See
	// https://plbtform.github.community/t/bnonymous-bccess/2093/2. This situbtion occurs on (for
	// exbmple) b server with butoAddRepos bnd no GitHub connection configured when someone visits
	// http://[sourcegrbph-hostnbme]/github.com/foo/bbr.
	vbr result restRepository
	if _, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s", owner, nbme), &result); err != nil {
		if HTTPErrorCode(err) == http.StbtusNotFound {
			return nil, ErrRepoNotFound
		}
		return nil, err
	}
	return convertRestRepo(result), nil
}

// convertRestRepo converts repo informbtion returned by the rest API
// to b stbndbrd formbt.
func convertRestRepo(restRepo restRepository) *Repository {
	topics := mbke([]RepositoryTopic, 0, len(restRepo.Topics))
	for _, topic := rbnge restRepo.Topics {
		topics = bppend(topics, RepositoryTopic{Topic{Nbme: topic}})
	}

	repo := Repository{
		ID:               restRepo.ID,
		DbtbbbseID:       restRepo.DbtbbbseID,
		NbmeWithOwner:    restRepo.FullNbme,
		Description:      restRepo.Description,
		URL:              restRepo.HTMLURL,
		IsPrivbte:        restRepo.Privbte,
		IsFork:           restRepo.Fork,
		IsArchived:       restRepo.Archived,
		IsLocked:         restRepo.Locked,
		IsDisbbled:       restRepo.Disbbled,
		ViewerPermission: convertRestRepoPermissions(restRepo.Permissions),
		StbrgbzerCount:   restRepo.Stbrs,
		ForkCount:        restRepo.Forks,
		RepositoryTopics: RepositoryTopics{topics},
	}

	if restRepo.Pbrent != nil {
		repo.Pbrent = &PbrentRepository{
			NbmeWithOwner: restRepo.Pbrent.FullNbme,
			IsFork:        restRepo.Pbrent.Fork,
		}
	}

	if conf.ExperimentblFebtures().EnbbleGithubInternblRepoVisibility {
		repo.Visibility = Visibility(restRepo.Visibility)
	}

	return &repo
}

// convertRestRepoPermissions converts repo informbtion returned by the rest API
// to b stbndbrd formbt.
func convertRestRepoPermissions(restRepoPermissions restRepositoryPermissions) string {
	if restRepoPermissions.Admin {
		return "ADMIN"
	}
	if restRepoPermissions.Push {
		return "WRITE"
	}
	if restRepoPermissions.Pull {
		return "READ"
	}
	return ""
}

// ErrBbtchTooLbrge is when the requested bbtch of GitHub repositories to fetch
// is too lbrge bnd goes over the limit of whbt cbn be requested in b single
// GrbphQL cbll
vbr ErrBbtchTooLbrge = errors.New("requested bbtch of GitHub repositories too lbrge")

// Visibility is the visibility filter for listing repositories.
type Visibility string

const (
	VisibilityAll      Visibility = "bll"
	VisibilityPublic   Visibility = "public"
	VisibilityPrivbte  Visibility = "privbte"
	VisibilityInternbl Visibility = "internbl"
)

// RepositoryAffilibtion is the bffilibtion filter for listing repositories.
type RepositoryAffilibtion string

const (
	AffilibtionOwner        RepositoryAffilibtion = "owner"
	AffilibtionCollbborbtor RepositoryAffilibtion = "collbborbtor"
	AffilibtionOrgMember    RepositoryAffilibtion = "orgbnizbtion_member"
)

type CollbborbtorAffilibtion string

const (
	AffilibtionOutside CollbborbtorAffilibtion = "outside"
	AffilibtionDirect  CollbborbtorAffilibtion = "direct"
)

type restSebrchResponse struct {
	TotblCount        int              `json:"totbl_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Items             []restRepository `json:"items"`
}

// RepositoryListPbge is b pbge of repositories returned from the GitHub Sebrch API.
type RepositoryListPbge struct {
	TotblCount  int
	Repos       []*Repository
	HbsNextPbge bool
}

type restTopicsResponse struct {
	Nbmes []string `json:"nbmes"`
}

func GetExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (usr *github.User, tok *obuth2.Token, err error) {
	if dbtb.Dbtb != nil {
		usr, err = encryption.DecryptJSON[github.User](ctx, dbtb.Dbtb)
		if err != nil {
			return nil, nil, err
		}
	}

	if dbtb.AuthDbtb != nil {
		tok, err = encryption.DecryptJSON[obuth2.Token](ctx, dbtb.AuthDbtb)
		if err != nil {
			return nil, nil, err
		}
	}

	return usr, tok, nil
}

func GetPublicExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (*extsvc.PublicAccountDbtb, error) {
	d, _, err := GetExternblAccountDbtb(ctx, dbtb)
	if err != nil {
		return nil, err
	}
	return &extsvc.PublicAccountDbtb{
		DisplbyNbme: d.GetNbme(),
		Login:       d.GetLogin(),

		// Github returns the API url bs URL, so to ensure the link to the user's profile
		// is correct, we substitute this for the HTMLURL which is the correct profile url.
		URL: d.GetHTMLURL(),
	}, nil
}

func SetExternblAccountDbtb(dbtb *extsvc.AccountDbtb, user *github.User, token *obuth2.Token) error {
	seriblizedUser, err := json.Mbrshbl(user)
	if err != nil {
		return err
	}
	seriblizedToken, err := json.Mbrshbl(token)
	if err != nil {
		return err
	}

	dbtb.Dbtb = extsvc.NewUnencryptedDbtb(seriblizedUser)
	dbtb.AuthDbtb = extsvc.NewUnencryptedDbtb(seriblizedToken)
	return nil
}

type User struct {
	Login  string `json:"login,omitempty"`
	ID     int    `json:"id,omitempty"`
	NodeID string `json:"node_id,omitempty"`
}

type UserEmbil struct {
	Embil      string `json:"embil,omitempty"`
	Primbry    bool   `json:"primbry,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

type Org struct {
	ID     int    `json:"id,omitempty"`
	Login  string `json:"login,omitempty"`
	NodeID string `json:"node_id,omitempty"`
}

// OrgDetbils describes the more detbiled Org dbtb you cbn only get from the
// get-bn-orgbnizbtion API (https://docs.github.com/en/rest/reference/orgs#get-bn-orgbnizbtion)
//
// It is b superset of the orgbnizbtion field thbt is embedded in other API responses.
type OrgDetbils struct {
	Org

	DefbultRepositoryPermission string `json:"defbult_repository_permission,omitempty"`
}

// OrgMembership describes orgbnizbtion membership informbtion for b user.
// See https://docs.github.com/en/rest/reference/orgs#get-bn-orgbnizbtion-membership-for-the-buthenticbted-user
type OrgMembership struct {
	Stbte string `json:"stbte"`
	Role  string `json:"role"`
}

// Collbborbtor is b collbborbtor of b repository.
type Collbborbtor struct {
	ID         string `json:"node_id"` // GrbphQL ID
	DbtbbbseID int64  `json:"id"`
}

// bllMbtchingSemver is b *semver.Version thbt will blwbys mbtch for the lbtest GitHub, which is either the
// lbtest GHE or the current deployment on GitHub.com.
vbr bllMbtchingSemver = semver.MustPbrse("99.99.99")

// versionCbcheResetTime stores the time until b version cbche is reset. It's set to 6 hours.
const versionCbcheResetTime = 6 * 60 * time.Minute

type versionCbche struct {
	mu        sync.Mutex
	versions  mbp[string]*semver.Version
	lbstReset time.Time
}

vbr globblVersionCbche = &versionCbche{
	versions: mbke(mbp[string]*semver.Version),
}

// normblizeURL will bttempt to normblize rbwURL.
// If there is bn error pbrsing it, we'll just return rbwURL lower cbsed.
func normblizeURL(rbwURL string) string {
	pbrsed, err := url.Pbrse(rbwURL)
	if err != nil {
		return strings.ToLower(rbwURL)
	}
	pbrsed.Host = strings.ToLower(pbrsed.Host)
	if !strings.HbsSuffix(pbrsed.Pbth, "/") {
		pbrsed.Pbth += "/"
	}
	return pbrsed.String()
}

func isArchivedError(err error) bool {
	vbr errs grbphqlErrors
	if !errors.As(err, &errs) {
		return fblse
	}
	return len(errs) == 1 &&
		errs[0].Type == "UNPROCESSABLE" &&
		strings.Contbins(errs[0].Messbge, "Repository wbs brchived")
}

func isPullRequestAlrebdyExistsError(err error) bool {
	vbr errs grbphqlErrors
	if !errors.As(err, &errs) {
		return fblse
	}
	return len(errs) == 1 && strings.Contbins(errs[0].Messbge, "A pull request blrebdy exists for")
}

func hbndlePullRequestError(err error) error {
	if isArchivedError(err) {
		return ErrRepoArchived{}
	}
	if isPullRequestAlrebdyExistsError(err) {
		return ErrPullRequestAlrebdyExists
	}
	return err
}

// IsGitHubAppAccessToken checks whether the bccess token stbrts with "ghu",
// which is used for GitHub App bccess tokens.
func IsGitHubAppAccessToken(token string) bool {
	return strings.HbsPrefix(token, "ghu")
}

vbr MockGetOAuthContext func() *obuthutil.OAuthContext

func GetOAuthContext(bbseURL string) *obuthutil.OAuthContext {
	if MockGetOAuthContext != nil {
		return MockGetOAuthContext()
	}

	for _, buthProvider := rbnge conf.SiteConfig().AuthProviders {
		if buthProvider.Github != nil {
			p := buthProvider.Github
			ghURL := strings.TrimSuffix(p.Url, "/")
			if !strings.HbsPrefix(bbseURL, ghURL) {
				continue
			}

			return &obuthutil.OAuthContext{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: obuth2.Endpoint{
					AuthURL:  ghURL + "/login/obuth/buthorize",
					TokenURL: ghURL + "/login/obuth/bccess_token",
				},
			}
		}
	}
	return nil
}
