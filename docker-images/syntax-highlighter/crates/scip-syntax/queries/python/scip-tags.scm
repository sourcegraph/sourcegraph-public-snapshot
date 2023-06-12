;; Allow (if_statement) since many conditional imports are created via if statements
[(while_statement) (for_statement) (with_statement)] @local

(import_statement name: (_) @descriptor.term)
(import_from_statement name: (_) @descriptor.term)

(class_definition name: (_) @descriptor.type body: (_) @scope)
(function_definition name: (_) @descriptor.method body: (_) @local)

;; foo = 1
(expression_statement (assignment left: (identifier) @descriptor.term))

;; foo, bar, baz = 1, 2, 3
(expression_statement (assignment left: (pattern_list (identifier) @descriptor.term)))
