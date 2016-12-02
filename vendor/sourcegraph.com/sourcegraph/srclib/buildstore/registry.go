package buildstore

import (
	"reflect"
	"strings"
)

var DataTypes = make(map[string]interface{})
var DataTypeNames = make(map[reflect.Type]string)

// RegisterDataType makes a build data type available by the provided name.
//
// When serializing build data of a certain type, the corresponding type name is
// included in the file basename. When deserializing build data from files, the
// file basename is parsed to determine the data type to deserialize into. For
// example, a data type with name "foo.v0" and type Foo{} would result in files
// named "whatever.foo.v0.json" being written, and those files would deserialize
// into Foo{} structs.
//
// The name should contain a version identifier so that types may be modified
// more easily; for example, name could be "graph.v0" (and a subsequent version
// could be registered as "graph.v1" with a different struct).
//
// The emptyInstance should be an zero value of the type (usually a nil pointer
// to a struct, or a nil slice or map) that holds the build data; it is used to
// instantiate instances dynamically.
//
// If RegisterDataType is called twice with the same type or type name, if name
// is empty, or if emptyInstance is nil, it panics
func RegisterDataType(name string, emptyInstance interface{}) {
	if _, dup := DataTypes[name]; dup {
		panic("build: RegisterDataType called twice for type name " + name)
	}
	if emptyInstance == nil {
		panic("build: RegisterDataType emptyInstance is nil")
	}
	DataTypes[name] = emptyInstance

	typ := reflect.TypeOf(emptyInstance)
	if _, dup := DataTypeNames[typ]; dup {
		panic("build: RegisterDataType called twice for type " + typ.String())
	}
	if name == "" {
		panic("build: RegisterDataType name is nil")
	}
	DataTypeNames[typ] = name
}

func DataTypeSuffix(data interface{}) string {
	var suffix string
	if s, ok := data.(string); ok {
		suffix = s
	} else {
		typ := reflect.TypeOf(data)
		name, registered := DataTypeNames[typ]
		if !registered {
			panic("buildstore: data type not registered: " + typ.String())
		}
		suffix = name
	}
	return suffix + ".json"
}

// DataType returns the data type name and empty instance (previously registered
// with RegisterDataType) to use for a build data file named filename. If no
// registered data type is found for filename, "" and nil are returned.
//
// For example, if a data type "foo.v0" was registered, a filename of
// "qux_foo.v0.json" would return "foo.v0" and the registered type.
func DataType(filename string) (string, interface{}) {
	for name, emptyInstance := range DataTypes {
		if strings.HasSuffix(filename, DataTypeSuffix(emptyInstance)) {
			return name, emptyInstance
		}
	}
	return "", nil
}
