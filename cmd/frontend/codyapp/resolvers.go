package codyapp

import (
	"context"
	"encoding/json"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type UpdateManifestResolver interface {
	Resolve(ctx context.Context) (*AppUpdateManifest, error)
}

type GCSManifestResolver struct {
	client       *storage.Client
	bucket       string
	manifestName string
}

type StaticManifestResolver struct {
	Manifest AppUpdateManifest
}

func NewGCSManifestResolver(ctx context.Context, bucket, manifestName string) (UpdateManifestResolver, error) {
	client, err := storage.NewClient(ctx, option.WithScopes(storage.ScopeReadOnly))
	if err != nil {
		return nil, err
	}

	return &GCSManifestResolver{
		client:       client,
		bucket:       bucket,
		manifestName: manifestName,
	}, nil
}

func (r *GCSManifestResolver) Resolve(ctx context.Context) (*AppUpdateManifest, error) {
	obj := r.client.Bucket(r.bucket).Object(r.manifestName)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	manifest := AppUpdateManifest{}
	err = json.Unmarshal(data, &manifest)
	return &manifest, err
}

func (r *StaticManifestResolver) Resolve(_ context.Context) (*AppUpdateManifest, error) {
	return &r.Manifest, nil
}
