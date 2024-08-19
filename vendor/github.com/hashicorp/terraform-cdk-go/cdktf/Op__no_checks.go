// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateOp_AddParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_AndParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_DivParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_EqParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_GtParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_GteParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_LtParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_LteParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_ModParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_MulParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_NegateParameters(expression interface{}) error {
	return nil
}

func validateOp_NeqParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_NotParameters(expression interface{}) error {
	return nil
}

func validateOp_OrParameters(left interface{}, right interface{}) error {
	return nil
}

func validateOp_SubParameters(left interface{}, right interface{}) error {
	return nil
}

