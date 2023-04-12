package main

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type capability func(payload string) (string, bool, error)

func newCapabilities(read readerFunc) map[string]capability {
	return map[string]capability{
		"respond": func(payload string) (string, bool, error) {
			fmt.Printf("Response: %s\n", payload)
			return "", false, nil
		},
		"error": func(payload string) (string, bool, error) {
			return "", false, errors.New(payload)
		},
		"clarify": func(payload string) (string, bool, error) {
			return read(fmt.Sprintf("%s: ", payload)), true, nil
		},

		//
		"chat-input": func(payload string) (string, bool, error) {
			return read(fmt.Sprintf("%s: ", payload)), true, nil
		},
	}
}
