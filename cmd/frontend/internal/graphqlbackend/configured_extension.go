package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cxp"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/sourcegraph/sourcegraph/xlang"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// configuredExtensionResolver implements the GraphQL type ConfiguredExtension.
type configuredExtensionResolver struct {
	extensionID string
	subject     *configurationSubject // nil if the extension is just being queried for capabilities
	enabled     bool

	// cache result because it is used by multiple fields
	once       sync.Once
	initResult *lsp.InitializeResult
	err        error
}

func (r *configuredExtensionResolver) Extension(ctx context.Context) (*registryExtensionMultiResolver, error) {
	return getExtensionByExtensionID(ctx, r.extensionID)
}

func (r *configuredExtensionResolver) ExtensionID() string { return r.extensionID }

func (r *configuredExtensionResolver) Subject() *configurationSubject { return r.subject }

func (r *configuredExtensionResolver) IsEnabled() bool { return r.enabled }

func (r *configuredExtensionResolver) ViewerCanConfigure(ctx context.Context) (bool, error) {
	if r.subject == nil {
		return false, nil
	}
	return r.subject.ViewerCanAdminister(ctx)
}

func (r *configuredExtensionResolver) getInitializeResult(ctx context.Context) (*lsp.InitializeResult, error) {
	do := func() (*lsp.InitializeResult, error) {
		mergedSettings, err := r.viewerMergedSettings(ctx)
		if err != nil {
			return nil, err
		}
		var mergedSettingsRaw *json.RawMessage
		if mergedSettings != nil {
			raw, err := json.Marshal(mergedSettings)
			if err != nil {
				return nil, err
			}
			mergedSettingsRaw = (*json.RawMessage)(&raw)
		}

		c, err := xlang.UnsafeNewDefaultClient()
		if err != nil {
			return nil, err
		}
		defer c.Close()

		var result lsp.InitializeResult
		err = c.Call(ctx, "initialize", cxp.ClientProxyInitializeParams{
			ClientProxyInitializeParams: lspext.ClientProxyInitializeParams{
				InitializeParams: lsp.InitializeParams{
					// TODO(extensions): dummy URI because xlang requires a URI
					RootURI: lsp.DocumentURI("git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06"),
				},
			},
			InitializationOptions: cxp.ClientProxyInitializationOptions{
				ClientProxyInitializationOptions: lspext.ClientProxyInitializationOptions{
					Mode: r.extensionID,
				},
				Settings: cxp.ExtensionSettings{Merged: mergedSettingsRaw},
			},
		}, &result)
		if err != nil {
			if e, ok := err.(*jsonrpc2.Error); ok && e.Code == 0 {
				// Remove noisy "jsonrpc2: code 0 message: " prefix from error message.
				err = errors.New(e.Message)
			}
			if strings.HasPrefix(err.Error(), "dial tcp ") && strings.HasSuffix(err.Error(), "connect: connection refused") {
				err = fmt.Errorf("unable to connect to extension's configured TCP address: %s", err)
			} else {
				err = errors.Wrap(err, "LSP initialize")
			}
			return nil, err
		}
		return &result, nil
	}
	r.once.Do(func() {
		r.initResult, r.err = do()
	})
	return r.initResult, r.err
}

func (r *configuredExtensionResolver) Capabilities(ctx context.Context) (*jsonValue, error) {
	result, err := r.getInitializeResult(ctx)
	if err != nil {
		return nil, err
	}
	return &jsonValue{value: result.Capabilities}, nil
}

func (r *configuredExtensionResolver) Contributions(ctx context.Context) (*jsonValue, error) {
	result, err := r.getInitializeResult(ctx)
	if err != nil {
		return nil, err
	}
	var contributions interface{}
	if m, ok := result.Capabilities.Experimental.(map[string]interface{}); ok {
		contributions, _ = m["contributions"]
	}
	return &jsonValue{value: contributions}, nil
}

func (r *configuredExtensionResolver) viewerMergedSettings(ctx context.Context) (*schema.ExtensionSettings, error) {
	merged, err := viewerMergedConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	var settings schema.Settings
	if err := json.Unmarshal([]byte(merged.Contents()), &settings); err != nil {
		return nil, err
	}

	if settings.Extensions == nil {
		return nil, nil
	}
	extensionSettings, ok := settings.Extensions[r.extensionID]
	if !ok {
		return nil, nil
	}
	return &extensionSettings, nil
}

func (r *configuredExtensionResolver) MergedSettings(ctx context.Context) (*jsonValue, error) {
	s, err := r.viewerMergedSettings(ctx)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, err
	}
	return &jsonValue{value: *s}, nil
}
