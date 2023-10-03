[(for_statement) (if_statement) (switch_statement) (while_statement) (block)] @scope

(assignment left: (identifier) @definition.var (#set! "strength" "weak"))
(global_operator (identifier) @definition.var (#set! "strength" "weak"))
(persistent_operator (identifier) @definition.var (#set! "strength" "weak"))

(function_definition) @scope
(function_definition
    name: (identifier) @definition.function
)
(function_arguments 
    (identifier) @definition.term
)
(function_output
    [
        (multioutput_variable
            (identifier) @definition.term
        )
        (identifier) @definition.term
    ]
)

(class_definition name: (identifier) @definition.type) @scope

(lambda) @scope
(lambda (arguments (identifier) @definition.term))

(identifier) @reference
