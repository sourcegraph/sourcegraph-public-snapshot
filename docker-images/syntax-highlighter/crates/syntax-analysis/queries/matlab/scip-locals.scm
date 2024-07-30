(function_definition) @scope.function
(lambda) @scope

(assignment
 left: (identifier) @definition.var
 (#set! "def_ref"))
(global_operator
 (identifier) @definition.var
 (#set! "def_ref"))
(persistent_operator
 (identifier) @definition.var
 (#set! "def_ref"))

;; MATLAB exports the _first_ function from a module as a non-local,
;; which this query matches using tree-sitter's anchor syntax (the
;; dot).
(source_file . (function_definition
 name: (identifier) @occurrence.skip))

(properties
 (property name: [(identifier) (property_name (identifier))] @occurrence.skip))
(methods
 (function_definition name: (identifier) @occurrence.skip))

(function_definition
 name: (identifier) @definition.function
 (#set! "hoist" "function"))

(function_arguments
 (identifier) @definition.term)

(function_output
    [
        (multioutput_variable
            (identifier) @definition.term
        )
        (identifier) @definition.term
    ]
)

(lambda (arguments (identifier) @definition.term))

(class_definition name: (identifier) @occurrence.skip)
(field_expression
 field: [(identifier)
         (function_call (identifier))] @occurrence.skip)
((identifier) @reference)
