package webhooks

import "reflect"

func nullable[T any](value T) *T {
	if reflect.ValueOf(value).IsZero() {
		return nil
	}
	return &value
}

func nullableMap[T, U any](value T, mapper func(T) U) *U {
	if reflect.ValueOf(value).IsZero() {
		return nil
	}
	mapped := mapper(value)
	return &mapped
}
