// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func (d *jsiiProxy_DynamicListTerraformIterator) validateDynamicParameters(attributes *map[string]interface{}) error {
	if attributes == nil {
		return fmt.Errorf("parameter attributes is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateForExpressionForListParameters(expression interface{}) error {
	if expression == nil {
		return fmt.Errorf("parameter expression is required, but nil was provided")
	}
	switch expression.(type) {
	case *string:
		// ok
	case string:
		// ok
	case IResolvable:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(expression) {
			return fmt.Errorf("parameter expression must be one of the allowed types: *string, IResolvable; received %#v (a %T)", expression, expression)
		}
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateForExpressionForMapParameters(keyExpression interface{}, valueExpression interface{}) error {
	if keyExpression == nil {
		return fmt.Errorf("parameter keyExpression is required, but nil was provided")
	}
	switch keyExpression.(type) {
	case *string:
		// ok
	case string:
		// ok
	case IResolvable:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(keyExpression) {
			return fmt.Errorf("parameter keyExpression must be one of the allowed types: *string, IResolvable; received %#v (a %T)", keyExpression, keyExpression)
		}
	}

	if valueExpression == nil {
		return fmt.Errorf("parameter valueExpression is required, but nil was provided")
	}
	switch valueExpression.(type) {
	case *string:
		// ok
	case string:
		// ok
	case IResolvable:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(valueExpression) {
			return fmt.Errorf("parameter valueExpression must be one of the allowed types: *string, IResolvable; received %#v (a %T)", valueExpression, valueExpression)
		}
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetAnyParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetAnyMapParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetBooleanParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetBooleanMapParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetListParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetMapParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetNumberParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetNumberListParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetNumberMapParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetStringParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validateGetStringMapParameters(attribute *string) error {
	if attribute == nil {
		return fmt.Errorf("parameter attribute is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DynamicListTerraformIterator) validatePluckPropertyParameters(property *string) error {
	if property == nil {
		return fmt.Errorf("parameter property is required, but nil was provided")
	}

	return nil
}

func validateDynamicListTerraformIterator_FromComplexListParameters(list interface{}, mapKeyAttributeName *string) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}
	switch list.(type) {
	case IResolvable:
		// ok
	case ComplexList:
		// ok
	case StringMapList:
		// ok
	case NumberMapList:
		// ok
	case BooleanMapList:
		// ok
	case AnyMapList:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(list) {
			return fmt.Errorf("parameter list must be one of the allowed types: IResolvable, ComplexList, StringMapList, NumberMapList, BooleanMapList, AnyMapList; received %#v (a %T)", list, list)
		}
	}

	if mapKeyAttributeName == nil {
		return fmt.Errorf("parameter mapKeyAttributeName is required, but nil was provided")
	}

	return nil
}

func validateDynamicListTerraformIterator_FromDataSourcesParameters(resource ITerraformResource) error {
	if resource == nil {
		return fmt.Errorf("parameter resource is required, but nil was provided")
	}

	return nil
}

func validateDynamicListTerraformIterator_FromListParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}
	switch list.(type) {
	case *[]*string:
		// ok
	case []*string:
		// ok
	case IResolvable:
		// ok
	case *[]*float64:
		// ok
	case []*float64:
		// ok
	case *[]interface{}:
		list := list.(*[]interface{})
		for idx_a33039, v := range *list {
			switch v.(type) {
			case *bool:
				// ok
			case bool:
				// ok
			case IResolvable:
				// ok
			default:
				if !_jsii_.IsAnonymousProxy(v) {
					return fmt.Errorf("parameter list[%#v] must be one of the allowed types: *bool, IResolvable; received %#v (a %T)", idx_a33039, v, v)
				}
			}
		}
	case []interface{}:
		list_ := list.([]interface{})
		list := &list_
		for idx_a33039, v := range *list {
			switch v.(type) {
			case *bool:
				// ok
			case bool:
				// ok
			case IResolvable:
				// ok
			default:
				if !_jsii_.IsAnonymousProxy(v) {
					return fmt.Errorf("parameter list[%#v] must be one of the allowed types: *bool, IResolvable; received %#v (a %T)", idx_a33039, v, v)
				}
			}
		}
	default:
		if !_jsii_.IsAnonymousProxy(list) {
			return fmt.Errorf("parameter list must be one of the allowed types: *[]*string, IResolvable, *[]*float64, *[]interface{}; received %#v (a %T)", list, list)
		}
	}

	return nil
}

func validateDynamicListTerraformIterator_FromMapParameters(map_ interface{}) error {
	if map_ == nil {
		return fmt.Errorf("parameter map_ is required, but nil was provided")
	}
	switch map_.(type) {
	case ComplexMap:
		// ok
	case *map[string]interface{}:
		// ok
	case map[string]interface{}:
		// ok
	case *map[string]*string:
		// ok
	case map[string]*string:
		// ok
	case *map[string]*float64:
		// ok
	case map[string]*float64:
		// ok
	case *map[string]*bool:
		// ok
	case map[string]*bool:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(map_) {
			return fmt.Errorf("parameter map_ must be one of the allowed types: ComplexMap, *map[string]interface{}, *map[string]*string, *map[string]*float64, *map[string]*bool; received %#v (a %T)", map_, map_)
		}
	}

	return nil
}

func validateDynamicListTerraformIterator_FromResourcesParameters(resource ITerraformResource) error {
	if resource == nil {
		return fmt.Errorf("parameter resource is required, but nil was provided")
	}

	return nil
}

func validateNewDynamicListTerraformIteratorParameters(list interface{}, mapKeyAttributeName *string) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}
	switch list.(type) {
	case *[]*string:
		// ok
	case []*string:
		// ok
	case IResolvable:
		// ok
	case *[]*float64:
		// ok
	case []*float64:
		// ok
	case *[]interface{}:
		list := list.(*[]interface{})
		for idx_a33039, v := range *list {
			switch v.(type) {
			case *bool:
				// ok
			case bool:
				// ok
			case IResolvable:
				// ok
			default:
				if !_jsii_.IsAnonymousProxy(v) {
					return fmt.Errorf("parameter list[%#v] must be one of the allowed types: *bool, IResolvable; received %#v (a %T)", idx_a33039, v, v)
				}
			}
		}
	case []interface{}:
		list_ := list.([]interface{})
		list := &list_
		for idx_a33039, v := range *list {
			switch v.(type) {
			case *bool:
				// ok
			case bool:
				// ok
			case IResolvable:
				// ok
			default:
				if !_jsii_.IsAnonymousProxy(v) {
					return fmt.Errorf("parameter list[%#v] must be one of the allowed types: *bool, IResolvable; received %#v (a %T)", idx_a33039, v, v)
				}
			}
		}
	default:
		if !_jsii_.IsAnonymousProxy(list) {
			return fmt.Errorf("parameter list must be one of the allowed types: *[]*string, IResolvable, *[]*float64, *[]interface{}; received %#v (a %T)", list, list)
		}
	}

	if mapKeyAttributeName == nil {
		return fmt.Errorf("parameter mapKeyAttributeName is required, but nil was provided")
	}

	return nil
}

