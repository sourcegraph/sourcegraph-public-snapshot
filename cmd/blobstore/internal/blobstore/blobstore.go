// Package blobstore is a service which exposes an S3-compatible API for object storage.
package blobstore

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

	initOnce      sync.Once
	bucketLocksMu sync.Mutex
	bucketLocks   map[string]*sync.RWMutex
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

	err := s.serve(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Error("serving request", sglog.Error(err))
		fmt.Fprintf(w, "blobstore: error: %v", err)
		return
	}
}

func (s *Service) serve(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	path := strings.FieldsFunc(r.URL.Path, func(r rune) bool { return r == '/' })
	switch r.Method {
	case "PUT":
		switch len(path) {
		case 1:
			// PUT /<bucket>
			// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateBucket.html
			if r.ContentLength != 0 {
				return errors.Newf("expected CreateBucket request to have content length 0: %s %s", r.Method, r.URL)
			}
			bucketName := path[0]
			if err := s.createBucket(ctx, bucketName); err != nil {
				if err == ErrBucketAlreadyExists {
					return writeS3Error(w, s3ErrorBucketAlreadyOwnedByYou, bucketName, err)
				}
				return errors.Wrap(err, "createBucket")
			}
			w.WriteHeader(http.StatusOK)
			return nil
		case 2:
			// PUT /<bucket>/<key>
			// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html
			// TODO(blobstore): implement me!
			return nil
		default:
			return errors.Newf("unsupported method: PUT request: %s", r.URL)
		}
	case "GET":
		if len(path) == 2 && r.URL.Query().Get("x-id") == "GetObject" {
			// GET /<bucket>/<key>?x-id=GetObject
			// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html
			// TODO(blobstore): implement me!
			w.WriteHeader(http.StatusNotFound)
			return nil
		}
		return errors.Newf("unexpected GET request: %s", r.URL)
	default:
		return errors.Newf("unexpected request: %s %s", r.Method, r.URL)
	}
}

var ErrBucketAlreadyExists = errors.New("bucket already exists")

func (s *Service) createBucket(ctx context.Context, name string) error {
	_ = ctx

	// Lock the bucket so nobody can read or write to the same bucket while we create it.
	bucketLock := s.bucketLock(name)
	bucketLock.Lock()
	defer bucketLock.Unlock()

	// Create the bucket storage directory.
	bucketDir := filepath.Join(s.DataDir, "buckets", name)
	if _, err := os.Stat(bucketDir); err == nil {
		return ErrBucketAlreadyExists
	}

	defer s.Log.Info("created bucket", sglog.String("name", name), sglog.String("dir", bucketDir))
	if err := os.Mkdir(bucketDir, os.ModePerm); err != nil {
		return errors.Wrap(err, "MkdirAll")
	}
	return nil
}

// returns a bucket-level lock which can be used for reading objects in a bucket, or in write-lock
// mode can be used to create or delete a bucket with the given name.
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

var (
	metricRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "blobstore_service_running",
		Help: "Number of running blobstore requests.",
	})
)
