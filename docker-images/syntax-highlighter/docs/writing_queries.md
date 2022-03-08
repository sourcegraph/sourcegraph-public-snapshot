# Writing Queries

General notes (will organize at a later date):

## Match Precedence

For cases where you have two things that can match, the first match that is
written in the query file will be the one selected by tree-sitter. For example:

```scheme
((identifier) @constant (#match? @constant "^[A-Z][A-Z\\d_]+$"))
(identifier) @variable
```

In this case, the `(identifier) @variable` will always match, so you want it to be put
after the one with a condition, so that it will be used as the fallback.

## Locals

(Probably worth writing something here, but will explore a bit more before writing)
