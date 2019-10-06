package nnz

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	nnzTestRun(t, testSpec{
		nnzValue: func() sql.Scanner { var v String; return &v },
		scanTests: []scanTest{
			{nil, String("")},
			{"", String("")},
			{"a", String("a")},
			{(*string)(nil), String("")},
			{stringPtr(""), String("")},
			{stringPtr("a"), String("a")},
			{[]byte(nil), String("")},
			{[]byte(""), String("")},
			{[]byte("a"), String("a")},
			{(*[]byte)(nil), String("")},
			{&[]byte{}, String("")},
			{&[]byte{'a'}, String("a")},
			{json.RawMessage(nil), String("")},
			{json.RawMessage(""), String("")},
			{json.RawMessage("a"), String("a")},
			{(*json.RawMessage)(nil), String("")},
			{&json.RawMessage{}, String("")},
			{&json.RawMessage{'a'}, String("a")},
		},
		scanErrorTests: []driver.Value{1},
		valueTests: map[driver.Valuer]interface{}{
			String(""):  nil,
			String("a"): "a",
		},
	})
}

func TestInt64(t *testing.T) {
	nnzTestRun(t, testSpec{
		nnzValue: func() sql.Scanner { var v Int64; return &v },
		scanTests: []scanTest{
			{nil, Int64(0)},
			{int(0), Int64(0)},
			{int(1), Int64(1)},
			{(*int)(nil), Int64(0)},
			{intPtr(0), Int64(0)},
			{intPtr(1), Int64(1)},
			{int32(0), Int64(0)},
			{int32(1), Int64(1)},
			{(*int32)(nil), Int64(0)},
			{int32Ptr(0), Int64(0)},
			{int32Ptr(1), Int64(1)},
			{int64(0), Int64(0)},
			{int64(1), Int64(1)},
			{(*int64)(nil), Int64(0)},
			{int64Ptr(0), Int64(0)},
			{int64Ptr(1), Int64(1)},
		},
		scanErrorTests: []driver.Value{"", uint(0), uint32(0), uint64(0)},
		valueTests: map[driver.Valuer]interface{}{
			Int64(0): nil,
			Int64(1): int64(1),
		},
	})
}

type testSpec struct {
	nnzValue       func() sql.Scanner
	scanTests      []scanTest
	scanErrorTests []driver.Value
	valueTests     map[driver.Valuer]interface{}
}

type scanTest struct {
	value driver.Value
	want  interface{}
}

func nnzTestRun(t *testing.T, spec testSpec) {
	t.Helper()

	for _, scanSpec := range spec.scanTests {
		nnzValue := spec.nnzValue()
		t.Run(fmt.Sprintf("scan %#v", scanSpec.value), func(t *testing.T) {
			if err := nnzValue.Scan(scanSpec.value); err != nil {
				t.Fatal(err)
			}
			if v := reflect.ValueOf(nnzValue).Elem().Interface(); !reflect.DeepEqual(v, scanSpec.want) {
				t.Errorf("got %#v, want %#v", v, scanSpec.want)
			}
		})
	}
	for _, value := range spec.scanErrorTests {
		nnzValue := spec.nnzValue()
		t.Run(fmt.Sprintf("scan error %#v", value), func(t *testing.T) {
			if err := nnzValue.Scan(value); err == nil {
				t.Fatal("want error")
			}
		})
	}
	for driverValuer, want := range spec.valueTests {
		t.Run(fmt.Sprintf("value %#v", driverValuer), func(t *testing.T) {
			v, err := driverValuer.Value()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(v, want) {
				t.Errorf("got %#v, want %#v", v, want)
			}
		})
	}
}

func TestInt32(t *testing.T) {
	tests := map[int32]driver.Value{
		0: nil,
		1: int64(1),
	}
	for i32, want := range tests {
		got := Int32(i32)
		if got != want {
			t.Errorf("%#v: want %#v, got %#v", i32, got, want)
		}
	}
}

func TestToInt32(t *testing.T) {
	tests := []struct {
		value interface{}
		want  int32
		err   bool
	}{
		{value: nil, want: 0},
		{value: 0, want: 0},
		{value: 1, want: 1},
		{value: "a", err: true},
	}
	for _, test := range tests {
		var i32 int32
		dest := ToInt32(&i32)
		if err := dest.Scan(test.value); !test.err && err != nil {
			t.Error(err)
		} else if test.err && err == nil {
			t.Fatal("want error")
		}
		if i32 != test.want {
			t.Errorf("%#v: got %d, want %d", test.value, i32, test.want)
		}
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		value json.RawMessage
		want  driver.Value
	}{
		{nil, nil},
		{json.RawMessage(nil), nil},
		{json.RawMessage("a"), []byte("a")},
	}
	for _, test := range tests {
		got := JSON(test.value)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%#v: want %#v, got %#v", test.value, got, test.want)
		}
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		value interface{}
		want  json.RawMessage
		err   bool
	}{
		{value: nil, want: json.RawMessage(nil)},
		{value: []byte(""), want: json.RawMessage(nil)},
		{value: "", want: json.RawMessage(nil)},
		{value: []byte("{}"), want: json.RawMessage("{}")},
		{value: "{}", want: json.RawMessage("{}")},
		{value: 1, err: true},
	}
	for _, test := range tests {
		var v json.RawMessage
		dest := ToJSON(&v)
		if err := dest.Scan(test.value); !test.err && err != nil {
			t.Error(err)
		} else if test.err && err == nil {
			t.Fatal("want error")
		}
		if !reflect.DeepEqual(v, test.want) {
			t.Errorf("%#v: got %#v, want %#v", test.value, v, test.want)
		}
	}
}

func stringPtr(v string) *string { return &v }
func intPtr(v int) *int          { return &v }
func int32Ptr(v int32) *int32    { return &v }
func int64Ptr(v int64) *int64    { return &v }
