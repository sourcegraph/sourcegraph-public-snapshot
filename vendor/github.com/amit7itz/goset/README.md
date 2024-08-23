# GoSet

🏆 **`goset` is a generic Go implementation of the Set data structure**

Inspired by Python's set, this project uses Go 1.18+ generics to implement all the methods you ever wanted for a Set of items!


## 📌 Install

```sh
go get github.com/amit7itz/goset@v1
```

## 🤓 Usage

**Import:**
```go
import "github.com/amit7itz/goset"
```

**Initialize:**
```go
// create an empty Set of integers
emptySet := goset.NewSet[int]()

// create a Set of strings with items 
mySet := goset.NewSet[string]("c", "d")

// create a Set of strings from a slice
lettersSet := goset.FromSlice([]string{"a", "b", "c", "d"})
```

**Use:**
```go
mySet.Add("c", "d", "e", "f", "g")
mySet.Discard("g")
println(mySet.String())
// Set[string]{"c", "d", "e", "f"}

intersectionSet := mySet.Intersection(lettersSet)
// Set[string]{"c", "d"}

unionSet := mySet.Union(lettersSet)
// Set[string]{"a", "b", "c", "d", "e", "f"}

naturals := goset.NewSet[int](1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
primes := goset.NewSet[int](2, 3, 5, 7)

primes.IsSubset(naturals)
// true

primes.Contains(8)
// false

items := primes.Items()
// []int{2, 3, 5, 7}
```
`Set` works on any comparable object, so you can use it also for structs!
(But don't use it for structs with pointer fields, it will hurt) 
```go
type Person struct {
    Name string
}
peopleSet := goset.NewSet(Person{Name: "Amit"}, Person{Name: "Amit"})
println(peopleSet.Len())
// 1
```

## 📖 Spec

Full documentation at GoDoc: https://godoc.org/github.com/amit7itz/goset

Constructors:
- NewSet
- FromSlice

Methods:
- Add
- Contains
- Copy
- Discard
- IsEmpty
- Items
- Len
- Pop
- Remove
- String
- Difference
- Equal
- Intersection
- IsDisjoint
- IsSubset
- IsSuperset
- SymmetricDifference
- Union
- Update


## 🤝 Contributing

Open source is awesome!
Feel free to fork the [project](https://github.com/amit7itz/goset), open [issues](https://github.com/amit7itz/goset/issues), and request new features!

## 🖋️ Authors

- Amit Itzkovitch

## 💫 Show your support

If you liked it then you should have put a ⭐ on it!
(And let your friends know, so they can enjoy it too!)