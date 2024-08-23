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

// Package module provides functions to read information out of MODULE.bazel files.

package module

import (
	"os"
	"path/filepath"

	"github.com/bazelbuild/buildtools/build"
)

// ExtractModuleToApparentNameMapping collects the mapping of module names (e.g. "rules_go") to
// user-configured apparent names (e.g. "my_rules_go") from the repos MODULE.bazel, if it exists.
// See https://bazel.build/external/module#repository_names_and_strict_deps for more information on
// apparent names.
func ExtractModuleToApparentNameMapping(repoRoot string) (func(string) string, error) {
	moduleFile, err := parseModuleFile(repoRoot)
	if err != nil {
		return nil, err
	}
	var moduleToApparentName map[string]string
	if moduleFile != nil {
		moduleToApparentName = collectApparentNames(moduleFile)
	} else {
		// If there is no MODULE.bazel file, return a function that always returns the empty string.
		// Languages will know to fall back to the WORKSPACE names of repos.
		moduleToApparentName = make(map[string]string)
	}

	return func(moduleName string) string {
		return moduleToApparentName[moduleName]
	}, nil
}

func parseModuleFile(repoRoot string) (*build.File, error) {
	path := filepath.Join(repoRoot, "MODULE.bazel")
	bytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return build.ParseModule(path, bytes)
}

// Collects the mapping of module names (e.g. "rules_go") to user-configured apparent names (e.g.
// "my_rules_go"). See https://bazel.build/external/module#repository_names_and_strict_deps for more
// information on apparent names.
func collectApparentNames(m *build.File) map[string]string {
	apparentNames := make(map[string]string)

	for _, dep := range m.Rules("") {
		if dep.Name() == "" {
			continue
		}
		if dep.Kind() != "module" && dep.Kind() != "bazel_dep" {
			continue
		}
		// We support module in addition to bazel_dep to handle language repos that use Gazelle to
		// manage their own BUILD files.
		if name := dep.AttrString("name"); name != "" {
			if repoName := dep.AttrString("repo_name"); repoName != "" {
				apparentNames[name] = repoName
			} else {
				apparentNames[name] = name
			}
		}
	}

	return apparentNames
}
