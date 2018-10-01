package goon

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shurcooL/go/reflectsource"
)

var config = struct {
	indent string
}{
	indent: "\t",
}

// dumpState contains information about the state of a dump operation.
type dumpState struct {
	w                io.Writer
	depth            int
	pointers         map[uintptr]int
	ignoreNextType   bool
	ignoreNextIndent bool
}

// indent performs indentation according to the depth level and cs.Indent
// option.
func (d *dumpState) indent() {
	if d.ignoreNextIndent {
		d.ignoreNextIndent = false
		return
	}
	d.w.Write(bytes.Repeat([]byte(config.indent), d.depth))
}

// unpackValue returns values inside of non-nil interfaces when possible.
// This is useful for data types like structs, arrays, slices, and maps which
// can contain varying types packed inside an interface.
func (d *dumpState) unpackValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

// dumpPtr handles formatting of pointers by indirecting them as necessary.
func (d *dumpState) dumpPtr(v reflect.Value) {
	// Remove pointers at or below the current depth from map used to detect
	// circular refs.
	for k, depth := range d.pointers {
		if depth >= d.depth {
			delete(d.pointers, k)
		}
	}

	// Figure out how many levels of indirection there are by dereferencing
	// pointers and unpacking interfaces down the chain while detecting circular
	// references.
	nilFound := false
	cycleFound := false
	indirects := 0
	ve := v
	for ve.Kind() == reflect.Ptr {
		if ve.IsNil() {
			nilFound = true
			break
		}
		indirects++
		addr := ve.Pointer()
		if pd, ok := d.pointers[addr]; ok && pd < d.depth {
			cycleFound = true
			indirects--
			break
		}
		d.pointers[addr] = d.depth

		ve = ve.Elem()
		if ve.Kind() == reflect.Interface {
			if ve.IsNil() {
				nilFound = true
				break
			}
			ve = ve.Elem()
		}
	}

	// Display type information.
	d.w.Write(bytes.Repeat(ampersandBytes, indirects))

	// Display dereferenced value.
	switch {
	case nilFound:
		d.w.Write(nilBytes)

	case cycleFound:
		d.w.Write(circularBytes)

	default:
		d.ignoreNextType = true
		d.dump(ve)
	}
}

// dump is the main workhorse for dumping a value.  It uses the passed reflect
// value to figure out what kind of object we are dealing with and formats it
// appropriately.  It is a recursive function, however circular data structures
// are detected and handled properly.
func (d *dumpState) dump(v reflect.Value) {
	// Handle invalid reflect values immediately.
	kind := v.Kind()
	if kind == reflect.Invalid {
		d.w.Write(invalidAngleBytes)
		return
	}

	// Handle pointers specially.
	if kind == reflect.Ptr {
		d.indent()
		d.w.Write(openParenBytes)
		d.w.Write([]byte(typeStringWithoutPackagePrefix(v)))
		d.w.Write(closeParenBytes)
		d.w.Write(openParenBytes)
		d.dumpPtr(v)
		d.w.Write(closeParenBytes)
		return
	}

	// Print type information unless already handled elsewhere.
	var shouldPrintClosingBr = false
	if !d.ignoreNextType {
		d.indent()
		d.w.Write(openParenBytes)
		d.w.Write([]byte(typeStringWithoutPackagePrefix(v)))
		d.w.Write(closeParenBytes)
		d.w.Write(openParenBytes)
		shouldPrintClosingBr = true
	}
	d.ignoreNextType = false

	if v.Type() == timeType {
		t := v.Interface().(time.Time)
		switch t.IsZero() {
		case false:
			var location string
			switch t.Location() {
			case time.UTC:
				location = "time.UTC"
			case time.Local:
				location = "time.Local"
			default:
				location = fmt.Sprintf("must(time.LoadLocation(%q))", t.Location().String())
			}
			fmt.Fprintf(d.w, "time.Date(%d, %d, %d, %d, %d, %d, %d, %s)", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), location)
		case true:
			d.w.Write([]byte("time.Time{}"))
		}
		goto AfterKindSwitch
	}

	switch kind {
	case reflect.Invalid:
		// Do nothing.  We should never get here since invalid has already
		// been handled above.

	case reflect.Bool:
		printBool(d.w, v.Bool())

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		printInt(d.w, v.Int(), 10)

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		printUint(d.w, v.Uint(), 10)

	case reflect.Float32:
		printFloat(d.w, v.Float(), 32)

	case reflect.Float64:
		printFloat(d.w, v.Float(), 64)

	case reflect.Complex64:
		printComplex(d.w, v.Complex(), 32)

	case reflect.Complex128:
		printComplex(d.w, v.Complex(), 64)

	case reflect.Array:
		d.w.Write([]byte(typeStringWithoutPackagePrefix(v)))
		d.w.Write(openBraceNewlineBytes)
		d.depth++
		for i := 0; i < v.Len(); i++ {
			d.dump(d.unpackValue(v.Index(i)))
			d.w.Write(commaNewlineBytes)
		}
		d.depth--
		d.indent()
		d.w.Write(closeBraceBytes)

	case reflect.Slice:
		if v.IsNil() {
			d.w.Write(nilBytes)
		} else {
			d.w.Write([]byte(typeStringWithoutPackagePrefix(v)))
			d.w.Write(openBraceNewlineBytes)
			d.depth++
			for i := 0; i < v.Len(); i++ {
				d.dump(d.unpackValue(v.Index(i)))
				d.w.Write(commaNewlineBytes)
			}
			d.depth--
			d.indent()
			d.w.Write(closeBraceBytes)
		}

	case reflect.String:
		d.w.Write([]byte(strconv.Quote(v.String())))

	case reflect.Interface:
		// If we got here, it's because interface is nil
		// See https://github.com/davecgh/go-spew/issues/12
		d.w.Write(nilBytes)

	case reflect.Ptr:
		// Do nothing.  We should never get here since pointers have already
		// been handled above.

	case reflect.Map:
		if v.IsNil() {
			d.w.Write(nilBytes)
		} else {
			d.w.Write([]byte(typeStringWithoutPackagePrefix(v)))
			d.w.Write(openBraceNewlineBytes)
			d.depth++
			keys := v.MapKeys()
			for _, key := range keys {
				d.dump(d.unpackValue(key))
				d.w.Write(colonSpaceBytes)
				d.ignoreNextIndent = true
				d.dump(d.unpackValue(v.MapIndex(key)))
				d.w.Write(commaNewlineBytes)
			}
			d.depth--
			d.indent()
			d.w.Write(closeBraceBytes)
		}

	case reflect.Struct:
		d.w.Write([]byte(typeStringWithoutPackagePrefix(v)))
		d.w.Write(openBraceBytes)
		d.depth++
		{
			vt := v.Type()
			numFields := v.NumField()
			if numFields > 0 {
				d.w.Write(newlineBytes)
			}
			for i := 0; i < numFields; i++ {
				d.indent()
				vtf := vt.Field(i)
				d.w.Write([]byte(vtf.Name))
				d.w.Write(colonSpaceBytes)
				d.ignoreNextIndent = true
				d.dump(d.unpackValue(v.Field(i)))
				d.w.Write(commaBytes)
				d.w.Write(newlineBytes)
			}
		}
		d.depth--
		d.indent()
		d.w.Write(closeBraceBytes)

	case reflect.Uintptr:
		printHexPtr(d.w, uintptr(v.Uint()))

	case reflect.Func:
		d.w.Write([]byte(reflectsource.GetFuncValueSourceAsString(v)))

	case reflect.UnsafePointer, reflect.Chan:
		printHexPtr(d.w, v.Pointer())

	// There were not any other types at the time this code was written, but
	// fall back to letting the default fmt package handle it in case any new
	// types are added.
	default:
		if v.CanInterface() {
			fmt.Fprintf(d.w, "%v", v.Interface())
		} else {
			fmt.Fprintf(d.w, "%v", v.String())
		}
	}
AfterKindSwitch:

	if shouldPrintClosingBr {
		d.w.Write(closeParenBytes)
	}
}

var timeType = reflect.TypeOf(time.Time{})

func typeStringWithoutPackagePrefix(v reflect.Value) string {
	//return v.Type().String()[len(v.Type().PkgPath())+1:]		// TODO: Error checking?
	//return v.Type().PkgPath()
	//return v.Type().String()
	//return v.Type().Name()

	/*x := v.Type().String()
	if strings.HasPrefix(x, "main.") {
		x = x[len("main."):]
	}
	return x*/

	px := v.Type().String()
	prefix := px[0 : len(px)-len(strings.TrimLeft(px, "*"))] // Split "**main.Lang" -> "**" and "main.Lang"
	x := px[len(prefix):]
	x = strings.TrimPrefix(x, "main.")
	x = strings.TrimPrefix(x, "goon_test.")
	return prefix + x

	/*x = string(debug.Stack())//GetLine(string(debug.Stack()), 0)
	//x = x[1:strings.Index(x, ":")]
	//spew.Printf(">%s<\n", x)
	//panic(nil)
	//st := string(debug.Stack())
	//debug.PrintStack()

	return x*/
}

// fdump is a helper function to consolidate the logic from the various public
// methods which take varying writers and config states.
func fdump(w io.Writer, a ...interface{}) {
	for _, arg := range a {
		d := dumpState{w: w}
		if arg == nil {
			d.w.Write(interfaceBytes)
			d.w.Write(nilParenBytes)
		} else {
			d.pointers = make(map[uintptr]int)
			d.dump(reflect.ValueOf(arg))
		}
		d.w.Write(newlineBytes)
	}
}

// bdump dumps to []byte.
func bdump(a ...interface{}) []byte {
	var buf bytes.Buffer
	fdump(&buf, a...)
	return gofmt(buf.Bytes())
}

func fdumpNamed(w io.Writer, names []string, a ...interface{}) {
	for argIndex, arg := range a {
		d := dumpState{w: w}
		if argIndex < len(names) {
			d.w.Write([]byte(names[argIndex]))
			d.w.Write([]byte(" = "))
		}
		if arg == nil {
			d.w.Write(interfaceBytes)
			d.w.Write(nilParenBytes)
		} else {
			d.pointers = make(map[uintptr]int)
			d.dump(reflect.ValueOf(arg))
		}
		if len(names) >= len(a) {
			d.w.Write(newlineBytes)
		} else {
			if argIndex < len(a)-1 {
				d.w.Write(commaNewlineBytes)
			} else {
				d.w.Write(newlineBytes)
			}
		}
	}
}

func bdumpNamed(names []string, a ...interface{}) []byte {
	var buf bytes.Buffer
	fdumpNamed(&buf, names, a...)
	return gofmt(buf.Bytes())
}

func gofmt(src []byte) []byte {
	formattedSrc, err := format.Source(src)
	if nil != err {
		return []byte("gofmt error (" + err.Error() + ")!\n" + string(src))
	}
	return formattedSrc
}
