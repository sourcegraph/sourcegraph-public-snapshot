//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package azopenai

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// EventReader streams events dynamically from an OpenAI endpoint.
type EventReader[T any] struct {
	reader  io.ReadCloser // Required for Closing
	scanner *bufio.Scanner
}

func newEventReader[T any](r io.ReadCloser) *EventReader[T] {
	return &EventReader[T]{reader: r, scanner: bufio.NewScanner(r)}
}

// Read reads the next event from the stream.
// Returns io.EOF when there are no further events.
func (er *EventReader[T]) Read() (T, error) {
	// https://html.spec.whatwg.org/multipage/server-sent-events.html
	for er.scanner.Scan() { // Scan while no error
		line := er.scanner.Text() // Get the line & interpret the event stream:

		if line == "" || line[0] == ':' { // If the line is blank or is a comment, skip it
			continue
		}

		if strings.Contains(line, ":") { // If the line contains a U+003A COLON character (:), process the field
			tokens := strings.SplitN(line, ":", 2)
			tokens[0], tokens[1] = strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
			var data T
			switch tokens[0] {
			case "data": // return the deserialized JSON object
				if tokens[1] == "[DONE]" { // If data is [DONE], end of stream was reached
					return data, io.EOF
				}
				err := json.Unmarshal([]byte(tokens[1]), &data)
				return data, err
			default: // Any other event type is an unexpected
				return data, errors.New("unexpected event type: " + tokens[0])
			}
			// Unreachable
		}
	}

	scannerErr := er.scanner.Err()

	if scannerErr == nil {
		return *new(T), errors.New("incomplete stream")
	}

	return *new(T), scannerErr
}

// Close closes the EventReader and any applicable inner stream state.
func (er *EventReader[T]) Close() error {
	return er.reader.Close()
}
