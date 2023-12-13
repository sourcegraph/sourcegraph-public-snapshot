(function_definition) @scope.function
(lambda) @scope
(class_definition
 name: (identifier) @definition.type
 (#set! "hoist" "global")) @scope

(assignment left: (identifier) @definition.var (#set! "reassignment_behavior" "oldest_is_definition"))
(global_operator (identifier) @definition.var (#set! "reassignment_behavior" "oldest_is_definition"))
(persistent_operator (identifier) @definition.var (#set! "reassignment_behavior" "oldest_is_definition"))

(function_definition
 name: (identifier) @definition.function
 (#set! "hoist" "function"))

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

(lambda (arguments (identifier) @definition.term))

(identifier) @reference
