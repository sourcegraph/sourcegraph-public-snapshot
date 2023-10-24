(program
    (package_declaration
        (_)
        @descriptor.namespace @scope
        (scoped_identifier)
        @descriptor.namespace @scope))

(class_declaration name: (_) @descriptor.type) @scope
(interface_declaration name: (_) @descriptor.type) @scope
(record_declaration name: (_) @descriptor.type) @scope
(enum_declaration name: (_) @descriptor.type) @scope

(method_declaration name: (_) @descriptor.method) @local
(constructor_declaration name: (_) @descriptor.method) @local

(field_declaration (variable_declarator name: (_) @descriptor.term))

; formal_parameters node is also used in method_declaration
; so we narrow the search to record_declaration children
(record_declaration
    (formal_parameters
        (formal_parameter
          name: (identifier) @descriptor.term)
    )
)
(enum_constant name: (_) @descriptor.term)
