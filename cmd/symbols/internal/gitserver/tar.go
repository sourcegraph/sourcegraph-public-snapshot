package gitserver

import (
	"archive/tar"
	"bytes"
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func CreateTestFetchTarFunc(tarContents map[string]string) func(context.Context, api.RepoName, api.CommitID, []string) (io.ReadCloser, error) {
	return func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
		var buffer bytes.Buffer
		tarWriter := tar.NewWriter(&buffer)

		for name, content := range tarContents {
			if paths != nil {
				found := false
				for _, path := range paths {
					if path == name {
						found = true
					}
				}
				if !found {
					continue
				}
			}

			tarHeader := &tar.Header{
				Name: name,
				Mode: 0o600,
				Size: int64(len(content)),
			}
			if err := tarWriter.WriteHeader(tarHeader); err != nil {
				return nil, err
			}
			if _, err := tarWriter.Write([]byte(content)); err != nil {
				return nil, err
			}
		}

		if err := tarWriter.Close(); err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
}
