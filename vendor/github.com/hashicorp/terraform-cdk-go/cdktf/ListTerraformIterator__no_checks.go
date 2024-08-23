// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (l *jsiiProxy_ListTerraformIterator) validateDynamicParameters(attributes *map[string]interface{}) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateForExpressionForListParameters(expression interface{}) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateForExpressionForMapParameters(keyExpression interface{}, valueExpression interface{}) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetAnyParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetAnyMapParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetBooleanParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetBooleanMapParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetListParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetMapParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetNumberParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetNumberListParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetNumberMapParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetStringParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validateGetStringMapParameters(attribute *string) error {
	return nil
}

func (l *jsiiProxy_ListTerraformIterator) validatePluckPropertyParameters(property *string) error {
	return nil
}

func validateListTerraformIterator_FromComplexListParameters(list interface{}, mapKeyAttributeName *string) error {
	return nil
}

func validateListTerraformIterator_FromDataSourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateListTerraformIterator_FromListParameters(list interface{}) error {
	return nil
}

func validateListTerraformIterator_FromMapParameters(map_ interface{}) error {
	return nil
}

func validateListTerraformIterator_FromResourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateNewListTerraformIteratorParameters(list interface{}) error {
	return nil
}

