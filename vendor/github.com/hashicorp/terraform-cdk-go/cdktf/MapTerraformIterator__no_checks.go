// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (m *jsiiProxy_MapTerraformIterator) validateDynamicParameters(attributes *map[string]interface{}) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateForExpressionForListParameters(expression interface{}) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateForExpressionForMapParameters(keyExpression interface{}, valueExpression interface{}) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetAnyParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetAnyMapParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetBooleanParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetBooleanMapParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetListParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetMapParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetNumberParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetNumberListParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetNumberMapParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetStringParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validateGetStringMapParameters(attribute *string) error {
	return nil
}

func (m *jsiiProxy_MapTerraformIterator) validatePluckPropertyParameters(property *string) error {
	return nil
}

func validateMapTerraformIterator_FromComplexListParameters(list interface{}, mapKeyAttributeName *string) error {
	return nil
}

func validateMapTerraformIterator_FromDataSourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateMapTerraformIterator_FromListParameters(list interface{}) error {
	return nil
}

func validateMapTerraformIterator_FromMapParameters(map_ interface{}) error {
	return nil
}

func validateMapTerraformIterator_FromResourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateNewMapTerraformIteratorParameters(map_ interface{}) error {
	return nil
}

