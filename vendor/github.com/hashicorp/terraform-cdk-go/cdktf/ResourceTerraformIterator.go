// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type ResourceTerraformIterator interface {
	TerraformIterator
	// Returns the current entry in the list or set that is being iterated over.
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

// The jsii proxy struct for ResourceTerraformIterator
type jsiiProxy_ResourceTerraformIterator struct {
	jsiiProxy_TerraformIterator
}

func (j *jsiiProxy_ResourceTerraformIterator) Key() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"key",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ResourceTerraformIterator) Value() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"value",
		&returns,
	)
	return returns
}


// Experimental.
func NewResourceTerraformIterator(element ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateNewResourceTerraformIteratorParameters(element); err != nil {
		panic(err)
	}
	j := jsiiProxy_ResourceTerraformIterator{}

	_jsii_.Create(
		"cdktf.ResourceTerraformIterator",
		[]interface{}{element},
		&j,
	)

	return &j
}

// Experimental.
func NewResourceTerraformIterator_Override(r ResourceTerraformIterator, element ITerraformResource) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.ResourceTerraformIterator",
		[]interface{}{element},
		r,
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
func ResourceTerraformIterator_FromComplexList(list interface{}, mapKeyAttributeName *string) DynamicListTerraformIterator {
	_init_.Initialize()

	if err := validateResourceTerraformIterator_FromComplexListParameters(list, mapKeyAttributeName); err != nil {
		panic(err)
	}
	var returns DynamicListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ResourceTerraformIterator",
		"fromComplexList",
		[]interface{}{list, mapKeyAttributeName},
		&returns,
	)

	return returns
}

// Creates a new iterator from a data source that has been created with the `for_each` argument.
// Experimental.
func ResourceTerraformIterator_FromDataSources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateResourceTerraformIterator_FromDataSourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ResourceTerraformIterator",
		"fromDataSources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

// Creates a new iterator from a list.
// Experimental.
func ResourceTerraformIterator_FromList(list interface{}) ListTerraformIterator {
	_init_.Initialize()

	if err := validateResourceTerraformIterator_FromListParameters(list); err != nil {
		panic(err)
	}
	var returns ListTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ResourceTerraformIterator",
		"fromList",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// Creates a new iterator from a map.
// Experimental.
func ResourceTerraformIterator_FromMap(map_ interface{}) MapTerraformIterator {
	_init_.Initialize()

	if err := validateResourceTerraformIterator_FromMapParameters(map_); err != nil {
		panic(err)
	}
	var returns MapTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ResourceTerraformIterator",
		"fromMap",
		[]interface{}{map_},
		&returns,
	)

	return returns
}

// Creates a new iterator from a resource that has been created with the `for_each` argument.
// Experimental.
func ResourceTerraformIterator_FromResources(resource ITerraformResource) ResourceTerraformIterator {
	_init_.Initialize()

	if err := validateResourceTerraformIterator_FromResourcesParameters(resource); err != nil {
		panic(err)
	}
	var returns ResourceTerraformIterator

	_jsii_.StaticInvoke(
		"cdktf.ResourceTerraformIterator",
		"fromResources",
		[]interface{}{resource},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) Dynamic(attributes *map[string]interface{}) IResolvable {
	if err := r.validateDynamicParameters(attributes); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"dynamic",
		[]interface{}{attributes},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) ForExpressionForList(expression interface{}) IResolvable {
	if err := r.validateForExpressionForListParameters(expression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"forExpressionForList",
		[]interface{}{expression},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) ForExpressionForMap(keyExpression interface{}, valueExpression interface{}) IResolvable {
	if err := r.validateForExpressionForMapParameters(keyExpression, valueExpression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"forExpressionForMap",
		[]interface{}{keyExpression, valueExpression},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetAny(attribute *string) IResolvable {
	if err := r.validateGetAnyParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"getAny",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetAnyMap(attribute *string) *map[string]interface{} {
	if err := r.validateGetAnyMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"getAnyMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetBoolean(attribute *string) IResolvable {
	if err := r.validateGetBooleanParameters(attribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"getBoolean",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetBooleanMap(attribute *string) *map[string]*bool {
	if err := r.validateGetBooleanMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		r,
		"getBooleanMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetList(attribute *string) *[]*string {
	if err := r.validateGetListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		r,
		"getList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetMap(attribute *string) *map[string]interface{} {
	if err := r.validateGetMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"getMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetNumber(attribute *string) *float64 {
	if err := r.validateGetNumberParameters(attribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		r,
		"getNumber",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetNumberList(attribute *string) *[]*float64 {
	if err := r.validateGetNumberListParameters(attribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		r,
		"getNumberList",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetNumberMap(attribute *string) *map[string]*float64 {
	if err := r.validateGetNumberMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		r,
		"getNumberMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetString(attribute *string) *string {
	if err := r.validateGetStringParameters(attribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		r,
		"getString",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) GetStringMap(attribute *string) *map[string]*string {
	if err := r.validateGetStringMapParameters(attribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		r,
		"getStringMap",
		[]interface{}{attribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) Keys() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"keys",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) PluckProperty(property *string) IResolvable {
	if err := r.validatePluckPropertyParameters(property); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"pluckProperty",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_ResourceTerraformIterator) Values() IResolvable {
	var returns IResolvable

	_jsii_.Invoke(
		r,
		"values",
		nil, // no parameters
		&returns,
	)

	return returns
}

