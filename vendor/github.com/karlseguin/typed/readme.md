## Typed

*Making `map[string]interface{}` slightly less painful*

It isn't always desirable or even possible to decode JSON into a structure. That, unfortunately, leaves us with writing ugly code around `map[string]interface{}` and `[]interace{}`.

This library hopes to make that slightly less painful.

## Install

    go get github.com/karlseguin/typed

## Note

While the library could be used with any `map[string]interface{}`, it *is* tailored to be used with `encoding/json`. Specifically, it won't ever try to convert data, except for integers (which `encoding/json` treats as floats).

## Usage:

A typed wrapper around a `map[string]interface{}` can be created in one of three ways:

```go
// directly from a map[string]interace{}
typed := typed.New(a_map)

// from a json []byte
typed, err := typed.Json(data)

// from a file containing JSON
typed, err := typed.JsonFile(path)
```

Once we have a typed wrapper, we can use various functions to navigate the structure:

- `Bool(key string) bool`
- `Int(key string) int`
- `Float(key string) float64`
- `String(key string) string`
- `Object(key string) typed.Type`
- `Interface(key string) interface{}`

- `BoolOr(key string, defaultValue bool) bool`
- `IntOr(key string, defaultValue int) int`
- `FloatOr(key string, defaultValue float64) float64`
- `StringOr(key string, defaultValue string) string`
- `ObjectOr(key string, defaultValue map[string]interface{}) typed.Type`
- `InterfaceOr(key string, defaultValue interface{}) interface{}`

- `BoolMust(key string) bool`
- `IntMust(key string) int`
- `FloatMust(key string) float`
- `StringMust(key string) string`
- `ObjectMust(key string) typed.Type`
- `InterfaceMust(key string) interface{}``

- `BoolIf(key string) (bool, bool)`
- `IntIf(key string) (int, bool)`
- `FloatIf(key string) (float, bool)`
- `StringIf(key string) (string, bool)`
- `ObjectIf(key string) (typed.Type, bool)`
- `InterfaceIf(key string) (interface{}, bool)`

We can also extract arrays via:

- `Bools(key string) []bool`
- `Ints(key string) []int`
- `Ints64(key string) []int64`
- `Floats(key string) []float64`
- `Strings(key string) []string`
- `StringsOr(key string, defaultValue []string) []string`
- `BoolsOr(key string, defaultValue []bool) []bool`
- `IntsOr(key string, defaultValue []int) []int`
- `Ints64Or(key string, defaultValue []int64) []int64`
- `FloatsOr(key string, defaultValue []float64) []float64`

We can extract nested objects, other as another typed wrapper, or as a raw `map[string]interface{}`:

- `Object(key string) Typed`
- `ObjectOr(key string, map[string]interface) Typed`
- `ObjectIf(key string) (Typed, bool)`
- `Objects(key string) []Typed`

- `Map(key string) map[string]interface{}`
- `MaoOr(key string, map[string]interface) map[string]interface{}`
- `MapIf(key string) (map[string]interface{}, bool)`
- `Maps(key string) []map[string]interface{}`

We can extract key value pairs:

- `StringBool(key string) map[string]bool`
- `StringInt(key string) map[string]int`
- `StringFloat(key string) map[string]float64`
- `StringString(key string) map[string]string`
- `StringObject(key string) map[string]Typed`

## Example

```go
package main

import (
  "fmt"
  "typed"
)

func main() {
  json := `
{
  "log": true,
  "name": "leto",
  "percentiles": [75, 85, 95],
  "server": {
    "port": 9001,
    "host": "localhost"
  },
  "cache": {
    "users": {"ttl": 5}
  },
  "blocked": {
    "10.10.1.1": true
  }
}`
  typed, err := typed.JsonString(json)
  if err != nil {
    fmt.Println(err)
  }

  fmt.Println(typed.Bool("log"))
  fmt.Println(typed.String("name"))
  fmt.Println(typed.Ints("percentiles"))
  fmt.Println(typed.FloatOr("load", 0.5))

  server := typed.Object("server")
  fmt.Println(server.Int("port"))
  fmt.Println(server.String("host"))

  fmt.Println(typed.Map("server"))

  fmt.Println(typed.StringObject("cache")["users"].Int("ttl"))
  fmt.Println(typed.StringBool("blocked")["10.10.1.1"])
}
```

# Misc

## To Bytes
`ToBytes(key string) ([]byte, error)` can be used to get the JSON data, as a []byte, from the Type. `KeyNotFound` will be returned if the key isn't valid.

Alternatively, `MustBytes(key string) []byte` can be used. It will panic on error.

## Root Array

Support for JSON document with arrays at their root is supported. Use the `JsonArray(data []byte)`, `JsonStringArray(data string)` and `JsonFileArray(path string)` to receive an `[]Typed`

If the array has primitive values, the syntax is a little awkward. The key for the field is "0":

```go
json := `[1, {"name": "leto"}, "spice"]`

typed, err := typed, err := typed.JsonStringArray(json)
if err != nil {
  panic(err)
}
println(typed[0].Int("0"))
println(typed[1].String("name"))
println(typed[2].String("0"))
```

## JsonWriter

The [JsonWriter](https://github.com/karlseguin/jsonwriter) library provides the opposite functionality: a lightweight approaching to writing JSON data.
