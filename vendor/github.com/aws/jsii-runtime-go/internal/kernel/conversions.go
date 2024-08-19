package kernel

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/jsii-runtime-go/internal/api"
)

var (
	anyType = reflect.TypeOf((*interface{})(nil)).Elem()
)

// CastAndSetToPtr accepts a pointer to any type and attempts to cast the value
// argument to be the same type. Then it sets the value of the pointer element
// to be the newly cast data. This is used to cast payloads from JSII to
// expected return types for Get and Invoke functions.
func (c *Client) CastAndSetToPtr(ptr interface{}, data interface{}) {
	ptrVal := reflect.ValueOf(ptr).Elem()
	dataVal := reflect.ValueOf(data)

	c.castAndSetToPtr(ptrVal, dataVal)
}

// castAndSetToPtr is the same as CastAndSetToPtr except it operates on the
// reflect.Value representation of the pointer and data.
func (c *Client) castAndSetToPtr(ptr reflect.Value, data reflect.Value) {
	if !data.IsValid() {
		// data will not be valid if was made from a nil value, as there would
		// not have been enough type information available to build a valid
		// reflect.Value. In such cases, we must craft the correctly-typed zero
		// value ourselves.
		data = reflect.Zero(ptr.Type())
	} else if ptr.Kind() == reflect.Ptr && ptr.IsNil() {
		// if ptr is a Pointer type and data is valid, initialize a non-nil pointer
		// type. Otherwise inner value is not-settable upon recursion. See third
		// law of reflection.
		// https://blog.golang.org/laws-of-reflection
		ptr.Set(reflect.New(ptr.Type().Elem()))
		c.castAndSetToPtr(ptr.Elem(), data)
		return
	} else if data.Kind() == reflect.Interface && !data.IsNil() {
		// If data is a non-nil interface, unwrap it to get it's dynamic value
		// type sorted out, so that further calls in this method don't have to
		// worry about this edge-case when reasoning on kinds.
		data = reflect.ValueOf(data.Interface())
	}

	if ref, isRef := castValToRef(data); isRef {
		// If return data is a jsii struct passed by reference, de-reference it all.
		if fields, _, isStruct := c.Types().StructFields(ptr.Type()); isStruct {
			for _, field := range fields {
				got, err := c.Get(GetProps{
					Property: field.Tag.Get("json"),
					ObjRef:   ref,
				})
				if err != nil {
					panic(err)
				}
				fieldVal := ptr.FieldByIndex(field.Index)
				c.castAndSetToPtr(fieldVal, reflect.ValueOf(got.Value))
			}
			return
		}

		targetType := ptr.Type()
		if typ, ok := c.Types().FindType(ref.TypeFQN()); ok && typ.AssignableTo(ptr.Type()) {
			// Specialize the return type to be the dynamic value type
			targetType = typ
		}

		// If it's currently tracked, return the current instance
		if object, ok := c.objects.GetObjectAs(ref.InstanceID, targetType); ok {
			ptr.Set(object)
			return
		}

		// If return data is jsii object references, add to objects table.
		if err := c.Types().InitJsiiProxy(ptr, targetType); err == nil {
			if err = c.RegisterInstance(ptr, ref); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
		return
	}

	if enumref, isEnum := castValToEnumRef(data); isEnum {
		member, err := c.Types().EnumMemberForEnumRef(enumref)
		if err != nil {
			panic(err)
		}

		ptr.Set(reflect.ValueOf(member))
		return
	}

	if date, isDate := castValToDate(data); isDate {
		ptr.Set(reflect.ValueOf(date))
		return
	}

	// maps
	if m, isMap := c.castValToMap(data, ptr.Type()); isMap {
		ptr.Set(m)
		return
	}

	// arrays
	if data.Kind() == reflect.Slice {
		len := data.Len()
		var slice reflect.Value
		if ptr.Kind() == reflect.Slice {
			slice = reflect.MakeSlice(ptr.Type(), len, len)
		} else {
			slice = reflect.MakeSlice(reflect.SliceOf(anyType), len, len)
		}

		// If return type is a slice, recursively cast elements
		for i := 0; i < len; i++ {
			c.castAndSetToPtr(slice.Index(i), data.Index(i))
		}

		ptr.Set(slice)
		return
	}

	ptr.Set(data)
}

// Accepts pointers to structs that implement interfaces and searches for an
// existing object reference in the kernel. If it exists, it casts it to an
// objref for the runtime. Recursively casts types that may contain nested
// object references.
func (c *Client) CastPtrToRef(dataVal reflect.Value) interface{} {
	if !dataVal.IsValid() {
		// dataVal is a 0-value, meaning we have no value available... We return
		// this to JavaScript as a "null" value.
		return nil
	}
	if (dataVal.Kind() == reflect.Interface || dataVal.Kind() == reflect.Ptr) && dataVal.IsNil() {
		return nil
	}

	// In case we got a time.Time value (or pointer to one).
	if wireDate, isDate := castPtrToDate(dataVal); isDate {
		return wireDate
	}

	switch dataVal.Kind() {
	case reflect.Map:
		result := api.WireMap{MapData: make(map[string]interface{})}

		iter := dataVal.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			val := iter.Value()
			result.MapData[key] = c.CastPtrToRef(val)
		}

		return result

	case reflect.Interface, reflect.Ptr:
		if valref, valHasRef := c.FindObjectRef(dataVal); valHasRef {
			return valref
		}

		// In case we got a pointer to a map, slice, enum, ...
		if elem := reflect.Indirect(dataVal.Elem()); elem.Kind() != reflect.Struct {
			return c.CastPtrToRef(elem)
		}

		if dataVal.Elem().Kind() == reflect.Struct {
			elemVal := dataVal.Elem()
			if fields, fqn, isStruct := c.Types().StructFields(elemVal.Type()); isStruct {
				data := make(map[string]interface{})
				for _, field := range fields {
					fieldVal := elemVal.FieldByIndex(field.Index)
					if (fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface) && fieldVal.IsNil() {
						// If there is the "field" tag, and it's "required", then panic since the value is nil.
						if requiredOrOptional, found := field.Tag.Lookup("field"); found && requiredOrOptional == "required" {
							panic(fmt.Sprintf("Field %v.%v is required, but has nil value", field.Type, field.Name))
						}
						continue
					}
					key := field.Tag.Get("json")
					data[key] = c.CastPtrToRef(fieldVal)
				}

				return api.WireStruct{
					StructDescriptor: api.StructDescriptor{
						FQN:    fqn,
						Fields: data,
					},
				}
			}
		} else if dataVal.Elem().Kind() == reflect.Ptr {
			// Typically happens when a struct pointer is passed into an interface{}
			// typed API (such as a place where a union is accepted).
			elemVal := dataVal.Elem()
			return c.CastPtrToRef(elemVal)
		}

		if ref, err := c.ManageObject(dataVal); err != nil {
			panic(err)
		} else {
			return ref
		}

	case reflect.Slice:
		refs := make([]interface{}, dataVal.Len())
		for i := 0; i < dataVal.Len(); i++ {
			refs[i] = c.CastPtrToRef(dataVal.Index(i))
		}
		return refs

	case reflect.String:
		if enumRef, isEnumRef := c.Types().TryRenderEnumRef(dataVal); isEnumRef {
			return enumRef
		}
	}
	return dataVal.Interface()
}

// castPtrToDate obtains an api.WireDate from the provided reflect.Value if it
// represents a time.Time or *time.Time value. It accepts both a pointer and
// direct value as a convenience (when passing time.Time through an interface{}
// parameter, having to unwrap it as a pointer is annoying and unneeded).
func castPtrToDate(data reflect.Value) (wireDate api.WireDate, ok bool) {
	var timestamp *time.Time
	if timestamp, ok = data.Interface().(*time.Time); !ok {
		var val time.Time
		if val, ok = data.Interface().(time.Time); ok {
			timestamp = &val
		}
	}
	if ok {
		wireDate.Timestamp = timestamp.Format(time.RFC3339Nano)
	}
	return
}

func castValToRef(data reflect.Value) (ref api.ObjectRef, ok bool) {
	if data.Kind() == reflect.Map {
		for _, k := range data.MapKeys() {
			// Finding values type requires extracting from reflect.Value
			// otherwise .Kind() returns `interface{}`
			v := reflect.ValueOf(data.MapIndex(k).Interface())

			if k.Kind() != reflect.String {
				continue
			}

			switch k.String() {
			case "$jsii.byref":
				if v.Kind() != reflect.String {
					ok = false
					return
				}
				ref.InstanceID = v.String()
				ok = true
			case "$jsii.interfaces":
				if v.Kind() != reflect.Slice {
					continue
				}
				ifaces := make([]api.FQN, v.Len())
				for i := 0; i < v.Len(); i++ {
					e := reflect.ValueOf(v.Index(i).Interface())
					if e.Kind() != reflect.String {
						ok = false
						return
					}
					ifaces[i] = api.FQN(e.String())
				}
				ref.Interfaces = ifaces
			}

		}
	}

	return ref, ok
}

// TODO: This should return a time.Time instead
func castValToDate(data reflect.Value) (date time.Time, ok bool) {
	if data.Kind() == reflect.Map {
		for _, k := range data.MapKeys() {
			v := reflect.ValueOf(data.MapIndex(k).Interface())
			if k.Kind() == reflect.String && k.String() == "$jsii.date" && v.Kind() == reflect.String {
				var err error
				date, err = time.Parse(time.RFC3339Nano, v.String())
				ok = (err == nil)
				break
			}
		}
	}

	return
}

func castValToEnumRef(data reflect.Value) (enum api.EnumRef, ok bool) {
	ok = false

	if data.Kind() == reflect.Map {
		for _, k := range data.MapKeys() {
			// Finding values type requires extracting from reflect.Value
			// otherwise .Kind() returns `interface{}`
			v := reflect.ValueOf(data.MapIndex(k).Interface())

			if k.Kind() == reflect.String && k.String() == "$jsii.enum" && v.Kind() == reflect.String {
				enum.MemberFQN = v.String()
				ok = true
				break
			}
		}
	}

	return
}

// castValToMap attempts converting the provided jsii wire value to a
// go map. This recognizes the "$jsii.map" object and does the necessary
// recursive value conversion.
func (c *Client) castValToMap(data reflect.Value, mapType reflect.Type) (m reflect.Value, ok bool) {
	ok = false

	if data.Kind() != reflect.Map || data.Type().Key().Kind() != reflect.String {
		return
	}

	if mapType.Kind() == reflect.Map && mapType.Key().Kind() != reflect.String {
		return
	}
	if mapType == anyType {
		mapType = reflect.TypeOf((map[string]interface{})(nil))
	}

	dataIter := data.MapRange()
	for dataIter.Next() {
		key := dataIter.Key().String()
		if key != "$jsii.map" {
			continue
		}

		// Finding value type requries extracting from reflect.Value
		// otherwise .Kind() returns `interface{}`
		val := reflect.ValueOf(dataIter.Value().Interface())
		if val.Kind() != reflect.Map {
			return
		}

		ok = true

		m = reflect.MakeMap(mapType)

		iter := val.MapRange()
		for iter.Next() {
			val := iter.Value()
			// Note: reflect.New(t) returns a pointer to a newly allocated t
			convertedVal := reflect.New(mapType.Elem()).Elem()
			c.castAndSetToPtr(convertedVal, val)

			m.SetMapIndex(iter.Key(), convertedVal)
		}
		return
	}
	return
}
