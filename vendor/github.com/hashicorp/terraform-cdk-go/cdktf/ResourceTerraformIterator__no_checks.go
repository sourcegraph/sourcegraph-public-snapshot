// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (r *jsiiProxy_ResourceTerraformIterator) validateDynamicParameters(attributes *map[string]interface{}) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateForExpressionForListParameters(expression interface{}) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateForExpressionForMapParameters(keyExpression interface{}, valueExpression interface{}) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetAnyParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetAnyMapParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetBooleanParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetBooleanMapParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetListParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetMapParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetNumberParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetNumberListParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetNumberMapParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetStringParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validateGetStringMapParameters(attribute *string) error {
	return nil
}

func (r *jsiiProxy_ResourceTerraformIterator) validatePluckPropertyParameters(property *string) error {
	return nil
}

func validateResourceTerraformIterator_FromComplexListParameters(list interface{}, mapKeyAttributeName *string) error {
	return nil
}

func validateResourceTerraformIterator_FromDataSourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateResourceTerraformIterator_FromListParameters(list interface{}) error {
	return nil
}

func validateResourceTerraformIterator_FromMapParameters(map_ interface{}) error {
	return nil
}

func validateResourceTerraformIterator_FromResourcesParameters(resource ITerraformResource) error {
	return nil
}

func validateNewResourceTerraformIteratorParameters(element ITerraformResource) error {
	return nil
}

