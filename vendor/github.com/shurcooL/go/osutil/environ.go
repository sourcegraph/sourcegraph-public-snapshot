// Package osutil offers a utility for manipulating a set of environment variables.
package osutil

import "strings"

// Environ is a slice of strings representing the environment, in the form "key=value".
type Environ []string

// Set environment variable key to value.
func (e *Environ) Set(key, value string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = key + "=" + value
			return
		}
	}
	// If we get here, it's because the key isn't already present, so add a new one.
	*e = append(*e, key+"="+value)
}

// Unset environment variable key.
func (e *Environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			return
		}
	}
}
