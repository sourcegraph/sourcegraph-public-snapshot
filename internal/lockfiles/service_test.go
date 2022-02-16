package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestService_FetchDependencies(t *testing.T) {
	s := &Service{
		GitArchive: func(c context.Context, repo api.RepoName, ao gitserver.ArchiveOptions) (io.ReadCloser, error) {
			var b bytes.Buffer
			zw := zip.NewWriter(&b)
			defer zw.Close()

			for file, data := range map[string]string{
				"client/package-lock.json": `{"dependencies": { "tree-sitter-cli": {"version": "0.20.4"} }}`,
				"web/package-lock.json":    `{"dependencies": { "nan": {"version": "2.15.0"} }}`,
			} {
				w, err := zw.Create(file)
				if err != nil {
					t.Fatal(err)
				}

				_, err = w.Write([]byte(data))
				if err != nil {
					t.Fatal(err)
				}
			}

			return io.NopCloser(&b), nil
		},
	}

	ctx := context.Background()
	got, err := s.FetchDependencies(ctx, "foo", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	want := []*Dependency{
		{Name: "nan", Version: "2.15.0", Kind: KindNPM},
		{Name: "tree-sitter-cli", Version: "0.20.4", Kind: KindNPM},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("dependency mismatch (-want +got):\n%s", diff)
	}
}
