// Copyright 2022-2023 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build go1.21

package otelconnect

import "context"

// afterFunc calls context.AfterFunc. Build tags ensure that this function
// is only compiled when the Go version is at least 1.21.
func afterFunc(ctx context.Context, f func()) (stop func() bool) {
	return context.AfterFunc(ctx, f)
}
