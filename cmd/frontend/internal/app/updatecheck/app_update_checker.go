package updatecheck

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RouteAppUpdateCheck is the name of the route that the Sourcegraph App will use to check if there are updates
const RouteAppUpdateCheck = "app.update.check"

// ManifestBucket the name of the bucket where the Sourcegraph App update manifest is stored
const ManifestBucket = "sourcegraph_app"

// ManifestName is the name of the manifest object that is in the ManifestBucket
const ManifestName = "update.test.manifest.json"

type AppVersion struct {
	Target  string
	Version string
	Arch    string
}

type AppUpdateManifest struct {
	Version   string      `json:"version"`
	Notes     string      `json:"notes"`
	PubDate   time.Time   `json:"pub_date"`
	Platforms AppPlatform `json:"platforms"`
}

type AppPlatform map[string]AppLocation

type AppLocation struct {
	Signature string `json:"signature"`
	Url       string `json:"url"`
}

type AppUpdateChecker struct {
	logger           log.Logger
	manifestResolver UpdateManifestResolver
}

type AppNoopUpdateChecker struct{}

type UpdateManifestResolver interface {
	Resolve(ctx context.Context) (*AppUpdateManifest, error)
}

type GCSManifestResolver struct {
	client       *storage.Client
	bucket       string
	manifestName string
}

type StaticManifestResolver struct {
	manifest AppUpdateManifest
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
	return &r.manifest, nil
}

func NewAppUpdateChecker(logger log.Logger, resolver UpdateManifestResolver) *AppUpdateChecker {
	return &AppUpdateChecker{
		logger:           logger,
		manifestResolver: resolver,
	}
}

func (a *AppVersion) validate() error {
	if a.Target == "" {
		return errors.New("target is empty")
	}
	if a.Version == "" {
		return errors.New("version is empty")
	}
	if a.Arch == "" {
		return errors.New("arch is empty")
	}
	return nil
}

func (checker *AppUpdateChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appClientVersion := readClientAppVersion(r.URL)
		if err := appClientVersion.validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		manifest, err := checker.manifestResolver.Resolve(ctx)
		if err != nil {
			checker.logger.Error("failed to resolve update manifest", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		checker.logger.Info("app update check", log.Object("App",
			log.String("target", appClientVersion.Target),
			log.String("version", appClientVersion.Version),
			log.String("arch", appClientVersion.Arch),
		))

		if !checker.canUpdate(appClientVersion, manifest) {
			// No update
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// notify the app client that they can update
		json.NewEncoder(w).Encode(manifest)
		w.WriteHeader(http.StatusOK)
	}
}

func readClientAppVersion(reqURL *url.URL) *AppVersion {
	queryValues := reqURL.Query()
	var appClientVersion = AppVersion{}
	for key, attr := range map[string]*string{
		"target":          &appClientVersion.Target,
		"current_version": &appClientVersion.Version,
		"arch":            &appClientVersion.Arch,
	} {
		if v, ok := queryValues[key]; ok && len(v) > 0 {
			*attr = v[0]
		}
	}

	return &appClientVersion
}

func (checker *AppUpdateChecker) canUpdate(client *AppVersion, manifest *AppUpdateManifest) bool {
	// very crude check, we should convert each version to semver here and do a proper check
	return client.Version != manifest.Version
}

func (checker *AppNoopUpdateChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// No update
		w.WriteHeader(http.StatusNoContent)
	}
}

func AppUpdateHandler(logger log.Logger) http.HandlerFunc {
	// We store the Sourcegraph App manifest in a GCS bucket
	resolver, err := NewGCSManifestResolver(context.Background(), ManifestBucket, ManifestName)
	if err != nil {
		logger.Error("failed to initialize GCS Manifest resolver. Using NoopUpdateChecker which will tell all clients that there are no updates", log.Error(err))
		return (&AppNoopUpdateChecker{}).Handler()
	} else {
		return NewAppUpdateChecker(logger, resolver).Handler()
	}
}
