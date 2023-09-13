package blobstore

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// HTTP handlers in this file implement a single S3 API route.
// Handlers should not implement e.g. two routes under the same handler.
// Handlers should be named after the S3 API they implement, and clearly link the S3 API doc.

// serveS3 serves an S3-compatible HTTP API.
func (s *Service) serveS3(w http.ResponseWriter, r *http.Request) error {
	path := strings.FieldsFunc(r.URL.Path, func(r rune) bool { return r == '/' })
	switch len(path) {
	case 1:
		bucketName := path[0]
		switch r.Method {
		case "GET":
			return s.serveListObjectsV2(w, r, bucketName)
		case "PUT":
			return s.serveCreateBucket(w, r, bucketName)
		case "POST":
			if r.URL.Query().Has("delete") {
				return s.serveDeleteObjects(w, r, bucketName)
			}
		}
	case 2:
		bucketName := path[0]
		objectName := path[1]
		switch r.Method {
		case "HEAD":
			return s.serveHeadObject(w, r, bucketName, objectName)
		case "GET":
			return s.serveGetObject(w, r, bucketName, objectName)
		case "PUT":
			if r.URL.Query().Has("partNumber") {
				if r.Header.Get("x-amz-copy-source") != "" {
					return s.serveUploadPartCopy(w, r, bucketName, objectName)
				}
				return s.serveUploadPart(w, r, bucketName, objectName)
			}
			return s.servePutObject(w, r, bucketName, objectName)
		case "POST":
			if r.URL.Query().Has("uploads") {
				return s.serveCreateMultipartUpload(w, r, bucketName, objectName)
			}
			return s.serveCompleteMultipartUpload(w, r, bucketName, objectName)
		case "DELETE":
			if r.URL.Query().Has("uploadId") {
				return s.serveAbortMultipartUpload(w, r, bucketName, objectName)
			}
			return s.serveDeleteObject(w, r, bucketName, objectName)
		}
	}
	return errors.Newf("unsupported method: %s request: %s", r.Method, r.URL)
}

// GET /<bucket>
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html
func (s *Service) serveListObjectsV2(w http.ResponseWriter, r *http.Request, bucketName string) error {
	prefix := r.URL.Query().Get("prefix")

	var contents []s3Object
	objects, err := s.listObjects(r.Context(), bucketName, prefix)
	if err != nil {
		return writeS3Error(w, s3ErrorNoSuchBucket, bucketName, err, http.StatusConflict)
	}
	for _, obj := range objects {
		contents = append(contents, s3Object{
			Key:          obj.Name,
			LastModified: obj.LastModified.Format(time.RFC3339Nano),
		})
	}
	return writeXML(w, http.StatusOK, s3ListBucketResult{
		Name:        bucketName,
		KeyCount:    len(contents),
		IsTruncated: false,
		Contents:    contents,
	})
}

// PUT /<bucket>
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateBucket.html
func (s *Service) serveCreateBucket(w http.ResponseWriter, r *http.Request, bucketName string) error {
	if r.ContentLength != 0 {
		return errors.Newf("expected CreateBucket request to have content length 0: %s %s", r.Method, r.URL)
	}
	if err := s.createBucket(r.Context(), bucketName); err != nil {
		if err == ErrBucketAlreadyExists {
			return writeS3Error(w, s3ErrorBucketAlreadyOwnedByYou, bucketName, err, http.StatusConflict)
		}
		return errors.Wrap(err, "createBucket")
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

// HEAD /<bucket>/<object>
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_HeadObject.html
func (s *Service) serveHeadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	// TODO(blobstore): HEAD should not need to actually read the entire file, implement this with os.Stat
	reader, err := s.getObject(r.Context(), bucketName, objectName)
	if err != nil {
		if err == ErrNoSuchKey {
			return writeS3Error(w, s3ErrorNoSuchKey, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "getObject")
	}
	defer reader.Close()
	var numBytes int
	for {
		var buf [1024 * 10]byte
		n, err := reader.Read(buf[:])
		numBytes += n
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrap(err, "Read")
		}
	}
	w.Header().Set("Content-Length", strconv.Itoa(numBytes))
	return nil
}

// GET /<bucket>/<object>
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html
func (s *Service) serveGetObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	reader, err := s.getObject(r.Context(), bucketName, objectName)
	if err != nil {
		if err == ErrNoSuchKey {
			return writeS3Error(w, s3ErrorNoSuchKey, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "getObject")
	}
	defer reader.Close()
	_, err = io.Copy(w, reader)
	return errors.Wrap(err, "Copy")
}

// PUT /<bucket>/<object>?uploadId=foobar&partNumber=123
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_UploadPartCopy.html
func (s *Service) serveUploadPartCopy(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	copySource := r.Header.Get("x-amz-copy-source")
	if copySource == "" {
		return errors.New("expected header: x-amz-copy-source")
	}
	partNumber, err := strconv.Atoi(r.URL.Query().Get("partNumber"))
	if err != nil {
		return errors.Wrap(err, "partNumber query parameter must be an integer")
	}
	uploadID := r.URL.Query().Get("uploadId")
	source := strings.SplitN(copySource, "/", 2)
	if len(source) != 2 {
		return errors.New("expected x-amz-copy-source header to have 2 components")
	}
	srcBucketName, srcObjectName := source[0], source[1]

	if r.Header.Get("x-amz-copy-source-range") != "" {
		return errors.New("x-amz-copy-source-range is not supported")
	}
	srcObjectReader, err := s.getObject(r.Context(), srcBucketName, srcObjectName)
	if err != nil {
		return errors.Wrap(err, "reading source object")
	}
	metadata, err := s.uploadPart(r.Context(), bucketName, objectName, uploadID, partNumber, srcObjectReader)
	if err != nil {
		if err == ErrNoSuchUpload {
			return writeS3Error(w, s3ErrorNoSuchUpload, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "uploadPart")
	}
	return writeXML(w, http.StatusOK, s3CopyPartResult{
		LastModified: metadata.LastModified.Format(time.RFC3339Nano),
	})
}

// PUT /<bucket>/<object>?uploadId=foobar&partNumber=123
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_UploadPart.html
func (s *Service) serveUploadPart(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	partNumber, err := strconv.Atoi(r.URL.Query().Get("partNumber"))
	if err != nil {
		return errors.Wrap(err, "partNumber query parameter must be an integer")
	}
	uploadID := r.URL.Query().Get("uploadId")
	_, err = s.uploadPart(r.Context(), bucketName, objectName, uploadID, partNumber, r.Body)
	if err != nil {
		if err == ErrNoSuchUpload {
			return writeS3Error(w, s3ErrorNoSuchUpload, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "uploadPart")
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

// PUT /<bucket>/<object>
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html
func (s *Service) servePutObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	if _, err := s.putObject(r.Context(), bucketName, objectName, r.Body); err != nil {
		if err == ErrNoSuchBucket {
			return writeS3Error(w, s3ErrorNoSuchBucket, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "putObject")
	}
	return nil
}

// POST /<bucket>/<object>?uploads=
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateMultipartUpload.html
func (s *Service) serveCreateMultipartUpload(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	if !r.URL.Query().Has("uploads") {
		return errors.New("expected ?uploads= query parameter")
	}
	if uploads := r.URL.Query().Get("uploads"); uploads != "" {
		return errors.New("?uploads query parameter must always be empty")
	}
	uploadID, err := s.createUpload(r.Context(), bucketName, objectName)
	if err != nil {
		return errors.Wrap(err, "createUpload")
	}
	if err := writeXML(w, http.StatusOK, s3InitiateMultipartUploadResult{
		Bucket:   bucketName,
		Key:      objectName,
		UploadId: uploadID,
	}); err != nil {
		return errors.Wrap(err, "writeXML")
	}
	return nil
}

// POST /<bucket>/<object>?uploadId=foobar
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CompleteMultipartUpload.html
func (s *Service) serveCompleteMultipartUpload(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	uploadID := r.URL.Query().Get("uploadId")
	if err := s.completeUpload(r.Context(), bucketName, objectName, uploadID); err != nil {
		if err == ErrNoSuchUpload {
			return writeS3Error(w, s3ErrorNoSuchUpload, bucketName, err, http.StatusNotFound)
		}
		if err == ErrInvalidPartOrder {
			return writeS3Error(w, s3ErrorInvalidPartOrder, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "completeUpload")
	}
	if err := writeXML(w, http.StatusOK, s3CompleteMultipartUploadResult{
		Bucket: bucketName,
		Key:    objectName,
	}); err != nil {
		return errors.Wrap(err, "writeXML")
	}
	return nil
}

// DELETE /<bucket>/<object>?uploadId=foobar
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_AbortMultipartUpload.html
func (s *Service) serveAbortMultipartUpload(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	uploadID := r.URL.Query().Get("uploadId")
	if uploadID == "" {
		return errors.New("expected ?uploadId query parameter")
	}
	if err := s.abortUpload(r.Context(), bucketName, objectName, uploadID); err != nil {
		if err == ErrNoSuchUpload {
			return writeS3Error(w, s3ErrorNoSuchUpload, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "abortUpload")
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

// DELETE /<bucket>/<object>
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObject.html
func (s *Service) serveDeleteObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) error {
	if err := s.deleteObject(r.Context(), bucketName, objectName); err != nil {
		if err == ErrNoSuchKey {
			return writeS3Error(w, s3ErrorNoSuchKey, bucketName, err, http.StatusNotFound)
		}
		return errors.Wrap(err, "deleteObject")
	}
	return nil
}

// POST /<bucket>?delete
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObjects.html
func (s *Service) serveDeleteObjects(_ http.ResponseWriter, r *http.Request, bucketName string) error {
	var req s3DeleteObjectsRequest
	defer r.Body.Close()
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.Wrap(err, "decoding XML request")
	}

	// TODO(blobstore): technically we should compile a list of errors, and respect req.Quiet in returning
	// error responses. See the S3 API docs above. But for now we just ignore errors, after all, what would
	// our client do with that info?
	for _, obj := range req.Object {
		objectName := obj.Key
		if err := s.deleteObject(r.Context(), bucketName, objectName); err != nil {
			if err == ErrNoSuchKey {
				continue
			}
			s.Log.Warn("error deleting object", sglog.String("key", bucketName+"/"+objectName), sglog.Error(err))
			continue
		}
	}
	return nil
}
