# Scoping of different programming languages

In order to find a small number of features that support as many languages as possible we analyzed the scoping behaviour of various programming languages.
This led us to the design of the current implementation of `src/locals.rs`.
The main axis along which we categorized languages was their "hoisting" behaviour.

## Hoisting

Hoisting means lifting definitions/names to a particular scope thus making them visible even for references lexically preceding them.

As an example, consider the following Python code:

```python
def outer():
    def inner():
        print(x)
    x = 10
    inner()
```

Calling `outer()` will print `10`, because Python _hoists_ local variables to the top of their enclosing function.
This means `inner` is able to reference `x`, even though it is referenced before the definition.

## Python does have hoisting

```python
x = 10 # def 1

def my_fun():
    print(f"global first {x}") # ref 2
    x = 20 # def 2
my_fun()
```

Results in the following error:

    Traceback (most recent call last):
      File "<stdin>", line 7, in <module>
      File "<stdin>", line 4, in my_fun
    UnboundLocalError: cannot access local variable 'x' where it is not associated with a value

Variables are not exactly lexically scoped.
While they can be referenced, assigning to a variable from a parent scope in a nested function actually creates a new local binding in that function instead.
This can be circumvented using the `global` or `nonlocal` keywords, but I don't know how I'd go about handling that in a language-agnostic fashion.

Notably Python hoists to the function level, so assigning a variable in a nested-if creates a new function scoped variable.

```python
y = 0 # def 0
x = 10 # def 4

def outer():
    y = 10 # def 1
    def inner():
      y = 20 # def 2
      print(f"{x} and {y} from inner") # ref 3, ref 2
    x = 40 # def 3
    inner()
    print(f"{x} and {y} from outer") # ref 3, ref 1
outer()
```

Results:

    40 and 20 from inner
    40 and 10 from outer


## JS/TS does have hoisting

Here's a program that shows that JS uses hoisting for both let and var, but var is hoisted to the nearest function scope, while let is block scoped.

```js
let y = 15 // def 1

function outer() {
    function inner() {
        { var y = 20 }; // def 2
        console.log(`${x} and ${y} from inner`) // ref 5, ref 2
    };

    function inner2() {
        { let y = 20 }; // def 3
        // Captures the y from _outer_ and not the global because of
        // hoisting
        console.log(`${x} and ${y} from inner2`) // ref 5, ref 4
    };

    let y = 10; // def 4
    var x = 40; // def 5
    inner();
    inner2();
    console.log(`${x} and ${y} from outer`); // ref 5, ref 4
}
outer()
```

Results:

    40 and 20 from inner
    40 and 10 from inner2
    40 and 10 from outer


## Java does not have hoisting, but we can treat it as if it did anyway

While Java does not exactly do hoisting, it also doesn't allow the kinds of shadowing one would need to exploit that fact.
We can treat it as if it did do (block scoped) hoisting, and we'd only 'get it wrong' for code that javac wouldn't accept anyway.


## Scala does not have hoisting, but we can treat it as if it did anyway

While Scala is fine with shadowing it still does not allow two different references to a variable of the same name in the same block.
This means it can also be treated as a language with (block scoped) hoisting.


## Go does not have hoisting

Here's a program showing how Go has full lexical scoping.

```go
package main

import "fmt"

var x = 10 // def 1

func myFun() {
        fmt.Printf("global first %v\n", x) // ref 1
        x = 40 // ref 1

        var x = 20 // def 2
        fmt.Printf("local first %v\n", x) // ref 2
        x = 30 // ref 2
        fmt.Printf("local second %v\n", x) // ref 2
}

func main() {
        myFun()
        fmt.Printf("global second %v\n", x) // ref 1
}
```

Results:

    global first 10
    local first 20
    local second 30
    global second 40


## Perl does not have hoisting

Here's a program showing how Perl has full lexical scoping (for `my` variables).
It also has full on dynamic scoping with the `local` keyword, but there's no hope for statically analyzing that.

```perl
my $x = 10; # def 1

sub my_fun {
    print "global first $x\n"; # ref 1
    $x = 40; # ref 1

    my $x = 20; # def 2
    print "local first $x\n"; # ref 2
    $x = 30; # ref 2
    print "local second $x\n"; # ref 2
}

my_fun();

print "global second: $x\n"; # ref 1
```

Results:

    global first 10
    local first 20
    local second 30
    global second: 40
