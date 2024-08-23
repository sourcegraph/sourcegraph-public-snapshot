// A programming model for software-defined state
package constructs

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"constructs.Construct",
		reflect.TypeOf((*Construct)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_Construct{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IConstruct)
			return &j
		},
	)
	_jsii_.RegisterEnum(
		"constructs.ConstructOrder",
		reflect.TypeOf((*ConstructOrder)(nil)).Elem(),
		map[string]interface{}{
			"PREORDER": ConstructOrder_PREORDER,
			"POSTORDER": ConstructOrder_POSTORDER,
		},
	)
	_jsii_.RegisterClass(
		"constructs.Dependable",
		reflect.TypeOf((*Dependable)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "dependencyRoots", GoGetter: "DependencyRoots"},
		},
		func() interface{} {
			return &jsiiProxy_Dependable{}
		},
	)
	_jsii_.RegisterClass(
		"constructs.DependencyGroup",
		reflect.TypeOf((*DependencyGroup)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "add", GoMethod: "Add"},
		},
		func() interface{} {
			j := jsiiProxy_DependencyGroup{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IDependable)
			return &j
		},
	)
	_jsii_.RegisterInterface(
		"constructs.IConstruct",
		reflect.TypeOf((*IConstruct)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
		},
		func() interface{} {
			j := jsiiProxy_IConstruct{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IDependable)
			return &j
		},
	)
	_jsii_.RegisterInterface(
		"constructs.IDependable",
		reflect.TypeOf((*IDependable)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_IDependable{}
		},
	)
	_jsii_.RegisterInterface(
		"constructs.IValidation",
		reflect.TypeOf((*IValidation)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "validate", GoMethod: "Validate"},
		},
		func() interface{} {
			return &jsiiProxy_IValidation{}
		},
	)
	_jsii_.RegisterStruct(
		"constructs.MetadataEntry",
		reflect.TypeOf((*MetadataEntry)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"constructs.MetadataOptions",
		reflect.TypeOf((*MetadataOptions)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"constructs.Node",
		reflect.TypeOf((*Node)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addDependency", GoMethod: "AddDependency"},
			_jsii_.MemberMethod{JsiiMethod: "addMetadata", GoMethod: "AddMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "addr", GoGetter: "Addr"},
			_jsii_.MemberMethod{JsiiMethod: "addValidation", GoMethod: "AddValidation"},
			_jsii_.MemberProperty{JsiiProperty: "children", GoGetter: "Children"},
			_jsii_.MemberProperty{JsiiProperty: "defaultChild", GoGetter: "DefaultChild"},
			_jsii_.MemberProperty{JsiiProperty: "dependencies", GoGetter: "Dependencies"},
			_jsii_.MemberMethod{JsiiMethod: "findAll", GoMethod: "FindAll"},
			_jsii_.MemberMethod{JsiiMethod: "findChild", GoMethod: "FindChild"},
			_jsii_.MemberMethod{JsiiMethod: "getAllContext", GoMethod: "GetAllContext"},
			_jsii_.MemberMethod{JsiiMethod: "getContext", GoMethod: "GetContext"},
			_jsii_.MemberProperty{JsiiProperty: "id", GoGetter: "Id"},
			_jsii_.MemberMethod{JsiiMethod: "lock", GoMethod: "Lock"},
			_jsii_.MemberProperty{JsiiProperty: "locked", GoGetter: "Locked"},
			_jsii_.MemberProperty{JsiiProperty: "metadata", GoGetter: "Metadata"},
			_jsii_.MemberProperty{JsiiProperty: "path", GoGetter: "Path"},
			_jsii_.MemberProperty{JsiiProperty: "root", GoGetter: "Root"},
			_jsii_.MemberProperty{JsiiProperty: "scope", GoGetter: "Scope"},
			_jsii_.MemberProperty{JsiiProperty: "scopes", GoGetter: "Scopes"},
			_jsii_.MemberMethod{JsiiMethod: "setContext", GoMethod: "SetContext"},
			_jsii_.MemberMethod{JsiiMethod: "tryFindChild", GoMethod: "TryFindChild"},
			_jsii_.MemberMethod{JsiiMethod: "tryGetContext", GoMethod: "TryGetContext"},
			_jsii_.MemberMethod{JsiiMethod: "tryRemoveChild", GoMethod: "TryRemoveChild"},
			_jsii_.MemberMethod{JsiiMethod: "validate", GoMethod: "Validate"},
		},
		func() interface{} {
			return &jsiiProxy_Node{}
		},
	)
}
