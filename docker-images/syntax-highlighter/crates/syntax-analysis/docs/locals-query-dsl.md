# The Query DSL

For specifying how to resolve local bindings we use the [tree-sitter query] language and a set of custom captures and properties.
The three main concepts are _scopes_, _definitions_, and _references_.

## Scopes

Scopes are specified by labeling a capture as a `@scope[.kind]`, where kind can be any string.
The optional scope kind can be used to [hoist][hoisting] definitions to scopes of that kind.
There is an implicit top-level scope that is of kind `"global"`

### Examples

```scm
(block @scope)
(function_declaration @scope.function)
```

## Definitions

Definitions are specified by labeling a capture as a `@definition`.

### Lexical scoping

```scm
(variable_definition (identifier) @definition)
```

The default behaviour for a definition is to be visible to all references that appear lexically _after_ the definition.

```js
print(my_var) // Will not be resolved

let my_var = 10
print(my_var) // Will be resolved
```

### Hoisting

For more details see [hoisting] in the scoping documentation.

If you want a definition to be _hoisted_ to the start of a scope instead, you can specify the kind of the nearest enclosing scope it should be hoisted to.

```scm
(function_definition
 (identifier) @definition
 #set! "hoist" "function")
```

The definition will be visible to the nearest enclosing scope with kind `function`.
If no such enclosing scope is found, the definition will be visible in the global scope.

```js
// Will be resolved as `global_func` will be visible at the `global` scope
global_func(10)

function global_func(x) {
  // Will be resolved as `local_func` is hoisted to the top of `global_func`'s scope
  local_func(10)
  function local_func(y) {
    print(y)
  }
}
```

### First assignment is Definition

Certain languages (Python, MATLAB etc.) do not have special syntactic forms for introducing variables.
Instead the first assignment of a variable is considered to be its definition, and all further ones are references.
To support this in our DSL you can can mark a `@definition` as a 'def_ref'.

```scm
(assignment
 (identifier) @definition
 #set! "def_ref")
```

If you also specify a hoist level, only existing assignments that match the current hoist-level will be considered when deciding between definition and reference.
As an example, here's how local variables in Python functions could be handled.

```scm
(python_assignment
 (identifier) @definition
 #set! "def_ref"
 #set! "hoist" "function")
```

```python
a = 10 # definition 1
def f():
  # This assignment gets hoisted to the `f` function, which means
  # it won't consider a binding in parent scopes
  a = 3 # definition 2
  if True:
    a = 4 # reference 2
a = 4 # reference 1
```

## References

References are specified by labeling a capture as a `@reference`.

```scm
(variable_expression (identifier) @reference)
```

If the type of symbol under reference can be inferred from syntactic information, you can
add a `"kind"` attribute to refine the descriptor that will be produced:

```scm
(method_invocation
  name: (identifier) @reference (#set! "kind" "method")
)
```

The kind will be resolved into a [descriptor suffix](https://github.com/sourcegraph/scip/blob/main/scip.proto#L211-L222)
in case this reference is resolved to a non-local symbol.

As references can be either local or non-local, you can use the value of `"kind"` descriptor to guide
the resolution:

1. `"<type>"` or absent - these references will first be resolved against definitions in current and parent scopes, and
   if that resolution fails, a non-local reference will be produced instead.

   Example:

   ```scm
   (package
     (identifier) @reference (#set! "kind" "namespace")
   )
   ```

2. `"global.<type>"` - these references will immediately emit a non-local reference, without attempting to
   do any resolution against local definitions.

   Example:

   ```scm
   (method_invocation
     name: (identifier) @reference (#set! "kind" "global.namespace")
   )
   ```

When resolving a reference against local definitions, non-hoisted definitions are only resolved if they are defined _before_ the reference.

## Skipping definitions & references

Because we want to exclude non-local definitions and references when collecting locals, it's possible to mark an occurence as `@occurrence.skip`.
This will make it so the occurrence is not included in the output and all future matches of it will be skipped.
It's important that you specify skip matches _before_ regular definition and reference matches.

```scm
;; Skip top-level var_spec definitions
(source_file (var_spec (identifier) @occurrence.skip))

;; Captures all var_spec definitions as definitions
(var_spec (identifier) @definition)

;; Skip field_access as references
(field_access field: (identifier) @occurrence.skip)

(identifier) @reference
```

```go
// Will be skipped
var top_level = 10

func main() {
    // Will be recorded as a local
    var local = 10

    // local will be recorded, but field will be skipped
    local.field
}
```


[hoisting]: ./locals-scoping.md#hoisting
[tree-sitter query]: https://tree-sitter.github.io/tree-sitter/using-parsers#pattern-matching-with-queries
