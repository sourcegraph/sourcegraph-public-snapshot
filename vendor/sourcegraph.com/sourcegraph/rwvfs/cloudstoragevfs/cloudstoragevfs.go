package cloudstoragevfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"
	"strings"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/rwvfs"

	"context"

	"golang.org/x/oauth2/google"
	"golang.org/x/tools/godoc/vfs"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/storage/v1"
)

const dirKind = "storage#objects"

var defaultClient *http.Client
var defaultClientErr error
var defaultClientOnce sync.Once

// NewDefault returns a partial RWVFS implementation backed by Google Cloud Storage.
func NewDefault(bucket string) (rwvfs.FileSystem, error) {
	defaultClientOnce.Do(func() {
		// Authentication is provided by the gcloud tool when running locally, and
		// by the associated service account when running on Compute Engine.
		defaultClient, defaultClientErr = google.DefaultClient(context.Background(), storage.DevstorageFullControlScope)
	})
	if defaultClientErr != nil {
		return nil, fmt.Errorf("unable to get default client: %v", defaultClientErr)
	}
	return New(defaultClient, bucket)
}

// New returns a partial RWVFS implementation backed by the Cloud Storage API.
func New(client *http.Client, bucket string) (rwvfs.FileSystem, error) {
	service, err := storage.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create storage service: %v", err)
	}

	return &cloudStorageVFS{
		service: service,
		bucket:  bucket,
	}, nil
}

type cloudStorageVFS struct {
	service *storage.Service
	bucket  string
}

func (fs *cloudStorageVFS) String() string {
	return "cloudstorage(" + fs.bucket + ")"
}

func (fs *cloudStorageVFS) Lstat(path string) (os.FileInfo, error) {
	fi, err := fs.stat(path)
	if err != nil {
		err = &os.PathError{"lstat", path, err}
	}
	return fi, err
}

func (fs *cloudStorageVFS) Stat(path string) (os.FileInfo, error) {
	fi, err := fs.stat(path)
	if err != nil {
		err = &os.PathError{"stat", path, err}
	}
	return fi, err
}

func (fs *cloudStorageVFS) stat(path string) (os.FileInfo, error) {
	path = cleanPath(path)
	obj, err := fs.service.Objects.Get(fs.bucket, path).Do()
	if err != nil {
		if isNotFound(err) {
			objs, err := fs.service.Objects.List(fs.bucket).Prefix(path).Delimiter("/").MaxResults(1).Do()
			if err != nil {
				return nil, err
			}
			if len(objs.Prefixes) == 0 {
				return nil, os.ErrNotExist
			}
			return fileInfo{&storage.Object{
				Name: pathpkg.Base(path),
				Kind: dirKind,
			}}, nil
		}
		return nil, err
	}
	return fileInfo{obj}, nil
}

type fileInfo struct {
	obj *storage.Object
}

func (fi fileInfo) Name() string {
	return fi.obj.Name
}

func (fi fileInfo) Size() int64 {
	return int64(fi.obj.Size)
}

func (fi fileInfo) Mode() os.FileMode {
	if fi.IsDir() {
		return os.ModeDir
	}
	return 0
}

func (fi fileInfo) ModTime() time.Time {
	t, _ := time.Parse(time.RFC3339, fi.obj.Updated)
	return t
}

func (fi fileInfo) IsDir() bool {
	return fi.obj.Kind == dirKind
}

func (fi fileInfo) Sys() interface{} {
	return nil
}

func (fs *cloudStorageVFS) Create(path string) (io.WriteCloser, error) {
	r, w := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		_, err := fs.service.Objects.Insert(fs.bucket, &storage.Object{Name: cleanPath(path)}).Media(r).Do()
		errCh <- err
	}()
	return &upload{w, errCh}, nil
}

type upload struct {
	io.WriteCloser
	errCh chan error
}

func (u *upload) Close() error {
	if err := u.WriteCloser.Close(); err != nil {
		return err
	}
	return <-u.errCh
}

func (fs *cloudStorageVFS) Open(path string) (vfs.ReadSeekCloser, error) {
	resp, err := fs.service.Objects.Get(fs.bucket, cleanPath(path)).Download()
	if err != nil {
		if isNotFound(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return download{bytes.NewReader(data)}, nil
}

type download struct {
	io.ReadSeeker
}

func (download) Close() error {
	return nil
}

func (fs *cloudStorageVFS) Remove(path string) error {
	return fs.service.Objects.Delete(fs.bucket, cleanPath(path)).Do()
}

func (fs *cloudStorageVFS) Mkdir(path string) error {
	return nil // Cloud Storage has no need for creating directories
}

func (fs *cloudStorageVFS) ReadDir(path string) ([]os.FileInfo, error) {
	path = cleanPath(path)
	if path != "" {
		path += "/"
	}

	var infos []os.FileInfo
	var nextPageToken string
	for {
		call := fs.service.Objects.List(fs.bucket).Prefix(path).Delimiter("/")
		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}
		objs, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, obj := range objs.Items {
			obj.Name = strings.TrimPrefix(obj.Name, path)
			infos = append(infos, fileInfo{obj})
		}
		for _, prefix := range objs.Prefixes {
			infos = append(infos, fileInfo{&storage.Object{
				Name: strings.TrimPrefix(prefix, path),
				Kind: dirKind,
			}})
		}
		if objs.NextPageToken == "" {
			break
		}
		nextPageToken = objs.NextPageToken
	}
	return infos, nil
}

func isNotFound(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		return apiErr.Code == http.StatusNotFound
	}
	return false
}

func cleanPath(path string) string {
	return strings.Trim(pathpkg.Clean("/"+path), "/")
}
