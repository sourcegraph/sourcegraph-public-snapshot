package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
)

func checkAttributeName(name string) {
	// starts w/ a A-Za-z followed by a A-Za-z0-9, a dollar sign, a hyphen or an underscore
	match, err := regexp.MatchString(`^[A-Za-z][\w$-]*$`, name)
	if err != nil {
		panic(err)
	}

	if !match {
		panic(fmt.Sprintf("invalid attribute name %q", name))
	}
}

// AttributeDataType is a single keyword indicating the derived data type from JSON.
type AttributeDataType struct {
	t attributeType
}

// AttributeTypeDecimal indicates that the data type is a real number with at least one digit to the left and right of the period.
// This is the default value.
func AttributeTypeDecimal() AttributeDataType {
	return AttributeDataType{t: attributeDataTypeDecimal}
}

// AttributeTypeInteger indicates that the data type is a whole number with no fractional digits or decimal.
func AttributeTypeInteger() AttributeDataType {
	return AttributeDataType{t: attributeDataTypeInteger}
}

// AttributeMutability is a single keyword indicating the circumstances under which the value of the attribute can be
// (re)defined.
type AttributeMutability struct {
	m attributeMutability
}

// AttributeMutabilityImmutable indicates that the attribute MAY be defined at resource creation (e.g., POST) or at
// record replacement via a request (e.g., a PUT). The attribute SHALL NOT be updated.
func AttributeMutabilityImmutable() AttributeMutability {
	return AttributeMutability{m: attributeMutabilityImmutable}
}

// AttributeMutabilityReadOnly indicates that the attribute SHALL NOT be modified.
func AttributeMutabilityReadOnly() AttributeMutability {
	return AttributeMutability{m: attributeMutabilityReadOnly}
}

// AttributeMutabilityReadWrite indicates that the attribute MAY be updated and read at any time.
// This is the default value.
func AttributeMutabilityReadWrite() AttributeMutability {
	return AttributeMutability{m: attributeMutabilityReadWrite}
}

// AttributeMutabilityWriteOnly indicates that the attribute MAY be updated at any time. Attribute values SHALL NOT
// be returned (e.g., because the value is a stored hash).
// Note: An attribute with a mutability of "writeOnly" usually also has a returned setting of "never".
func AttributeMutabilityWriteOnly() AttributeMutability {
	return AttributeMutability{m: attributeMutabilityWriteOnly}
}

// AttributeReferenceType is a single keyword indicating the reference type of the SCIM resource that may be referenced.
// This attribute is only applicable for attributes that are of type "reference".
type AttributeReferenceType string

const (
	// AttributeReferenceTypeExternal indicates that the resource is an external resource.
	AttributeReferenceTypeExternal AttributeReferenceType = "external"
	// AttributeReferenceTypeURI indicates that the reference is to a service endpoint or an identifier.
	AttributeReferenceTypeURI AttributeReferenceType = "uri"
)

// AttributeReturned is a single keyword indicating the circumstances under which an attribute and associated values are
// returned in response to a GET request or in response to a PUT, POST, or PATCH request.
type AttributeReturned struct {
	r attributeReturned
}

// AttributeReturnedAlways indicates that the attribute is always returned.
func AttributeReturnedAlways() AttributeReturned {
	return AttributeReturned{r: attributeReturnedAlways}
}

// AttributeReturnedDefault indicates that the attribute is returned by default in all SCIM operation responses
// where attribute values are returned.
func AttributeReturnedDefault() AttributeReturned {
	return AttributeReturned{r: attributeReturnedDefault}
}

// AttributeReturnedNever indicates that the attribute is never returned.
func AttributeReturnedNever() AttributeReturned {
	return AttributeReturned{r: attributeReturnedNever}
}

// AttributeReturnedRequest indicates that the attribute is returned in response to any PUT, POST, or PATCH
// operations if the attribute was specified by the client (for example, the attribute was modified).
func AttributeReturnedRequest() AttributeReturned {
	return AttributeReturned{r: attributeReturnedRequest}
}

// AttributeUniqueness is a single keyword value that specifies how the service provider enforces uniqueness of attribute values.
type AttributeUniqueness struct {
	u attributeUniqueness
}

// AttributeUniquenessGlobal indicates that the value SHOULD be globally unique (e.g., an email address, a GUID, or
// other value). No two resources on any server SHOULD possess the same value.
func AttributeUniquenessGlobal() AttributeUniqueness {
	return AttributeUniqueness{u: attributeUniquenessGlobal}
}

// AttributeUniquenessNone indicates that the values are not intended to be unique in any way.
// This is the default value.
func AttributeUniquenessNone() AttributeUniqueness {
	return AttributeUniqueness{u: attributeUniquenessNone}
}

// AttributeUniquenessServer indicates that the value SHOULD be unique within the context of the current SCIM
// endpoint (or tenancy).  No two resources on the same server SHOULD possess the same value.
func AttributeUniquenessServer() AttributeUniqueness {
	return AttributeUniqueness{u: attributeUniquenessServer}
}

type attributeMutability int

const (
	attributeMutabilityReadWrite attributeMutability = iota
	attributeMutabilityImmutable
	attributeMutabilityReadOnly
	attributeMutabilityWriteOnly
)

func (a attributeMutability) MarshalJSON() ([]byte, error) {
	switch a {
	case attributeMutabilityImmutable:
		return json.Marshal("immutable")
	case attributeMutabilityReadOnly:
		return json.Marshal("readOnly")
	case attributeMutabilityWriteOnly:
		return json.Marshal("writeOnly")
	default:
		return json.Marshal("readWrite")
	}
}

type attributeReturned int

const (
	attributeReturnedDefault attributeReturned = iota
	attributeReturnedAlways
	attributeReturnedNever
	attributeReturnedRequest
)

func (a attributeReturned) MarshalJSON() ([]byte, error) {
	switch a {
	case attributeReturnedAlways:
		return json.Marshal("always")
	case attributeReturnedNever:
		return json.Marshal("never")
	case attributeReturnedRequest:
		return json.Marshal("request")
	default:
		return json.Marshal("default")
	}
}

type attributeType int

const (
	attributeDataTypeDecimal attributeType = iota
	attributeDataTypeInteger

	attributeDataTypeBinary
	attributeDataTypeBoolean
	attributeDataTypeComplex
	attributeDataTypeDateTime
	attributeDataTypeReference
	attributeDataTypeString
)

func (a attributeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a attributeType) String() string {
	switch a {
	case attributeDataTypeDecimal:
		return "decimal"
	case attributeDataTypeInteger:
		return "integer"
	case attributeDataTypeBinary:
		return "binary"
	case attributeDataTypeBoolean:
		return "boolean"
	case attributeDataTypeComplex:
		return "complex"
	case attributeDataTypeDateTime:
		return "dateTime"
	case attributeDataTypeReference:
		return "reference"
	default:
		return "string"
	}
}

type attributeUniqueness int

const (
	attributeUniquenessNone attributeUniqueness = iota
	attributeUniquenessGlobal
	attributeUniquenessServer
)

func (a attributeUniqueness) MarshalJSON() ([]byte, error) {
	switch a {
	case attributeUniquenessGlobal:
		return json.Marshal("global")
	case attributeUniquenessServer:
		return json.Marshal("server")
	default:
		return json.Marshal("none")
	}
}
