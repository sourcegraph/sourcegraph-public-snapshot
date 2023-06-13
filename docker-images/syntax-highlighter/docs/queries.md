# scip-tags.scm

## Captures:

- `@descriptor.SUFFIX`
  - Used to create a new descriptor for the query match.
  - For example, `@descriptor.term` will create a new term descriptor with the string contents of whatever node is captured
  - Can use more than one of these per match.
    - For example, in Go, when you declare a method, you would do `func (thing *MyThing) ThisFunc() {}`. In this case, you want to associate
      `ThisFunc` with the struct `MyThing. You can do that via the following query:

```
(method_declaration
  receiver: (parameter_list
               (parameter_declaration type: (type_identifier) @descriptor.type))
  name: (field_identifier) @descriptor.method)
```

- `@scope`

  - Used to create a new scope, with whatever descriptors are defined by this query.
  - This allows namespacing nested elements

- `@enclosing`

  - Used primarily for the `/symbols` endpoint, but gives the enclosing range for a particular symbol.
  - Does not need to be used with `@scope` because `@scope` already gives us the range already.
    - TODO: In the future, it may be possible we come up with some scenarios for this,
      but I haven't found any use cases for it at the moment

- `@local`
  - Use this to ignore any new symbols that might be generated within this block.
  - Future improvement would hope that we just skip parsing / matching on this block, but I don't think that's
    feasible at the moment. For now it just notices the match and skips doing anymore work on it.

## Predicates

- `(#filter! @node "node-kind-1" "node-kind-2" ...)`
  - `#filter!` can be used to make certain cases NOT match. This is useful for if you want anything EXCEPT some case to match.

For example:

```scheme
;; {{{ Handle multiple scenarios of literal objects at top level
;; var X = { key: value }
;;           ^^^ X.key
;;
;;   First query makes sure to make a method
;;   Second query collects the rest of the options as a term
;;     (best effort method detection)
;;
(object
  (pair
    key: (property_identifier) @descriptor.method
    value: [(function) (arrow_function)]))

((object
   (pair
     key: (property_identifier) @descriptor.term
     value: (_) @_value_type))
 (#filter! @_value_type "function" "arrow_function"))
;; }}}
```

- `(#transform! "regex-string" "regex-replacement")`
  - `#transform!` can be used to take the last descriptor of a match and generat new identifiers from it. This is useful when
    a language feature mangles a name in some new way, but it's predicatble in text.

```scheme
;; attr_accessor :bar -> bar, bar=
((call
   method: (identifier) @_attr_accessor
   arguments: (argument_list (simple_symbol) @descriptor.method))
 (#eq? @_attr_accessor "attr_accessor")
 (#transform! ":(.*)" "$1")
 (#transform! ":(.*)" "$1="))
```
