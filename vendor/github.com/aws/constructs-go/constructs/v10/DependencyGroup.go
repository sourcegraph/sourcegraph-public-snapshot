package constructs

import (
	_init_ "github.com/aws/constructs-go/constructs/v10/jsii"
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// A set of constructs to be used as a dependable.
//
// This class can be used when a set of constructs which are disjoint in the
// construct tree needs to be combined to be used as a single dependable.
// Experimental.
type DependencyGroup interface {
	IDependable
	// Add a construct to the dependency roots.
	// Experimental.
	Add(scopes ...IDependable)
}

// The jsii proxy struct for DependencyGroup
type jsiiProxy_DependencyGroup struct {
	jsiiProxy_IDependable
}

// Experimental.
func NewDependencyGroup(deps ...IDependable) DependencyGroup {
	_init_.Initialize()

	args := []interface{}{}
	for _, a := range deps {
		args = append(args, a)
	}

	j := jsiiProxy_DependencyGroup{}

	_jsii_.Create(
		"constructs.DependencyGroup",
		args,
		&j,
	)

	return &j
}

// Experimental.
func NewDependencyGroup_Override(d DependencyGroup, deps ...IDependable) {
	_init_.Initialize()

	args := []interface{}{}
	for _, a := range deps {
		args = append(args, a)
	}

	_jsii_.Create(
		"constructs.DependencyGroup",
		args,
		d,
	)
}

func (d *jsiiProxy_DependencyGroup) Add(scopes ...IDependable) {
	args := []interface{}{}
	for _, a := range scopes {
		args = append(args, a)
	}

	_jsii_.InvokeVoid(
		d,
		"add",
		args,
	)
}

