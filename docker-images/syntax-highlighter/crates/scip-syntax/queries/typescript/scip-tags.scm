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

(interface_declaration name: (_) @descriptor.type body: (_) @scope)
(interface_declaration
    (object_type
        [
            (method_signature (property_identifier) @descriptor.method)
            (property_signature (property_identifier) @descriptor.term)
        ]
    )
)

(class_declaration name: (_) @descriptor.type body: (_) @scope)
(class_declaration
    (class_body
        [
            (public_field_definition name: (_) @descriptor.term)
            (method_definition name: (_) @descriptor.method body: (_) @local)
        ]
    )
)

(enum_declaration name: (_) @descriptor.type body: (_) @scope)
(enum_declaration
    (enum_body
        (property_identifier) @descriptor.term
    )
)

(module name: (string (string_fragment) @descriptor.namespace) body: (_) @scope)

[(if_statement) (while_statement) (for_statement) (do_statement)] @local
