/* Copyright 2018 The Bazel Authors. All rights reserved.

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

package golang

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/language"
)

func importReposFromModules(args language.ImportReposArgs) language.ImportReposResult {
	// run go list in the dir where go.mod is located
	data, err := goListModules(filepath.Dir(args.Path))
	if err != nil {
		return language.ImportReposResult{Error: processGoListError(err, data)}
	}

	pathToModule, err := extractModules(data)
	if err != nil {
		return language.ImportReposResult{Error: err}
	}

	// Load sums from go.sum. Ideally, they're all there.
	goSumPath := filepath.Join(filepath.Dir(args.Path), "go.sum")
	data, _ = ioutil.ReadFile(goSumPath)
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		fields := bytes.Fields(line)
		if len(fields) != 3 {
			continue
		}
		path, version, sum := string(fields[0]), string(fields[1]), string(fields[2])
		if strings.HasSuffix(version, "/go.mod") {
			continue
		}
		if mod, ok := pathToModule[path+"@"+version]; ok {
			mod.Sum = sum
		}
	}

	pathToModule, err = fillMissingSums(pathToModule)
	if err != nil {
		return language.ImportReposResult{Error: fmt.Errorf("finding module sums: %v", err)}
	}

	return language.ImportReposResult{Gen: toRepositoryRules(pathToModule)}
}
