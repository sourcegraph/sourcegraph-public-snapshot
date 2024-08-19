package typeregistry

import (
	"fmt"
	"reflect"

	"github.com/aws/jsii-runtime-go/internal/api"
)

// TypeRegistry is used to record runtime type information about the loaded
// modules, which is later used to correctly convert objects received from the
// JavaScript process into native go values.
type TypeRegistry struct {
	// fqnToType is used to obtain the native go type for a given jsii fully
	// qualified type name. The kind of type being returned depends on what the
	// FQN represents... This will be the second argument of provided to a
	// register* function.
	// enums are not included
	fqnToType map[api.FQN]registeredType

	// fqnToEnumMember maps enum member FQNs (e.g. "jsii-calc.StringEnum/A") to
	// the corresponding go const for this member.
	fqnToEnumMember map[string]interface{}

	// typeToEnumFQN maps Go enum type ("StringEnum") to the corresponding jsii
	// enum FQN (e.g. "jsii-calc.StringEnum")
	typeToEnumFQN map[reflect.Type]api.FQN

	// typeToInterfaceFQN maps Go interface type ("SomeInterface") to the
	// corresponding jsii interface FQN (e.g: "jsii-calc.SomeInterface")
	typeToInterfaceFQN map[reflect.Type]api.FQN

	// structInfo maps registered struct types to all their fields, and possibly a validator
	structInfo map[reflect.Type]registeredStruct

	// proxyMakers map registered interface types to a proxy maker function.
	proxyMakers map[reflect.Type]func() interface{}

	// typeMembers maps each class or interface FQN to the set of members it
	// implements in the form of api.Override values.
	typeMembers map[api.FQN][]api.Override
}

type anonymousProxy struct{ _ int } // Padded so it's not 0-sized

// New creates a new type registry.
func New() *TypeRegistry {
	registry := TypeRegistry{
		fqnToType:          make(map[api.FQN]registeredType),
		fqnToEnumMember:    make(map[string]interface{}),
		typeToEnumFQN:      make(map[reflect.Type]api.FQN),
		typeToInterfaceFQN: make(map[reflect.Type]api.FQN),
		structInfo:         make(map[reflect.Type]registeredStruct),
		proxyMakers:        make(map[reflect.Type]func() interface{}),
		typeMembers:        make(map[api.FQN][]api.Override),
	}

	// Ensure we can initialize proxies for `interface{}` when a method returns `any`.
	registry.proxyMakers[reflect.TypeOf((*interface{})(nil)).Elem()] = func() interface{} {
		return &anonymousProxy{}
	}

	return &registry
}

// IsAnonymousProxy tells whether the value v is an anonymous object proxy, or
// a pointer to one.
func (t *TypeRegistry) IsAnonymousProxy(v interface{}) bool {
	_, ok := v.(*anonymousProxy)
	if !ok {
		_, ok = v.(anonymousProxy)
	}
	return ok
}

// StructFields returns the list of fields associated with a jsii struct type,
// the jsii fully qualified type name, and a boolean telling whether the
// provided type was a registered jsii struct type.
func (t *TypeRegistry) StructFields(typ reflect.Type) (fields []reflect.StructField, fqn api.FQN, ok bool) {
	var info registeredStruct
	if info, ok = t.structInfo[typ]; !ok {
		return
	}

	fqn = info.FQN
	fields = make([]reflect.StructField, len(info.Fields))
	copy(fields, info.Fields)
	return
}

// FindType returns the registered type corresponding to the provided jsii FQN.
func (t *TypeRegistry) FindType(fqn api.FQN) (typ reflect.Type, ok bool) {
	var reg registeredType
	if reg, ok = t.fqnToType[fqn]; ok {
		typ = reg.Type
	}
	return
}

// InitJsiiProxy initializes a jsii proxy value at the provided pointer. It
// returns an error if the pointer does not have a value of a registered
// proxyable type (that is, a class or interface type).
func (t *TypeRegistry) InitJsiiProxy(val reflect.Value, valType reflect.Type) error {
	switch valType.Kind() {
	case reflect.Interface:
		if maker, ok := t.proxyMakers[valType]; ok {
			made := maker()
			val.Set(reflect.ValueOf(made))
			return nil
		}
		return fmt.Errorf("unable to make an instance of unregistered interface %v", valType)

	case reflect.Struct:
		if !val.IsZero() {
			return fmt.Errorf("refusing to initialize non-zero-value struct %v", val)
		}
		numField := valType.NumField()
		for i := 0; i < numField; i++ {
			field := valType.Field(i)
			if field.Name == "_" {
				// Ignore any padding
				continue
			}
			if !field.Anonymous {
				return fmt.Errorf("refusing to initialize non-anonymous field %v of %v", field.Name, val)
			}
			if err := t.InitJsiiProxy(val.Field(i), field.Type); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("unable to make an instance of %v (neither a struct nor interface)", valType)
	}
}

// EnumMemberForEnumRef returns the go enum member corresponding to a jsii fully
// qualified enum member name (e.g: "jsii-calc.StringEnum/A"). If no enum member
// was registered (via registerEnum) for the provided enumref, an error is
// returned.
func (t *TypeRegistry) EnumMemberForEnumRef(ref api.EnumRef) (interface{}, error) {
	if member, ok := t.fqnToEnumMember[ref.MemberFQN]; ok {
		return member, nil
	}
	return nil, fmt.Errorf("no enum member registered for %v", ref.MemberFQN)
}

// TryRenderEnumRef returns an enumref if the provided value corresponds to a
// registered enum type. The returned enumref is nil if the provided enum value
// is a zero-value (i.e: "").
func (t *TypeRegistry) TryRenderEnumRef(value reflect.Value) (ref *api.EnumRef, isEnumRef bool) {
	if value.Kind() != reflect.String {
		isEnumRef = false
		return
	}

	if enumFQN, ok := t.typeToEnumFQN[value.Type()]; ok {
		isEnumRef = true
		if memberName := value.String(); memberName != "" {
			ref = &api.EnumRef{MemberFQN: fmt.Sprintf("%v/%v", enumFQN, memberName)}
		} else {
			ref = nil
		}
	} else {
		isEnumRef = false
	}

	return
}

func (t *TypeRegistry) InterfaceFQN(typ reflect.Type) (fqn api.FQN, found bool) {
	fqn, found = t.typeToInterfaceFQN[typ]
	return
}
