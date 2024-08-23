// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type ListTerraformIterator interface {
	TerraformIterator
	// Returns the currently entry in the list or set that is being iterated over.
	//
	// For lists this is the same as `iterator.value`. If you need the index,
	// use count via `TerraformCount`:
	// https://developer.hashicorp.com/terraform/cdktf/concepts/iterators#using-count
	// Experimental.
	Key() interface{}
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

// The jsii proxy struct for ListTerraformIterator
type jsiiProxy_ListTerraformIterator struct {
	jsiiProxy_TerraformIterator
}

func (j *jsiiProxy_ListTerraformIterator) Key() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"key",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ListTerraformIterator) Value() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"value",
		&returns,
	)
	return returns
}


// Experimental.
func NewListTerraformIterator(list interface{}) ListTerraformIterator {
	_init_.Initialize()

	if err := validateNewListTerraformIteratorParameters(list); err != nil {
		panic(err)
	}
	j := jsiiProxy_ListTerraformIterator{}

	_jsii_.Create(
		"cdktf.ListTerraformIterator",
		[]interface{}{list},
		&j,
	)

	return &j
}

// Experimental.
func NewListTerraformIterator_Override(l ListTerraformIterator, list interface{}) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.ListTerraformIterator",
		[]interface{}{list},
		l,
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
func ListTerraformIterator_FromComplexList(list interface{}, mapKeyAttributeName *string) DynamicListTerraformIterator {
	_init_.Initialize()

	if err := validateListTerraformIterator_FromComplexListParameters(list, mapKeyAttributeName); err != nil {
		panic(err)
	}
	var returns DynamicListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ListTerraformIterator",
		"fromComplexList",
		[]interface{}{list, mapKeyAttributeName},
		&returns,
	)

	return returns
}

// Creates a new iterator from a data source that has been created with the `for_each` argument.
// Experimental.
func ListTerraformIterator_FromDataSources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateListTerraformIterator_FromDataSourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ListTerraformIterator",
		"fromDataSources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

// Creates a new iterator from a list.
// Experimental.
func ListTerraformIterator_FromList(list interface{}) ListTerraformIterator {
	_init_.Initialize()

	if err := validateListTerraformIterator_FromListParameters(list); err != nil {
		panic(err)
	}
	var returns ListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ListTerraformIterator",
		"fromList",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// Creates a new iterator from a map.
// Experimental.
func ListTerraformIterator_FromMap(map_ interface{}) MapTerraformIterator {
	_init_.Initialize()

	if err := validateListTerraformIterator_FromMapParameters(map_); err != nil {
		panic(err)
	}
	var returns MapTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ListTerraformIterator",
		"fromMap",
		[]interface{}{map_},
		&returns,
	)

	return returns
}

// Creates a new iterator from a resource that has been created with the `for_each` argument.
// Experimental.
func ListTerraformIterator_FromResources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateListTerraformIterator_FromResourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ListTerraformIterator",
		"fromResources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) Dynamic(attributes *map[string]interface{}) IResolvable {
	if err := l.validateDynamicParameters(attributes); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"dynamic",
		[]interface{}{attributes},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) ForExpressionForList(expression interface{}) IResolvable {
	if err := l.validateForExpressionForListParameters(expression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"forExpressionForList",
		[]interface{}{expression},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) ForExpressionForMap(keyExpression interface{}, valueExpression interface{}) IResolvable {
	if err := l.validateForExpressionForMapParameters(keyExpression, valueExpression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"forExpressionForMap",
		[]interface{}{keyExpression, valueExpression},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetAny(attribute *string) IResolvable {
	if err := l.validateGetAnyParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"getAny",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetAnyMap(attribute *string) *map[string]interface{} {
	if err := l.validateGetAnyMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		l,
		"getAnyMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetBoolean(attribute *string) IResolvable {
	if err := l.validateGetBooleanParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"getBoolean",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetBooleanMap(attribute *string) *map[string]*bool {
	if err := l.validateGetBooleanMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		l,
		"getBooleanMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetList(attribute *string) *[]*string {
	if err := l.validateGetListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		l,
		"getList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetMap(attribute *string) *map[string]interface{} {
	if err := l.validateGetMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		l,
		"getMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetNumber(attribute *string) *float64 {
	if err := l.validateGetNumberParameters(attribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		l,
		"getNumber",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetNumberList(attribute *string) *[]*float64 {
	if err := l.validateGetNumberListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		l,
		"getNumberList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetNumberMap(attribute *string) *map[string]*float64 {
	if err := l.validateGetNumberMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		l,
		"getNumberMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetString(attribute *string) *string {
	if err := l.validateGetStringParameters(attribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		l,
		"getString",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) GetStringMap(attribute *string) *map[string]*string {
	if err := l.validateGetStringMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		l,
		"getStringMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) Keys() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"keys",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) PluckProperty(property *string) IResolvable {
	if err := l.validatePluckPropertyParameters(property); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"pluckProperty",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_ListTerraformIterator) Values() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		l,
		"values",
		nil, // no parameters
		&returns,
	)

	return returns
}

