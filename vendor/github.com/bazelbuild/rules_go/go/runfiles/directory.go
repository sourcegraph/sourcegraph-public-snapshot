// Copyright 2020, 2021, 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runfiles

import "path/filepath"

// Directory specifies the location of the runfiles directory.  You can pass
// this as an option to New.  If unset or empty, use the value of the
// environmental variable RUNFILES_DIR.
type Directory string

func (d Directory) new(sourceRepo SourceRepo) (*Runfiles, error) {
	r := &Runfiles{
		impl: d,
		env: []string{
			directoryVar + "=" + string(d),
			legacyDirectoryVar + "=" + string(d),
		},
		sourceRepo: string(sourceRepo),
	}
	err := r.loadRepoMapping()
	return r, err
}

func (d Directory) path(s string) (string, error) {
	return filepath.Join(string(d), filepath.FromSlash(s)), nil
}
