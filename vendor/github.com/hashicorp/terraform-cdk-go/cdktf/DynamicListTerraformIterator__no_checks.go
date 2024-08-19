// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DynamicListTerraformIterator) validateDynamicParameters(attributes *map[string]interface{}) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateForExpressionForListParameters(expression interface{}) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateForExpressionForMapParameters(keyExpression interface{}, valueExpression interface{}) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetAnyParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetAnyMapParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetBooleanParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetBooleanMapParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetListParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetMapParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetNumberParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetNumberListParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetNumberMapParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetStringParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetStringMapParameters(attribute *string) error {
	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validatePluckPropertyParameters(property *string) error {
	return nil
}

func validateDynamicListTerraformIterator_FromComplexListParameters(list interface{}, mapKeyAttributeName *string) error {
	return nil
}

func validateDynamicListTerraformIterator_FromDataSourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateDynamicListTerraformIterator_FromListParameters(list interface{}) error {
	return nil
}

func validateDynamicListTerraformIterator_FromMapParameters(map_ interface{}) error {
	return nil
}

func validateDynamicListTerraformIterator_FromResourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateNewDynamicListTerraformIteratorParameters(list interface{}, mapKeyAttributeName *string) error {
	return nil
}

