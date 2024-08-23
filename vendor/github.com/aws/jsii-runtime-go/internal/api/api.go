package api

import (
	"fmt"
	"regexp"
)

// FQN represents a fully-qualified type name in the jsii type system.
type FQN string

// Override is a public interface implementing a private method `isOverride`
// implemented by the private custom type `override`. This is embedded by
// MethodOverride and PropertyOverride to simulate the union type of Override =
// MethodOverride | PropertyOverride.
type Override interface {
	GoName() string
	isOverride()
}

type override struct{}

func (o override) isOverride() {}

// MethodOverride is used to register a "go-native" implementation to be
// substituted to the default javascript implementation on the created object.
type MethodOverride struct {
	override

	JsiiMethod string `json:"method"`
	GoMethod   string `json:"cookie"`
}

func (m MethodOverride) GoName() string {
	return m.GoMethod
}

// PropertyOverride is used to register a "go-native" implementation to be
// substituted to the default javascript implementation on the created object.
type PropertyOverride struct {
	override

	JsiiProperty string `json:"property"`
	GoGetter     string `json:"cookie"`
}

func (m PropertyOverride) GoName() string {
	return m.GoGetter
}

func IsMethodOverride(value Override) bool {
	switch value.(type) {
	case MethodOverride, *MethodOverride:
		return true
	default:
		return false
	}
}

func IsPropertyOverride(value Override) bool {
	switch value.(type) {
	case PropertyOverride, *PropertyOverride:
		return true
	default:
		return false
	}
}

type ObjectRef struct {
	InstanceID string `json:"$jsii.byref"`
	Interfaces []FQN  `json:"$jsii.interfaces,omitempty"`
}

func (o *ObjectRef) TypeFQN() FQN {
	re := regexp.MustCompile(`^(.+)@(\d+)$`)
	if parts := re.FindStringSubmatch(o.InstanceID); parts == nil {
		panic(fmt.Errorf("invalid instance id: %#v", o.InstanceID))
	} else {
		return FQN(parts[1])
	}
}

type EnumRef struct {
	MemberFQN string `json:"$jsii.enum"`
}

type WireDate struct {
	Timestamp string `json:"$jsii.date"`
}

type WireMap struct {
	MapData map[string]interface{} `json:"$jsii.map"`
}

type WireStruct struct {
	StructDescriptor `json:"$jsii.struct"`
}

type StructDescriptor struct {
	FQN    FQN                    `json:"fqn"`
	Fields map[string]interface{} `json:"data"`
}
