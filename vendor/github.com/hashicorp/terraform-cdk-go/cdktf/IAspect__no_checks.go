// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (i *jsiiProxy_IAspect) validateVisitParameters(node constructs.IConstruct) error {
	return nil
}

