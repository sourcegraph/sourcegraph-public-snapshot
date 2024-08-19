package constructs

import (
	_init_ "github.com/aws/constructs-go/constructs/v10/jsii"
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Represents the building block of the construct graph.
//
// All constructs besides the root construct must be created within the scope of
// another construct.
type Construct interface {
	IConstruct
	// The tree node.
	Node() Node
	// Returns a string representation of this construct.
	ToString() *string
}

// The jsii proxy struct for Construct
type jsiiProxy_Construct struct {
	jsiiProxy_IConstruct
}

func (j *jsiiProxy_Construct) Node() Node {
	var returns Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}


// Creates a new construct node.
func NewConstruct(scope Construct, id *string) Construct {
	_init_.Initialize()

	if err := validateNewConstructParameters(scope, id); err != nil {
		panic(err)
	}
	j := jsiiProxy_Construct{}

	_jsii_.Create(
		"constructs.Construct",
		[]interface{}{scope, id},
		&j,
	)

	return &j
}

// Creates a new construct node.
func NewConstruct_Override(c Construct, scope Construct, id *string) {
	_init_.Initialize()

	_jsii_.Create(
		"constructs.Construct",
		[]interface{}{scope, id},
		c,
	)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Construct_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateConstruct_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"constructs.Construct",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Construct) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

