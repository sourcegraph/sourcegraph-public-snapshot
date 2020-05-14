package codeintel

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/hashicorp/go-multierror"
)

type UploadIndexOpts struct {
	Endpoint            string
	AccessToken         string
	Repo                string
	Commit              string
	Root                string
	Indexer             string
	GitHubToken         string
	File                string
	MaxPayloadSizeBytes int
}

// ErrUnauthorized occurs when the upload endpoint returns a 401 response.
var ErrUnauthorized = errors.New("unauthorized upload")

// UploadIndex uploads the given index file to the upload endpoint. If the upload
// file is large, it may be split into multiple chunks locally and uploaded over
// multiple requests.
func UploadIndex(opts UploadIndexOpts) (string, error) {
	fileInfo, err := os.Stat(opts.File)
	if err != nil {
		return "", err
	}

	if fileInfo.Size() <= int64(opts.MaxPayloadSizeBytes) {
		id, err := uploadIndex(opts)
		if err != nil {
			return "", err
		}
		return uploadIDToGraphQLID(id), nil
	}

	id, err := uploadMultipartIndex(opts)
	if err != nil {
		return "", err
	}
	return uploadIDToGraphQLID(id), nil
}

// uploadIndex constructs a GraphQL-compatible identifier from the raw identifier returned
// from the upload endpoint.
func uploadIDToGraphQLID(uploadID int) string {
	return string(base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(`LSIFUpload:"%d"`, uploadID))))
}

// uploadIndex performs a single request  to the upload endpoint. The new upload id is returned.
func uploadIndex(opts UploadIndexOpts) (id int, err error) {
	f, err := os.Open(opts.File)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	args := requestArgs{
		endpoint:    opts.Endpoint,
		accessToken: opts.AccessToken,
		repo:        opts.Repo,
		commit:      opts.Commit,
		root:        opts.Root,
		indexer:     opts.Indexer,
		gitHubToken: opts.GitHubToken,
	}

	if err := makeUploadRequest(args, gzipReader(f), &id); err != nil {
		return 0, err
	}

	return id, nil
}

// uploadMultipartIndex splits the index file into chunks small enough to upload, then
// performs a series of request to the upload endpoint. The new upload id is returned.
func uploadMultipartIndex(opts UploadIndexOpts) (id int, err error) {
	files, cleanup, err := splitFile(opts.File, opts.MaxPayloadSizeBytes)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = cleanup(err)
	}()

	setupArgs := requestArgs{
		endpoint:    opts.Endpoint,
		accessToken: opts.AccessToken,
		repo:        opts.Repo,
		commit:      opts.Commit,
		root:        opts.Root,
		indexer:     opts.Indexer,
		gitHubToken: opts.GitHubToken,
		multiPart:   true,
		numParts:    len(files),
	}
	if err := makeUploadRequest(setupArgs, nil, &id); err != nil {
		return 0, err
	}

	for i, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return 0, err
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				err = multierror.Append(err, closeErr)
			}
		}()

		uploadArgs := requestArgs{
			endpoint:    opts.Endpoint,
			accessToken: opts.AccessToken,
			uploadID:    id,
			index:       i,
		}
		if err := makeUploadRequest(uploadArgs, gzipReader(f), nil); err != nil {
			return 0, err
		}
	}

	finalizeArgs := requestArgs{
		endpoint:    opts.Endpoint,
		accessToken: opts.AccessToken,
		uploadID:    id,
		done:        true,
	}
	if err = makeUploadRequest(finalizeArgs, nil, nil); err != nil {
		return 0, err
	}

	return id, nil
}

// requestArgs are a superset of the values that can be supplied in the query string of the
// upload endpoint. The endpoint and access token fields must be set on every request, but the
// remaining fields must be set when appropriate by the caller of makeUploadRequest.
type requestArgs struct {
	endpoint    string
	accessToken string
	repo        string
	commit      string
	root        string
	indexer     string
	gitHubToken string
	multiPart   bool
	numParts    int
	uploadID    int
	index       int
	done        bool
}

// EncodeQuery constructs a query string from the args.
func (args requestArgs) EncodeQuery() string {
	qs := newQueryValues()
	qs.SetOptionalString("repository", args.repo)
	qs.SetOptionalString("commit", args.commit)
	qs.SetOptionalString("root", args.root)
	qs.SetOptionalString("indexerName", args.indexer)
	qs.SetOptionalString("github_token", args.gitHubToken)
	qs.SetOptionalBool("multiPart", args.multiPart)
	qs.SetOptionalInt("numParts", args.numParts)
	qs.SetOptionalInt("uploadId", args.uploadID)
	qs.SetOptionalBool("done", args.done)

	// Do not set an index of zero unless we're uploading a part
	if args.uploadID != 0 && !args.multiPart && !args.done {
		qs.SetInt("index", args.index)
	}

	return qs.Encode()
}

// makeUploadRequest performs an HTTP POST to the upload endpoint. The query string of the request
// is constructed from the given request args and the body of the request is the unmodified reader.
// If target is a non-nil pointer, it will be assigned the value of the upload identifier present
// in the response body.
func makeUploadRequest(args requestArgs, payload io.Reader, target *int) error {
	url, err := url.Parse(args.endpoint + "/.api/lsif/upload")
	if err != nil {
		return err
	}
	url.RawQuery = args.EncodeQuery()

	req, err := http.NewRequest("POST", url.String(), payload)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-ndjson+lsif")
	if args.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", args.accessToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}

		return fmt.Errorf("unexpected status code: %d\n\n%s", resp.StatusCode, body)
	}

	if target != nil {
		var respPayload struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(body, &respPayload); err != nil {
			return err
		}

		id, err := strconv.Atoi(respPayload.ID)
		if err != nil {
			return err
		}

		*target = id
	}

	return nil
}
