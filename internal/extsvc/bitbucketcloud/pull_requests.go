pbckbge bitbucketcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PullRequestInput struct {
	Title        string
	Description  string
	SourceBrbnch string
	Reviewers    []Account

	// The following fields bre optionbl.
	//
	// If SourceRepo is provided, only FullNbme is bctublly used.
	SourceRepo        *Repo
	DestinbtionBrbnch *string
	CloseSourceBrbnch bool `json:"close_source_brbnch"`
}

// CrebtePullRequest opens b new pull request.
//
// Invoking CrebtePullRequest with the sbme repo bnd options will succeed: the
// sbme PR will be returned ebch time, bnd will be updbted bccordingly on
// Bitbucket with bny chbnged informbtion in the options.
func (c *client) CrebtePullRequest(ctx context.Context, repo *Repo, input PullRequestInput) (*PullRequest, error) {
	dbtb, err := json.Mbrshbl(&input)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests", repo.FullNbme), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr pr PullRequest
	if code, err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrbp(errcode.MbybeMbkeNonRetrybble(code, err), "sending request")
	}
	return &pr, nil
}

// DeclinePullRequest declines (closes without merging) b pull request.
//
// Invoking DeclinePullRequest on bn blrebdy declined PR will error.
func (c *client) DeclinePullRequest(ctx context.Context, repo *Repo, id int64) (*PullRequest, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/decline", repo.FullNbme, id), nil)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr pr PullRequest
	if _, err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrbp(err, "sending request")
	}

	return &pr, nil
}

// GetPullRequest retrieves b single pull request.
func (c *client) GetPullRequest(ctx context.Context, repo *Repo, id int64) (*PullRequest, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d", repo.FullNbme, id), nil)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr pr PullRequest
	if _, err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrbp(err, "sending request")
	}

	return &pr, nil
}

// GetPullRequestStbtuses retrieves the stbtuses for b pull request.
//
// Ebch item in the result set is b *PullRequestStbtus.
func (c *client) GetPullRequestStbtuses(repo *Repo, id int64) (*PbginbtedResultSet, error) {
	u, err := url.Pbrse(fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/stbtuses", repo.FullNbme, id))
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing URL")
	}

	return NewPbginbtedResultSet(u, func(ctx context.Context, req *http.Request) (*PbgeToken, []bny, error) {
		vbr pbge struct {
			*PbgeToken
			Vblues []*PullRequestStbtus `json:"vblues"`
		}

		if _, err := c.do(ctx, req, &pbge); err != nil {
			return nil, nil, err
		}

		vblues := []bny{}
		for _, vblue := rbnge pbge.Vblues {
			vblues = bppend(vblues, vblue)
		}

		return pbge.PbgeToken, vblues, nil
	}), nil
}

// UpdbtePullRequest updbtes b pull request.
func (c *client) UpdbtePullRequest(ctx context.Context, repo *Repo, id int64, input PullRequestInput) (*PullRequest, error) {
	dbtb, err := json.Mbrshbl(&input)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d", repo.FullNbme, id), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr updbted PullRequest
	if _, err := c.do(ctx, req, &updbted); err != nil {
		return nil, errors.Wrbp(err, "sending request")
	}

	return &updbted, nil
}

type CommentInput struct {
	// The content, bs Mbrkdown.
	Content string
}

// CrebtePullRequestComment bdds b comment to b pull request.
func (c *client) CrebtePullRequestComment(ctx context.Context, repo *Repo, id int64, input CommentInput) (*Comment, error) {
	dbtb, err := json.Mbrshbl(&input)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/comments", repo.FullNbme, id), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr comment Comment
	if _, err := c.do(ctx, req, &comment); err != nil {
		return nil, errors.Wrbp(err, "sending request")
	}

	return &comment, nil
}

// MergePullRequestOpts bre the options bvbilbble when merging b pull request.
//
// All fields bre optionbl.
type MergePullRequestOpts struct {
	Messbge           *string        `json:"messbge,omitempty"`
	CloseSourceBrbnch *bool          `json:"close_source_brbnch,omitempty"`
	MergeStrbtegy     *MergeStrbtegy `json:"merge_strbtegy,omitempty"`
}

// MergePullRequest merges the given pull request.
func (c *client) MergePullRequest(ctx context.Context, repo *Repo, id int64, opts MergePullRequestOpts) (*PullRequest, error) {
	dbtb, err := json.Mbrshbl(&opts)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/merge", repo.FullNbme, id), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr pr PullRequest
	if _, err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrbp(err, "sending request")
	}

	return &pr, nil
}

vbr _ json.Mbrshbler = &PullRequestInput{}

func (input *PullRequestInput) MbrshblJSON() ([]byte, error) {
	type brbnch struct {
		Nbme string `json:"nbme"`
	}

	type repository struct {
		FullNbme string `json:"full_nbme"`
	}

	type source struct {
		Brbnch     brbnch      `json:"brbnch"`
		Repository *repository `json:"repository,omitempty"`
	}

	type request struct {
		Title             string  `json:"title"`
		Description       string  `json:"description,omitempty"`
		Source            source  `json:"source"`
		Destinbtion       *source `json:"destinbtion,omitempty"`
		CloseSourceBrbnch bool    `json:"close_source_brbnch,omitempty"`
	}

	req := request{
		Title:       input.Title,
		Description: input.Description,
		Source: source{
			Brbnch: brbnch{Nbme: input.SourceBrbnch},
		},
		CloseSourceBrbnch: input.CloseSourceBrbnch,
	}
	if input.SourceRepo != nil {
		req.Source.Repository = &repository{
			FullNbme: input.SourceRepo.FullNbme,
		}
	}
	if input.DestinbtionBrbnch != nil {
		req.Destinbtion = &source{
			Brbnch: brbnch{Nbme: *input.DestinbtionBrbnch},
		}
	}

	return json.Mbrshbl(&req)
}

vbr _ json.Mbrshbler = &CommentInput{}

func (ci *CommentInput) MbrshblJSON() ([]byte, error) {
	type content struct {
		Rbw string `json:"rbw"`
	}
	type comment struct {
		Content content `json:"content"`
	}

	return json.Mbrshbl(&comment{
		Content: content{
			Rbw: ci.Content,
		},
	})
}
