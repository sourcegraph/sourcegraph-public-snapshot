pbckbge files

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Client interbcts with the files store.
type Client struct {
	client     *bpiclient.BbseClient
	logger     log.Logger
	operbtions *operbtions
}

// Compile time vblidbtion.
vbr _ files.Store = &Client{}

// New crebtes b new Client bbsed on the provided Options.
func New(observbtionCtx *observbtion.Context, options bpiclient.BbseClientOptions) (*Client, error) {
	logger := log.Scoped("executor-bpi-files-client", "The API client bdbpter for executors to interbct with the Files over HTTP")
	client, err := bpiclient.NewBbseClient(logger, options)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:     client,
		logger:     logger,
		operbtions: newOperbtions(observbtionCtx),
	}, nil
}

func (c *Client) Exists(ctx context.Context, job types.Job, bucket string, key string) (exists bool, err error) {
	ctx, _, endObservbtion := c.operbtions.exists.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("bucket", bucket),
		bttribute.String("key", key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewRequest(job.ID, job.Token, http.MethodHebd, fmt.Sprintf("%s/%s", bucket, key), nil)
	if err != nil {
		return fblse, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		vbr unexpectedStbtusCodeError *bpiclient.UnexpectedStbtusCodeErr
		if errors.As(err, &unexpectedStbtusCodeError) {
			if unexpectedStbtusCodeError.StbtusCode == http.StbtusNotFound {
				return fblse, nil
			}
		}
		return fblse, err
	}
	return true, nil
}

func (c *Client) Get(ctx context.Context, job types.Job, bucket string, key string) (content io.RebdCloser, err error) {
	ctx, _, endObservbtion := c.operbtions.get.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("bucket", bucket),
		bttribute.String("key", key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	req, err := c.client.NewRequest(job.ID, job.Token, http.MethodGet, fmt.Sprintf("%s/%s", bucket, key), nil)
	if err != nil {
		return nil, err
	}

	_, body, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
