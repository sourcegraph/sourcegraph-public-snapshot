package httpapi

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"google.golang.org/api/option"
)

var (
	gcsClient *storage.Client
	gcsConfig *schema.CodyUpload
)

func init() {
	conf.Watch(func() {
		cf := conf.Get()
		if cf.Dotcom == nil || cf.Dotcom.CodyUpload == nil || cf.Dotcom.CodyUpload.Bucket == "" || cf.Dotcom.CodyUpload.CredentialsJSON == "" {
			gcsConfig = nil
			return
		}

		if cf.Dotcom.CodyUpload.Bucket != gcsConfig.Bucket || cf.Dotcom.CodyUpload.CredentialsJSON != gcsConfig.CredentialsJSON {
			gcsConfig = cf.Dotcom.CodyUpload
			if newGCSClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(gcsConfig.CredentialsJSON))); err == nil {
				gcsClient = newGCSClient
			} else {
				log15.Error("Failed to create new GCS client for Cody uploads", "err", err)
				gcsClient = nil
			}
		}
	})
}

func handleCodySaveTranscript(w http.ResponseWriter, req *http.Request) (err error) {
	if gcsClient == nil {
		return nil
	}
	ctx := req.Context()

	// TODO(beyang): read this from request
	const objectKey = "my.object.path"

	// ctx, cancel := context.WithCancel(ctx)
	// TODO(beyang): is cancel invoked correctly in this function?
	// defer cancel()

	bucketWriter := gcsClient.Bucket(gcsConfig.Bucket).Object(objectKey).NewWriter(ctx)
	defer func() {
		if closeErr := bucketWriter.Close(); closeErr != nil {
			log.Printf("########## closeErr: %v", closeErr)
			err = errors.Append(err, errors.Wrap(closeErr, "failed to close bucket writer"))

			// cancel()
		}
	}()

	_, bucketErr := bucketWriter.Write([]byte("This is an object"))
	// log.Printf("# bucketErr: %v", bucketErr)
	if bucketErr != nil {
		return bucketErr
	}

	w.Write([]byte(`{ "message": "hello world 3" }`))
	return nil
}
