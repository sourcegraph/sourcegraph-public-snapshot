(program
    (package_declaration
        (_)
        @descriptor.namespace @scope
        (scoped_identifier)
        @descriptor.namespace @scope))

(class_declaration name: (_) @descriptor.type) @scope
(interface_declaration name: (_) @descriptor.type) @scope
(enum_declaration name: (_) @descriptor.type) @scope

(method_declaration name: (_) @descriptor.method) @local
(constructor_declaration name: (_) @descriptor.method) @local

(field_declaration (variable_declarator name: (_) @descriptor.term))
(enum_constant name: (_) @descriptor.term)
