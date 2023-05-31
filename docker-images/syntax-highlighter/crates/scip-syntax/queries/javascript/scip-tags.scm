(namespace_import (identifier) @descriptor.term)
(named_imports
    [
        (import_specifier alias: (_) @descriptor.term)
        (import_specifier name: (_) @descriptor.term !alias)
    ]
)

(function_declaration (identifier) @descriptor.method body: (_) @local)
(lexical_declaration (variable_declarator name: (identifier) @descriptor.term))
(variable_declaration (variable_declarator name: (identifier) @descriptor.term))

(class_declaration name: (_) @descriptor.type body: (_) @scope)
(class_declaration
    (class_body
        [
            (method_definition name: (_) @descriptor.method body: (_) @local)
        ]
    )
)

[(if_statement) (while_statement) (for_statement) (do_statement)] @local
