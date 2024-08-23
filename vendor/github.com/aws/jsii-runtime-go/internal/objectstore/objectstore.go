package objectstore

import (
	"fmt"
	"reflect"

	"github.com/aws/jsii-runtime-go/internal/api"
)

// stringSet is a set of strings, implemented as a map from string to an
// arbitrary 0-width value.
type stringSet map[string]struct{}

// ObjectStore tracks object instances for which an identifier has been
// associated. Object to instanceID association is tracked using the object
// memory address (aka pointer value) in order to not have issues with go's
// standard object equality rules (we need distinct - but possibly equal) object
// instances to be considered as separate entities for our purposes.
type ObjectStore struct {
	// objectToID associates an object's memory address (pointer value) with an
	// instanceID. This includes aliases (anonymous embedded values) of objects
	// passed to the Register method.
	objectToID map[uintptr]string

	// idToObject associates an instanceID with the first reflect.Value instance
	// that represents the top-level object that was registered with the
	// instanceID first via the Register method.
	idToObject map[string]reflect.Value

	// idToObjects associates an instanceID with the reflect.Value instances that
	// represent the top-level objects that were registered with the instanceID
	// via the Register method.
	idToObjects map[string]map[reflect.Value]struct{}

	// idToInterfaces associates an instanceID with the set of interfaces that it
	// is known to implement.
	//
	// Incorrect use of the UnsafeCast function may result in an instance's
	// interface list containing interfaces that it does not actually implement.
	idToInterfaces map[string]stringSet
}

// New initializes a new ObjectStore.
func New() *ObjectStore {
	return &ObjectStore{
		objectToID:     make(map[uintptr]string),
		idToObject:     make(map[string]reflect.Value),
		idToObjects:    make(map[string]map[reflect.Value]struct{}),
		idToInterfaces: make(map[string]stringSet),
	}
}

// Register associates the provided value with the given instanceID. It also
// registers any anonymously embedded value (transitively) against the same
// instanceID, so that methods promoted from those resolve the correct
// instanceID, too.
//
// Returns an error if the provided value is not a pointer value; if the value
// or any of it's (transitively) anonymous embeds have already been registered
// against a different instanceID; of if the provided instanceID was already
// associated to a different value.
//
// The call is idempotent: calling Register again with the same value and
// instanceID does not result in an error.
func (o *ObjectStore) Register(value reflect.Value, objectRef api.ObjectRef) error {
	var err error
	if value, err = canonicalValue(value); err != nil {
		return err
	}
	ptr := value.Pointer()

	if existing, found := o.objectToID[ptr]; found {
		if existing == objectRef.InstanceID {
			o.mergeInterfaces(objectRef)
			return nil
		}
		return fmt.Errorf("attempting to register %v as %v, but it was already registered as %v", value, objectRef.InstanceID, existing)
	}

	aliases := findAliases(value)

	if existing, found := o.idToObjects[objectRef.InstanceID]; found {
		if _, found := existing[value]; found {
			o.mergeInterfaces(objectRef)
			return nil
		}
		// Value already exists (e.g: a constructor made a callback with "this"
		// passed as an argument). We make the current value(s) an alias of the new
		// one.
		for existing := range existing {
			aliases = append(aliases, existing)
		}
	}

	for _, alias := range aliases {
		ptr := alias.Pointer()
		if existing, found := o.objectToID[ptr]; found && existing != objectRef.InstanceID {
			return fmt.Errorf("value %v is embedded in %v which has ID %v, but was already assigned %v", alias.String(), value.String(), objectRef.InstanceID, existing)
		}
	}

	o.objectToID[ptr] = objectRef.InstanceID
	// Only add to idToObject if this is the first time this InstanceID is registered
	if _, found := o.idToObject[objectRef.InstanceID]; !found {
		o.idToObject[objectRef.InstanceID] = value
	}
	if _, found := o.idToObjects[objectRef.InstanceID]; !found {
		o.idToObjects[objectRef.InstanceID] = make(map[reflect.Value]struct{})
	}
	o.idToObjects[objectRef.InstanceID][value] = struct{}{}
	for _, alias := range aliases {
		o.objectToID[alias.Pointer()] = objectRef.InstanceID
	}

	o.mergeInterfaces(objectRef)

	return nil
}

// mergeInterfaces adds all interfaces carried by the provided objectRef to the
// tracking set for the objectRef's InstanceID. Does nothing if no interfaces
// are designated on the objectRef.
func (o *ObjectStore) mergeInterfaces(objectRef api.ObjectRef) {
	// If we don't have interfaces, we have nothing to do...
	if objectRef.Interfaces == nil {
		return
	}

	// Find or create the interface list for the relevant InstanceID
	var interfaces stringSet
	if list, found := o.idToInterfaces[objectRef.InstanceID]; found {
		interfaces = list
	} else {
		interfaces = make(stringSet)
		o.idToInterfaces[objectRef.InstanceID] = interfaces
	}

	// Add any missing interface to the list.
	for _, iface := range objectRef.Interfaces {
		interfaces[string(iface)] = struct{}{}
	}
}

// InstanceID attempts to determine the instanceID associated with the provided
// value, if any. Returns the existing instanceID and a boolean informing
// whether an instanceID was already found or not.
//
// The InstanceID method is safe to call with values that are not track-able in
// an ObjectStore (i.e: non-pointer values, primitive values, etc...).
func (o *ObjectStore) InstanceID(value reflect.Value) (instanceID string, found bool) {
	var err error
	if value, err = canonicalValue(value); err == nil {
		ptr := value.Pointer()
		instanceID, found = o.objectToID[ptr]
	}
	return
}

// Interfaces returns the set of interfaces associated with the provided
// instanceID.
//
// It returns a nil slice in case the instancceID is invalid, or if it does not
// have any associated interfaces.
func (o *ObjectStore) Interfaces(instanceID string) []api.FQN {
	if set, found := o.idToInterfaces[instanceID]; found {
		interfaces := make([]api.FQN, 0, len(set))
		for iface := range set {
			interfaces = append(interfaces, api.FQN(iface))
		}
		return interfaces
	} else {
		return nil
	}
}

// GetObject attempts to retrieve the object value associated with the given
// instanceID. Returns the existing value and a boolean informing whether a
// value was associated with this instanceID or not.
//
// The GetObject method is safe to call with an instanceID that was never
// registered with the ObjectStore.
func (o *ObjectStore) GetObject(instanceID string) (value reflect.Value, found bool) {
	value, found = o.idToObject[instanceID]
	return
}

// GetObjectAs attempts to retrieve the object value associated with the given
// instanceID, compatible with the given type. Returns the existing value and a
// boolean informing whether a value was associated with this instanceID and
// compatible with this type or not.
//
// The GetObjectAs method is safe to call with an instanceID that was never
// registered with the ObjectStore.
func (o *ObjectStore) GetObjectAs(instanceID string, typ reflect.Type) (value reflect.Value, found bool) {
	found = false
	if values, exists := o.idToObjects[instanceID]; exists {
		for value = range values {
			if value.Type().AssignableTo(typ) {
				value = value.Convert(typ)
				found = true
				return
			}
		}
	}
	return
}

// canonicalValue ensures the same reference is always considered for object
// identity (especially in maps), so that we don't get surprised by pointer to
// struct versus struct value versus opaque interface value, etc...
func canonicalValue(value reflect.Value) (reflect.Value, error) {
	if value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct {
		return value, nil
	}
	// If this is a pointer to something, de-references it.
	result := reflect.ValueOf(reflect.Indirect(value).Interface())

	if result.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("illegal argument: %v is not a pointer", result.String())
	}

	return result, nil
}

// findAliases traverses the provided object value to recursively identify all
// anonymous embedded values, which will then be registered against the same
// instanceID as the embedding value.
//
// This function assumes the provided value is either a reflect.Struct or a
// pointer to a reflect.Struct (possibly as a reflect.Interface). Calling with
// a nil value, or a value that is not ultimately a reflect.Struct may result
// in panic.
func findAliases(value reflect.Value) []reflect.Value {
	var result []reflect.Value

	// Indirect so we always work on the pointer referree
	value = reflect.Indirect(value)

	t := value.Type()
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		f := t.Field(i)

		// Ignore non-anonymous fields (including padding)
		if !f.Anonymous {
			continue
		}

		fv := value.FieldByIndex(f.Index)
		if fv.Kind() == reflect.Interface {
			// If an interface, de-reference to get to the struct type.
			fv = reflect.ValueOf(fv.Interface())
		}
		if fv.Kind() == reflect.Struct {
			// If a struct, get the address of the member.
			fv = fv.Addr()
		}

		result = append(result, fv)
		// Recurse down to collect nested aliases
		result = append(result, findAliases(fv)...)
	}

	return result
}
