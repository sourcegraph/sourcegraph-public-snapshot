package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestParMap2(t *testing.T) {
	vs := []string{"baz", "f", "to", "cadoodle"}
	r, _ := ParMap(func(s string) (int, error) {
		return len(s), nil
	}, vs)
	results := r.([]int)

	expected := []int{3, 1, 2, 8}
	if !reflect.DeepEqual(expected, results) {
		t.Errorf("expected %v, got %v", expected, results)
	}
}

func BenchmarkParMap(b *testing.B) {
	l := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		l[i] = i
	}
	ParMap(func(i int) (int, error) {
		return i + 2, nil
	}, l)
}
