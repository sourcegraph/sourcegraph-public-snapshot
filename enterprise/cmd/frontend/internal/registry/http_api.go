package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/registry"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	frontendregistry.HandleRegistry = handleRegistry
}

// Funcs called by serveRegistry to get registry data. If fakeRegistryData is set, it is used as
// the data source instead of the database.
var (
	registryList = func(ctx context.Context, opt dbExtensionsListOptions) ([]*registry.Extension, error) {
		vs, err := dbExtensions{}.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		xs := make([]*registry.Extension, 0, len(vs))
		for _, v := range vs {
			x, err := toRegistryAPIExtension(ctx, v)
			if err != nil {
				continue
			}

			// To be safe, ensure that the JSON can be safely unmarshaled by API clients. If not,
			// skip this extension.
			if x.Manifest != nil {
				var o schema.SourcegraphExtensionManifest
				if err := jsonc.Unmarshal(*x.Manifest, &o); err != nil {
					continue
				}
			}

			xs = append(xs, x)
		}
		return xs, nil
	}

	registryGetByUUID = func(ctx context.Context, uuid string) (*registry.Extension, error) {
		x, err := dbExtensions{}.GetByUUID(ctx, uuid)
		if err != nil {
			return nil, err
		}
		return toRegistryAPIExtension(ctx, x)
	}

	registryGetByExtensionID = func(ctx context.Context, extensionID string) (*registry.Extension, error) {
		x, err := dbExtensions{}.GetByExtensionID(ctx, extensionID)
		if err != nil {
			return nil, err
		}
		return toRegistryAPIExtension(ctx, x)
	}
)

func toRegistryAPIExtension(ctx context.Context, v *dbExtension) (*registry.Extension, error) {
	manifest, publishedAt, err := getExtensionManifestWithBundleURL(ctx, v.NonCanonicalExtensionID, v.ID, "release")
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimSuffix(conf.Get().ExternalURL, "/")
	return &registry.Extension{
		UUID:        v.UUID,
		ExtensionID: v.NonCanonicalExtensionID,
		Publisher: registry.Publisher{
			Name: v.Publisher.NonCanonicalName,
			URL:  baseURL + frontendregistry.PublisherExtensionsURL(v.Publisher.UserID != 0, v.Publisher.OrgID != 0, v.Publisher.NonCanonicalName),
		},
		Name:        v.Name,
		Manifest:    manifest,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		PublishedAt: publishedAt,
		URL:         baseURL + frontendregistry.ExtensionURL(v.NonCanonicalExtensionID),
	}, nil
}

// handleRegistry serves the external HTTP API for the extension registry.
func handleRegistry(w http.ResponseWriter, r *http.Request) (err error) {
	if conf.Extensions() == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	builder := honey.Builder("registry")
	builder.AddField("api_version", r.Header.Get("Accept"))
	builder.AddField("url", r.URL.String())
	ev := builder.NewEvent()
	defer func() {
		ev.AddField("success", err == nil)
		if err == nil {
			registryRequestsSuccessCounter.Inc()
		} else {
			registryRequestsErrorCounter.Inc()
			ev.AddField("error", err.Error())
		}
		ev.Send()
	}()

	// Identify this response as coming from the registry API.
	w.Header().Set(registry.MediaTypeHeaderName, registry.MediaType)
	w.Header().Set("Vary", registry.MediaTypeHeaderName)

	// Validate API version.
	if v := r.Header.Get("Accept"); v != registry.AcceptHeader {
		http.Error(w, fmt.Sprintf("invalid Accept header: expected %q", registry.AcceptHeader), http.StatusBadRequest)
		return nil
	}

	// This handler can be mounted at either /.internal or /.api.
	urlPath := r.URL.Path
	switch {
	case strings.HasPrefix(urlPath, "/.internal"):
		urlPath = strings.TrimPrefix(urlPath, "/.internal")
	case strings.HasPrefix(urlPath, "/.api"):
		urlPath = strings.TrimPrefix(urlPath, "/.api")
	}

	const extensionsPath = "/registry/extensions"
	var result interface{}
	switch {
	case urlPath == extensionsPath:
		query := r.URL.Query().Get("q")
		ev.AddField("query", query)
		var opt dbExtensionsListOptions
		opt.Query, opt.Category, opt.Tag = parseExtensionQuery(query)
		xs, err := registryList(r.Context(), opt)
		if err != nil {
			return err
		}
		ev.AddField("results_count", len(xs))
		result = xs

	case strings.HasPrefix(urlPath, extensionsPath+"/"):
		var (
			spec = strings.TrimPrefix(urlPath, extensionsPath+"/")
			x    *registry.Extension
			err  error
		)
		switch {
		case strings.HasPrefix(spec, "uuid/"):
			x, err = registryGetByUUID(r.Context(), strings.TrimPrefix(spec, "uuid/"))
		case strings.HasPrefix(spec, "extension-id/"):
			x, err = registryGetByExtensionID(r.Context(), strings.TrimPrefix(spec, "extension-id/"))
		default:
			w.WriteHeader(http.StatusNotFound)
			return nil
		}
		if x == nil || err != nil {
			if x == nil || errcode.IsNotFound(err) {
				w.Header().Set("Cache-Control", "max-age=5, private")
				http.Error(w, "extension not found", http.StatusNotFound)
				return nil
			}
			return err
		}
		ev.AddField("extension-id", x.ExtensionID)
		result = x

	default:
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	w.Header().Set("Cache-Control", "max-age=30, private")
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

var (
	registryRequestsSuccessCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "registry",
		Name:      "requests_success",
		Help:      "Number of successful requests (HTTP 200) to the HTTP registry API",
	})
	registryRequestsErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "registry",
		Name:      "requests_error",
		Help:      "Number of failed (non-HTTP 200) requests to the HTTP registry API",
	})
)

func init() {
	prometheus.MustRegister(registryRequestsSuccessCounter)
	prometheus.MustRegister(registryRequestsErrorCounter)
}

func init() {
	// Allow providing fake registry data for local dev (intended for use in local dev only).
	//
	// If FAKE_REGISTRY is set and refers to a valid JSON file (of []*registry.Extension), is used
	// by serveRegistry (instead of the DB) as the source for registry data.
	path := os.Getenv("FAKE_REGISTRY")
	if path == "" {
		return
	}

	readFakeExtensions := func() ([]*registry.Extension, error) {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var xs []*registry.Extension
		if err := json.Unmarshal(data, &xs); err != nil {
			return nil, err
		}
		return xs, nil
	}

	registryList = func(ctx context.Context, opt dbExtensionsListOptions) ([]*registry.Extension, error) {
		xs, err := readFakeExtensions()
		if err != nil {
			return nil, err
		}
		return frontendregistry.FilterRegistryExtensions(xs, opt.Query), nil
	}
	registryGetByUUID = func(ctx context.Context, uuid string) (*registry.Extension, error) {
		xs, err := readFakeExtensions()
		if err != nil {
			return nil, err
		}
		return frontendregistry.FindRegistryExtension(xs, "uuid", uuid), nil
	}
	registryGetByExtensionID = func(ctx context.Context, extensionID string) (*registry.Extension, error) {
		xs, err := readFakeExtensions()
		if err != nil {
			return nil, err
		}
		return frontendregistry.FindRegistryExtension(xs, "extensionID", extensionID), nil
	}
}
