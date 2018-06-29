package proxy

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	insecureAllowExtensionTargets = strings.Fields(env.Get("INSECURE_ALLOW_EXTENSION_TARGETS", "x.sgdev.org:* *.x.sgdev.org:* *://x.sgdev.org:* *://x.sgdev.org:*/* *://x.sgdev.org/*", "whitelist of globs specifying allowed TCP host:port and WebSocket URL targets for extensions (INSECURE: allows users to cause the server to communicate with these targets; include only targets that are known to be in the external network)"))
	insecureAllowExtensionExec, _ = strconv.ParseBool(env.Get("INSECURE_EXTENSION_EXEC", "false", "allow extensions to specify a command to exec (INSECURE: allows users to run arbitrary shell commands)"))
)

func allowExtensionTarget(allowedPatterns []string, target string) bool {
	for _, pattern := range allowedPatterns {
		if ok, _ := path.Match(pattern, target); ok {
			return true
		}
	}
	return false
}

// 🚨 SECURITY: This lets users on the Sourcegraph site create extensions and cause Sourcegraph to
// communicate with external hosts/URLs specified in INSECURE_ALLOW_EXTENSION_TARGETS (or, when
// INSECURE_EXTENSION_EXEC is true, run arbitrary commands).
//
// The threats to protect against are (1) granting an attacker effective access to the internal
// network where this Sourcegraph site runs (by executing the attacker's extensions, which open
// sockets to internal addresses); and (2) causing abuse to external servers that originates from
// this Sourcegraph site.
//
// To protect against this, targets must be explicitly whitelisted in
// INSECURE_ALLOW_EXTENSION_TARGETS.
func lookupExtension(ctx context.Context, extensionID string) (jsonrpc2.ObjectStream, error) {
	if conf.Platform() == nil {
		return nil, nil
	}

	manifest, err := api.InternalClient.Extension(ctx, extensionID)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("extension %q", extensionID))
	}
	switch {
	case manifest.Platform.Tcp != nil:
		if !allowExtensionTarget(insecureAllowExtensionTargets, manifest.Platform.Tcp.Address) {
			return nil, fmt.Errorf("unable to use extension %q: TCP target %q is not allowed", extensionID, manifest.Platform.Tcp.Address)
		}
		return tcpServer(manifest.Platform.Tcp.Address)()
	case manifest.Platform.Websocket != nil:
		if !allowExtensionTarget(insecureAllowExtensionTargets, manifest.Platform.Websocket.Url) {
			return nil, fmt.Errorf("unable to use extension %q: WebSocket target %q is not allowed", extensionID, manifest.Platform.Websocket.Url)
		}
		return webSocketServer(manifest.Platform.Websocket.Url)()
	case manifest.Platform.Exec != nil:
		// 🚨 SECURITY: This allows any extension publisher to run arbitrary commands on this
		// server, which is highly insecure but useful for local dev. Only allow it when the
		// INSECURE_EXTENSION_EXEC env var is true.
		if insecureAllowExtensionExec {
			return execServer(manifest.Platform.Exec.Command, nil)()
		}
		return nil, fmt.Errorf("unable to use extension %q: exec is disabled", extensionID)
	default:
		return nil, fmt.Errorf("unable to use extension %q: no supported platform in manifest (supported platforms are: exec tcp websocket)", extensionID)
	}
}

func getInitializationOptionsForExtension(ctx context.Context, extensionID string) (map[string]interface{}, error) {
	// TODO(extensions): This should allow for using extra initializationOptions that a user
	// specifies in their user settings for the extension.
	if conf.Platform() == nil {
		return nil, nil
	}
	manifest, err := api.InternalClient.Extension(ctx, extensionID)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("extension %q", extensionID))
	}
	if manifest.Args == nil {
		return nil, nil
	}
	return *manifest.Args, nil
}
