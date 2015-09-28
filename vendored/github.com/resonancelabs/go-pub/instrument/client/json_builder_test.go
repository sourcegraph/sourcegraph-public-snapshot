package client

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	. "testing"
)

func testJsonEqual(t *T, expectedJson string, obj interface{}) {
	cleanActual, _ := json.Marshal(obj)
	tmp := map[string]interface{}{}
	json.Unmarshal([]byte(expectedJson), &tmp)
	cleanExpected, _ := json.Marshal(tmp)
	if string(cleanActual) != string(cleanExpected) {
		t.Errorf("expected != actual.\n\nexpected: %v\n\nactual: %v\n", string(cleanExpected), string(cleanActual))
	}
}

func TestValueToJSON(t *T) {

	checkJson := func(inObj interface{}, expected interface{}, fieldMaxBytes, totalMaxBytes int) {
		// Convert to thrift, then convert that to a json-friendly interface{}.
		rval, err := ValueToSanitizedJSONString(inObj, fieldMaxBytes, totalMaxBytes)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		var obj interface{}
		err = json.Unmarshal([]byte(rval), &obj)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if !reflect.DeepEqual(obj, expected) {
			fmt.Printf("\n\nActual:   %v\n\nExpected: %v\n\n", obj, expected)
			t.Error("Structures differ")
		}
	}

	checkJsonLength := func(inObj interface{}, fieldMaxBytes, totalMaxBytes, expectedMinBytes, expectedMaxBytes int) {
		// Convert to thrift, then convert that to a json-friendly interface{}.
		rval, err := ValueToSanitizedJSONString(inObj, fieldMaxBytes, totalMaxBytes)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		if len(rval) < expectedMinBytes || len(rval) > expectedMaxBytes {
			t.Errorf("Length constraint not satisfied: %v < %v < %v\nActual JSON: %v",
				expectedMinBytes, len(rval), expectedMaxBytes, rval)
		}
	}

	type wackyStruct struct {
		Exported     float32
		unexported   int8
		AGenericMap  map[string]interface{}
		AFloatKeyMap map[float64]interface{}
		SomeFunc     func()
		Interface    interface{}
	}

	emptyMap := map[string]string{}
	var nilMap map[string]string = nil

	ptrToEmptyMap := &emptyMap
	ptrToNilMap := &nilMap

	wacky := &wackyStruct{
		Exported:   3.14,
		unexported: 42,
		AGenericMap: map[string]interface{}{
			"key1": "one",
			"key2": 2,
			"key3": &wackyStruct{},
		},
		AFloatKeyMap: map[float64]interface{}{
			3.14: "pi",
			2.71: struct {
				Name string
				Type string
			}{"e", "transcendental"},
			0.0:          emptyMap,
			0.1:          &emptyMap,
			0.2:          nilMap,
			0.3:          &nilMap,
			0.4:          ptrToEmptyMap,
			0.5:          &ptrToEmptyMap,
			0.6:          ptrToNilMap,
			0.7:          &ptrToNilMap,
			math.Inf(-1): "-infinity",
			math.Inf(+1): "+infinity",
		},
		SomeFunc: func() {
			fmt.Printf("don't panic!")
		},
	}

	expectedResult := map[string]interface{}{
		"AFloatKeyMap": map[string]interface{}{
			"+Inf": "+infinity",
			"-Inf": "-infinity",
			"0":    make(map[string]interface{}),
			"0.1":  make(map[string]interface{}),
			"0.2":  nil,
			"0.3":  nil,
			"0.4":  make(map[string]interface{}),
			"0.5":  make(map[string]interface{}),
			"0.6":  nil,
			"0.7":  nil,
			"2.71": map[string]interface{}{
				"Name": "e",
				"Type": "transcendental",
			},
			"3.14": "pi",
		},
		"AGenericMap": map[string]interface{}{
			"key1": "one",
			"key2": "2",
			"key3": map[string]interface{}{
				"AFloatKeyMap": nil,
				"AGenericMap":  nil,
				"Exported":     "0",
				"Interface":    nil,
				"SomeFunc":     "\u003cfunc\u003e",
			},
		},
		"Exported":  "3.14",
		"Interface": nil,
		"SomeFunc":  "\u003cfunc\u003e",
	}

	checkJson(wacky, expectedResult, 100, 1000) // plenty of maxBytes headroom

	fieldTruncatedResult := map[string]interface{}{
		// Note how all dynamic-length strings longer than 8 chars are truncated.
		"AFloatKe…": map[string]interface{}{
			"+Inf": "+infinit…",
			"-Inf": "-infinit…",
			"0":    make(map[string]interface{}),
			"0.1":  make(map[string]interface{}),
			"0.2":  nil,
			"0.3":  nil,
			"0.4":  make(map[string]interface{}),
			"0.5":  make(map[string]interface{}),
			"0.6":  nil,
			"0.7":  nil,
			"2.71": map[string]interface{}{
				"Name": "e",
				"Type": "transcen…",
			},
			"3.14": "pi",
		},
		"AGeneric…": map[string]interface{}{
			"key1": "one",
			"key2": "2",
			"key3": map[string]interface{}{
				"AFloatKe…": nil,
				"AGeneric…": nil,
				"Exported":  "0",
				"Interfac…": nil,
				"SomeFunc":  "\u003cfunc\u003e",
			},
		},
		"Exported":  "3.14",
		"Interfac…": nil,
		"SomeFunc":  "\u003cfunc\u003e",
	}
	checkJson(wacky, fieldTruncatedResult, 8, 1000) // not enough per-field headroom

	// These strings have a lot of filler, so even a maxBytes val of 25 can
	// result in a considerably longer JSON-encoded final product. E.g., for
	// the totalMaxBytes=25 case, we get something like this:
	//
	//    {"AGenericMap":{"key1":"on…","…":"…"},"Exported":"3.14","…":"…"}
	//
	// It's 74 runes, but only 33 are actual content (including ellipses!).
	checkJsonLength(wacky, 1024 /* plenty of field headroom */, 25, 25, 100)
	checkJsonLength(wacky, 1024 /* plenty of field headroom */, 125, 125, 310)

	// Now try restricting both field and total message length for a single
	// really long string.
	longStr64K := strings.Repeat("-", 1024*64)
	checkJsonLength(longStr64K, 100, 1024*128, 100, 110) // limit field length
	checkJsonLength(longStr64K, 1024*128, 100, 100, 110) // limit total length

	justNil := interface{}(nil)
	checkJson(justNil, nil, 100, 1000)

	justNilMap := (map[string]string)(nil)
	checkJson(justNilMap, nil, 100, 1000)

	justNilInMap := map[string]interface{}{"foo": nil}
	checkJson(justNilInMap, justNilInMap, 100, 1000)

	justInterfaceNil := interface{}(interface{}(nil))
	checkJson(justInterfaceNil, nil, 100, 1000)

	justPointerNil := (*int)(nil)
	checkJson(justPointerNil, nil, 100, 1000)

	type LinkNode struct {
		Value int
		Next  *LinkNode
	}

	objectA := &LinkNode{1, nil}
	objectB := &LinkNode{2, nil}
	objectA.Next = objectB
	objectB.Next = objectA

	// TODO: should test that this does not follow the circular reference
	// (which it does at the moment; that hasn't been addressed)
	_, err := ValueToSanitizedJSONString(objectA, 1024, 4096)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
