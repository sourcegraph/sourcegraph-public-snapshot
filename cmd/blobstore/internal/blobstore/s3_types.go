package blobstore

import (
	"encoding/xml"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	s3ErrorBucketAlreadyOwnedByYou = "BucketAlreadyOwnedByYou"
	s3ErrorNoSuchBucket            = "NoSuchBucket"
	s3ErrorNoSuchKey               = "NoSuchKey"
	s3ErrorNoSuchUpload            = "NoSuchUpload"
	s3ErrorInvalidPartOrder        = "InvalidPartOrder"
)

type s3Error struct {
	XMLName xml.Name `xml:"Error"`
	Code    string
}

type s3Message struct {
	XMLName xml.Name `xml:"Message"`
	Message string   `xml:",chardata"`
}

type s3BucketName struct {
	XMLName    xml.Name `xml:"BucketName"`
	BucketName string   `xml:",chardata"`
}

type s3InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string
	Key      string // Object name only
	UploadId string // opaque string ID like "b008a2ef-4ced-48ea-92bf-d6abddbf06ef"
}

type s3CopyPartResult struct {
	XMLName        xml.Name `xml:"CopyPartResult"`
	ETag           string
	LastModified   string
	ChecksumCRC32  string
	ChecksumCRC32C string
	ChecksumSHA1   string
	ChecksumSHA256 string
}

type s3CompleteMultipartUploadResult struct {
	XMLName        xml.Name `xml:"CompleteMultipartUploadResult"`
	Bucket, Key    string
	ETag           string
	ChecksumCRC32  string
	ChecksumCRC32C string
	ChecksumSHA1   string
	ChecksumSHA256 string
}

type s3ObjectOwner struct {
	DisplayName string
	ID          string
}

type s3Object struct {
	XMLName      xml.Name `xml:"Contents"`
	Key          string
	LastModified string
	Owner        s3ObjectOwner
	Size         int
	StorageClass string
}

type s3ListBucketResult struct {
	XMLName               xml.Name `xml:"ListBucketResult"`
	IsTruncated           bool
	Name                  string
	Prefix                string
	Delimiter             string
	MaxKeys               int
	KeyCount              int
	Contents              []s3Object
	ContinuationToken     string
	NextContinuationToken string
	StartAfter            string
}

type s3ObjectIdentifier struct {
	XMLName   xml.Name `xml:"Object"`
	Key       string
	VersionId string
}

type s3DeleteObjectsRequest struct {
	XMLName xml.Name `xml:"Delete"`
	Object  []s3ObjectIdentifier
	Quiet   bool
}

func writeS3Error(w http.ResponseWriter, code, bucketName string, err error, statusCode int) error {
	return writeXML(w, statusCode,
		s3Error{Code: code},
		s3Message{Message: err.Error()},
		s3BucketName{BucketName: bucketName},
	)
}

func writeXML(w http.ResponseWriter, statusCode int, values ...any) error {
	w.Header().Set("Content-Type", "application/xml;charset=utf-8")
	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return errors.Wrap(err, "writing XML header")
	}

	enc := xml.NewEncoder(w)
	for _, v := range values {
		if err := enc.Encode(v); err != nil {
			return errors.Wrap(err, "Encode")
		}
	}
	return nil
}
