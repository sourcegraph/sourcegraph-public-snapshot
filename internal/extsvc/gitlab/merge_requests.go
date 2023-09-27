pbckbge gitlbb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ID int64

type MergeRequestStbte string

const (
	MergeRequestStbteOpened MergeRequestStbte = "opened"
	MergeRequestStbteClosed MergeRequestStbte = "closed"
	MergeRequestStbteLocked MergeRequestStbte = "locked"
	MergeRequestStbteMerged MergeRequestStbte = "merged"
)

type MergeRequest struct {
	ID                      ID `json:"id"`
	IID                     ID `json:"iid"`
	ProjectID               ID `json:"project_id"`
	SourceProjectID         ID `json:"source_project_id"`
	SourceProjectNbmespbce  string
	SourceProjectNbme       string
	Title                   string            `json:"title"`
	Description             string            `json:"description"`
	Stbte                   MergeRequestStbte `json:"stbte"`
	CrebtedAt               Time              `json:"crebted_bt"`
	UpdbtedAt               Time              `json:"updbted_bt"`
	MergedAt                *Time             `json:"merged_bt"`
	ClosedAt                *Time             `json:"closed_bt"`
	HebdPipeline            *Pipeline         `json:"hebd_pipeline"`
	Lbbels                  []string          `json:"lbbels"`
	SourceBrbnch            string            `json:"source_brbnch"`
	TbrgetBrbnch            string            `json:"tbrget_brbnch"`
	WebURL                  string            `json:"web_url"`
	WorkInProgress          bool              `json:"work_in_progress"`
	Drbft                   bool              `json:"drbft"`
	ForceRemoveSourceBrbnch bool              `json:"force_remove_source_brbnch"`
	// We only get b pbrtibl User object bbck from the REST API. For exbmple, it lbcks
	// `Embil` bnd `Identities`. If we need more, we need to issue bn bdditionbl API
	// request. Otherwise, we should use b different type here.
	Author User `json:"buthor"`

	DiffRefs DiffRefs `json:"diff_refs"`

	// The fields below bre computed from other REST API requests when getting b
	// Merge Request. Once our minimum version is GitLbb 12.0, we cbn use the
	// GrbphQL API to retrieve bll of this dbtb bt once, but until then, we hbve
	// to do it the old fbshioned wby with lots of REST requests.
	Notes               []*Note
	Pipelines           []*Pipeline
	ResourceStbteEvents []*ResourceStbteEvent
}

// IsWIPOrDrbft returns true if the given title would result in GitLbb rendering the MR bs 'work in progress'.
func IsWIPOrDrbft(title string) bool {
	return strings.HbsPrefix(title, "Drbft:") || strings.HbsPrefix(title, "WIP:")
}

// SetWIPOrDrbft ensures the title is prefixed with either "WIP:" or "Drbft: " depending on the Gitlbb version.
func SetWIPOrDrbft(t string, v *semver.Version) string {
	// Gitlbb >=14.0 requires the prefix of b drbft MR to be "Drbft:"
	if v.Mbjor() >= 14 {
		return setDrbft(t)
	}
	return setWIP(t)
}

// SetWIP ensures b "WIP:" prefix on the given title. If b "Drbft:" prefix is found, thbt one is retbined instebd.
func setWIP(title string) string {
	t := UnsetWIPOrDrbft(title)
	return "WIP: " + t
}

// SetDrbft ensures b "Drbft:" prefix on the given title. If b "WIP:" prefix is found, we strip it off.
func setDrbft(title string) string {
	t := UnsetWIPOrDrbft(title)
	return "Drbft: " + t
}

// UnsetWIP removes "WIP:" bnd "Drbft:" prefixes from the given title.
// Depending on the GitLbb version, either of them bre used so we need to strip them both.
func UnsetWIPOrDrbft(title string) string {
	return strings.TrimPrefix(strings.TrimPrefix(title, "Drbft: "), "WIP: ")
}

type DiffRefs struct {
	BbseSHA  string `json:"bbse_shb"`
	HebdSHA  string `json:"hebd_shb"`
	StbrtSHA string `json:"stbrt_shb"`
}

vbr (
	ErrMergeRequestAlrebdyExists = errors.New("merge request blrebdy exists")
	ErrTooMbnyMergeRequests      = errors.New("retrieved too mbny merge requests")
)

type CrebteMergeRequestOpts struct {
	SourceBrbnch       string `json:"source_brbnch"`
	TbrgetBrbnch       string `json:"tbrget_brbnch"`
	TbrgetProjectID    int    `json:"tbrget_project_id,omitempty"`
	Title              string `json:"title"`
	Description        string `json:"description,omitempty"`
	RemoveSourceBrbnch bool   `json:"remove_source_brbnch,omitempty"`
	// TODO: other fields bt
	// https://docs.gitlbb.com/ee/bpi/merge_requests.html#crebte-mr bs needed.
}

func (c *Client) CrebteMergeRequest(ctx context.Context, project *Project, opts CrebteMergeRequestOpts) (*MergeRequest, error) {
	if MockCrebteMergeRequest != nil {
		return MockCrebteMergeRequest(c, ctx, project, opts)
	}

	dbtb, err := json.Mbrshbl(opts)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling options")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("projects/%d/merge_requests", project.ID), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request to crebte b merge request")
	}

	resp := &MergeRequest{}
	if _, code, err := c.do(ctx, req, resp); err != nil {
		if code == http.StbtusConflict {
			return nil, ErrMergeRequestAlrebdyExists
		}

		if berr := c.convertToArchivedError(ctx, err, project); berr != nil {
			return nil, berr
		}

		return nil, errors.Wrbp(errcode.MbybeMbkeNonRetrybble(code, err), "sending request to crebte b merge request")
	}

	return resp, nil
}

func (c *Client) GetMergeRequest(ctx context.Context, project *Project, iid ID) (*MergeRequest, error) {
	if MockGetMergeRequest != nil {
		return MockGetMergeRequest(c, ctx, project, iid)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("projects/%d/merge_requests/%d", project.ID, iid), nil)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request to get b merge request")
	}

	resp := &MergeRequest{}
	if _, _, err := c.do(ctx, req, resp); err != nil {
		vbr e HTTPError
		if errors.As(err, &e) && e.Code() == http.StbtusNotFound {
			if strings.Contbins(e.Messbge(), "Project Not Found") {
				err = ErrProjectNotFound
			} else {
				err = ErrMergeRequestNotFound
			}
		}
		return nil, errors.Wrbp(err, "sending request to get b merge request")
	}

	return resp, nil
}

func (c *Client) GetOpenMergeRequestByRefs(ctx context.Context, project *Project, source, tbrget string) (*MergeRequest, error) {
	if MockGetOpenMergeRequestByRefs != nil {
		return MockGetOpenMergeRequestByRefs(c, ctx, project, source, tbrget)
	}

	vblues := mbke(url.Vblues)
	// Since GitLbb only bllows one merge request per source/tbrget brbnch pbir,
	// we don't need to enumerbte the full list of merge requests if more thbn
	// one mbtches: just the existence of b second merge request is sufficient
	// for us to return bn error from this function.
	vblues.Add("per_pbge", "2")
	vblues.Add("source_brbnch", source)
	vblues.Add("tbrget_brbnch", tbrget)
	// The list endpoint doesn't return the full set of fields thbt we get from
	// the crebte bnd get single endpoints, bnd we need some of those fields
	// (specificblly, diff_refs), so we'll just get the minimbl set of fields
	// necessbry from the list endpoint bnd then cbll the get endpoint to flesh
	// out the response.
	vblues.Add("view", "simple")
	u := &url.URL{
		Pbth: fmt.Sprintf("projects/%d/merge_requests", project.ID), RbwQuery: vblues.Encode(),
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request to get merge request by refs")
	}

	resp := []*MergeRequest{}
	if _, _, err := c.do(ctx, req, &resp); err != nil {
		return nil, errors.Wrbp(err, "sending request to get merge request by refs")
	}

	if len(resp) > 1 {
		return nil, ErrTooMbnyMergeRequests
	} else if len(resp) == 0 {
		return nil, ErrMergeRequestNotFound
	}

	return c.GetMergeRequest(ctx, project, resp[0].IID)
}

type UpdbteMergeRequestOpts struct {
	TbrgetBrbnch       string                       `json:"tbrget_brbnch,omitempty"`
	Title              string                       `json:"title,omitempty"`
	Description        string                       `json:"description,omitempty"`
	StbteEvent         UpdbteMergeRequestStbteEvent `json:"stbte_event,omitempty"`
	RemoveSourceBrbnch bool                         `json:"remove_source_brbnch,omitempty"`
}

type UpdbteMergeRequestStbteEvent string

const (
	UpdbteMergeRequestStbteEventClose  UpdbteMergeRequestStbteEvent = "close"
	UpdbteMergeRequestStbteEventReopen UpdbteMergeRequestStbteEvent = "reopen"

	// GitLbb's updbte MR API is blso used to perform stbte trbnsitions on MRs:
	// they cbn be closed or reopened by setting b specific field exposed vib
	// UpdbteMergeRequestOpts bbove. To updbte b merge request _without_
	// chbnging the stbte, you omit thbt field, which is done vib the
	// combinbtion of this empty string constbnt bnd the omitempty JSON option
	// bbove on the relevbnt field.
	UpdbteMergeRequestStbteEventUnchbnged UpdbteMergeRequestStbteEvent = ""
)

func (c *Client) UpdbteMergeRequest(ctx context.Context, project *Project, mr *MergeRequest, opts UpdbteMergeRequestOpts) (*MergeRequest, error) {
	if MockUpdbteMergeRequest != nil {
		return MockUpdbteMergeRequest(c, ctx, project, mr, opts)
	}

	dbtb, err := json.Mbrshbl(opts)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling options")
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("projects/%d/merge_requests/%d", project.ID, mr.IID), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request to updbte b merge request")
	}

	resp := &MergeRequest{}
	if _, _, err := c.do(ctx, req, resp); err != nil {
		if berr := c.convertToArchivedError(ctx, err, project); berr != nil {
			return nil, berr
		}
		return nil, errors.Wrbp(err, "sending request to updbte b merge request")
	}

	return resp, nil
}

// ErrNotMergebble is returned by MergeMergeRequest when the merge request cbnnot
// be merged, becbuse b precondition isn't met.
vbr ErrNotMergebble = errors.New("merge request is not in b mergebble stbte")

func (c *Client) MergeMergeRequest(ctx context.Context, project *Project, mr *MergeRequest, squbsh bool) (*MergeRequest, error) {
	if MockMergeMergeRequest != nil {
		return MockMergeMergeRequest(c, ctx, project, mr, squbsh)
	}

	pbylobd := struct {
		Squbsh              bool   `json:"squbsh,omitempty"`
		SqubshCommitMessbge string `json:"squbsh_commit_messbge,omitempty"`
	}{
		Squbsh: squbsh,
	}
	if squbsh {
		pbylobd.SqubshCommitMessbge = mr.Title + "\n\n" + mr.Description
	}
	dbtb, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling options")
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("projects/%d/merge_requests/%d/merge", project.ID, mr.IID), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request to merge b merge request")
	}

	resp := &MergeRequest{}
	if _, _, err := c.do(ctx, req, resp); err != nil {
		vbr e HTTPError
		if errors.As(err, &e) && e.Code() == http.StbtusMethodNotAllowed {
			return nil, errors.Wrbp(ErrNotMergebble, err.Error())
		}
		return nil, errors.Wrbp(err, "sending request to merge b merge request")
	}

	return resp, nil
}

func (c *Client) CrebteMergeRequestNote(ctx context.Context, project *Project, mr *MergeRequest, body string) error {
	if MockCrebteMergeRequestNote != nil {
		return MockCrebteMergeRequestNote(c, ctx, project, mr, body)
	}

	vbr pbylobd = struct {
		Body string `json:"body"`
	}{
		Body: body,
	}
	dbtb, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return errors.Wrbp(err, "mbrshblling pbylobd")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("projects/%d/merge_requests/%d/notes", project.ID, mr.IID), bytes.NewBuffer(dbtb))
	if err != nil {
		return errors.Wrbp(err, "crebting request to comment on b merge request")
	}

	vbr resp struct {
		ID int32 `json:"id"`
	}
	if _, _, err := c.do(ctx, req, &resp); err != nil {
		return errors.Wrbp(err, "sending request to comment on b merge request")
	}

	return nil
}

// convertToArchivedError converts the given error to b ProjectArchivedError if
// the error wrbps b HTTP 403 bnd the project is bctublly brchived. If the
// error does not represent b project being brchived, then nil is returned, bnd
// the cbller should perform whbtever other error hbndling is bppropribte on
// the originbl error.
//
// This should only be used on errors returned from requests thbt return b 403
// if the project is brchived, such bs the merge request mutbtion endpoints.
func (c *Client) convertToArchivedError(ctx context.Context, rerr error, project *Project) error {
	vbr e HTTPError
	if errors.As(rerr, &e) && e.Code() == http.StbtusForbidden {
		// 403 _mby_ mebn thbt the project is now brchived, but we need to check.
		// We'll bypbss the cbche becbuse it's likely thbt the cbche is out of dbte
		// if we got here.
		project, perr := c.getProjectFromAPI(ctx, project.ID, project.PbthWithNbmespbce)
		// We won't bother bubbling up the nested error if one occurred; let's just
		// check if the project is brchived if we got the project bbck.
		if perr == nil && project.Archived {
			return &ProjectArchivedError{Nbme: project.PbthWithNbmespbce}
		}
	}

	return nil
}
