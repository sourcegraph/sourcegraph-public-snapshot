sqlf [![Build Status](https://travis-ci.org/keegancsmith/sqlf.svg?branch=master)](https://travis-ci.org/) [![GoDoc](https://godoc.org/github.com/keegancsmith/sqlf?status.svg)](https://godoc.org/github.com/keegancsmith/sqlf)
======

Generate parameterized SQL statements in Go, sprintf Style.

```go
q := sqlf.Sprintf("SELECT * FROM users WHERE country = %s AND age > %d", "US", 27);
rows, err := db.Query(q.Query(sqlf.SimpleBindVar), q.Args()...) // db is a database/sql.DB
```

`sqlf.Sprintf` does not return a string. It returns `*sqlf.Query` which has
methods for a parameterized SQL query and arguments. You then pass that to
`db.Query`, `db.Exec`, etc. This is not like using `fmt.Sprintf`, which could
expose you to malformed SQL or SQL injection attacks.

`sqlf.Query` can be passed as an argument to `sqlf.Sprintf`. It will "flatten"
the query string, while preserving the correct variable binding. This allows
you to easily compose and build SQL queries. See the below examples to find
out more.

```go
// This is an example which shows off embedding SQL, which simplifies building
// complicated SQL queries
name := "John"
age, offset := 27, 100
where := sqlf.Sprintf("name=%s AND age=%d", name, age)
limit := sqlf.Sprintf("%d OFFSET %d", 10, offset)
q := sqlf.Sprintf("SELECT name FROM users WHERE %s LIMIT %s", where, limit)
fmt.Println(q.Query(sqlf.PostgresBindVar))
fmt.Println(q.Args())
// Output: SELECT name FROM users WHERE name=$1 AND age=$2 LIMIT $3 OFFSET $4
// [John 27 10 100]
```

Another common task is joining conditionals with `AND` or `OR`. sqlf
simplifies this task with `sqlf.Join`:

```go
// Our inputs
min_quantity := 100
name_filters := []string{"apple", "orange", "coffee"}

var conds []*sqlf.Query
for _, filter := range name_filters {
    conds = append(conds, sqlf.Sprintf("name LIKE %s", "%"+filter+"%"))
}
sub_query := sqlf.Sprintf("SELECT product_id FROM order_item WHERE quantity > %d", min_quantity)
q := sqlf.Sprintf("SELECT name FROM product WHERE id IN (%s) AND (%s)", sub_query, sqlf.Join(conds, "OR"))

fmt.Println(q.Query(sqlf.PostgresBindVar))
fmt.Println(q.Args())
// Output: SELECT name FROM product WHERE id IN (SELECT product_id FROM order_item WHERE quantity > $1) AND (name LIKE $2 OR name LIKE $3 OR name LIKE $4)
// [100 %apple% %orange% %coffee%]
```

See https://godoc.org/github.com/keegancsmith/sqlf for more information.
