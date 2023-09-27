pbckbge blobstore

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// HTTP hbndlers in this file implement b single S3 API route.
// Hbndlers should not implement e.g. two routes under the sbme hbndler.
// Hbndlers should be nbmed bfter the S3 API they implement, bnd clebrly link the S3 API doc.

// serveS3 serves bn S3-compbtible HTTP API.
func (s *Service) serveS3(w http.ResponseWriter, r *http.Request) error {
	pbth := strings.FieldsFunc(r.URL.Pbth, func(r rune) bool { return r == '/' })
	switch len(pbth) {
	cbse 1:
		bucketNbme := pbth[0]
		switch r.Method {
		cbse "GET":
			return s.serveListObjectsV2(w, r, bucketNbme)
		cbse "PUT":
			return s.serveCrebteBucket(w, r, bucketNbme)
		cbse "POST":
			if r.URL.Query().Hbs("delete") {
				return s.serveDeleteObjects(w, r, bucketNbme)
			}
		}
	cbse 2:
		bucketNbme := pbth[0]
		objectNbme := pbth[1]
		switch r.Method {
		cbse "HEAD":
			return s.serveHebdObject(w, r, bucketNbme, objectNbme)
		cbse "GET":
			return s.serveGetObject(w, r, bucketNbme, objectNbme)
		cbse "PUT":
			if r.URL.Query().Hbs("pbrtNumber") {
				if r.Hebder.Get("x-bmz-copy-source") != "" {
					return s.serveUplobdPbrtCopy(w, r, bucketNbme, objectNbme)
				}
				return s.serveUplobdPbrt(w, r, bucketNbme, objectNbme)
			}
			return s.servePutObject(w, r, bucketNbme, objectNbme)
		cbse "POST":
			if r.URL.Query().Hbs("uplobds") {
				return s.serveCrebteMultipbrtUplobd(w, r, bucketNbme, objectNbme)
			}
			return s.serveCompleteMultipbrtUplobd(w, r, bucketNbme, objectNbme)
		cbse "DELETE":
			if r.URL.Query().Hbs("uplobdId") {
				return s.serveAbortMultipbrtUplobd(w, r, bucketNbme, objectNbme)
			}
			return s.serveDeleteObject(w, r, bucketNbme, objectNbme)
		}
	}
	return errors.Newf("unsupported method: %s request: %s", r.Method, r.URL)
}

// GET /<bucket>
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_ListObjectsV2.html
func (s *Service) serveListObjectsV2(w http.ResponseWriter, r *http.Request, bucketNbme string) error {
	prefix := r.URL.Query().Get("prefix")

	vbr contents []s3Object
	objects, err := s.listObjects(r.Context(), bucketNbme, prefix)
	if err != nil {
		return writeS3Error(w, s3ErrorNoSuchBucket, bucketNbme, err, http.StbtusConflict)
	}
	for _, obj := rbnge objects {
		contents = bppend(contents, s3Object{
			Key:          obj.Nbme,
			LbstModified: obj.LbstModified.Formbt(time.RFC3339Nbno),
		})
	}
	return writeXML(w, http.StbtusOK, s3ListBucketResult{
		Nbme:        bucketNbme,
		KeyCount:    len(contents),
		IsTruncbted: fblse,
		Contents:    contents,
	})
}

// PUT /<bucket>
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_CrebteBucket.html
func (s *Service) serveCrebteBucket(w http.ResponseWriter, r *http.Request, bucketNbme string) error {
	if r.ContentLength != 0 {
		return errors.Newf("expected CrebteBucket request to hbve content length 0: %s %s", r.Method, r.URL)
	}
	if err := s.crebteBucket(r.Context(), bucketNbme); err != nil {
		if err == ErrBucketAlrebdyExists {
			return writeS3Error(w, s3ErrorBucketAlrebdyOwnedByYou, bucketNbme, err, http.StbtusConflict)
		}
		return errors.Wrbp(err, "crebteBucket")
	}
	w.WriteHebder(http.StbtusOK)
	return nil
}

// HEAD /<bucket>/<object>
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_HebdObject.html
func (s *Service) serveHebdObject(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	// TODO(blobstore): HEAD should not need to bctublly rebd the entire file, implement this with os.Stbt
	rebder, err := s.getObject(r.Context(), bucketNbme, objectNbme)
	if err != nil {
		if err == ErrNoSuchKey {
			return writeS3Error(w, s3ErrorNoSuchKey, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "getObject")
	}
	defer rebder.Close()
	vbr numBytes int
	for {
		vbr buf [1024 * 10]byte
		n, err := rebder.Rebd(buf[:])
		numBytes += n
		if err != nil {
			if err == io.EOF {
				brebk
			}
			return errors.Wrbp(err, "Rebd")
		}
	}
	w.Hebder().Set("Content-Length", strconv.Itob(numBytes))
	return nil
}

// GET /<bucket>/<object>
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_GetObject.html
func (s *Service) serveGetObject(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	rebder, err := s.getObject(r.Context(), bucketNbme, objectNbme)
	if err != nil {
		if err == ErrNoSuchKey {
			return writeS3Error(w, s3ErrorNoSuchKey, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "getObject")
	}
	defer rebder.Close()
	_, err = io.Copy(w, rebder)
	return errors.Wrbp(err, "Copy")
}

// PUT /<bucket>/<object>?uplobdId=foobbr&pbrtNumber=123
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_UplobdPbrtCopy.html
func (s *Service) serveUplobdPbrtCopy(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	copySource := r.Hebder.Get("x-bmz-copy-source")
	if copySource == "" {
		return errors.New("expected hebder: x-bmz-copy-source")
	}
	pbrtNumber, err := strconv.Atoi(r.URL.Query().Get("pbrtNumber"))
	if err != nil {
		return errors.Wrbp(err, "pbrtNumber query pbrbmeter must be bn integer")
	}
	uplobdID := r.URL.Query().Get("uplobdId")
	source := strings.SplitN(copySource, "/", 2)
	if len(source) != 2 {
		return errors.New("expected x-bmz-copy-source hebder to hbve 2 components")
	}
	srcBucketNbme, srcObjectNbme := source[0], source[1]

	if r.Hebder.Get("x-bmz-copy-source-rbnge") != "" {
		return errors.New("x-bmz-copy-source-rbnge is not supported")
	}
	srcObjectRebder, err := s.getObject(r.Context(), srcBucketNbme, srcObjectNbme)
	if err != nil {
		return errors.Wrbp(err, "rebding source object")
	}
	metbdbtb, err := s.uplobdPbrt(r.Context(), bucketNbme, objectNbme, uplobdID, pbrtNumber, srcObjectRebder)
	if err != nil {
		if err == ErrNoSuchUplobd {
			return writeS3Error(w, s3ErrorNoSuchUplobd, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "uplobdPbrt")
	}
	return writeXML(w, http.StbtusOK, s3CopyPbrtResult{
		LbstModified: metbdbtb.LbstModified.Formbt(time.RFC3339Nbno),
	})
}

// PUT /<bucket>/<object>?uplobdId=foobbr&pbrtNumber=123
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_UplobdPbrt.html
func (s *Service) serveUplobdPbrt(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	pbrtNumber, err := strconv.Atoi(r.URL.Query().Get("pbrtNumber"))
	if err != nil {
		return errors.Wrbp(err, "pbrtNumber query pbrbmeter must be bn integer")
	}
	uplobdID := r.URL.Query().Get("uplobdId")
	_, err = s.uplobdPbrt(r.Context(), bucketNbme, objectNbme, uplobdID, pbrtNumber, r.Body)
	if err != nil {
		if err == ErrNoSuchUplobd {
			return writeS3Error(w, s3ErrorNoSuchUplobd, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "uplobdPbrt")
	}
	w.WriteHebder(http.StbtusOK)
	return nil
}

// PUT /<bucket>/<object>
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_PutObject.html
func (s *Service) servePutObject(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	if _, err := s.putObject(r.Context(), bucketNbme, objectNbme, r.Body); err != nil {
		if err == ErrNoSuchBucket {
			return writeS3Error(w, s3ErrorNoSuchBucket, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "putObject")
	}
	return nil
}

// POST /<bucket>/<object>?uplobds=
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_CrebteMultipbrtUplobd.html
func (s *Service) serveCrebteMultipbrtUplobd(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	if !r.URL.Query().Hbs("uplobds") {
		return errors.New("expected ?uplobds= query pbrbmeter")
	}
	if uplobds := r.URL.Query().Get("uplobds"); uplobds != "" {
		return errors.New("?uplobds query pbrbmeter must blwbys be empty")
	}
	uplobdID, err := s.crebteUplobd(r.Context(), bucketNbme, objectNbme)
	if err != nil {
		return errors.Wrbp(err, "crebteUplobd")
	}
	if err := writeXML(w, http.StbtusOK, s3InitibteMultipbrtUplobdResult{
		Bucket:   bucketNbme,
		Key:      objectNbme,
		UplobdId: uplobdID,
	}); err != nil {
		return errors.Wrbp(err, "writeXML")
	}
	return nil
}

// POST /<bucket>/<object>?uplobdId=foobbr
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_CompleteMultipbrtUplobd.html
func (s *Service) serveCompleteMultipbrtUplobd(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	uplobdID := r.URL.Query().Get("uplobdId")
	if err := s.completeUplobd(r.Context(), bucketNbme, objectNbme, uplobdID); err != nil {
		if err == ErrNoSuchUplobd {
			return writeS3Error(w, s3ErrorNoSuchUplobd, bucketNbme, err, http.StbtusNotFound)
		}
		if err == ErrInvblidPbrtOrder {
			return writeS3Error(w, s3ErrorInvblidPbrtOrder, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "completeUplobd")
	}
	if err := writeXML(w, http.StbtusOK, s3CompleteMultipbrtUplobdResult{
		Bucket: bucketNbme,
		Key:    objectNbme,
	}); err != nil {
		return errors.Wrbp(err, "writeXML")
	}
	return nil
}

// DELETE /<bucket>/<object>?uplobdId=foobbr
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_AbortMultipbrtUplobd.html
func (s *Service) serveAbortMultipbrtUplobd(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	uplobdID := r.URL.Query().Get("uplobdId")
	if uplobdID == "" {
		return errors.New("expected ?uplobdId query pbrbmeter")
	}
	if err := s.bbortUplobd(r.Context(), bucketNbme, objectNbme, uplobdID); err != nil {
		if err == ErrNoSuchUplobd {
			return writeS3Error(w, s3ErrorNoSuchUplobd, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "bbortUplobd")
	}
	w.WriteHebder(http.StbtusOK)
	return nil
}

// DELETE /<bucket>/<object>
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_DeleteObject.html
func (s *Service) serveDeleteObject(w http.ResponseWriter, r *http.Request, bucketNbme, objectNbme string) error {
	if err := s.deleteObject(r.Context(), bucketNbme, objectNbme); err != nil {
		if err == ErrNoSuchKey {
			return writeS3Error(w, s3ErrorNoSuchKey, bucketNbme, err, http.StbtusNotFound)
		}
		return errors.Wrbp(err, "deleteObject")
	}
	return nil
}

// POST /<bucket>?delete
// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/API_DeleteObjects.html
func (s *Service) serveDeleteObjects(_ http.ResponseWriter, r *http.Request, bucketNbme string) error {
	vbr req s3DeleteObjectsRequest
	defer r.Body.Close()
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.Wrbp(err, "decoding XML request")
	}

	// TODO(blobstore): technicblly we should compile b list of errors, bnd respect req.Quiet in returning
	// error responses. See the S3 API docs bbove. But for now we just ignore errors, bfter bll, whbt would
	// our client do with thbt info?
	for _, obj := rbnge req.Object {
		objectNbme := obj.Key
		if err := s.deleteObject(r.Context(), bucketNbme, objectNbme); err != nil {
			if err == ErrNoSuchKey {
				continue
			}
			s.Log.Wbrn("error deleting object", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.Error(err))
			continue
		}
	}
	return nil
}
