package constructs

import (
	_init_ "github.com/aws/constructs-go/constructs/v10/jsii"
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Trait for IDependable.
//
// Traits are interfaces that are privately implemented by objects. Instead of
// showing up in the public interface of a class, they need to be queried
// explicitly. This is used to implement certain framework features that are
// not intended to be used by Construct consumers, and so should be hidden
// from accidental use.
//
// Example:
//   // Usage
//   const roots = Dependable.of(construct).dependencyRoots;
//
//   // Definition
//   Dependable.implement(construct, {
//         dependencyRoots: [construct],
//   });
//
// Experimental.
type Dependable interface {
	// The set of constructs that form the root of this dependable.
	//
	// All resources under all returned constructs are included in the ordering
	// dependency.
	// Experimental.
	DependencyRoots() *[]IConstruct
}

// The jsii proxy struct for Dependable
type jsiiProxy_Dependable struct {
	_ byte // padding
}

func (j *jsiiProxy_Dependable) DependencyRoots() *[]IConstruct {
	var returns *[]IConstruct
	_jsii_.Get(
		j,
		"dependencyRoots",
		&returns,
	)
	return returns
}


// Experimental.
func NewDependable_Override(d Dependable) {
	_init_.Initialize()

	_jsii_.Create(
		"constructs.Dependable",
		nil, // no parameters
		d,
	)
}

// Return the matching Dependable for the given class instance.
// Deprecated: use `of`.
func Dependable_Get(instance IDependable) Dependable {
	_init_.Initialize()

	if err := validateDependable_GetParameters(instance); err != nil {
		panic(err)
	}
	var returns Dependable

	_jsii_.StaticInvoke(
		"constructs.Dependable",
		"get",
		[]interface{}{instance},
		&returns,
	)

	return returns
}

// Turn any object into an IDependable.
// Experimental.
func Dependable_Implement(instance IDependable, trait Dependable) {
	_init_.Initialize()

	if err := validateDependable_ImplementParameters(instance, trait); err != nil {
		panic(err)
	}
	_jsii_.StaticInvokeVoid(
		"constructs.Dependable",
		"implement",
		[]interface{}{instance, trait},
	)
}

// Return the matching Dependable for the given class instance.
// Experimental.
func Dependable_Of(instance IDependable) Dependable {
	_init_.Initialize()

	if err := validateDependable_OfParameters(instance); err != nil {
		panic(err)
	}
	var returns Dependable

	_jsii_.StaticInvoke(
		"constructs.Dependable",
		"of",
		[]interface{}{instance},
		&returns,
	)

	return returns
}

