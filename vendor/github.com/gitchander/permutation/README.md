# permutation
Simple permutation package for golang

## Install
```bash
go get github.com/gitchander/permutation
```

## Usage

#### permutations of int slice:
```go
package main

import (
	"fmt"

	prmt "github.com/gitchander/permutation"
)

func main() {
	a := []int{1, 2, 3}
	p := prmt.New(prmt.IntSlice(a))
	for p.Next() {
		fmt.Println(a)
	}
}
```
result:
```
[1 2 3]
[2 1 3]
[3 1 2]
[1 3 2]
[2 3 1]
[3 2 1]
```

#### permutations of string slice:
```go
package main

import (
	"fmt"

	prmt "github.com/gitchander/permutation"
)

func main() {
	a := []string{"alpha", "beta", "gamma"}
	p := prmt.New(prmt.StringSlice(a))
	for p.Next() {
		fmt.Println(a)
	}
}
```
result:
```
[alpha beta gamma]
[beta alpha gamma]
[gamma alpha beta]
[alpha gamma beta]
[beta gamma alpha]
[gamma beta alpha]
```

#### permutation use of AnySlice:
```go
a := []interface{}{-1, "control", 9.3}

data, err := prmt.NewAnySlice(a)
if err != nil {
	log.Fatal(err)
}

p := prmt.New(data)
for p.Next() {
	fmt.Println(a)
}
```
or use MustAnySlice (panic if error):
```go
a := []int{1, 2}
p := prmt.New(prmt.MustAnySlice(a))
for p.Next() {
	fmt.Println(a)
}
```

#### usage permutation.Interface
```go
package main

import (
	"fmt"

	prmt "github.com/gitchander/permutation"
)

type Person struct {
	Name string
	Age  int
}

type PersonSlice []Person

func (ps PersonSlice) Len() int      { return len(ps) }
func (ps PersonSlice) Swap(i, j int) { ps[i], ps[j] = ps[j], ps[i] }

func main() {
	a := []Person{
		{Name: "one", Age: 1},
		{Name: "two", Age: 2},
		{Name: "three", Age: 3},
	}
	p := prmt.New(PersonSlice(a))
	for p.Next() {
		fmt.Println(a)
	}
}
```
result:
```
[{one 1} {two 2} {three 3}]
[{two 2} {one 1} {three 3}]
[{three 3} {one 1} {two 2}]
[{one 1} {three 3} {two 2}]
[{two 2} {three 3} {one 1}]
[{three 3} {two 2} {one 1}]
```
