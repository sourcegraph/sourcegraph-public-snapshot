# unidecode

[![Go Reference](https://pkg.go.dev/badge/github.com/gosimple/unidecode.svg)](https://pkg.go.dev/github.com/gosimple/unidecode)
[![Tests](https://github.com/gosimple/unidecode/actions/workflows/tests.yml/badge.svg)](https://github.com/gosimple/unidecode/actions/workflows/tests.yml)

Unicode transliterator in Golang - Replaces non-ASCII characters with their
ASCII approximations.

Fork of https://github.com/rainycape/unidecode

## Example

```go
package main

import (
	"fmt"

	"github.com/gosimple/unidecode"
)

func main() {
	decoded := unidecode.Unidecode("Łódź")
	fmt.Println(decoded)
	// Output: Lodz
}
```

### Requests or bugs?

<https://github.com/gosimple/unidecode/issues>

## Installation

```shell
go get -u github.com/gosimple/unidecode
```

## Benchmark

```shell
go test -run=NONE -bench=. -benchmem -count=6 ./... > old.txt
# make changes
go test -run=NONE -bench=. -benchmem -count=6 ./... > new.txt

go install golang.org/x/perf/cmd/benchstat@latest

benchstat old.txt new.txt
```

## Add new characters

1. Edit `table.txt` file.
2. Rebuild `table.go` file:

   ```go
   go run ./make_table.go
   ```
