package azuredevops

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	PullRequestBuildStatusStateSucceeded     PullRequestStatusState = "succeeded"
	PullRequestBuildStatusStateError         PullRequestStatusState = "error"
	PullRequestBuildStatusStateFailed        PullRequestStatusState = "failed"
	PullRequestBuildStatusStatePending       PullRequestStatusState = "pending"
	PullRequestBuildStatusStateNotApplicable PullRequestStatusState = "notApplicable"
	PullRequestBuildStatusStateNotSet        PullRequestStatusState = "notSet"

	PullRequestStatusActive    PullRequestStatus = "active"
	PullRequestStatusAbandoned PullRequestStatus = "abandoned"
	PullRequestStatusCompleted PullRequestStatus = "completed"
	PullRequestStatusNotSet    PullRequestStatus = "notSet"

	PullRequestMergeStrategySquash        PullRequestMergeStrategy = "squash"
	PullRequestMergeStrategyRebase        PullRequestMergeStrategy = "rebase"
	PullRequestMergeStrategyRebaseMerge   PullRequestMergeStrategy = "rebaseMerge"
	PullRequestMergeStrategyNoFastForward PullRequestMergeStrategy = "notFastForward"
)

type Org struct {
	ID   string `json:"accountId"`
	URI  string `json:"accountUri"`
	Name string `json:"accountName"`
}

type ListAuthorizedUserOrgsResponse struct {
	Count int   `json:"count"`
	Value []Org `json:"value"`
}

type OrgProjectRepoArgs struct {
	Org          string
	Project      string
	RepoNameOrID string
}

// ListRepositoriesByProjectOrOrgArgs defines options to be set on the ListRepositories methods' calls.
type ListRepositoriesByProjectOrOrgArgs struct {
	// Should be in the form of 'org/project' for projects and 'org' for orgs.
	ProjectOrOrgName string
}

type ForkRepositoryInput struct {
	Name             string                              `json:"name"`
	Project          ForkRepositoryInputProject          `json:"project"`
	ParentRepository ForkRepositoryInputParentRepository `json:"parentRepository"`
}

type ForkRepositoryInputParentRepository struct {
	ID      string                     `json:"id"`
	Project ForkRepositoryInputProject `json:"project"`
}

type ForkRepositoryInputProject struct {
	ID string `json:"id"`
}

type ListRepositoriesResponse struct {
	Value []Repository `json:"value"`
	Count int          `json:"count"`
}

type ListRefsResponse struct {
	Value []Ref `json:"value"`
	Count int   `json:"count"`
}

type Ref struct {
	Name      string      `json:"name"`
	CommitSHA string      `json:"objectId"`
	Creator   CreatorInfo `json:"creator"`
}

type CreatePullRequestInput struct {
	SourceRefName     string                        `json:"sourceRefName"`
	TargetRefName     string                        `json:"targetRefName"`
	Title             string                        `json:"title"`
	Description       string                        `json:"description"`
	Reviewers         []Reviewer                    `json:"reviewers"`
	ForkSource        *ForkRef                      `json:"forkSource"`
	IsDraft           bool                          `json:"isDraft"`
	CompletionOptions *PullRequestCompletionOptions `json:"completionOptions"`
}

type ForkRef struct {
	Repository Repository `json:"repository"`
	Name       string     `json:"name"`
	URl        string     `json:"url"`
}

type Reviewer struct {
	// Vote represents the status of a review on Azure DevOps. Here are possible values for Vote:
	//
	//   10: approved
	//   5 : approved with suggestions
	//   0 : no vote
	//  -5 : waiting for author
	//  -10: rejected
	Vote        int    `json:"vote"`
	ID          string `json:"id"`
	HasDeclined bool   `json:"hasDeclined"`
	IsRequired  bool   `json:"isRequired"`
	UniqueName  string `json:"uniqueName"`
}

type PullRequestCommonArgs struct {
	PullRequestID string
	Org           string
	Project       string
	RepoNameOrID  string
}

type PullRequest struct {
	Repository            Repository        `json:"repository"`
	ID                    int               `json:"pullRequestId"`
	CodeReviewID          int               `json:"codeReviewId"`
	Status                PullRequestStatus `json:"status"`
	CreationDate          time.Time         `json:"creationDate"`
	Title                 string            `json:"title"`
	Description           string            `json:"description"`
	CreatedBy             CreatorInfo       `json:"createdBy"`
	SourceRefName         string            `json:"sourceRefName"`
	TargetRefName         string            `json:"targetRefName"`
	MergeStatus           string            `json:"mergeStatus"`
	MergeID               string            `json:"mergeId"`
	LastMergeSourceCommit PullRequestCommit `json:"lastMergeSourceCommit"`
	LastMergeTargetCommit PullRequestCommit `json:"lastMergeTargetCommit"`
	SupportsIterations    bool              `json:"supportsIterations"`
	ArtifactID            string            `json:"artifactId"`
	Reviewers             []Reviewer        `json:"reviewers"`
	ForkSource            *ForkRef          `json:"forkSource"`
	URL                   string            `json:"url"`
	IsDraft               bool              `json:"isDraft"`
}

type PullRequestCommit struct {
	CommitID string `json:"commitId"`
	URL      string `json:"url"`
}

type PullRequestUpdateInput struct {
	Status                *PullRequestStatus            `json:"status"`
	Title                 *string                       `json:"title"`
	Description           *string                       `json:"description"`
	MergeOptions          *PullRequestMergeOptions      `json:"mergeOptions"`
	LastMergeSourceCommit *PullRequestCommit            `json:"lastMergeSourceCommit"`
	TargetRefName         *string                       `json:"targetRefName"`
	IsDraft               *bool                         `json:"isDraft"`
	CompletionOptions     *PullRequestCompletionOptions `json:"completionOptions"`
	// ADO does not seem to support updating Source ref name, only TargetRefName which needs to be explicitly enabled.
}

type PullRequestStatus string
type PullRequestMergeStrategy string

type PullRequestMergeOptions struct {
	ConflictAuthorshipCommits  *bool `json:"conflictAuthorshipCommits"`
	DetectRenameFalsePositives *bool `json:"detectRenameFalsePositives"`
	DisableRenames             *bool `json:"disableRenames"`
}

type PullRequestCompleteInput struct {
	CommitID           string
	MergeStrategy      *PullRequestMergeStrategy
	DeleteSourceBranch bool
}

type PullRequestCompletionOptions struct {
	MergeStrategy      PullRequestMergeStrategy `json:"mergeStrategy,omitempty"`
	DeleteSourceBranch bool                     `json:"deleteSourceBranch,omitempty"`
	MergeCommitMessage string                   `json:"mergeCommitMessage"`
}

type PullRequestCommentInput struct {
	Comments []PullRequestCommentForInput `json:"Comments"`
}

type PullRequestCommentResponse struct {
	ID            int                             `json:"id"`
	Comments      []PullRequestCommentForResponse `json:"Comments"`
	PublishedDate time.Time                       `json:"publishedDate"`
	LastUpdatedOn time.Time                       `json:"lastUpdatedOn"`
	IsDeleted     bool                            `json:"isDeleted"`
}

type PullRequestCommentForInput struct {
	ParentCommentID int    `json:"parentCommentId"`
	Content         string `json:"content"`
	CommentType     int    `json:"commentType"`
}
type PullRequestCommentForResponse struct {
	ID            int64     `json:"id"`
	PublishedDate time.Time `json:"publishedDate"`
	LastUpdatedOn time.Time `json:"lastUpdatedOn"`
	Content       string    `json:"content"`
	CommentType   string    `json:"commentType"`
}

type PullRequestStatuses struct {
	Value []PullRequestBuildStatus
	Count int
}

type PullRequestBuildStatus struct {
	ID           int                    `json:"id"`
	State        PullRequestStatusState `json:"state"`
	Description  string                 `json:"description"`
	CreationDate time.Time              `json:"creationDate"`
	UpdateDate   time.Time              `json:"updatedDate"`
	CreatedBy    CreatorInfo            `json:"createdBy"`
}

type PullRequestStatusState string

type Repository struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CloneURL   string  `json:"remoteURL"`
	APIURL     string  `json:"url"`
	SSHURL     string  `json:"sshUrl"`
	WebURL     string  `json:"webUrl"`
	IsDisabled bool    `json:"isDisabled"`
	IsFork     bool    `json:"isFork"`
	Project    Project `json:"project"`
}

type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Revision   int    `json:"revision"`
	Visibility string `json:"visibility"`
	URL        string `json:"url"`
}

func (p Repository) GetOrganization() (string, error) {
	u, err := url.Parse(p.APIURL)
	if err != nil {
		return "", err
	}

	splitPath := strings.SplitN(u.Path, "/", 3)
	if len(splitPath) != 3 {
		return "", errors.Errorf("unable to parse Azure DevOps organization from repo URL: %s", p.APIURL)
	}

	return splitPath[1], nil
}

func (p Repository) Namespace() string {
	return p.Project.Name
}

type Profile struct {
	ID           string    `json:"id"`
	DisplayName  string    `json:"displayName"`
	EmailAddress string    `json:"emailAddress"`
	LastChanged  time.Time `json:"timestamp"`
	PublicAlias  string    `json:"publicAlias"`
}

type CreatorInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	URL         string `json:"url"`
	ImageURL    string `json:"imageUrl"`
}

type HTTPError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

// Error returns a minimal string of the HTTP error with the status code and the URL.
//
// It does not write HTTPError.Body as part of the error string as the Azure DevOPs API returns raw
// HTML for 4xx errors and non 200 OK responses inspite of using the "application/json" header in
// the request. It does return JSON for 200 OK responses though.
//
// The body would only add noise to logs and error messages that also may bubble up to the user
// interface which makes for a bad user experience. For our usecases in debugging, the status
// code and URL should provide plenty of information on what is going contrary to expectations.
// In the worst case, we should reproduce the error by manually sending a curl request that
// causes an error and inspecting the HTML output.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("Azure DevOps API HTTP error: code=%d url=%q", e.StatusCode, e.URL)
}
