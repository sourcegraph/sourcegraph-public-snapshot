// Package blobstore is a service which exposes an S3-compatible API for object storage.
package blobstore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Service is the blobstore service. It is an http.Handler.
type Service struct {
	DataDir        string
	Log            log.Logger
	ObservationCtx *observation.Context

	initOnce              sync.Once
	bucketLocksMu         sync.Mutex
	bucketLocks           map[string]*sync.RWMutex
	mutatePendingUploadMu sync.Mutex
	MockObjectAge         map[string]time.Time
}

func (s *Service) init() {
	s.initOnce.Do(func() {
		s.bucketLocks = map[string]*sync.RWMutex{}

		if err := os.MkdirAll(filepath.Join(s.DataDir, "buckets"), os.ModePerm); err != nil {
			s.Log.Fatal("cannot create buckets directory:", sglog.Error(err))
		}
	})
}

// ServeHTTP handles HTTP based search requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.init()
	metricRunning.Inc()
	defer metricRunning.Dec()

	err := s.serveS3(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Error("serving request", sglog.Error(err))
		fmt.Fprintf(w, "blobstore: error: %v", err)
		return
	}
}

var (
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrNoSuchBucket        = errors.New("no such bucket")
	ErrNoSuchKey           = errors.New("no such key")
	ErrNoSuchUpload        = errors.New("no such upload")
	ErrInvalidPartOrder    = errors.New("invalid part order")
)

func (s *Service) createBucket(ctx context.Context, name string) error {
	_ = ctx

	// Lock the bucket so nobody can read or write to the same bucket while we create it.
	bucketLock := s.bucketLock(name)
	bucketLock.Lock()
	defer bucketLock.Unlock()

	// Create the bucket storage directory.
	bucketDir := s.bucketDir(name)
	if _, err := os.Stat(bucketDir); err == nil {
		return ErrBucketAlreadyExists
	}

	defer s.Log.Info("created bucket", sglog.String("name", name), sglog.String("dir", bucketDir))
	if err := os.Mkdir(bucketDir, os.ModePerm); err != nil {
		return errors.Wrap(err, "MkdirAll")
	}
	return nil
}

type objectMetadata struct {
	LastModified time.Time
	Name         string
}

func (s *Service) putObject(ctx context.Context, bucketName, objectName string, data io.ReadCloser) (*objectMetadata, error) {
	defer data.Close()
	_ = ctx

	// Ensure the bucket cannot be created/deleted while we look at it.
	bucketLock := s.bucketLock(bucketName)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	// Does the bucket exist?
	bucketDir := s.bucketDir(bucketName)
	if _, err := os.Stat(bucketDir); err != nil {
		return nil, ErrNoSuchBucket
	}

	// Write the object, relying on an atomic filesystem rename operation to prevent any parallel
	// read/write issues.
	//
	// Note that the bucket lock guarantees the bucket (folder) cannot be created/deleted, but does NOT
	// guarantee that nobody else is writing/deleting/reading the same object (file) within the bucket.
	tmpFile, err := os.CreateTemp(bucketDir, "*-"+objectFileName(objectName)+".tmp")
	if err != nil {
		return nil, errors.Wrap(err, "creating tmp file")
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	if _, err := io.Copy(tmpFile, data); err != nil {
		return nil, errors.Wrap(err, "copying data into tmp file")
	}
	// Ensure file bytes are on disk before renaming
	// see https://github.com/sourcegraph/sourcegraph/pull/46972#discussion_r1088293666
	if err := tmpFile.Sync(); err != nil {
		return nil, errors.Wrap(err, "sync tmp file")
	}
	objectFile := s.objectFilePath(bucketName, objectName)
	tmpFile.Close()
	if err := os.Rename(tmpFile.Name(), objectFile); err != nil {
		return nil, errors.Wrap(err, "renaming object file")
	}
	// fsync the directory to ensure the rename is recorded
	// see https://github.com/sourcegraph/sourcegraph/pull/46972#discussion_r1088293666
	if err := fsync(s.bucketDir(bucketName)); err != nil {
		return nil, errors.Wrap(err, "sync bucket dir")
	}
	s.Log.Debug("put object", sglog.String("key", bucketName+"/"+objectName))

	age := time.Now().UTC() // logically right now, no reason to consult filesystem
	if mock, ok := s.MockObjectAge[objectName]; ok {
		age = mock
	}
	return &objectMetadata{
		LastModified: age,
		Name:         objectName,
	}, nil
}

func (s *Service) getObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	_ = ctx

	// Ensure the bucket cannot be created/deleted while we look at it.
	bucketLock := s.bucketLock(bucketName)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	// Read the object
	// Note that we return an io.ReadCloser here, so f.Close is intentionally NOT called.
	objectFile := s.objectFilePath(bucketName, objectName)
	f, err := os.Open(objectFile)
	if err != nil {
		s.Log.Debug("get object", sglog.String("key", bucketName+"/"+objectName), sglog.Error(err))
		if os.IsNotExist(err) {
			return nil, ErrNoSuchKey
		}
		return nil, errors.Wrap(err, "Open")
	}
	s.Log.Debug("get object", sglog.String("key", bucketName+"/"+objectName))
	return f, nil
}

func (s *Service) deleteObject(ctx context.Context, bucketName, objectName string) error {
	_ = ctx

	// Ensure the bucket cannot be created/deleted while we look at it.
	bucketLock := s.bucketLock(bucketName)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	// Delete the object
	objectFile := s.objectFilePath(bucketName, objectName)
	if err := os.Remove(objectFile); err != nil {
		if os.IsNotExist(err) {
			return ErrNoSuchKey
		}
		return errors.Wrap(err, "Remove")
	}
	s.Log.Debug("delete object", sglog.String("key", bucketName+"/"+objectName))
	return nil
}

func (s *Service) listObjects(_ context.Context, bucketName string, prefix string) ([]objectMetadata, error) {

	// Ensure the bucket cannot be created/deleted while we look at it.
	bucketLock := s.bucketLock(bucketName)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	entries, err := os.ReadDir(s.bucketDir(bucketName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoSuchBucket
		}
		return nil, errors.Wrap(err, "ReadDir")
	}

	var objects []objectMetadata
	for _, entry := range entries {
		objectName := fnameToObjectName(entry.Name())

		// Skip objects that don't match the prefix.
		if !strings.HasPrefix(objectName, prefix) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			s.Log.Warn("error listing objects in bucket (ignoring)", sglog.String("key", bucketName+"/"+objectName), sglog.Error(err))
			continue
		}
		age := info.ModTime().UTC()
		if mock, ok := s.MockObjectAge[objectName]; ok {
			age = mock
		}
		objects = append(objects, objectMetadata{
			Name:         objectName,
			LastModified: age,
		})
	}
	return objects, nil
}

// Returns a bucket-level lock
//
// When locked for reading, you have shared access to the bucket, for reading/writing objects to it.
// The bucket cannot be created or deleted while you hold a read lock.
//
// When locked for writing, you have exclusive ownership of the entire bucket.
func (s *Service) bucketLock(name string) *sync.RWMutex {
	s.bucketLocksMu.Lock()
	defer s.bucketLocksMu.Unlock()

	lock, ok := s.bucketLocks[name]
	if !ok {
		lock = &sync.RWMutex{}
		s.bucketLocks[name] = lock
	}
	return lock
}

func (s *Service) bucketDir(name string) string {
	return filepath.Join(s.DataDir, "buckets", name)
}

func (s *Service) objectFilePath(bucketName, objectName string) string {
	return filepath.Join(s.DataDir, "buckets", bucketName, objectFileName(objectName))
}

// An object name may not be a valid file path, and may include slashes. We need to keep a flat
// directory structure <bucket>/<object> and so we URL encode the object name. Note that object
// listing requests require us to be able to get the original object name back, and require that
// we be able to perform prefix matching on object keys.
func objectFileName(objectName string) string {
	return url.QueryEscape(objectName)
}

func fnameToObjectName(fname string) string {
	v, _ := url.QueryUnescape(fname)
	return v
}

var metricRunning = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "blobstore_service_running",
	Help: "Number of running blobstore requests.",
})
