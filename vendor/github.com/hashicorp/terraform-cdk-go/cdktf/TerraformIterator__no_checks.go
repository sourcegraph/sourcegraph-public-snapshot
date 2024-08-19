// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformIterator) validateDynamicParameters(attributes *map[string]interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateForExpressionForListParameters(expression interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateForExpressionForMapParameters(keyExpression interface{}, valueExpression interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetAnyParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetAnyMapParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetBooleanParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetBooleanMapParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetListParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetMapParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetNumberParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetNumberListParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetNumberMapParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetStringParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validateGetStringMapParameters(attribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformIterator) validatePluckPropertyParameters(property *string) error {
	return nil
}

func validateTerraformIterator_FromComplexListParameters(list interface{}, mapKeyAttributeName *string) error {
	return nil
}

func validateTerraformIterator_FromDataSourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateTerraformIterator_FromListParameters(list interface{}) error {
	return nil
}

func validateTerraformIterator_FromMapParameters(map_ interface{}) error {
	return nil
}

func validateTerraformIterator_FromResourcesParameters(resource ITerraformResource) error {
	return nil
}

