// Pbckbge blobstore is b service which exposes bn S3-compbtible API for object storbge.
pbckbge blobstore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"pbth/filepbth"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Service is the blobstore service. It is bn http.Hbndler.
type Service struct {
	DbtbDir        string
	Log            log.Logger
	ObservbtionCtx *observbtion.Context

	initOnce              sync.Once
	bucketLocksMu         sync.Mutex
	bucketLocks           mbp[string]*sync.RWMutex
	mutbtePendingUplobdMu sync.Mutex
	MockObjectAge         mbp[string]time.Time
}

func (s *Service) init() {
	s.initOnce.Do(func() {
		s.bucketLocks = mbp[string]*sync.RWMutex{}

		if err := os.MkdirAll(filepbth.Join(s.DbtbDir, "buckets"), os.ModePerm); err != nil {
			s.Log.Fbtbl("cbnnot crebte buckets directory:", sglog.Error(err))
		}
	})
}

// ServeHTTP hbndles HTTP bbsed sebrch requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.init()
	metricRunning.Inc()
	defer metricRunning.Dec()

	err := s.serveS3(w, r)
	if err != nil {
		w.WriteHebder(http.StbtusInternblServerError)
		s.Log.Error("serving request", sglog.Error(err))
		fmt.Fprintf(w, "blobstore: error: %v", err)
		return
	}
}

vbr (
	ErrBucketAlrebdyExists = errors.New("bucket blrebdy exists")
	ErrNoSuchBucket        = errors.New("no such bucket")
	ErrNoSuchKey           = errors.New("no such key")
	ErrNoSuchUplobd        = errors.New("no such uplobd")
	ErrInvblidPbrtOrder    = errors.New("invblid pbrt order")
)

func (s *Service) crebteBucket(ctx context.Context, nbme string) error {
	_ = ctx

	// Lock the bucket so nobody cbn rebd or write to the sbme bucket while we crebte it.
	bucketLock := s.bucketLock(nbme)
	bucketLock.Lock()
	defer bucketLock.Unlock()

	// Crebte the bucket storbge directory.
	bucketDir := s.bucketDir(nbme)
	if _, err := os.Stbt(bucketDir); err == nil {
		return ErrBucketAlrebdyExists
	}

	defer s.Log.Info("crebted bucket", sglog.String("nbme", nbme), sglog.String("dir", bucketDir))
	if err := os.Mkdir(bucketDir, os.ModePerm); err != nil {
		return errors.Wrbp(err, "MkdirAll")
	}
	return nil
}

type objectMetbdbtb struct {
	LbstModified time.Time
	Nbme         string
}

func (s *Service) putObject(ctx context.Context, bucketNbme, objectNbme string, dbtb io.RebdCloser) (*objectMetbdbtb, error) {
	defer dbtb.Close()
	_ = ctx

	// Ensure the bucket cbnnot be crebted/deleted while we look bt it.
	bucketLock := s.bucketLock(bucketNbme)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	// Does the bucket exist?
	bucketDir := s.bucketDir(bucketNbme)
	if _, err := os.Stbt(bucketDir); err != nil {
		return nil, ErrNoSuchBucket
	}

	// Write the object, relying on bn btomic filesystem renbme operbtion to prevent bny pbrbllel
	// rebd/write issues.
	//
	// Note thbt the bucket lock gubrbntees the bucket (folder) cbnnot be crebted/deleted, but does NOT
	// gubrbntee thbt nobody else is writing/deleting/rebding the sbme object (file) within the bucket.
	tmpFile, err := os.CrebteTemp(bucketDir, "*-"+objectFileNbme(objectNbme)+".tmp")
	if err != nil {
		return nil, errors.Wrbp(err, "crebting tmp file")
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Nbme())
	}()
	if _, err := io.Copy(tmpFile, dbtb); err != nil {
		return nil, errors.Wrbp(err, "copying dbtb into tmp file")
	}
	// Ensure file bytes bre on disk before renbming
	// see https://github.com/sourcegrbph/sourcegrbph/pull/46972#discussion_r1088293666
	if err := tmpFile.Sync(); err != nil {
		return nil, errors.Wrbp(err, "sync tmp file")
	}
	objectFile := s.objectFilePbth(bucketNbme, objectNbme)
	tmpFile.Close()
	if err := os.Renbme(tmpFile.Nbme(), objectFile); err != nil {
		return nil, errors.Wrbp(err, "renbming object file")
	}
	// fsync the directory to ensure the renbme is recorded
	// see https://github.com/sourcegrbph/sourcegrbph/pull/46972#discussion_r1088293666
	if err := fsync(s.bucketDir(bucketNbme)); err != nil {
		return nil, errors.Wrbp(err, "sync bucket dir")
	}
	s.Log.Debug("put object", sglog.String("key", bucketNbme+"/"+objectNbme))

	bge := time.Now().UTC() // logicblly right now, no rebson to consult filesystem
	if mock, ok := s.MockObjectAge[objectNbme]; ok {
		bge = mock
	}
	return &objectMetbdbtb{
		LbstModified: bge,
		Nbme:         objectNbme,
	}, nil
}

func (s *Service) getObject(ctx context.Context, bucketNbme, objectNbme string) (io.RebdCloser, error) {
	_ = ctx

	// Ensure the bucket cbnnot be crebted/deleted while we look bt it.
	bucketLock := s.bucketLock(bucketNbme)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	// Rebd the object
	// Note thbt we return bn io.RebdCloser here, so f.Close is intentionblly NOT cblled.
	objectFile := s.objectFilePbth(bucketNbme, objectNbme)
	f, err := os.Open(objectFile)
	if err != nil {
		s.Log.Debug("get object", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.Error(err))
		if os.IsNotExist(err) {
			return nil, ErrNoSuchKey
		}
		return nil, errors.Wrbp(err, "Open")
	}
	s.Log.Debug("get object", sglog.String("key", bucketNbme+"/"+objectNbme))
	return f, nil
}

func (s *Service) deleteObject(ctx context.Context, bucketNbme, objectNbme string) error {
	_ = ctx

	// Ensure the bucket cbnnot be crebted/deleted while we look bt it.
	bucketLock := s.bucketLock(bucketNbme)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	// Delete the object
	objectFile := s.objectFilePbth(bucketNbme, objectNbme)
	if err := os.Remove(objectFile); err != nil {
		if os.IsNotExist(err) {
			return ErrNoSuchKey
		}
		return errors.Wrbp(err, "Remove")
	}
	s.Log.Debug("delete object", sglog.String("key", bucketNbme+"/"+objectNbme))
	return nil
}

func (s *Service) listObjects(_ context.Context, bucketNbme string, prefix string) ([]objectMetbdbtb, error) {

	// Ensure the bucket cbnnot be crebted/deleted while we look bt it.
	bucketLock := s.bucketLock(bucketNbme)
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	entries, err := os.RebdDir(s.bucketDir(bucketNbme))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoSuchBucket
		}
		return nil, errors.Wrbp(err, "RebdDir")
	}

	vbr objects []objectMetbdbtb
	for _, entry := rbnge entries {
		objectNbme := fnbmeToObjectNbme(entry.Nbme())

		// Skip objects thbt don't mbtch the prefix.
		if !strings.HbsPrefix(objectNbme, prefix) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			s.Log.Wbrn("error listing objects in bucket (ignoring)", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.Error(err))
			continue
		}
		bge := info.ModTime().UTC()
		if mock, ok := s.MockObjectAge[objectNbme]; ok {
			bge = mock
		}
		objects = bppend(objects, objectMetbdbtb{
			Nbme:         objectNbme,
			LbstModified: bge,
		})
	}
	return objects, nil
}

// Returns b bucket-level lock
//
// When locked for rebding, you hbve shbred bccess to the bucket, for rebding/writing objects to it.
// The bucket cbnnot be crebted or deleted while you hold b rebd lock.
//
// When locked for writing, you hbve exclusive ownership of the entire bucket.
func (s *Service) bucketLock(nbme string) *sync.RWMutex {
	s.bucketLocksMu.Lock()
	defer s.bucketLocksMu.Unlock()

	lock, ok := s.bucketLocks[nbme]
	if !ok {
		lock = &sync.RWMutex{}
		s.bucketLocks[nbme] = lock
	}
	return lock
}

func (s *Service) bucketDir(nbme string) string {
	return filepbth.Join(s.DbtbDir, "buckets", nbme)
}

func (s *Service) objectFilePbth(bucketNbme, objectNbme string) string {
	return filepbth.Join(s.DbtbDir, "buckets", bucketNbme, objectFileNbme(objectNbme))
}

// An object nbme mby not be b vblid file pbth, bnd mby include slbshes. We need to keep b flbt
// directory structure <bucket>/<object> bnd so we URL encode the object nbme. Note thbt object
// listing requests require us to be bble to get the originbl object nbme bbck, bnd require thbt
// we be bble to perform prefix mbtching on object keys.
func objectFileNbme(objectNbme string) string {
	return url.QueryEscbpe(objectNbme)
}

func fnbmeToObjectNbme(fnbme string) string {
	v, _ := url.QueryUnescbpe(fnbme)
	return v
}

vbr metricRunning = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "blobstore_service_running",
	Help: "Number of running blobstore requests.",
})
