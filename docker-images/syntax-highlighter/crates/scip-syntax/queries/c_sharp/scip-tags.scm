(using_directive (qualified_name) @descriptor.type)

(class_declaration name: (_) @descriptor.type) @scope
(interface_declaration name: (_) @descriptor.type) @scope
(enum_declaration name: (_) @descriptor.type) @scope
(struct_declaration name: (_) @descriptor.type) @scope
(namespace_declaration name: (_) @descriptor.namespace) @scope

; Counter-intuitive name; it can actually be global
(local_function_statement name: (_) @descriptor.method)
(method_declaration name: (_) @descriptor.method)
(constructor_declaration name: (_) @descriptor.method)

(block) @local

(field_declaration (variable_declaration (variable_declarator (identifier) @descriptor.term)))
(event_field_declaration (variable_declaration (variable_declarator (identifier) @descriptor.term)))
(property_declaration name: (identifier) @descriptor.term)
(enum_member_declaration name: (_) @descriptor.term)
(delegate_declaration name: (identifier) @descriptor.method)
