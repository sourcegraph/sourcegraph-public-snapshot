(import_statement name: (_) @descriptor.term)
(import_from_statement name: (_) @descriptor.term)

(class_definition name: (_) @descriptor.type body: (_) @scope)
(function_definition name: (_) @descriptor.method body: (_) @local)
(expression_statement (assignment left: (identifier) @descriptor.term))

[(if_statement) (while_statement) (for_statement) (with_statement)] @local
