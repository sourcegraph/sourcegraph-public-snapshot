package jsii

import (
	"reflect"

	"github.com/aws/jsii-runtime-go/internal/api"
	"github.com/aws/jsii-runtime-go/runtime"
)

// Deprecated: FQN represents a fully-qualified type name in the jsii type system.
type FQN api.FQN

// Deprecated: Member is a runtime descriptor for a class or interface member
type Member interface {
	asRuntimeMember() runtime.Member
}

// Deprecated: MemberMethod is a runtime descriptor for a class method (implementation of Member)
type MemberMethod api.MethodOverride

func (m MemberMethod) asRuntimeMember() runtime.Member {
	return runtime.MemberMethod(m)
}

// Deprecated: MemberProperty is a runtime descriptor for a class or interface property (implementation of Member)
type MemberProperty api.PropertyOverride

func (m MemberProperty) asRuntimeMember() runtime.Member {
	return runtime.MemberProperty(m)
}

// Deprecated: Load ensures a npm package is loaded in the jsii kernel.
func Load(name string, version string, tarball []byte) {
	runtime.Load(name, version, tarball)
}

// Deprecated: RegisterClass associates a class fully qualified name to the specified class
// interface, member list, and proxy maker function. Panics if class is not a go
// interface, or if the provided fqn was already used to register a different type.
func RegisterClass(fqn FQN, class reflect.Type, members []Member, maker func() interface{}) {
	rm := make([]runtime.Member, len(members))
	for i, m := range members {
		rm[i] = m.asRuntimeMember()
	}
	runtime.RegisterClass(runtime.FQN(fqn), class, rm, maker)
}

// Deprecated: RegisterEnum associates an enum's fully qualified name to the specified enum
// type, and members. Panics if enum is not a reflect.String type, any value in
// the provided members map is of a type other than enum, or if the provided
// fqn was already used to register a different type.
func RegisterEnum(fqn FQN, enum reflect.Type, members map[string]interface{}) {
	runtime.RegisterEnum(runtime.FQN(fqn), enum, members)
}

// Deprecated: RegisterInterface associates an interface's fully qualified name to the
// specified interface type, member list, and proxy maker function. Panics if iface is not
// an interface, or if the provided fqn was already used to register a different type.
func RegisterInterface(fqn FQN, iface reflect.Type, members []Member, maker func() interface{}) {
	rm := make([]runtime.Member, len(members))
	for i, m := range members {
		rm[i] = m.asRuntimeMember()
	}
	runtime.RegisterInterface(runtime.FQN(fqn), iface, rm, maker)
}

// Deprecated: RegisterStruct associates a struct's fully qualified name to the specified
// struct type. Panics if strct is not a struct, or if the provided fqn was
// already used to register a different type.
func RegisterStruct(fqn FQN, strct reflect.Type) {
	runtime.RegisterStruct(runtime.FQN(fqn), strct)
}

// Deprecated: InitJsiiProxy initializes a jsii proxy instance at the provided pointer.
// Panics if the pointer cannot be initialized to a proxy instance (i.e: the
// element of it is not a registered jsii interface or class type).
func InitJsiiProxy(ptr interface{}) {
	runtime.InitJsiiProxy(ptr)
}

// Deprecated: Create will construct a new JSII object within the kernel runtime. This is
// called by jsii object constructors.
func Create(fqn FQN, args []interface{}, inst interface{}) {
	runtime.Create(runtime.FQN(fqn), args, inst)
}

// Deprecated: Invoke will call a method on a jsii class instance. The response will be
// decoded into the expected return type for the method being called.
func Invoke(obj interface{}, method string, args []interface{}, ret interface{}) {
	runtime.Invoke(obj, method, args, ret)
}

// Deprecated: InvokeVoid will call a void method on a jsii class instance.
func InvokeVoid(obj interface{}, method string, args []interface{}) {
	runtime.InvokeVoid(obj, method, args)
}

// Deprecated: StaticInvoke will call a static method on a given jsii class. The response
// will be decoded into the expected return type for the method being called.
func StaticInvoke(fqn FQN, method string, args []interface{}, ret interface{}) {
	runtime.StaticInvoke(runtime.FQN(fqn), method, args, ret)
}

// Deprecated: StaticInvokeVoid will call a static void method on a given jsii class.
func StaticInvokeVoid(fqn FQN, method string, args []interface{}) {
	runtime.StaticInvokeVoid(runtime.FQN(fqn), method, args)
}

// Deprecated: Get reads a property value on a given jsii class instance. The response
// should be decoded into the expected type of the property being read.
func Get(obj interface{}, property string, ret interface{}) {
	runtime.Get(obj, property, ret)
}

// Deprecated: StaticGet reads a static property value on a given jsii class. The response
// should be decoded into the expected type of the property being read.
func StaticGet(fqn FQN, property string, ret interface{}) {
	runtime.StaticGet(runtime.FQN(fqn), property, ret)
}

// Deprecated: Set writes a property on a given jsii class instance. The value should match
// the type of the property being written, or the jsii kernel will crash.
func Set(obj interface{}, property string, value interface{}) {
	runtime.Set(obj, property, value)
}

// Deprecated: StaticSet writes a static property on a given jsii class. The value should
// match the type of the property being written, or the jsii kernel will crash.
func StaticSet(fqn FQN, property string, value interface{}) {
	runtime.StaticSet(runtime.FQN(fqn), property, value)
}
