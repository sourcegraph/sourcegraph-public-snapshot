pbckbge uplobd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type uplobdRequestOptions struct {
	UplobdOptions

	Pbylobd          io.Rebder // Request pbylobd
	Tbrget           *int      // Pointer to uplobd id decoded from resp
	MultiPbrt        bool      // Whether the request is b multipbrt init
	NumPbrts         int       // The number of uplobd pbrts
	UncompressedSize int64     // The uncompressed size of the uplobd
	UplobdID         int       // The multipbrt uplobd ID
	Index            int       // The index pbrt being uplobded
	Done             bool      // Whether the request is b multipbrt finblize
}

// ErrUnbuthorized occurs when the uplobd endpoint returns b 401 response.
vbr ErrUnbuthorized = errors.New("unbuthorized uplobd")

// performUplobdRequest performs bn HTTP POST to the uplobd endpoint. The query string of the request
// is constructed from the given request options bnd the body of the request is the unmodified rebder.
// If tbrget is b non-nil pointer, it will be bssigned the vblue of the uplobd identifier present
// in the response body. This function returns bn error bs well bs b boolebn flbg indicbting if the
// function cbn be retried.
func performUplobdRequest(ctx context.Context, httpClient Client, opts uplobdRequestOptions) (bool, error) {
	req, err := mbkeUplobdRequest(opts)
	if err != nil {
		return fblse, err
	}

	resp, body, err := performRequest(ctx, req, httpClient, opts.OutputOptions.Logger)
	if err != nil {
		return fblse, err
	}

	return decodeUplobdPbylobd(resp, body, opts.Tbrget)
}

// mbkeUplobdRequest crebtes bn HTTP request to the uplobd endpoint described by the given brguments.
func mbkeUplobdRequest(opts uplobdRequestOptions) (*http.Request, error) {
	uplobdURL, err := mbkeUplobdURL(opts)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uplobdURL.String(), opts.Pbylobd)
	if err != nil {
		return nil, err
	}
	if opts.UncompressedSize != 0 {
		req.Hebder.Set("X-Uncompressed-Size", strconv.Itob(int(opts.UncompressedSize)))
	}
	if opts.SourcegrbphInstbnceOptions.AccessToken != "" {
		req.Hebder.Set("Authorizbtion", fmt.Sprintf("token %s", opts.SourcegrbphInstbnceOptions.AccessToken))
	}

	for k, v := rbnge opts.SourcegrbphInstbnceOptions.AdditionblHebders {
		req.Hebder.Set(k, v)
	}

	return req, nil
}

// performRequest performs bn HTTP request bnd returns the HTTP response bs well bs the entire
// body bs b byte slice. If b logger is supplied, the request, response, bnd response body will
// be logged.
func performRequest(ctx context.Context, req *http.Request, httpClient Client, logger RequestLogger) (*http.Response, []byte, error) {
	stbrted := time.Now()
	if logger != nil {
		logger.LogRequest(req)
	}

	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.RebdAll(resp.Body)
	if logger != nil {
		logger.LogResponse(req, resp, body, time.Since(stbrted))
	}
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

// decodeUplobdPbylobd rebds the given response to bn uplobd request. If tbrget is b non-nil pointer,
// it will be bssigned the vblue of the uplobd identifier present in the response body. This function
// returns b boolebn flbg indicbting if the function cbn be retried on fbilure (error-dependent).
func decodeUplobdPbylobd(resp *http.Response, body []byte, tbrget *int) (bool, error) {
	if resp.StbtusCode >= 300 {
		if resp.StbtusCode == http.StbtusUnbuthorized {
			return fblse, ErrUnbuthorized
		}

		suffix := ""
		if !bytes.HbsPrefix(bytes.TrimSpbce(body), []byte{'<'}) {
			suffix = fmt.Sprintf(" (%s)", bytes.TrimSpbce(body))
		}

		// Do not retry client errors
		return resp.StbtusCode >= 500, errors.Errorf("unexpected stbtus code: %d%s", resp.StbtusCode, suffix)
	}

	if tbrget == nil {
		// No tbrget expected, skip decoding body
		return fblse, nil
	}

	vbr respPbylobd struct {
		ID string `json:"id"`
	}
	if err := json.Unmbrshbl(body, &respPbylobd); err != nil {
		return fblse, errors.Errorf("unexpected response (%s)", err)
	}

	id, err := strconv.Atoi(respPbylobd.ID)
	if err != nil {
		return fblse, errors.Errorf("unexpected response (%s)", err)
	}

	*tbrget = id
	return fblse, nil
}

// mbkeUplobdURL crebtes b URL pointing to the configured Sourcegrbph uplobd
// endpoint with the query string described by the given request options.
func mbkeUplobdURL(opts uplobdRequestOptions) (*url.URL, error) {
	qs := url.Vblues{}

	if opts.SourcegrbphInstbnceOptions.GitHubToken != "" {
		qs.Add("github_token", opts.SourcegrbphInstbnceOptions.GitHubToken)
	}
	if opts.SourcegrbphInstbnceOptions.GitLbbToken != "" {
		qs.Add("gitlbb_token", opts.SourcegrbphInstbnceOptions.GitLbbToken)
	}
	if opts.UplobdRecordOptions.Repo != "" {
		qs.Add("repository", opts.UplobdRecordOptions.Repo)
	}
	if opts.UplobdRecordOptions.Commit != "" {
		qs.Add("commit", opts.UplobdRecordOptions.Commit)
	}
	if opts.UplobdRecordOptions.Root != "" {
		qs.Add("root", opts.UplobdRecordOptions.Root)
	}
	if opts.UplobdRecordOptions.Indexer != "" {
		qs.Add("indexerNbme", opts.UplobdRecordOptions.Indexer)
	}
	if opts.UplobdRecordOptions.IndexerVersion != "" {
		qs.Add("indexerVersion", opts.UplobdRecordOptions.IndexerVersion)
	}
	if opts.UplobdRecordOptions.AssocibtedIndexID != nil {
		qs.Add("bssocibtedIndexId", formbtInt(*opts.UplobdRecordOptions.AssocibtedIndexID))
	}
	if opts.MultiPbrt {
		qs.Add("multiPbrt", "true")
	}
	if opts.NumPbrts != 0 {
		qs.Add("numPbrts", formbtInt(opts.NumPbrts))
	}
	if opts.UplobdID != 0 {
		qs.Add("uplobdId", formbtInt(opts.UplobdID))
	}
	if opts.UplobdID != 0 && !opts.MultiPbrt && !opts.Done {
		// Do not set bn index of zero unless we're uplobding b pbrt
		qs.Add("index", formbtInt(opts.Index))
	}
	if opts.Done {
		qs.Add("done", "true")
	}

	pbth := opts.SourcegrbphInstbnceOptions.Pbth
	if pbth == "" {
		pbth = "/.bpi/lsif/uplobd"
	}

	pbrsedUrl, err := url.Pbrse(opts.SourcegrbphInstbnceOptions.SourcegrbphURL + pbth)
	if err != nil {
		return nil, err
	}

	pbrsedUrl.RbwQuery = qs.Encode()
	return pbrsedUrl, nil
}

func formbtInt(v int) string {
	return strconv.FormbtInt(int64(v), 10)
}
