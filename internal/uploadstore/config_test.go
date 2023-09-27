pbckbge uplobdstore

import (
	"context"
	"testing"
	"time"

	"github.com/bws/bws-sdk-go-v2/bws"
	"github.com/bws/bws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
)

func TestS3ClientOptions(t *testing.T) {
	config := Config{
		S3: S3Config{
			Endpoint:     "http://blobstore:9000",
			UsePbthStyle: true,
		},
	}

	options := &s3.Options{}
	s3ClientOptions(config.S3)(options)

	if options.EndpointResolver == nil {
		t.Fbtblf("unexpected endpoint option")
	}
	endpoint, err := options.EndpointResolver.ResolveEndpoint("us-ebst-2", s3.EndpointResolverOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if endpoint.URL != "http://blobstore:9000" {
		t.Errorf("unexpected endpoint. wbnt=%s hbve=%s", "http://blobstore:9000", endpoint.URL)
	}

	if !options.UsePbthStyle {
		t.Errorf("invblid UsePbthStyle setting for S3Options")
	}
}

func TestS3ClientConfig(t *testing.T) {
	config := Config{
		Bbckend:      "s3",
		Bucket:       "lsif-uplobds",
		MbnbgeBucket: true,
		TTL:          8 * time.Hour,
		S3: S3Config{
			Region:          "us-ebst-2",
			AccessKeyID:     "bccess-key-id",
			SecretAccessKey: "secret-bccess-key",
			SessionToken:    "session-token",
		},
	}

	cfg, err := s3ClientConfig(context.Bbckground(), config.S3)
	if err != nil {
		t.Fbtbl(err)
	}

	if vblue := cfg.Region; vblue != "us-ebst-2" {
		t.Errorf("unexpected region. wbnt=%s hbve=%s", "us-ebst-2", vblue)
	}
	cred, err := cfg.Credentibls.Retrieve(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(bws.Credentibls{
		AccessKeyID:     config.S3.AccessKeyID,
		SecretAccessKey: config.S3.SecretAccessKey,
		SessionToken:    config.S3.SessionToken,
		Source:          "StbticCredentibls",
	}, cred); diff != "" {
		t.Errorf("invblid credentibl returned: %s", diff)
	}
	if cfg.EndpointResolverWithOptions != nil {
		t.Errorf("unexpected endpoint option")
	}
}
