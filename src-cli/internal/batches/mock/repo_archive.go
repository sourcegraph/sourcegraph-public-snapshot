package mock

import (
	"archive/zip"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"testing"
)

type RepoArchive struct {
	RepoName string
	Commit   string
	Path     string
	Files    map[string]string
}

func NewZipArchivesMux(t *testing.T, callback http.HandlerFunc, archives ...RepoArchive) *http.ServeMux {
	mux := http.NewServeMux()

	for _, archive := range archives {
		files := archive.Files
		path := fmt.Sprintf("/%s@%s/-/raw", archive.RepoName, archive.Commit)
		if archive.Path != "" {
			path = path + "/" + archive.Path
		}

		downloadName := filepath.Base(archive.RepoName)
		mediaType := mime.FormatMediaType("Attachment", map[string]string{
			"filename": downloadName,
		})

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", mediaType)

			zipWriter := zip.NewWriter(w)
			for name, body := range files {
				f, err := zipWriter.Create(name)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := f.Write([]byte(body)); err != nil {
					t.Errorf("failed to write body for %s to zip: %s", name, err)
				}

				if callback != nil {
					callback(w, r)
				}
			}
			if err := zipWriter.Close(); err != nil {
				t.Fatalf("closing zipWriter failed: %s", err)
			}
		})
	}

	return mux
}

type middleware func(http.Handler) http.Handler

type MockRepoAdditionalFiles struct {
	RepoName        string
	Commit          string
	AdditionalFiles map[string]string
}

func HandleAdditionalFiles(mux *http.ServeMux, files MockRepoAdditionalFiles, middle middleware) {
	for name, content := range files.AdditionalFiles {
		path := fmt.Sprintf("/%s@%s/-/raw/%s", files.RepoName, files.Commit, name)
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")

			w.Write([]byte(content))
		}

		if middle != nil {
			mux.Handle(path, middle(http.HandlerFunc(handler)))
		} else {
			mux.HandleFunc(path, handler)
		}
	}
}
