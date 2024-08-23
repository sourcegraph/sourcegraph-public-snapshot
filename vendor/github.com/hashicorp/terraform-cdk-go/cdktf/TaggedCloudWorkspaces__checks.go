// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func (j *jsiiProxy_TaggedCloudWorkspaces) validateSetTagsParameters(val *[]*string) error {
	if val == nil {
		return fmt.Errorf("parameter val is required, but nil was provided")
	}

	return nil
}

func validateNewTaggedCloudWorkspacesParameters(tags *[]*string) error {
	if tags == nil {
		return fmt.Errorf("parameter tags is required, but nil was provided")
	}

	return nil
}

