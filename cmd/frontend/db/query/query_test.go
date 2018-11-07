package query_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/query"
)

func ExampleEval() {
	rq, err := query.Eval(
		query.And(
			query.Or("foo", false),
			query.Not("bar"),
			true),
		func(q query.Q) (*sqlf.Query, error) {
			p, ok := q.(string)
			if !ok {
				return nil, errors.Errorf("unexpected token in query: %q", q)
			}
			return sqlf.Sprintf("name LIKE %s", "%"+p+"%"), nil
		})
	if err != nil {
		log.Fatal("unexpected error", err)
	}

	fmt.Println(rq.Query(sqlf.PostgresBindVar))
	fmt.Println(rq.Args())
	fmt.Println(query.Print(rq))
	// Output: ((name LIKE $1 OR FALSE) AND NOT(name LIKE $2) AND TRUE)
	// [%foo% %bar%]
	// ((name LIKE "%foo%" OR FALSE) AND NOT(name LIKE "%bar%") AND TRUE)
}

func TestEval_cornercase(t *testing.T) {
	cases := []struct {
		q    query.Q
		want string
	}{{
		q:    query.And(),
		want: "TRUE",
	}, {
		q:    query.And("test"),
		want: `("test")`,
	}, {
		q:    query.Or(),
		want: "FALSE",
	}, {
		q:    query.Or("test"),
		want: `("test")`,
	}}
	for _, c := range cases {
		got := query.Print(c.q)
		if got != c.want {
			t.Errorf("got %s, want %s", got, c.want)
		}
	}
}

func TestEval_error(t *testing.T) {
	_, err := query.Eval(
		query.And(
			query.Or(
				query.Not("bar"),
				false),
			query.Not(true)),
		func(q query.Q) (*sqlf.Query, error) {
			return nil, errors.New("42")
		})
	if err == nil {
		t.Fatal("expected error")
	}
}
