// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Code generated by "pdata/internal/cmd/pdatagen/main.go". DO NOT EDIT.
// To regenerate this file run "make genpdata".

package internal

import (
	otlpcommon "go.opentelemetry.io/collector/pdata/internal/data/protogen/common/v1"
)

type InstrumentationScope struct {
	orig  *otlpcommon.InstrumentationScope
	state *State
}

func GetOrigInstrumentationScope(ms InstrumentationScope) *otlpcommon.InstrumentationScope {
	return ms.orig
}

func GetInstrumentationScopeState(ms InstrumentationScope) *State {
	return ms.state
}

func NewInstrumentationScope(orig *otlpcommon.InstrumentationScope, state *State) InstrumentationScope {
	return InstrumentationScope{orig: orig, state: state}
}

func GenerateTestInstrumentationScope() InstrumentationScope {
	orig := otlpcommon.InstrumentationScope{}
	state := StateMutable
	tv := NewInstrumentationScope(&orig, &state)
	FillTestInstrumentationScope(tv)
	return tv
}

func FillTestInstrumentationScope(tv InstrumentationScope) {
	tv.orig.Name = "test_name"
	tv.orig.Version = "test_version"
	FillTestMap(NewMap(&tv.orig.Attributes, tv.state))
	tv.orig.DroppedAttributesCount = uint32(17)
}
