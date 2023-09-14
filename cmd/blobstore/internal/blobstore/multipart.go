package blobstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/russellhaering/gosaml2/uuid"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// the suffixed bucket name used to store pending multipart uploads
const multipartUploadsBucketSuffix = "---uploads"

type pendingUpload struct {
	BucketName, ObjectName string
	Parts                  []int
}

func (p *pendingUpload) reader() io.ReadCloser {
	data, _ := json.Marshal(p)
	return io.NopCloser(bytes.NewReader(data))
}

// Part numbers must be consecutively ordered, but can start/end at any number.
// Returns the min/max found in p.Parts.
func (p *pendingUpload) partNumberRange() (min, max int) {
	max = -1
	min = -1
	for _, partNumber := range p.Parts {
		if max == -1 || partNumber > max {
			max = partNumber
		}
		if min == -1 || partNumber < min {
			min = partNumber
		}
	}
	return min, max
}

func decodePendingUpload(r io.ReadCloser) (*pendingUpload, error) {
	defer r.Close()
	var v pendingUpload
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return nil, errors.Wrap(err, "Decode")
	}
	return &v, nil
}

func (s *Service) createUpload(ctx context.Context, bucketName, objectName string) (uploadID string, err error) {
	// Create the bucket which will hold multipart uploads for the named bucket.

	if err := s.createBucket(ctx, bucketName+multipartUploadsBucketSuffix); err != nil && err != ErrBucketAlreadyExists {
		return "", errors.Wrap(err, "createBucket")
	}

	// Create the upload descriptor object, which represents the upload, time it was created,
	// if it exists, how many parts have been uploaded so far, etc.
	uploadID = uuid.NewV4().String()
	upload := pendingUpload{BucketName: bucketName, ObjectName: objectName}
	if err := s.upsertPendingUpload(ctx, bucketName, uploadID, &upload); err != nil {
		return "", errors.Wrap(err, "upsertPendingUpload")
	}
	s.Log.Debug("createUpload", sglog.String("key", bucketName+"/"+objectName), sglog.String("uploadID", uploadID))
	return uploadID, nil
}

func (s *Service) getPendingUpload(ctx context.Context, bucketName, uploadID string) (*pendingUpload, error) {
	uploadObjectName := uploadID
	reader, err := s.getObject(ctx, bucketName+multipartUploadsBucketSuffix, uploadObjectName)
	if err != nil {
		if err == ErrNoSuchKey {
			return nil, ErrNoSuchUpload
		}
		return nil, errors.Wrap(err, "fetching upload object")
	}

	upload, err := decodePendingUpload(reader)
	if err != nil {
		return nil, errors.Wrap(err, "decodePendingUpload")
	}
	return upload, nil
}

// Upserts a pending upload descriptor object (which describes that the upload exists, time it was
// created, how many parts have been uploaded so far, etc.)
//
// This method must only be called when creating the object, as otherwise it would be racy with
// mutatePendingUpload.
func (s *Service) upsertPendingUpload(ctx context.Context, bucketName, uploadID string, upload *pendingUpload) error {
	uploadObjectName := uploadID
	_, err := s.putObject(ctx, bucketName+multipartUploadsBucketSuffix, uploadObjectName, upload.reader())
	return err
}

// Atomically mutates a pending upload descriptor object (which describes that the upload exists,
// time it was created, how many parts have been uploaded so far, etc.)
//
// This function holds a mutex to ensure that between the time the object is read, mutated, and
// written - that nobody else mutates the object and changes are lost.
func (s *Service) mutatePendingUploadAtomic(ctx context.Context, bucketName, uploadID string, mutate func(*pendingUpload)) error {
	s.mutatePendingUploadMu.Lock()
	defer s.mutatePendingUploadMu.Unlock()

	upload, err := s.getPendingUpload(ctx, bucketName, uploadID)
	if err != nil {
		return err
	}
	mutate(upload)
	if err := s.upsertPendingUpload(ctx, bucketName, uploadID, upload); err != nil {
		return errors.Wrap(err, "upsertPendingUpload")
	}
	return nil
}

func (s *Service) uploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, data io.ReadCloser) (*objectMetadata, error) {
	defer data.Close()

	// Add the new part number to the upload descriptor
	if err := s.mutatePendingUploadAtomic(ctx, bucketName, uploadID, func(upload *pendingUpload) {
		upload.Parts = append(upload.Parts, partNumber)
	}); err != nil {
		return nil, err
	}

	partObjectName := fmt.Sprintf("%v---%v", uploadID, partNumber)
	metadata, err := s.putObject(ctx, bucketName+multipartUploadsBucketSuffix, partObjectName, data)
	if err != nil {
		return nil, errors.Wrap(err, "putObject")
	}

	s.Log.Debug("uploadPart", sglog.String("key", bucketName+"/"+objectName), sglog.String("uploadID", uploadID), sglog.Int("partNumber", partNumber))
	return metadata, nil
}

func (s *Service) completeUpload(ctx context.Context, bucketName, objectName, uploadID string) error {
	upload, err := s.getPendingUpload(ctx, bucketName, uploadID)
	if err != nil {
		return err
	}
	minPartNumber, maxPartNumber := upload.partNumberRange()

	// Open the parts of the upload.
	var partReaders []io.Reader
	var partClosers []io.Closer
	defer func() {
		// Close all the opened parts.
		for _, closer := range partClosers {
			closer.Close()
		}

		// Delete the upload, if we fail past here there is no recovering the upload.
		if err := s.deletePendingUpload(ctx, bucketName, objectName, uploadID, minPartNumber, maxPartNumber); err != nil {
			s.Log.Error(
				"deleting pending multi-part upload failed",
				sglog.String("key", bucketName+"/"+objectName),
				sglog.String("uploadID", uploadID),
				sglog.Error(err),
			)
		}
	}()
	for partNumber := minPartNumber; partNumber <= maxPartNumber; partNumber++ {
		partObjectName := fmt.Sprintf("%v---%v", uploadID, partNumber)
		part, err := s.getObject(ctx, bucketName+multipartUploadsBucketSuffix, partObjectName)
		if err != nil {
			if err == ErrNoSuchKey {
				return ErrInvalidPartOrder
			}
			return errors.Wrap(err, "fetching part")
		}
		partReaders = append(partReaders, part)
		partClosers = append(partClosers, part)
	}

	// Create the composed object.
	_, err = s.putObject(ctx, bucketName, objectName, io.NopCloser(io.MultiReader(partReaders...)))
	if err != nil {
		return errors.Wrap(err, "creating composed object")
	}

	s.Log.Debug("completeUpload", sglog.String("key", bucketName+"/"+objectName), sglog.String("uploadID", uploadID), sglog.Int("parts", len(partReaders)))
	return nil
}

func (s *Service) deletePendingUpload(ctx context.Context, bucketName, objectName, uploadID string, minPartNumber, maxPartNumber int) error {
	uploadBucketName := bucketName + multipartUploadsBucketSuffix

	var deleteErrors error
	if err := s.deleteObject(ctx, uploadBucketName, uploadID); err != nil {
		deleteErrors = errors.Append(deleteErrors, err)
	}
	for partNumber := minPartNumber; partNumber <= maxPartNumber; partNumber++ {
		partObjectName := fmt.Sprintf("%v---%v", uploadID, partNumber)
		if err := s.deleteObject(ctx, uploadBucketName, partObjectName); err != nil {
			deleteErrors = errors.Append(deleteErrors, err)
		}
	}
	if deleteErrors != nil {
		return deleteErrors
	}
	s.Log.Debug("deletePendingUpload", sglog.String("key", bucketName+"/"+objectName), sglog.String("uploadID", uploadID))
	return nil
}

func (s *Service) abortUpload(ctx context.Context, bucketName, objectName, uploadID string) error {
	upload, err := s.getPendingUpload(ctx, bucketName, uploadID)
	if err != nil {
		return err
	}
	minPartNumber, maxPartNumber := upload.partNumberRange()

	// Delete the upload
	if err := s.deletePendingUpload(ctx, bucketName, objectName, uploadID, minPartNumber, maxPartNumber); err != nil {
		s.Log.Error(
			"deleting pending multi-part upload failed",
			sglog.String("key", bucketName+"/"+objectName),
			sglog.String("uploadID", uploadID),
			sglog.Error(err),
		)
	}

	s.Log.Debug("abortUpload", sglog.String("key", bucketName+"/"+objectName), sglog.String("uploadID", uploadID))
	return nil
}
