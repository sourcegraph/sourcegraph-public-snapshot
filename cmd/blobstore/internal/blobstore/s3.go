package blobstore

import (
	"encoding/xml"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var s3ErrorBucketAlreadyOwnedByYou = "BucketAlreadyOwnedByYou"

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

func writeS3Error(w http.ResponseWriter, code, bucketName string, err error) error {
	w.Header().Set("Content-Type", "application/xml;charset=utf-8")
	w.WriteHeader(http.StatusConflict)

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return errors.Wrap(err, "writing XML header")
	}

	enc := xml.NewEncoder(w)
	if err := enc.Encode(s3Error{Code: code}); err != nil {
		return errors.Wrap(err, "Encode")
	}
	if err := enc.Encode(s3Message{Message: err.Error()}); err != nil {
		return errors.Wrap(err, "Encode")
	}
	if err := enc.Encode(s3BucketName{BucketName: bucketName}); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}
