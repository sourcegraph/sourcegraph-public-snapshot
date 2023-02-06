package azuredevops

import (
	"net/url"
	"time"
)

const (
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
	SourceRefName string     `json:"sourceRefName"`
	TargetRefName string     `json:"targetRefName"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Reviewers     []Reviewer `json:"reviewers"`
}

type Reviewer struct {
	ID string `json:"id"`
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
	Status                string            `json:"status"`
	CreationDate          string            `json:"creationDate"`
	Title                 string            `json:"title"`
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
}

type PullRequestCommit struct {
	CommitID string `json:"commitId"`
	URL      string `json:"url"`
}

type PullRequestReviewer struct {
	ID          string `json:"id"`
	ReviewerURL string `json:"reviewerUrl"`
	Vote        int    `json:"vote"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	URL         string `json:"url"`
	ImageURL    string `json:"imageUrl"`
}

type PullRequestUpdateInput struct {
	Status                *PullRequestStatus       `json:"status"`
	Title                 *string                  `json:"title"`
	Description           *string                  `json:"description"`
	MergeOptions          *PullRequestMergeOptions `json:"mergeOptions"`
	LastMergeSourceCommit *PullRequestCommitRef    `json:"lastMergeSourceCommit"`
	TargetRefName         *string                  `json:"targetRefName"`
	// ADO does not seem to support updating Source ref name, only TargetRefName which needs to be explicitly enabled.
}

type PullRequestStatus string
type PullRequestMergeStrategy string

type PullRequestMergeOptions struct {
	ConflictAuthorshipCommits  *bool `json:"conflictAuthorshipCommits"`
	DetectRenameFalsePositives *bool `json:"detectRenameFalsePositives"`
	DisableRenames             *bool `json:"disableRenames"`
}

type PullRequestCommitRef struct {
	CommitID string `json:"commitId"`
}

type PullRequestCompletionOptions struct {
	MergeStrategy      PullRequestMergeStrategy `json:"mergeStrategy"`
	DeleteSourceBranch bool                     `json:"deleteSourceBranch"`
	MergeCommitMessage string                   `json:"mergeCommitMessage"`
	SquashMerge        bool                     `json:"squashMerge"`
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
	ParentCommitID int    `json:"parentCommitId"`
	Content        string `json:"content"`
	CommentType    int    `json:"commentType"`
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
	CreationDate string                 `json:"creationDate"`
	UpdateDate   string                 `json:"updateDate"`
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
	Project    Project `json:"project"`
}

type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Revision   int    `json:"revision"`
	Visibility string `json:"visibility"`
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

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}
