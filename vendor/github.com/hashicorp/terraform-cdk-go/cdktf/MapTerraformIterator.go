// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type MapTerraformIterator interface {
	TerraformIterator
	// Returns the key of the current entry in the map that is being iterated over.
	// Experimental.
	Key() *string
	// Returns the value of the current item iterated over.
	// Experimental.
	Value() interface{}
	// Creates a dynamic expression that can be used to loop over this iterator in a dynamic block.
	//
	// As this returns an IResolvable you might need to wrap the output in
	// a Token, e.g. `Token.asString`.
	// See https://developer.hashicorp.com/terraform/cdktf/concepts/iterators#using-iterators-for-list-attributes
	// Experimental.
	Dynamic(attributes *map[string]interface{}) IResolvable
	// Creates a for expression that results in a list.
	//
	// This method allows you to create every possible for expression, but requires more knowledge about
	// Terraform's for expression syntax.
	// For the most common use cases you can use keys(), values(), and pluckProperty() instead.
	//
	// You may write any valid Terraform for each expression, e.g.
	// `TerraformIterator.fromList(myIteratorSourceVar).forExpressionForList("val.foo if val.bar == true")`
	// will result in `[ for key, val in var.myIteratorSource: val.foo if val.bar == true ]`.
	//
	// As this returns an IResolvable you might need to wrap the output in
	// a Token, e.g. `Token.asString`.
	// Experimental.
	ForExpressionForList(expression interface{}) IResolvable
	// Creates a for expression that results in a map.
	//
	// This method allows you to create every possible for expression, but requires more knowledge about
	// Terraforms for expression syntax.
	// For the most common use cases you can use keys(), values(), and pluckProperty instead.
	//
	// You may write any valid Terraform for each expression, e.g.
	// `TerraformIterator.fromMap(myIteratorSourceVar).forExpressionForMap("key", "val.foo if val.bar == true")`
	// will result in `{ for key, val in var.myIteratorSource: key => val.foo if val.bar == true }`.
	//
	// As this returns an IResolvable you might need to wrap the output in
	// a Token, e.g. `Token.asString`.
	// Experimental.
	ForExpressionForMap(keyExpression interface{}, valueExpression interface{}) IResolvable
	// Returns: the given attribute of the current item iterated over as any.
	// Experimental.
	GetAny(attribute *string) IResolvable
	// Returns: the given attribute of the current item iterated over as a map of any.
	// Experimental.
	GetAnyMap(attribute *string) *map[string]interface{}
	// Returns: the given attribute of the current item iterated over as a boolean.
	// Experimental.
	GetBoolean(attribute *string) IResolvable
	// Returns: the given attribute of the current item iterated over as a map of booleans.
	// Experimental.
	GetBooleanMap(attribute *string) *map[string]*bool
	// Returns: the given attribute of the current item iterated over as a (string) list.
	// Experimental.
	GetList(attribute *string) *[]*string
	// Returns: the given attribute of the current item iterated over as a map.
	// Experimental.
	GetMap(attribute *string) *map[string]interface{}
	// Returns: the given attribute of the current item iterated over as a number.
	// Experimental.
	GetNumber(attribute *string) *float64
	// Returns: the given attribute of the current item iterated over as a number list.
	// Experimental.
	GetNumberList(attribute *string) *[]*float64
	// Returns: the given attribute of the current item iterated over as a map of numbers.
	// Experimental.
	GetNumberMap(attribute *string) *map[string]*float64
	// Returns: the given attribute of the current item iterated over as a string.
	// Experimental.
	GetString(attribute *string) *string
	// Returns: the given attribute of the current item iterated over as a map of strings.
	// Experimental.
	GetStringMap(attribute *string) *map[string]*string
	// Creates a for expression that maps the iterators to its keys.
	//
	// For lists these would be the indices, for maps the keys.
	// As this returns an IResolvable you might need to wrap the output in
	// a Token, e.g. `Token.asString`.
	// Experimental.
	Keys() IResolvable
	// Creates a for expression that accesses the key on each element of the iterator.
	//
	// As this returns an IResolvable you might need to wrap the output in
	// a Token, e.g. `Token.asString`.
	// Experimental.
	PluckProperty(property *string) IResolvable
	// Creates a for expression that maps the iterators to its value in case it is a map.
	//
	// For lists these would stay the same.
	// As this returns an IResolvable you might need to wrap the output in
	// a Token, e.g. `Token.asString`.
	// Experimental.
	Values() IResolvable
}

// The jsii proxy struct for MapTerraformIterator
type jsiiProxy_MapTerraformIterator struct {
	jsiiProxy_TerraformIterator
}

func (j *jsiiProxy_MapTerraformIterator) Key() *string {
	var returns *string
	_jsii_.Get(
		j,
		"key",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MapTerraformIterator) Value() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"value",
		&returns,
	)
	return returns
}


// Experimental.
func NewMapTerraformIterator(map_ interface{}) MapTerraformIterator {
	_init_.Initialize()

	if err := validateNewMapTerraformIteratorParameters(map_); err != nil {
		panic(err)
	}
	j := jsiiProxy_MapTerraformIterator{}

	_jsii_.Create(
		"cdktf.MapTerraformIterator",
		[]interface{}{map_},
		&j,
	)

	return &j
}

// Experimental.
func NewMapTerraformIterator_Override(m MapTerraformIterator, map_ interface{}) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.MapTerraformIterator",
		[]interface{}{map_},
		m,
	)
}

// Creates a new iterator from a complex list.
//
// One example for this would be a list of maps.
// The list will be converted into a map with the mapKeyAttributeName as the key.
//
// Example:
//   const cert = new AcmCertificate(this, "cert", {
//      domainName: "example.com",
//      validationMethod: "DNS",
//    });
//
//   const dvoIterator = TerraformIterator.fromComplexList(
//     cert.domainValidationOptions,
//     "domain_name"
//   );
//
//   new Route53Record(this, "record", {
//     allowOverwrite: true,
//     name: dvoIterator.getString("name"),
//     records: [dvoIterator.getString("record")],
//     ttl: 60,
//     type: dvoIterator.getString("type"),
//     zoneId: Token.asString(dataAwsRoute53ZoneExample.zoneId),
//     forEach: dvoIterator,
//   });
//
// Experimental.
func MapTerraformIterator_FromComplexList(list interface{}, mapKeyAttributeName *string) DynamicListTerraformIterator {
	_init_.Initialize()

	if err := validateMapTerraformIterator_FromComplexListParameters(list, mapKeyAttributeName); err != nil {
		panic(err)
	}
	var returns DynamicListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.MapTerraformIterator",
		"fromComplexList",
		[]interface{}{list, mapKeyAttributeName},
		&returns,
	)

	return returns
}

// Creates a new iterator from a data source that has been created with the `for_each` argument.
// Experimental.
func MapTerraformIterator_FromDataSources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateMapTerraformIterator_FromDataSourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.MapTerraformIterator",
		"fromDataSources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

// Creates a new iterator from a list.
// Experimental.
func MapTerraformIterator_FromList(list interface{}) ListTerraformIterator {
	_init_.Initialize()

	if err := validateMapTerraformIterator_FromListParameters(list); err != nil {
		panic(err)
	}
	var returns ListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.MapTerraformIterator",
		"fromList",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// Creates a new iterator from a map.
// Experimental.
func MapTerraformIterator_FromMap(map_ interface{}) MapTerraformIterator {
	_init_.Initialize()

	if err := validateMapTerraformIterator_FromMapParameters(map_); err != nil {
		panic(err)
	}
	var returns MapTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.MapTerraformIterator",
		"fromMap",
		[]interface{}{map_},
		&returns,
	)

	return returns
}

// Creates a new iterator from a resource that has been created with the `for_each` argument.
// Experimental.
func MapTerraformIterator_FromResources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateMapTerraformIterator_FromResourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.MapTerraformIterator",
		"fromResources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) Dynamic(attributes *map[string]interface{}) IResolvable {
	if err := m.validateDynamicParameters(attributes); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"dynamic",
		[]interface{}{attributes},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) ForExpressionForList(expression interface{}) IResolvable {
	if err := m.validateForExpressionForListParameters(expression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"forExpressionForList",
		[]interface{}{expression},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) ForExpressionForMap(keyExpression interface{}, valueExpression interface{}) IResolvable {
	if err := m.validateForExpressionForMapParameters(keyExpression, valueExpression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"forExpressionForMap",
		[]interface{}{keyExpression, valueExpression},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetAny(attribute *string) IResolvable {
	if err := m.validateGetAnyParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"getAny",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetAnyMap(attribute *string) *map[string]interface{} {
	if err := m.validateGetAnyMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		m,
		"getAnyMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetBoolean(attribute *string) IResolvable {
	if err := m.validateGetBooleanParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"getBoolean",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetBooleanMap(attribute *string) *map[string]*bool {
	if err := m.validateGetBooleanMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		m,
		"getBooleanMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetList(attribute *string) *[]*string {
	if err := m.validateGetListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		m,
		"getList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetMap(attribute *string) *map[string]interface{} {
	if err := m.validateGetMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		m,
		"getMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetNumber(attribute *string) *float64 {
	if err := m.validateGetNumberParameters(attribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		m,
		"getNumber",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetNumberList(attribute *string) *[]*float64 {
	if err := m.validateGetNumberListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		m,
		"getNumberList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetNumberMap(attribute *string) *map[string]*float64 {
	if err := m.validateGetNumberMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		m,
		"getNumberMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetString(attribute *string) *string {
	if err := m.validateGetStringParameters(attribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		m,
		"getString",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) GetStringMap(attribute *string) *map[string]*string {
	if err := m.validateGetStringMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		m,
		"getStringMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) Keys() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"keys",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) PluckProperty(property *string) IResolvable {
	if err := m.validatePluckPropertyParameters(property); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"pluckProperty",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MapTerraformIterator) Values() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		m,
		"values",
		nil, // no parameters
		&returns,
	)

	return returns
}

