(assignment left: (identifier) @definition.var (#set! "reassignment_behavior" "oldest_is_definition"))
(global_operator (identifier) @definition.var (#set! "reassignment_behavior" "oldest_is_definition"))
(persistent_operator (identifier) @definition.var (#set! "reassignment_behavior" "oldest_is_definition"))

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
