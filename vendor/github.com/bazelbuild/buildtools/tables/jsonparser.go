/*
Copyright 2017 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tables

import (
	"encoding/json"
	"io/ioutil"
)

type Definitions struct {
	IsLabelArg                      map[string]bool
	LabelDenylist                   map[string]bool
	IsListArg                       map[string]bool
	IsSortableListArg               map[string]bool
	SortableDenylist                map[string]bool
	SortableAllowlist               map[string]bool
	NamePriority                    map[string]int
	StripLabelLeadingSlashes        bool
	ShortenAbsoluteLabelsToRelative bool
}

// ParseJSONDefinitions reads and parses JSON table definitions from file.
func ParseJSONDefinitions(file string) (Definitions, error) {
	var definitions Definitions

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return definitions, err
	}

	err = json.Unmarshal(data, &definitions)
	return definitions, err
}

// ParseAndUpdateJSONDefinitions reads definitions from file and merges or
// overrides the values in memory.
func ParseAndUpdateJSONDefinitions(file string, merge bool) error {
	definitions, err := ParseJSONDefinitions(file)
	if err != nil {
		return err
	}

	if merge {
		MergeTables(definitions.IsLabelArg, definitions.LabelDenylist, definitions.IsListArg, definitions.IsSortableListArg, definitions.SortableDenylist, definitions.SortableAllowlist, definitions.NamePriority, definitions.StripLabelLeadingSlashes, definitions.ShortenAbsoluteLabelsToRelative)
	} else {
		OverrideTables(definitions.IsLabelArg, definitions.LabelDenylist, definitions.IsListArg, definitions.IsSortableListArg, definitions.SortableDenylist, definitions.SortableAllowlist, definitions.NamePriority, definitions.StripLabelLeadingSlashes, definitions.ShortenAbsoluteLabelsToRelative)
	}
	return nil
}
