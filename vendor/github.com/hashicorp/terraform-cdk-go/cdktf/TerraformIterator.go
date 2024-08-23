// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type TerraformIterator interface {
	ITerraformIterator
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

// The jsii proxy struct for TerraformIterator
type jsiiProxy_TerraformIterator struct {
	jsiiProxy_ITerraformIterator
}

// Experimental.
func NewTerraformIterator_Override(t TerraformIterator) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformIterator",
		nil, // no parameters
		t,
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
func TerraformIterator_FromComplexList(list interface{}, mapKeyAttributeName *string) DynamicListTerraformIterator {
	_init_.Initialize()

	if err := validateTerraformIterator_FromComplexListParameters(list, mapKeyAttributeName); err != nil {
		panic(err)
	}
	var returns DynamicListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.TerraformIterator",
		"fromComplexList",
		[]interface{}{list, mapKeyAttributeName},
		&returns,
	)

	return returns
}

// Creates a new iterator from a data source that has been created with the `for_each` argument.
// Experimental.
func TerraformIterator_FromDataSources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateTerraformIterator_FromDataSourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.TerraformIterator",
		"fromDataSources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

// Creates a new iterator from a list.
// Experimental.
func TerraformIterator_FromList(list interface{}) ListTerraformIterator {
	_init_.Initialize()

	if err := validateTerraformIterator_FromListParameters(list); err != nil {
		panic(err)
	}
	var returns ListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.TerraformIterator",
		"fromList",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// Creates a new iterator from a map.
// Experimental.
func TerraformIterator_FromMap(map_ interface{}) MapTerraformIterator {
	_init_.Initialize()

	if err := validateTerraformIterator_FromMapParameters(map_); err != nil {
		panic(err)
	}
	var returns MapTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.TerraformIterator",
		"fromMap",
		[]interface{}{map_},
		&returns,
	)

	return returns
}

// Creates a new iterator from a resource that has been created with the `for_each` argument.
// Experimental.
func TerraformIterator_FromResources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateTerraformIterator_FromResourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.TerraformIterator",
		"fromResources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) Dynamic(attributes *map[string]interface{}) IResolvable {
	if err := t.validateDynamicParameters(attributes); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"dynamic",
		[]interface{}{attributes},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) ForExpressionForList(expression interface{}) IResolvable {
	if err := t.validateForExpressionForListParameters(expression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"forExpressionForList",
		[]interface{}{expression},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) ForExpressionForMap(keyExpression interface{}, valueExpression interface{}) IResolvable {
	if err := t.validateForExpressionForMapParameters(keyExpression, valueExpression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"forExpressionForMap",
		[]interface{}{keyExpression, valueExpression},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetAny(attribute *string) IResolvable {
	if err := t.validateGetAnyParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"getAny",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetAnyMap(attribute *string) *map[string]interface{} {
	if err := t.validateGetAnyMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		t,
		"getAnyMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetBoolean(attribute *string) IResolvable {
	if err := t.validateGetBooleanParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"getBoolean",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetBooleanMap(attribute *string) *map[string]*bool {
	if err := t.validateGetBooleanMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		t,
		"getBooleanMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetList(attribute *string) *[]*string {
	if err := t.validateGetListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		t,
		"getList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetMap(attribute *string) *map[string]interface{} {
	if err := t.validateGetMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		t,
		"getMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetNumber(attribute *string) *float64 {
	if err := t.validateGetNumberParameters(attribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		t,
		"getNumber",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetNumberList(attribute *string) *[]*float64 {
	if err := t.validateGetNumberListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		t,
		"getNumberList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetNumberMap(attribute *string) *map[string]*float64 {
	if err := t.validateGetNumberMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		t,
		"getNumberMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetString(attribute *string) *string {
	if err := t.validateGetStringParameters(attribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		t,
		"getString",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) GetStringMap(attribute *string) *map[string]*string {
	if err := t.validateGetStringMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		t,
		"getStringMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) Keys() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"keys",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) PluckProperty(property *string) IResolvable {
	if err := t.validatePluckPropertyParameters(property); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"pluckProperty",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformIterator) Values() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		t,
		"values",
		nil, // no parameters
		&returns,
	)

	return returns
}

