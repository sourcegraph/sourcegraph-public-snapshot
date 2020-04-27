package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// BundleManagerClient is the interface to the precise-code-intel-bundle-manager service.
type BundleManagerClient interface {
	// BundleClient creates a client that can answer intelligence queries for a single dump.
	BundleClient(bundleID int) BundleClient

	// SendUpload transfers a raw LSIF upload to the bundle manager to be stored on disk.
	SendUpload(ctx context.Context, bundleID int, r io.Reader) error

	// GetUploads retrieves a raw LSIF upload from disk. The file is written to a file in the
	// given directory with a random filename. The generated filename is returned on success.
	GetUpload(ctx context.Context, bundleID int, dir string) (string, error)

	// SendDB transfers a converted databse to the bundle manager to be stored on disk.
	SendDB(ctx context.Context, bundleID int, r io.Reader) error
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
	body, err := c.request(ctx, "POST", fmt.Sprintf("uploads/%d", bundleID), r)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

// GetUploads retrieves a raw LSIF upload from disk. The file is written to a file in the
// given directory with a random filename. The generated filename is returned on success.
func (c *bundleManagerClientImpl) GetUpload(ctx context.Context, bundleID int, dir string) (_ string, err error) {
	body, err := c.request(ctx, "GET", fmt.Sprintf("uploads/%d", bundleID), nil)
	if err != nil {
		return "", err
	}
	defer body.Close()

	f, err := openRandomFile(dir)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	if _, err := io.Copy(f, body); err != nil {
		return "", err
	}

	return f.Name(), nil
}

// SendDB transfers a converted databse to the bundle manager to be stored on disk.
func (c *bundleManagerClientImpl) SendDB(ctx context.Context, bundleID int, r io.Reader) error {
	body, err := c.request(ctx, "POST", fmt.Sprintf("dbs/%d", bundleID), r)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}

func (c *bundleManagerClientImpl) request(ctx context.Context, method, path string, body io.Reader) (io.ReadCloser, error) {
	url, err := url.Parse(fmt.Sprintf("%s/%s", c.bundleManagerURL, path))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	// TODO - use context
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func openRandomFile(dir string) (*os.File, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return os.Create(filepath.Join(dir, uuid.String()))
}
