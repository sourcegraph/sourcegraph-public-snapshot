/* Copyright 2023 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package language

import "context"

// LifecycleManager allows an extension to initialize and
// free up resources at different points. Extensions should
// embed BaseLifecycleManager instead of implementing this
// interface directly.
type LifecycleManager interface {
	FinishableLanguage
	Before(ctx context.Context)
	AfterResolvingDeps(ctx context.Context)
}

var _ LifecycleManager = (*BaseLifecycleManager)(nil)

type BaseLifecycleManager struct{}

func (m *BaseLifecycleManager) Before(ctx context.Context) {}

func (m *BaseLifecycleManager) DoneGeneratingRules() {}

func (m *BaseLifecycleManager) AfterResolvingDeps(ctx context.Context) {}
