package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/go-enry/go-enry/v2/regex"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type FileHandler struct {
	logger     sglog.Logger
	operations *Operations
	db         database.DB
}

func NewFileHandler(operations *Operations, db database.DB) *FileHandler {
	return &FileHandler{
		logger:     sglog.Scoped("FileHandler", "Embeddings file REST API handler"),
		operations: operations,
		db:         db,
	}
}

// This is a random guess on the max upload size to expect
const maxUploadSize = 10 << 20 // 10MB

func (h *FileHandler) Upload() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		statusCode, err := h.upload(r)

		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(statusCode)
		w.Write([]byte("Data upload successful"))
	})
}

func (h *FileHandler) upload(r *http.Request) (statusCode int, err error) {
	ctx := r.Context()

	if ok := isSiteAdmin(ctx, h.logger, h.db); !ok {
		statusCode = http.StatusForbidden
		err = errors.New("must be site admin")
		return
	}

	// ParseMultipartForm parses the whole request body and stores the max size into memory. The rest of the body is
	// stored in temporary files on disk. The reason for parsing the whole request in one go is because data cannot be
	// "streamed" or "appended" to the bytea type column. Data for the bytea column must be inserted in one go.
	//
	// When we move to using a blob store (Blobstore/S3/GCS), we can stream the parts instead. This means we won't need to
	// parse the entire request body up front. We will be able to iterate over and write the parts/chunks one at a time
	// - thus avoiding putting everything into memory.
	err = r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			statusCode = http.StatusBadRequest
			err = errors.New("request payload exceeds 10MB limit")
			return
		}
		statusCode = http.StatusInternalServerError
		err = errors.Wrap(err, "parsing request")
		return
	}

	err = h.uploadEmbeddingPluginFile(ctx, r)
	if err != nil {
		statusCode = http.StatusBadRequest
		return
	}

	return http.StatusOK, nil
}

var pathValidationRegex = regex.MustCompile("[.]{2}|[\\\\]")

const (
	ZipHeader = "PK"
	TarHeader = "ustar"
)

func (h *FileHandler) uploadEmbeddingPluginFile(ctx context.Context, r *http.Request) error {
	file, _, err := r.FormFile("archive")
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert multipart file into an io.ReaderAt
	ra, ok := file.(io.ReaderAt)
	if !ok {
		return errors.New("invalid archive")
	}

	var size int64
	if s, err := file.Seek(0, io.SeekEnd); err == nil {
		size = s
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return errors.New("unable to read archive")
	}

	// Read the file using zip.NewReader
	reader, err := zip.NewReader(ra, size)
	if err != nil {
		return err
	}

	return h.db.WithTransact(ctx, func(tx database.DB) error {
		store := tx.EmbeddingPluginFiles()

		// Iterate through the files in the archive, just print them out for now
		for _, f := range reader.File {
			// When you compress a directory in macOS using its built-in compress (zip) functionality,
			// it includes extra files that start with "._" (dot underscore), and they are usually stored inside a __MACOSX directory.
			// These "._" files are part of Apple's way of storing extra information about files.
			// These files store metadata that the macOS operating system uses for various purposes.
			if f.FileInfo().IsDir() || strings.HasPrefix(f.Name, "__MACOSX/") || pathValidationRegex.MatchString(f.Name) {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				return err
			}

			defer rc.Close()

			// Read the content into a buffer
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(rc)
			if err != nil {
				return err
			}

			_, err = store.Create(ctx, f.Name, buf.Bytes(), 1)
			if err != nil {
				return errors.Wrap(err, "error saving embedding plugin file to database")
			}
		}

		return nil
	})
}

func isSiteAdmin(ctx context.Context, logger sglog.Logger, db database.DB) bool {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}

		logger.Error("failed to get current user", sglog.Error(err))
		return false
	}

	return user != nil && user.SiteAdmin
}
