// Copyright 2020-2023 Buf Technologies, Inc.
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

package bufwire

import (
	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
)

type moduleConfigSet struct {
	moduleConfigs []ModuleConfig
	workspace     bufmodule.Workspace
}

func newModuleConfigSet(moduleConfigs []ModuleConfig, workspace bufmodule.Workspace) *moduleConfigSet {
	return &moduleConfigSet{
		moduleConfigs: moduleConfigs,
		workspace:     workspace,
	}
}

func (m *moduleConfigSet) ModuleConfigs() []ModuleConfig {
	return m.moduleConfigs
}

func (m *moduleConfigSet) Workspace() bufmodule.Workspace {
	return m.workspace
}
