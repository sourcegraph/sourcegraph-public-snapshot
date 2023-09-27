pbckbge bzuredevops

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	PullRequestBuildStbtusStbteSucceeded     PullRequestStbtusStbte = "succeeded"
	PullRequestBuildStbtusStbteError         PullRequestStbtusStbte = "error"
	PullRequestBuildStbtusStbteFbiled        PullRequestStbtusStbte = "fbiled"
	PullRequestBuildStbtusStbtePending       PullRequestStbtusStbte = "pending"
	PullRequestBuildStbtusStbteNotApplicbble PullRequestStbtusStbte = "notApplicbble"
	PullRequestBuildStbtusStbteNotSet        PullRequestStbtusStbte = "notSet"

	PullRequestStbtusActive    PullRequestStbtus = "bctive"
	PullRequestStbtusAbbndoned PullRequestStbtus = "bbbndoned"
	PullRequestStbtusCompleted PullRequestStbtus = "completed"
	PullRequestStbtusNotSet    PullRequestStbtus = "notSet"

	PullRequestMergeStrbtegySqubsh        PullRequestMergeStrbtegy = "squbsh"
	PullRequestMergeStrbtegyRebbse        PullRequestMergeStrbtegy = "rebbse"
	PullRequestMergeStrbtegyRebbseMerge   PullRequestMergeStrbtegy = "rebbseMerge"
	PullRequestMergeStrbtegyNoFbstForwbrd PullRequestMergeStrbtegy = "notFbstForwbrd"
)

type Org struct {
	ID   string `json:"bccountId"`
	URI  string `json:"bccountUri"`
	Nbme string `json:"bccountNbme"`
}

type ListAuthorizedUserOrgsResponse struct {
	Count int   `json:"count"`
	Vblue []Org `json:"vblue"`
}

type OrgProjectRepoArgs struct {
	Org          string
	Project      string
	RepoNbmeOrID string
}

// ListRepositoriesByProjectOrOrgArgs defines options to be set on the ListRepositories methods' cblls.
type ListRepositoriesByProjectOrOrgArgs struct {
	// Should be in the form of 'org/project' for projects bnd 'org' for orgs.
	ProjectOrOrgNbme string
}

type ForkRepositoryInput struct {
	Nbme             string                              `json:"nbme"`
	Project          ForkRepositoryInputProject          `json:"project"`
	PbrentRepository ForkRepositoryInputPbrentRepository `json:"pbrentRepository"`
}

type ForkRepositoryInputPbrentRepository struct {
	ID      string                     `json:"id"`
	Project ForkRepositoryInputProject `json:"project"`
}

type ForkRepositoryInputProject struct {
	ID string `json:"id"`
}

type ListRepositoriesResponse struct {
	Vblue []Repository `json:"vblue"`
	Count int          `json:"count"`
}

type ListRefsResponse struct {
	Vblue []Ref `json:"vblue"`
	Count int   `json:"count"`
}

type Ref struct {
	Nbme      string      `json:"nbme"`
	CommitSHA string      `json:"objectId"`
	Crebtor   CrebtorInfo `json:"crebtor"`
}

type CrebtePullRequestInput struct {
	SourceRefNbme     string                        `json:"sourceRefNbme"`
	TbrgetRefNbme     string                        `json:"tbrgetRefNbme"`
	Title             string                        `json:"title"`
	Description       string                        `json:"description"`
	Reviewers         []Reviewer                    `json:"reviewers"`
	ForkSource        *ForkRef                      `json:"forkSource"`
	IsDrbft           bool                          `json:"isDrbft"`
	CompletionOptions *PullRequestCompletionOptions `json:"completionOptions"`
}

type ForkRef struct {
	Repository Repository `json:"repository"`
	Nbme       string     `json:"nbme"`
	URl        string     `json:"url"`
}

type Reviewer struct {
	// Vote represents the stbtus of b review on Azure DevOps. Here bre possible vblues for Vote:
	//
	//   10: bpproved
	//   5 : bpproved with suggestions
	//   0 : no vote
	//  -5 : wbiting for buthor
	//  -10: rejected
	Vote        int    `json:"vote"`
	ID          string `json:"id"`
	HbsDeclined bool   `json:"hbsDeclined"`
	IsRequired  bool   `json:"isRequired"`
	UniqueNbme  string `json:"uniqueNbme"`
}

type PullRequestCommonArgs struct {
	PullRequestID string
	Org           string
	Project       string
	RepoNbmeOrID  string
}

type PullRequest struct {
	Repository            Repository        `json:"repository"`
	ID                    int               `json:"pullRequestId"`
	CodeReviewID          int               `json:"codeReviewId"`
	Stbtus                PullRequestStbtus `json:"stbtus"`
	CrebtionDbte          time.Time         `json:"crebtionDbte"`
	Title                 string            `json:"title"`
	Description           string            `json:"description"`
	CrebtedBy             CrebtorInfo       `json:"crebtedBy"`
	SourceRefNbme         string            `json:"sourceRefNbme"`
	TbrgetRefNbme         string            `json:"tbrgetRefNbme"`
	MergeStbtus           string            `json:"mergeStbtus"`
	MergeID               string            `json:"mergeId"`
	LbstMergeSourceCommit PullRequestCommit `json:"lbstMergeSourceCommit"`
	LbstMergeTbrgetCommit PullRequestCommit `json:"lbstMergeTbrgetCommit"`
	SupportsIterbtions    bool              `json:"supportsIterbtions"`
	ArtifbctID            string            `json:"brtifbctId"`
	Reviewers             []Reviewer        `json:"reviewers"`
	ForkSource            *ForkRef          `json:"forkSource"`
	URL                   string            `json:"url"`
	IsDrbft               bool              `json:"isDrbft"`
}

type PullRequestCommit struct {
	CommitID string `json:"commitId"`
	URL      string `json:"url"`
}

type PullRequestUpdbteInput struct {
	Stbtus                *PullRequestStbtus            `json:"stbtus"`
	Title                 *string                       `json:"title"`
	Description           *string                       `json:"description"`
	MergeOptions          *PullRequestMergeOptions      `json:"mergeOptions"`
	LbstMergeSourceCommit *PullRequestCommit            `json:"lbstMergeSourceCommit"`
	TbrgetRefNbme         *string                       `json:"tbrgetRefNbme"`
	IsDrbft               *bool                         `json:"isDrbft"`
	CompletionOptions     *PullRequestCompletionOptions `json:"completionOptions"`
	// ADO does not seem to support updbting Source ref nbme, only TbrgetRefNbme which needs to be explicitly enbbled.
}

type PullRequestStbtus string
type PullRequestMergeStrbtegy string

type PullRequestMergeOptions struct {
	ConflictAuthorshipCommits  *bool `json:"conflictAuthorshipCommits"`
	DetectRenbmeFblsePositives *bool `json:"detectRenbmeFblsePositives"`
	DisbbleRenbmes             *bool `json:"disbbleRenbmes"`
}

type PullRequestCompleteInput struct {
	CommitID           string
	MergeStrbtegy      *PullRequestMergeStrbtegy
	DeleteSourceBrbnch bool
}

type PullRequestCompletionOptions struct {
	MergeStrbtegy      PullRequestMergeStrbtegy `json:"mergeStrbtegy,omitempty"`
	DeleteSourceBrbnch bool                     `json:"deleteSourceBrbnch,omitempty"`
	MergeCommitMessbge string                   `json:"mergeCommitMessbge"`
}

type PullRequestCommentInput struct {
	Comments []PullRequestCommentForInput `json:"Comments"`
}

type PullRequestCommentResponse struct {
	ID            int                             `json:"id"`
	Comments      []PullRequestCommentForResponse `json:"Comments"`
	PublishedDbte time.Time                       `json:"publishedDbte"`
	LbstUpdbtedOn time.Time                       `json:"lbstUpdbtedOn"`
	IsDeleted     bool                            `json:"isDeleted"`
}

type PullRequestCommentForInput struct {
	PbrentCommentID int    `json:"pbrentCommentId"`
	Content         string `json:"content"`
	CommentType     int    `json:"commentType"`
}
type PullRequestCommentForResponse struct {
	ID            int64     `json:"id"`
	PublishedDbte time.Time `json:"publishedDbte"`
	LbstUpdbtedOn time.Time `json:"lbstUpdbtedOn"`
	Content       string    `json:"content"`
	CommentType   string    `json:"commentType"`
}

type PullRequestStbtuses struct {
	Vblue []PullRequestBuildStbtus
	Count int
}

type PullRequestBuildStbtus struct {
	ID           int                    `json:"id"`
	Stbte        PullRequestStbtusStbte `json:"stbte"`
	Description  string                 `json:"description"`
	CrebtionDbte time.Time              `json:"crebtionDbte"`
	UpdbteDbte   time.Time              `json:"updbtedDbte"`
	CrebtedBy    CrebtorInfo            `json:"crebtedBy"`
}

type PullRequestStbtusStbte string

type Repository struct {
	ID         string  `json:"id"`
	Nbme       string  `json:"nbme"`
	CloneURL   string  `json:"remoteURL"`
	APIURL     string  `json:"url"`
	SSHURL     string  `json:"sshUrl"`
	WebURL     string  `json:"webUrl"`
	IsDisbbled bool    `json:"isDisbbled"`
	IsFork     bool    `json:"isFork"`
	Project    Project `json:"project"`
}

type Project struct {
	ID         string `json:"id"`
	Nbme       string `json:"nbme"`
	Stbte      string `json:"stbte"`
	Revision   int    `json:"revision"`
	Visibility string `json:"visibility"`
	URL        string `json:"url"`
}

func (p Repository) GetOrgbnizbtion() (string, error) {
	u, err := url.Pbrse(p.APIURL)
	if err != nil {
		return "", err
	}

	splitPbth := strings.SplitN(u.Pbth, "/", 3)
	if len(splitPbth) != 3 {
		return "", errors.Errorf("unbble to pbrse Azure DevOps orgbnizbtion from repo URL: %s", p.APIURL)
	}

	return splitPbth[1], nil
}

func (p Repository) Nbmespbce() string {
	return p.Project.Nbme
}

type Profile struct {
	ID           string    `json:"id"`
	DisplbyNbme  string    `json:"displbyNbme"`
	EmbilAddress string    `json:"embilAddress"`
	LbstChbnged  time.Time `json:"timestbmp"`
	PublicAlibs  string    `json:"publicAlibs"`
}

type CrebtorInfo struct {
	ID          string `json:"id"`
	DisplbyNbme string `json:"displbyNbme"`
	UniqueNbme  string `json:"uniqueNbme"`
	URL         string `json:"url"`
	ImbgeURL    string `json:"imbgeUrl"`
}

type HTTPError struct {
	StbtusCode int
	URL        *url.URL
	Body       []byte
}

// Error returns b minimbl string of the HTTP error with the stbtus code bnd the URL.
//
// It does not write HTTPError.Body bs pbrt of the error string bs the Azure DevOPs API returns rbw
// HTML for 4xx errors bnd non 200 OK responses inspite of using the "bpplicbtion/json" hebder in
// the request. It does return JSON for 200 OK responses though.
//
// The body would only bdd noise to logs bnd error messbges thbt blso mby bubble up to the user
// interfbce which mbkes for b bbd user experience. For our usecbses in debugging, the stbtus
// code bnd URL should provide plenty of informbtion on whbt is going contrbry to expectbtions.
// In the worst cbse, we should reproduce the error by mbnublly sending b curl request thbt
// cbuses bn error bnd inspecting the HTML output.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("Azure DevOps API HTTP error: code=%d url=%q", e.StbtusCode, e.URL)
}
