package yaml

import (
	"errors"
	"reflect"
	"strings"

	"github.com/flynn/go-shlex"
	"gopkg.in/yaml.v2"
)

// A Command is represented in YAML as either a string (in which case
// it is split on whitespace using shlex.Split to be converted into a
// Go string slice) or as an array of strings (in which case it is
// converted directly to a Go string slice).
type Command []string

func (c *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try unmarshaling the string form first.
	var s string
	if err := unmarshal(&s); err == nil {
		parts, err := shlex.Split(s)
		if err != nil {
			return err
		}
		*c = parts
		return nil
	}

	// Otherwise, unmarshal the array form.
	return unmarshal((*[]string)(c))
}

// A MapEqualSlice is represented in YAML as an ordered map. In Go it
// is represented as a "key=val" string for each map entry.
//
// The ordering is retained so that the MapEqualSlice can be
// remarshaled after modifications without other unintended changes to
// the YAML representation.
type MapEqualSlice []string

func (s MapEqualSlice) MarshalYAML() (interface{}, error) {
	m := make(yaml.MapSlice, len(s))
	for i, v := range s {
		kv := strings.SplitN(v, "=", 2)
		m[i].Key = kv[0]
		if len(kv) == 2 {
			m[i].Value = kv[1]
		}
	}
	return m, nil
}

func (s *MapEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Allow a YAML representation of MapEqualSlice as an array of
	// "key=val" strings.
	if err := unmarshal((*[]string)(s)); err == nil {
		return nil
	}

	var m yaml.MapSlice
	if err := unmarshal(&m); err != nil {
		return err
	}

	*s = make(MapEqualSlice, len(m))
	for i, e := range m {
		k, ok := e.Key.(string)
		if !ok {
			return errors.New("non-string key")
		}
		v, ok := e.Value.(string)
		if !ok {
			return errors.New("non-string value")
		}

		(*s)[i] = k + "=" + v
	}

	return nil
}

// Stringorslice represents a string or an array of strings.
// TODO use docker/docker/pkg/stringutils.StrSlice once 1.9.x is released.
type Stringorslice []string

// MarshalYAML implements the Marshaller interface.
func (s Stringorslice) MarshalYAML() (interface{}, error) {
	return s, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []string
	err := unmarshal(&sliceType)
	if err == nil {
		*s = sliceType
		return nil
	}

	var stringType string
	err = unmarshal(&stringType)
	if err == nil {
		sliceType = make([]string, 0, 1)
		*s = append(sliceType, string(stringType))
		return nil
	}
	return err
}

// Builds is an ordered map of builds.
type Builds []BuildItem

func (b Builds) MarshalYAML() (interface{}, error) {
	// Back-compat: Allow a single-build section.
	if len(b) == 1 && b[0].Key == "" {
		return b[0].Build, nil
	}

	// Usual case: multiple build sections.
	return marshalOrderedMapItemsYAML(b)
}

func (b *Builds) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Back-compat: Allow a single-build section.
	var build Build
	if err := unmarshal(&build); err != nil {
		return err
	}
	if build.Image != "" {
		*b = Builds{{Build: build}}
		return nil
	}

	// Usual case: multiple build sections.
	return unmarshalOrderedMapItemsYAML(b, unmarshal)
}

// A BuildItem is a key-value pair, used in an Builds ordered map.
type BuildItem struct {
	Key string
	Build
}

// Containers is an ordered map of containers.
type Containers []ContainerItem

func (c Containers) MarshalYAML() (interface{}, error) {
	return marshalOrderedMapItemsYAML(c)
}

func (c *Containers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshalOrderedMapItemsYAML(c, unmarshal); err != nil {
		return err
	}

	// Set defaults.
	for i, ci := range *c {
		if ci.Container.Image == "" {
			(*c)[i].Container.Image = ci.Key
		}
	}

	return nil
}

// A ContainerItem is a key-value pair, used in an Containers ordered
// map.
type ContainerItem struct {
	Key string
	Container
}

// Plugins is an ordered map of plugins.
type Plugins []PluginItem

func (p Plugins) MarshalYAML() (interface{}, error) {
	return marshalOrderedMapItemsYAML(p)
}

func (p *Plugins) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshalOrderedMapItemsYAML(p, unmarshal); err != nil {
		return err
	}

	// Set defaults.
	for i, pi := range *p {
		if pi.Plugin.Container.Image == "" {
			(*p)[i].Plugin.Container.Image = pi.Key
		}
	}

	return nil
}

// A PluginItem is a key-value pair, used in an Plugins ordered map.
type PluginItem struct {
	Key string
	Plugin
}

// marshalOrderedMapItemsYAML takes []XyzItem and returns an
// equivalent yaml.MapSlice using reflection.
func marshalOrderedMapItemsYAML(itemSlice interface{}) (interface{}, error) {
	v := reflect.ValueOf(itemSlice)
	if v.Kind() != reflect.Slice {
		panic("itemSlice must be a slice, not " + v.Type().String())
	}

	yamlMap := make(yaml.MapSlice, v.Len())
	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)

		keyV := e.FieldByName("Key")
		if keyV.Kind() != reflect.String {
			panic("key must be a string, not " + keyV.Type().String())
		}

		yamlMap[i] = yaml.MapItem{
			Key:   keyV.String(),
			Value: e.Field(1).Interface(),
		}
	}

	return yamlMap, nil
}

func unmarshalOrderedMapItemsYAML(itemSlice interface{}, unmarshal func(interface{}) error) error {
	var m yaml.MapSlice
	if err := unmarshal(&m); err != nil {
		return err
	}
	return convertOrderedMapItemsFromYAML(itemSlice, m)
}

// convertOrderedMapItemsFromYAML converts a YAML ordered map in src
// into itemSlice, which must be of type *[]XyzItem.
func convertOrderedMapItemsFromYAML(itemSlice interface{}, src yaml.MapSlice) error {
	v := reflect.ValueOf(itemSlice)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		panic("itemSlice must be a pointer to slice, not " + v.Type().String())
	}

	v.Elem().Set(reflect.MakeSlice(v.Elem().Type(), len(src), len(src)))
	for i, e := range src {
		ei := v.Elem().Index(i)

		// Create the XyzItem.
		item := reflect.New(ei.Type()).Elem()
		item.FieldByName("Key").SetString(e.Key.(string))

		// Re-unmarshal the value into its actual type (in the XyzItem
		// value field).
		yamlBytes, err := yaml.Marshal(e.Value)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlBytes, item.Field(1).Addr().Interface()); err != nil {
			return err
		}

		ei.Set(item)
	}

	return nil
}
