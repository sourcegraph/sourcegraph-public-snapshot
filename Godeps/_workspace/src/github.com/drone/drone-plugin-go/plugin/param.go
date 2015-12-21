package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var Stdin *ParamSet

func init() {
	// defaults to stdin
	Stdin = NewParamSet(os.Stdin)

	// check for params after the double dash
	// in the command string
	for i, argv := range os.Args {
		if argv == "--" {
			arg := os.Args[i+1]
			buf := bytes.NewBufferString(arg)
			Stdin = NewParamSet(buf)
			break
		}
	}
}

// this init function is deprecated, but I'm keeping it
// around just in case it proves useful in the future.
func deprecated_init() {
	// if piping from stdin we can just exit
	// and use the default Stdin value
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return
	}

	// check for params after the double dash
	// in the command string
	for i, argv := range os.Args {
		if argv == "--" {
			arg := os.Args[i+1]
			buf := bytes.NewBufferString(arg)
			Stdin = NewParamSet(buf)
			return
		}
	}

	// else use the first variable in the list
	if len(os.Args) > 1 {
		buf := bytes.NewBufferString(os.Args[1])
		Stdin = NewParamSet(buf)
	}
}

type ParamSet struct {
	reader io.Reader
	params map[string]interface{}
}

func NewParamSet(reader io.Reader) *ParamSet {
	var p = new(ParamSet)
	p.reader = reader
	p.params = map[string]interface{}{}
	return p
}

// Param defines a parameter with the specified name.
func (p ParamSet) Param(name string, value interface{}) {
	p.params[name] = value
}

// Parse parses parameter definitions from the map.
func (p ParamSet) Parse() error {
	raw := map[string]json.RawMessage{}
	err := json.NewDecoder(p.reader).Decode(&raw)
	if err != nil {
		return err
	}

	for key, val := range p.params {
		data, ok := raw[key]
		if !ok {
			continue
		}
		err := json.Unmarshal(data, val)
		if err != nil {
			return fmt.Errorf("Unable to unarmshal %s. %s", key, err)
		}
	}

	return nil
}

// Unmarshal parses the JSON payload from the command
// arguments and unmarshal into a value pointed to by v.
func (p ParamSet) Unmarshal(v interface{}) error {
	return json.NewDecoder(p.reader).Decode(v)
}

// Param defines a parameter with the specified name.
func Param(name string, value interface{}) {
	Stdin.Param(name, value)
}

// Parse parses parameter definitions from the map.
func Parse() error {
	return Stdin.Parse()
}

// Unmarshal parses the JSON payload from the command
// arguments and unmarshal into a value pointed to by v.
func Unmarshal(v interface{}) error {
	return Stdin.Unmarshal(v)
}

// Unmarshal parses the JSON payload from the command
// arguments and unmarshal into a value pointed to by v.
func MustUnmarshal(v interface{}) error {
	return Stdin.Unmarshal(v)
}

// MustParse parses parameter definitions from the map
// and panics if there is a parsing error.
func MustParse() {
	err := Parse()
	if err != nil {
		panic(err)
	}
}
