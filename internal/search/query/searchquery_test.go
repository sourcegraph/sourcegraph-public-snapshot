package query

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search/query/types"
)

func TestQuery_IsCaseSensitive(t *testing.T) {
	conf := types.Config{
		FieldTypes: map[string]types.FieldType{
			FieldCase: {Literal: types.BoolType, Quoted: types.BoolType, Singular: true},
		},
	}

	t.Run("yes", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "case:yes")
		if err != nil {
			t.Fatal(err)
		}
		if !query.IsCaseSensitive() {
			t.Error("IsCaseSensitive() == false, want true")
		}
	})

	t.Run("no (explicit)", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "case:no")
		if err != nil {
			t.Fatal(err)
		}
		if query.IsCaseSensitive() {
			t.Error("IsCaseSensitive() == true, want false")
		}
	})

	t.Run("no (default)", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "")
		if err != nil {
			t.Fatal(err)
		}
		if query.IsCaseSensitive() {
			t.Error("IsCaseSensitive() == true, want false")
		}
	})
}

func TestQuery_RegexpPatterns(t *testing.T) {
	conf := types.Config{
		FieldTypes: map[string]types.FieldType{
			"r": regexpNegatableFieldType,
			"s": {Literal: types.RegexpType, Quoted: types.StringType},
		},
	}

	t.Run("for regexp field", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "r:a r:b -r:c")
		if err != nil {
			t.Fatal(err)
		}
		v, nv := query.RegexpPatterns("r")
		if want := []string{"a", "b"}; !reflect.DeepEqual(v, want) {
			t.Errorf("got values %q, want %q", v, want)
		}
		if want := []string{"c"}; !reflect.DeepEqual(nv, want) {
			t.Errorf("got negated values %q, want %q", nv, want)
		}
	})

	t.Run("for unrecognized field", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "")
		if err != nil {
			t.Fatal(err)
		}
		checkPanic(t, "no such field: z", func() {
			query.RegexpPatterns("z")
		})
	})

	t.Run("for non-regexp field", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "s:a")
		if err != nil {
			t.Fatal(err)
		}
		checkPanic(t, "field is not always regexp-typed: s", func() {
			query.RegexpPatterns("s")
		})
	})
}

func TestQuery_StringValues(t *testing.T) {
	conf := types.Config{
		FieldTypes: map[string]types.FieldType{
			"s": {Literal: types.StringType, Quoted: types.StringType, Negatable: true},
			"r": {Literal: types.RegexpType, Quoted: types.StringType},
		},
	}

	t.Run("for string field", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "s:a s:b -s:c")
		if err != nil {
			t.Fatal(err)
		}
		v, nv := query.StringValues("s")
		if want := []string{"a", "b"}; !reflect.DeepEqual(v, want) {
			t.Errorf("got values %q, want %q", v, want)
		}
		if want := []string{"c"}; !reflect.DeepEqual(nv, want) {
			t.Errorf("got negated values %q, want %q", nv, want)
		}
	})

	t.Run("for unrecognized field", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "")
		if err != nil {
			t.Fatal(err)
		}
		checkPanic(t, "no such field: z", func() {
			query.StringValues("z")
		})
	})

	t.Run("for non-string field", func(t *testing.T) {
		query, err := parseAndCheck(&conf, "r:a")
		if err != nil {
			t.Fatal(err)
		}
		checkPanic(t, "field is not always string-typed: r", func() {
			query.StringValues("r")
		})
	})
}

func TestQuery_CaseInsensitiveFields(t *testing.T) {
	query, err := ParseAndCheck("repoHasFile:foo")
	if err != nil {
		t.Fatal(err)
	}

	values, _ := query.RegexpPatterns(FieldRepoHasFile)
	if len(values) != 1 || values[0] != "foo" {
		t.Errorf("unexpected values: want {\"foo\"}, got %v", values)
	}

	if got, want := query.Query.String(), `repohasfile~"foo"`; got != want {
		t.Errorf("unexpected parsed query:\ngot:  %s\nwant: %s", got, want)
	}
}

func checkPanic(t *testing.T, msg string, f func()) {
	t.Helper()
	defer func() {
		if e := recover(); e == nil {
			t.Error("no panic")
		} else if e.(string) != msg {
			t.Errorf("got panic %q, want %q", e, msg)
		}
	}()
	f()
}
