package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// BundleManagerClient is the interface to the precise-code-intel-bundle-manager service.
type BundleManagerClient interface {
	// BundleClient creates a client that can answer intelligence queries for a single dump.
	BundleClient(bundleID int) BundleClient

	// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk.
	SendUpload(ctx context.Context, bundleID int, r io.Reader) error
}

type bundleManagerClientImpl struct {
	bundleManagerURL string
}

var _ BundleManagerClient = &bundleManagerClientImpl{}

func New(bundleManagerURL string) BundleManagerClient {
	return &bundleManagerClientImpl{bundleManagerURL: bundleManagerURL}
}

// BundleClient creates a client that can answer intelligence queries for a single dump.
func (c *bundleManagerClientImpl) BundleClient(bundleID int) BundleClient {
	return &bundleClientImpl{
		bundleManagerURL: c.bundleManagerURL,
		bundleID:         bundleID,
	}
}

// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk.
func (c *bundleManagerClientImpl) SendUpload(ctx context.Context, bundleID int, r io.Reader) error {
	url, err := url.Parse(fmt.Sprintf("%s/uploads/%d", c.bundleManagerURL, bundleID))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url.String(), r)
	if err != nil {
		return err
	}

	// TODO - use context
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
