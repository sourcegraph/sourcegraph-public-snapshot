// Package env provides types to handle step environments in batch specs.
package env

import (
	"encoding/json"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Environment represents an environment used for a batch step, which may
// require values to be resolved from the outer environment the executor is
// running within.
type Environment struct {
	vars []variable
}

// MarshalJSON marshals the environment.
func (e Environment) MarshalJSON() ([]byte, error) {
	if e.vars == nil {
		return []byte(`{}`), nil
	}

	// For compatibility with older versions of Sourcegraph, if all environment
	// variables have static values defined, we'll encode to the object variant.
	if e.IsStatic() {
		vars := make(map[string]string, len(e.vars))
		for _, v := range e.vars {
			vars[v.name] = *v.value
		}

		return json.Marshal(vars)
	}

	// Otherwise, we have to return the array variant.
	return json.Marshal(e.vars)
}

// UnmarshalJSON unmarshals an environment from one of the two supported JSON
// forms: an array, or a string→string object.
func (e *Environment) UnmarshalJSON(data []byte) error {
	// data is either an array or object. (Or invalid.) Let's start by trying to
	// unmarshal it as an array.
	if err := json.Unmarshal(data, &e.vars); err == nil {
		return nil
	}

	// It's an object, then. We need to put it into a map, then convert it into
	// an array of variables.
	kv := make(map[string]string)
	if err := json.Unmarshal(data, &kv); err != nil {
		return err
	}

	e.vars = make([]variable, len(kv))
	i := 0
	for k, v := range kv {
		copy := v
		e.vars[i].name = k
		e.vars[i].value = &copy
		i++
	}

	return nil
}

// UnmarshalYAML unmarshals an environment from one of the two supported YAML
// forms: an array, or a string→string object.
func (e *Environment) UnmarshalYAML(unmarshal func(any) error) error {
	// data is either an array or object. (Or invalid.) Let's start by trying to
	// unmarshal it as an array.
	if err := unmarshal(&e.vars); err == nil {
		return nil
	}

	// It's an object, then. As above, we need to convert this via a map.
	kv := make(map[string]string)
	if err := unmarshal(&kv); err != nil {
		return err
	}

	e.vars = make([]variable, len(kv))
	i := 0
	for k, v := range kv {
		copy := v
		e.vars[i].name = k
		e.vars[i].value = &copy
		i++
	}

	return nil
}

// IsStatic returns true if the environment doesn't depend on any outer
// environment variables.
//
// Put another way: if this function returns true, then Resolve() will always
// return the same map for the environment.
func (e Environment) IsStatic() bool {
	for _, v := range e.vars {
		if v.value == nil {
			return false
		}
	}
	return true
}

// OuterVars returns the list of environment variables that depend on any
// environment variable defined in the global env.
func (e Environment) OuterVars() []string {
	outer := []string{}
	for _, v := range e.vars {
		if v.value == nil {
			outer = append(outer, v.name)
		}
	}
	return outer
}

// Resolve resolves the environment, using values from the given outer
// environment to fill in environment values as needed. If an environment
// variable doesn't exist in the outer environment, then an empty string will be
// used as the value.
//
// outer must be an array of strings in the form `KEY=VALUE`. Generally
// speaking, this will be the return value from os.Environ().
func (e Environment) Resolve(outer []string) (map[string]string, error) {
	// Convert the given outer environment into a map.
	omap := make(map[string]string, len(outer))
	for _, v := range outer {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("unable to parse environment variable %q", v)
		}
		omap[kv[0]] = kv[1]
	}

	// Now we can iterate over our own environment and fill in the missing
	// values.
	resolved := make(map[string]string, len(e.vars))
	for _, v := range e.vars {
		if v.value == nil {
			// We don't bother checking if v.name exists in omap here because
			// the default behaviour is what we want anyway: we'll get an empty
			// string (since that's the zero value for a string), and that is
			// the desired outcome if the environment variable isn't set.
			resolved[v.name] = omap[v.name]
		} else {
			resolved[v.name] = *v.value
		}
	}

	return resolved, nil
}

// Equal verifies if two environments are equal.
func (e Environment) Equal(other Environment) bool {
	return cmp.Equal(e.mapify(), other.mapify())
}

func (e Environment) mapify() map[string]*string {
	m := make(map[string]*string, len(e.vars))
	for _, v := range e.vars {
		m[v.name] = v.value
	}

	return m
}
