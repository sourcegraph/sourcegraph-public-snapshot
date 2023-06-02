(class_declaration name: (_) @descriptor.type) @scope
(interface_declaration name: (_) @descriptor.type) @scope
(enum_declaration name: (_) @descriptor.type) @scope
(namespace_declaration name: (_) @descriptor.type) @scope

(method_declaration name: (_) @descriptor.method)
(constructor_declaration name: (_) @descriptor.method)

(field_declaration (variable_declaration (variable_declarator (identifier) @descriptor.term)))
(property_declaration (identifier) @descriptor.term)
(enum_member_declaration name: (_) @descriptor.term)
