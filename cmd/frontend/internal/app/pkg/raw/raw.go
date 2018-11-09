package raw

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/golang/gddo/httputil"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

// ErrRequestHandled is returned by RepoProvider.GetRepoAndRev under various
// circumstances under which the HTTP request was handled / served.
var ErrRequestHandled = errors.New("request served")

// ErrRepoCloning is returned by RepoProvider.GetRepoAndRev when the repository
// being requested is cloning.
var ErrRepoCloning = errors.New("repository cloning")

// RepoProvider is used by the Handler to retrieve the repository and revision
// to use for serving the request.
//
// TODO: It really sucks that this logic is so %#@! complicated!
type RepoProvider interface {
	// GetRepoAndRev is invoked by the Handler to retrieve the repository and
	// revision to use for serving the request (as specified in the mux vars
	// "Repo" and "Rev").
	//
	// This method is responsible for serving the request (and returning
	// ErrRequestHandled) as appropriate when:
	//
	// - The user is not authorized to view the repository.
	// - The repository has been renamed.
	// - The repository does not exist.
	// - The repository is not clonable.
	// - The revision does not exist.
	//
	// Additionally, in the case of the repository cloning, it must return
	// ErrRepoCloning.
	GetRepoAndRev(w http.ResponseWriter, r *http.Request) (*types.Repo, api.CommitID, error)
}

// RepoProviderFunc is a function that implements the RepoProvider interface.
type RepoProviderFunc func(w http.ResponseWriter, r *http.Request) (*types.Repo, api.CommitID, error)

// GetRepoAndRev implements the RepoProvider interface.
func (rp RepoProviderFunc) GetRepoAndRev(w http.ResponseWriter, r *http.Request) (*types.Repo, api.CommitID, error) {
	return rp(w, r)
}

// Handler is an http.Handler that can return errors.
type Handler func(w http.ResponseWriter, r *http.Request) error

// NewHandler returns a handler which serves raw repository content requests.
// The same handler is exposed both via the UI and via the API routes because
// we want to serve both user (cookie / browser-based downloads) AND machine
// (none-cookie / access token) generated requests.
//
// Examples:
//
// Get a plaintext dir listing:
//     http://localhost:3080/github.com/gorilla/mux/-/raw/
//
// Get a file's contents (as text/plain, images will not be rendered by browsers):
//     http://localhost:3080/github.com/gorilla/mux/-/raw/mux.go
//     http://localhost:3080/github.com/sourcegraph/sourcegraph/-/raw/ui/assets/img/bg-hero.png
//
// Get a zip archive of a repository:
//     curl -H 'Accept: application/zip' http://localhost:3080/.api/github.com/gorilla/mux/-/raw/ -o repo.zip
//
// Get a tar archive of a repository:
//     curl -H 'Accept: application/x-tar' http://localhost:3080/.api/github.com/gorilla/mux/-/raw/ -o repo.tar
//
// Get a zip/tar archive of a _subdirectory_ of a repository:
//     curl -H 'Accept: application/zip' http://localhost:3080/.api/github.com/gorilla/mux/-/raw/.github -o repo-subdir.zip
//
// Get a zip/tar archive of a _file_ in a repository:
//     curl -H 'Accept: application/zip' http://localhost:3080/.api/github.com/gorilla/mux/-/raw/mux.go -o repo-file.zip
//
// Authenticate using an access token:
//     curl -H 'Accept: application/zip' http://localhost:3080/.api/github.com/gorilla/mux/-/raw/?token=fe70a9eeffc8ea7b1edf7c67095c143d1ada7e1b -o repo.zip
//
// Download an archive without specifying an Accept header (e.g. download via browser):
//     curl -O -J http://localhost:3080/github.com/gorilla/mux/-/raw?format=zip
//
// Known issues:
//
// - For security reasons, all non-archive files (e.g. code, images, binaries) are served with a Content-Type of text/plain.
// - Symlinks probably do not work well in the text/plain code path (i.e. when not requesting a zip/tar archive).
// - This route would ideally be using strict slashes, in order for us to support symlinks via HTTP redirects.
//
func NewHandler(repoProvider RepoProvider) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var (
			repo     *types.Repo
			commitID api.CommitID
			err      error
		)
		for {
			repo, commitID, err = repoProvider.GetRepoAndRev(w, r)
			if err == ErrRequestHandled {
				return nil
			}
			if err == ErrRepoCloning {
				time.Sleep(5 * time.Second)
				continue
			}
			if err != nil {
				return err
			}
			break
		}

		requestedPath := mux.Vars(r)["Path"]
		if !strings.HasPrefix(requestedPath, "/") {
			requestedPath = "/" + requestedPath
		}

		const (
			textPlain       = "text/plain"
			applicationZip  = "application/zip"
			applicationXTar = "application/x-tar"
		)

		// Negotiate the content type.
		contentTypeOffers := []string{textPlain, applicationZip, applicationXTar}
		defaultOffer := textPlain
		contentType := httputil.NegotiateContentType(r, contentTypeOffers, defaultOffer)

		// Allow users to override the negotiated content type so that e.g. browser
		// users can easily download tar/zip archives by adding ?format=zip etc. to
		// the URL.
		switch r.URL.Query().Get("format") {
		case "zip":
			contentType = applicationZip
		case "tar":
			contentType = applicationXTar
		}

		switch contentType {
		case applicationZip, applicationXTar:
			// Set the proper filename field, so that downloading "/github.com/gorilla/mux/-/raw"
			// gives us a "mux.zip" file (e.g. when downloading via a browser).
			ext := ".zip"
			if contentType == applicationXTar {
				ext = ".tar"
			}
			downloadName := path.Base(string(repo.Name)) + ext
			w.Header().Set("Content-Disposition", mime.FormatMediaType("Attachment", map[string]string{"filename": downloadName}))

			format := vfsutil.ArchiveFormatZip
			if contentType == applicationXTar {
				format = vfsutil.ArchiveFormatTar
			}
			relativePath := strings.TrimPrefix(requestedPath, "/")
			if relativePath == "" {
				relativePath = "."
			}
			f, _, err := vfsutil.GitServerFetchArchive(r.Context(), vfsutil.ArchiveOpts{
				Repo:         repo.Name,
				Commit:       commitID,
				Format:       format,
				RelativePath: relativePath,
			})
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(w, f)
			return err

		default:
			// This case also applies for defaultOffer. Note that this is preferred
			// over e.g. a 406 status code, according to the MDN:
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/406

			// 🚨 SECURITY: Files are served under the same Sourcegraph domain, and
			// may contain arbitrary contents (JS/HTML files, SVGs with JS in them,
			// malware in the form of .exe, etc). Serving with any other content
			// type is extremely dangerous unless we can guarantee the contents of
			// the file ourselves. GitHub, Wikipedia, and Facebook all use a
			// separate domain for exactly this reason (e.g. raw.githubusercontent.com).
			//
			// See also:
			//
			// - https://security.stackexchange.com/a/11779
			// - https://security.stackexchange.com/a/12916
			// - https://www.owasp.org/index.php/Unrestricted_File_Upload
			// - https://wiki.mozilla.org/WebAppSec/Secure_Coding_Guidelines#Uploads
			//
			// We try to protect against:
			//
			// - Serving user-uploaded malicious JS/HTML, SVGs with JS, etc. in a
			//   browser-interpreted form (not as literal text/plain content),
			//   which would introduce XSS, session-cookie stealing, etc.
			// - Serving user-uploaded malware, etc. which would flag our domain as
			//   untrustworthy by Google, etc. (We do serve such malware, but only
			//   with content type text/plain).
			//
			// We do NOT try to protect against:
			//
			// - Vulnerabilities in old browser versions / old IE versions that do
			//   not respect "nosniff".
			// - Vulnerabilities in Flash or Java (modern browsers should not run
			//   them).
			//
			// Note: We do not use a Content-Disposition attachment here because we
			// want files to be viewed in the browser only AND because doing so
			// would mean that we are literally serving malware to users
			// (i.e. browsers will auto-download it and not treat it as text).
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")

			archiveFS := vfsutil.NewGitServer(repo.Name, commitID)
			defer archiveFS.Close()
			fi, err := archiveFS.Lstat(r.Context(), requestedPath)
			if err != nil {
				return err
			}
			if fi.IsDir() {
				infos, err := archiveFS.ReadDir(r.Context(), requestedPath)
				if err != nil {
					return err
				}
				var names []string
				for _, info := range infos {
					names = append(names, info.Name())
				}
				result := strings.Join(names, "\n")
				fmt.Fprintf(w, "%s", template.HTMLEscapeString(result))
				return nil
			}

			// File
			f, err := archiveFS.Open(r.Context(), requestedPath)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(w, f)
			return err
		}
	}
}
