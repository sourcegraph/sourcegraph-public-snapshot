// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// package localhostgate defines a feature gate that controls whether server-like receivers and extensions use localhost as the default host for their endpoints.
// This package is duplicated across core and contrib to avoid exposing the feature gate as part of the public API.
// To do this we define a `registerOrLoad` helper and try to register the gate in both modules.
// IMPORTANT NOTE: ANY CHANGES TO THIS PACKAGE MUST BE MIRRORED IN THE CONTRIB COUNTERPART.
package localhostgate // import "go.opentelemetry.io/collector/internal/localhostgate"

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/featuregate"
)

const UseLocalHostAsDefaultHostID = "component.UseLocalHostAsDefaultHost"

// UseLocalHostAsDefaultHostfeatureGate is the feature gate that controls whether
// server-like receivers and extensions such as the OTLP receiver use localhost as the default host for their endpoints.
var UseLocalHostAsDefaultHostfeatureGate = mustRegisterOrLoad(
	featuregate.GlobalRegistry(),
	UseLocalHostAsDefaultHostID,
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("controls whether server-like receivers and extensions such as the OTLP receiver use localhost as the default host for their endpoints"),
)

// mustRegisterOrLoad tries to register the feature gate and loads it if it already exists.
// It panics on any other error.
func mustRegisterOrLoad(reg *featuregate.Registry, id string, stage featuregate.Stage, opts ...featuregate.RegisterOption) *featuregate.Gate {
	gate, err := reg.Register(id, stage, opts...)

	if errors.Is(err, featuregate.ErrAlreadyRegistered) {
		// Gate is already registered; find it.
		// Only a handful of feature gates are registered, so it's fine to iterate over all of them.
		reg.VisitAll(func(g *featuregate.Gate) {
			if g.ID() == id {
				gate = g
				return
			}
		})
	} else if err != nil {
		panic(err)
	}

	return gate
}

// EndpointForPort gets the endpoint for a given port using localhost or 0.0.0.0 depending on the feature gate.
func EndpointForPort(port int) string {
	host := "localhost"
	if !UseLocalHostAsDefaultHostfeatureGate.IsEnabled() {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// LogAboutUseLocalHostAsDefault logs about the upcoming change from 0.0.0.0 to localhost on server-like components.
func LogAboutUseLocalHostAsDefault(logger *zap.Logger) {
	if !UseLocalHostAsDefaultHostfeatureGate.IsEnabled() {
		logger.Warn(
			"The default endpoints for all servers in components will change to use localhost instead of 0.0.0.0 in a future version. Use the feature gate to preview the new default.",
			zap.String("feature gate ID", UseLocalHostAsDefaultHostID),
		)
	}
}
