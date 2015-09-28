package dbutil

import (
	"reflect"

	"github.com/jmoiron/sqlx"
)

func ColumnNames(vt reflect.Type) []string {
start:
	switch vt.Kind() {
	case reflect.Ptr, reflect.Slice:
		vt = vt.Elem()
		goto start
	case reflect.Struct:
		break
	}

	var cols []string
	for i := 0; i < vt.NumField(); i++ {
		f := vt.Field(i)
		name := f.Tag.Get("db")
		if name == "-" {
			continue
		}
		if f.Anonymous {
			cols = append(cols, ColumnNames(f.Type)...)
		} else {
			if name == "" {
				name = sqlx.NameMapper(f.Name)
			}
			cols = append(cols, name)
		}
	}
	return cols
}

func ColumnValues(vv reflect.Value) []interface{} {
start:
	vt := vv.Type()
	switch vt.Kind() {
	case reflect.Ptr, reflect.Slice:
		vt = vt.Elem()
		vv = vv.Elem()
		goto start
	case reflect.Struct:
		break
	}

	var vals []interface{}
	for i := 0; i < vt.NumField(); i++ {
		ft := vt.Field(i)
		fv := vv.Field(i)
		name := ft.Tag.Get("db")
		if name == "-" {
			continue
		}
		if ft.Anonymous {
			vals = append(vals, ColumnValues(fv)...)
		} else {
			vals = append(vals, fv.Interface())
		}
	}
	return vals
}
