package typed

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

var (
	// Used by ToBytes to indicate that the key was not
	// present in the type
	KeyNotFound = errors.New("Key not found")
	Empty       = Typed(nil)
)

// A Typed type helper for accessing a map
type Typed map[string]interface{}

// Wrap the map into a Typed
func New(m map[string]interface{}) Typed {
	return Typed(m)
}

// Create a Typed helper from the given JSON bytes
func Json(data []byte) (Typed, error) {
	return JsonReader(bytes.NewReader(data))
}

// Create a Typed helper from the given JSON bytes, panics on error
func Must(data []byte) Typed {
	m, err := Json(data)
	if err != nil {
		panic(err)
	}
	return m
}

// Create a Typed helper from the given JSON stream
func JsonReader(reader io.Reader) (Typed, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	var m map[string]interface{}
	err := decoder.Decode(&m)
	return Typed(m), err
}

// Create a Typed helper from the given JSON string
func JsonString(data string) (Typed, error) {
	return JsonReader(strings.NewReader(data))
}

// Create a Typed helper from the JSON within a file
func JsonFile(path string) (Typed, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Json(data)
}

// Create an array of Typed helpers
// Used for when the root is an array which contains objects
func JsonArray(data []byte) ([]Typed, error) {
	return JsonReaderArray(bytes.NewReader(data))
}

// Create an array of Typed helpers given JSON stream
func JsonReaderArray(reader io.Reader) ([]Typed, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	var m []interface{}
	err := decoder.Decode(&m)
	if err != nil {
		return nil, err
	}
	l := len(m)
	if l == 0 {
		return nil, nil
	}
	typed := make([]Typed, l)
	for i := 0; i < l; i++ {
		value := m[i]
		if t, ok := value.(map[string]interface{}); ok {
			typed[i] = t
		} else {
			typed[i] = map[string]interface{}{"0": value}
		}
	}
	return typed, nil
}

// Create an array of Typed helpers from a string
// Used for when the root is an array which contains objects
func JsonStringArray(data string) ([]Typed, error) {
	return JsonReaderArray(strings.NewReader(data))
}

// Create an array of Typed helpers from a file
// Used for when the root is an array which contains objects
func JsonFileArray(path string) ([]Typed, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return JsonArray(data)
}

func (t Typed) Keys() []string {
	keys := make([]string, len(t))
	i := 0
	for k := range t {
		keys[i] = k
		i++
	}
	return keys
}

// Returns a boolean at the key, or false if it
// doesn't exist, or if it isn't a bool
func (t Typed) Bool(key string) bool {
	return t.BoolOr(key, false)
}

// Returns a boolean at the key, or the specified
// value if it doesn't exist or isn't a bool
func (t Typed) BoolOr(key string, d bool) bool {
	if value, exists := t.BoolIf(key); exists {
		return value
	}
	return d
}

// Returns a bool or panics
func (t Typed) BoolMust(key string) bool {
	b, exists := t.BoolIf(key)
	if exists == false {
		panic("expected boolean value for " + key)
	}
	return b
}

// Returns a boolean at the key and whether
// or not the key existed and the value was a bolean
func (t Typed) BoolIf(key string) (bool, bool) {
	value, exists := t[key]
	if exists == false {
		return false, false
	}
	if n, ok := value.(bool); ok {
		return n, true
	}
	return false, false
}

func (t Typed) Int(key string) int {
	return t.IntOr(key, 0)
}

// Returns a int at the key, or the specified
// value if it doesn't exist or isn't a int
func (t Typed) IntOr(key string, d int) int {
	if value, exists := t.IntIf(key); exists {
		return value
	}
	return d
}

// Returns an int or panics
func (t Typed) IntMust(key string) int {
	i, exists := t.IntIf(key)
	if exists == false {
		panic("expected int value for " + key)
	}
	return i
}

// Returns an int at the key and whether
// or not the key existed and the value was an int
func (t Typed) IntIf(key string) (int, bool) {
	value, exists := t[key]
	if exists == false {
		return 0, false
	}

	switch t := value.(type) {
	case int:
		return t, true
	case int16:
		return int(t), true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	case string:
		i, err := strconv.Atoi(t)
		return i, err == nil
	case json.Number:
		i, err := t.Int64()
		return int(i), err == nil
	}
	return 0, false
}

func (t Typed) Float(key string) float64 {
	return t.FloatOr(key, 0)
}

// Returns a float at the key, or the specified
// value if it doesn't exist or isn't a float
func (t Typed) FloatOr(key string, d float64) float64 {
	if value, exists := t.FloatIf(key); exists {
		return value
	}
	return d
}

// Returns an float or panics
func (t Typed) FloatMust(key string) float64 {
	f, exists := t.FloatIf(key)
	if exists == false {
		panic("expected float value for " + key)
	}
	return f
}

// Returns an float at the key and whether
// or not the key existed and the value was an float
func (t Typed) FloatIf(key string) (float64, bool) {
	value, exists := t[key]
	if exists == false {
		return 0, false
	}
	switch t := value.(type) {
	case float64:
		return t, true
	case string:
		f, err := strconv.ParseFloat(t, 10)
		return f, err == nil
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	}
	return 0, false
}

func (t Typed) String(key string) string {
	return t.StringOr(key, "")
}

// Returns a string at the key, or the specified
// value if it doesn't exist or isn't a string
func (t Typed) StringOr(key string, d string) string {
	if value, exists := t.StringIf(key); exists {
		return value
	}
	return d
}

// Returns an string or panics
func (t Typed) StringMust(key string) string {
	s, exists := t.StringIf(key)
	if exists == false {
		panic("expected string value for " + key)
	}
	return s
}

// Returns an string at the key and whether
// or not the key existed and the value was an string
func (t Typed) StringIf(key string) (string, bool) {
	value, exists := t[key]
	if exists == false {
		return "", false
	}
	if n, ok := value.(string); ok {
		return n, true
	}
	return "", false
}

func (t Typed) Time(key string) time.Time {
	return t.TimeOr(key, time.Now())
}

// Returns a time at the key, or the specified
// value if it doesn't exist or isn't a time
func (t Typed) TimeOr(key string, d time.Time) time.Time {
	if value, exists := t.TimeIf(key); exists {
		return value
	}
	return d
}

// Returns a time.Time or panics
func (t Typed) TimeMust(key string) time.Time {
	tt, exists := t.TimeIf(key)
	if exists == false {
		panic("expected time.Time value for " + key)
	}
	return tt
}

// Returns an time.time at the key and whether
// or not the key existed and the value was a time.Time
func (t Typed) TimeIf(key string) (time.Time, bool) {
	value, exists := t[key]
	if exists == false {
		return time.Time{}, false
	}
	if n, ok := value.(time.Time); ok {
		return n, true
	}
	return time.Time{}, false
}

// Returns a Typed helper at the key
// If the key doesn't exist, a default Typed helper
// is returned (which will return default values for
// any subsequent sub queries)
func (t Typed) Object(key string) Typed {
	o := t.ObjectOr(key, nil)
	if o == nil {
		return Typed(nil)
	}
	return o
}

// Returns a Typed helper at the key or the specified
// default if the key doesn't exist or if the key isn't
// a map[string]interface{}
func (t Typed) ObjectOr(key string, d map[string]interface{}) Typed {
	if value, exists := t.ObjectIf(key); exists {
		return value
	}
	return Typed(d)
}

// Returns an typed object or panics
func (t Typed) ObjectMust(key string) Typed {
	t, exists := t.ObjectIf(key)
	if exists == false {
		panic("expected map for " + key)
	}
	return t
}

// Returns a Typed helper at the key and whether
// or not the key existed and the value was an map[string]interface{}
func (t Typed) ObjectIf(key string) (Typed, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	switch t := value.(type) {
	case map[string]interface{}:
		return Typed(t), true
	case Typed:
		return t, true
	}
	return nil, false
}

func (t Typed) Interface(key string) interface{} {
	return t.InterfaceOr(key, nil)
}

// Returns a string at the key, or the specified
// value if it doesn't exist or isn't a strin
func (t Typed) InterfaceOr(key string, d interface{}) interface{} {
	if value, exists := t.InterfaceIf(key); exists {
		return value
	}
	return d
}

// Returns an interface or panics
func (t Typed) InterfaceMust(key string) interface{} {
	i, exists := t.InterfaceIf(key)
	if exists == false {
		panic("expected map for " + key)
	}
	return i
}

// Returns an string at the key and whether
// or not the key existed and the value was an string
func (t Typed) InterfaceIf(key string) (interface{}, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	return value, true
}

// Returns a map[string]interface{} at the key
// or a nil map if the key doesn't exist or if the key isn't
// a map[string]interface
func (t Typed) Map(key string) map[string]interface{} {
	return t.MapOr(key, nil)
}

// Returns a map[string]interface{} at the key
// or the specified default if the key doesn't exist
// or if the key isn't a map[string]interface
func (t Typed) MapOr(key string, d map[string]interface{}) map[string]interface{} {
	if value, exists := t.MapIf(key); exists {
		return value
	}
	return d
}

// Returns a map[string]interface at the key and whether
// or not the key existed and the value was an map[string]interface{}
func (t Typed) MapIf(key string) (map[string]interface{}, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	if n, ok := value.(map[string]interface{}); ok {
		return n, true
	}
	return nil, false
}

// Returns an slice of boolean, or an nil slice
func (t Typed) Bools(key string) []bool {
	return t.BoolsOr(key, nil)
}

// Returns an slice of boolean, or the specified slice
func (t Typed) BoolsOr(key string, d []bool) []bool {
	n, ok := t.BoolsIf(key)
	if ok {
		return n
	}
	return d
}

// Returns a boolean slice + true if valid
// Returns nil + false otherwise
// (returns nil+false if one of the values is not a valid boolean)
func (t Typed) BoolsIf(key string) ([]bool, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	if n, ok := value.([]bool); ok {
		return n, true
	}
	if a, ok := value.([]interface{}); ok {
		l := len(a)
		n := make([]bool, l)
		var ok bool
		for i := 0; i < l; i++ {
			if n[i], ok = a[i].(bool); ok == false {
				return n, false
			}
		}
		return n, true
	}
	return nil, false
}

// Returns an slice of ints, or the specified slice
// Some conversion is done to handle the fact that JSON ints
// are represented as floats.
func (t Typed) Ints(key string) []int {
	return t.IntsOr(key, nil)
}

// Returns an slice of ints, or the specified slice
// if the key doesn't exist or isn't a valid []int.
// Some conversion is done to handle the fact that JSON ints
// are represented as floats.
func (t Typed) IntsOr(key string, d []int) []int {
	n, ok := t.IntsIf(key)
	if ok {
		return n
	}
	return d
}

// Returns a int slice + true if valid
// Returns nil + false otherwise
// (returns nil+false if one of the values is not a valid int)
func (t Typed) IntsIf(key string) ([]int, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	if n, ok := value.([]int); ok {
		return n, true
	}
	if a, ok := value.([]interface{}); ok {
		l := len(a)
		if l == 0 {
			return nil, false
		}

		n := make([]int, l)
		for i := 0; i < l; i++ {
			switch t := a[i].(type) {
			case int:
				n[i] = t
			case float64:
				n[i] = int(t)
			case string:
				_i, err := strconv.Atoi(t)
				if err != nil {
					return n, false
				}
				n[i] = _i
			default:
				return n, false
			}
		}
		return n, true
	}
	return nil, false
}

// Returns an slice of ints64, or the specified slice
// Some conversion is done to handle the fact that JSON ints
// are represented as floats.
func (t Typed) Ints64(key string) []int64 {
	return t.Ints64Or(key, nil)
}

// Returns an slice of ints, or the specified slice
// if the key doesn't exist or isn't a valid []int.
// Some conversion is done to handle the fact that JSON ints
// are represented as floats.
func (t Typed) Ints64Or(key string, d []int64) []int64 {
	n, ok := t.Ints64If(key)
	if ok {
		return n
	}
	return d
}

// Returns a boolean slice + true if valid
// Returns nil + false otherwise
// (returns nil+false if one of the values is not a valid boolean)
func (t Typed) Ints64If(key string) ([]int64, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	if n, ok := value.([]int64); ok {
		return n, true
	}
	if a, ok := value.([]interface{}); ok {
		l := len(a)
		if l == 0 {
			return nil, false
		}

		n := make([]int64, l)
		for i := 0; i < l; i++ {
			switch t := a[i].(type) {
			case int64:
				n[i] = t
			case float64:
				n[i] = int64(t)
			case int:
				n[i] = int64(t)
			case string:
				_i, err := strconv.ParseInt(t, 10, 10)
				if err != nil {
					return n, false
				}
				n[i] = _i
			case json.Number:
				_i, err := t.Int64()
				if err != nil {
					return n, false
				}
				n[i] = _i
			default:
				return n, false
			}
		}
		return n, true
	}
	return nil, false
}

// Returns an slice of floats, or a nil slice
func (t Typed) Floats(key string) []float64 {
	return t.FloatsOr(key, nil)
}

// Returns an slice of floats, or the specified slice
// if the key doesn't exist or isn't a valid []float64
func (t Typed) FloatsOr(key string, d []float64) []float64 {
	n, ok := t.FloatsIf(key)
	if ok {
		return n
	}
	return d
}

// Returns a float slice + true if valid
// Returns nil + false otherwise
// (returns nil+false if one of the values is not a valid float)
func (t Typed) FloatsIf(key string) ([]float64, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	if n, ok := value.([]float64); ok {
		return n, true
	}
	if a, ok := value.([]interface{}); ok {
		l := len(a)
		n := make([]float64, l)
		for i := 0; i < l; i++ {
			switch t := a[i].(type) {
			case float64:
				n[i] = t
			case string:
				f, err := strconv.ParseFloat(t, 10)
				if err != nil {
					return n, false
				}
				n[i] = f
			case json.Number:
				f, err := t.Float64()
				if err != nil {
					return n, false
				}
				n[i] = f
			default:
				return n, false
			}
		}
		return n, true
	}
	return nil, false
}

// Returns an slice of strings, or a nil slice
func (t Typed) Strings(key string) []string {
	return t.StringsOr(key, nil)
}

// Returns an slice of strings, or the specified slice
// if the key doesn't exist or isn't a valid []string
func (t Typed) StringsOr(key string, d []string) []string {
	n, ok := t.StringsIf(key)
	if ok {
		return n
	}
	return d
}

// Returns a string slice + true if valid
// Returns nil + false otherwise
// (returns nil+false if one of the values is not a valid string)
func (t Typed) StringsIf(key string) ([]string, bool) {
	value, exists := t[key]
	if exists == false {
		return nil, false
	}
	if n, ok := value.([]string); ok {
		return n, true
	}
	if a, ok := value.([]interface{}); ok {
		l := len(a)
		n := make([]string, l)
		var ok bool
		for i := 0; i < l; i++ {
			if n[i], ok = a[i].(string); ok == false {
				return n, false
			}
		}
		return n, true
	}
	return nil, false
}

// Returns an slice of Typed helpers, or a nil slice
func (t Typed) Objects(key string) []Typed {
	value, _ := t.ObjectsIf(key)
	return value
}

// Returns a slice of Typed helpers and true if exists, otherwise; nil and false.
func (t Typed) ObjectsIf(key string) ([]Typed, bool) {
	value, exists := t[key]
	if exists == true {
		switch t := value.(type) {
		case []interface{}:
			l := len(t)
			n := make([]Typed, l)
			for i := 0; i < l; i++ {
				switch it := t[i].(type) {
				case map[string]interface{}:
					n[i] = Typed(it)
				case Typed:
					n[i] = it
				}
			}
			return n, true
		case []map[string]interface{}:
			l := len(t)
			n := make([]Typed, l)
			for i := 0; i < l; i++ {
				n[i] = Typed(t[i])
			}
			return n, true
		case []Typed:
			return t, true
		}
	}
	return nil, false
}

func (t Typed) ObjectsMust(key string) []Typed {
	value, exists := t.ObjectsIf(key)
	if exists == false {
		panic("expected objects value for " + key)
	}

	return value
}

// Returns an slice of map[string]interfaces, or a nil slice
func (t Typed) Maps(key string) []map[string]interface{} {
	value, exists := t[key]
	if exists == true {
		if a, ok := value.([]interface{}); ok {
			l := len(a)
			n := make([]map[string]interface{}, l)
			for i := 0; i < l; i++ {
				n[i] = a[i].(map[string]interface{})
			}
			return n
		}
	}
	return nil
}

// Returns an map[string]bool
func (t Typed) StringBool(key string) map[string]bool {
	raw, ok := t.getmap(key)
	if ok == false {
		return nil
	}
	m := make(map[string]bool, len(raw))
	for k, value := range raw {
		m[k] = value.(bool)
	}
	return m
}

// Returns an map[string]int
// Some work is done to handle the fact that JSON ints
// are represented as floats.
func (t Typed) StringInt(key string) map[string]int {
	raw, ok := t.getmap(key)
	if ok == false {
		return nil
	}
	m := make(map[string]int, len(raw))
	for k, value := range raw {
		switch t := value.(type) {
		case int:
			m[k] = t
		case float64:
			m[k] = int(t)
		case string:
			i, err := strconv.Atoi(t)
			if err != nil {
				return nil
			}
			m[k] = i
		case json.Number:
			i, err := t.Int64()
			if err != nil {
				return nil
			}
			m[k] = int(i)
		}
	}
	return m
}

// Returns an map[string]float64
func (t Typed) StringFloat(key string) map[string]float64 {
	raw, ok := t.getmap(key)
	if ok == false {
		return nil
	}
	m := make(map[string]float64, len(raw))
	for k, value := range raw {
		switch t := value.(type) {
		case float64:
			m[k] = t
		case string:
			f, err := strconv.ParseFloat(t, 10)
			if err != nil {
				return nil
			}
			m[k] = f
		case json.Number:
			f, err := t.Float64()
			if err != nil {
				return nil
			}
			m[k] = f
		default:
			return nil
		}
	}
	return m
}

// Returns an map[string]string
func (t Typed) StringString(key string) map[string]string {
	raw, ok := t.getmap(key)
	if ok == false {
		return nil
	}
	m := make(map[string]string, len(raw))
	for k, value := range raw {
		m[k] = value.(string)
	}
	return m
}

// Returns an map[string]Typed
func (t Typed) StringObject(key string) map[string]Typed {
	raw, ok := t.getmap(key)
	if ok == false {
		return nil
	}
	m := make(map[string]Typed, len(raw))
	for k, value := range raw {
		m[k] = Typed(value.(map[string]interface{}))
	}
	return m
}

// Marhals the type into a []byte.
// If key isn't valid, KeyNotFound is returned.
// If the typed doesn't represent valid JSON, a relevant
// JSON error is returned
func (t Typed) ToBytes(key string) ([]byte, error) {
	var o interface{}
	if len(key) == 0 {
		o = t
	} else {
		exists := false
		o, exists = t[key]
		if exists == false {
			return nil, KeyNotFound
		}
	}
	if o == nil {
		return nil, nil
	}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (t Typed) MustBytes(key string) []byte {
	data, err := t.ToBytes(key)
	if err != nil {
		panic(err)
	}
	return data
}

func (t Typed) Exists(key string) bool {
	_, exists := t[key]
	return exists
}

func (t Typed) getmap(key string) (raw map[string]interface{}, exists bool) {
	value, exists := t[key]
	if exists == false {
		return
	}
	raw, exists = value.(map[string]interface{})
	return
}
