(class_declaration name: (_) @descriptor.type @kind.class) @scope
(interface_declaration name: (_) @descriptor.type @kind.interface) @scope
(enum_declaration name: (_) @descriptor.type @kind.enum) @scope
(struct_declaration name: (_) @descriptor.type @kind.struct) @scope
(namespace_declaration name: (_) @descriptor.namespace @kind.namespace) @scope

; Counter-intuitive name; it can actually be global
(local_function_statement name: (_) @descriptor.method @kind.function)
(method_declaration name: (_) @descriptor.method @kind.method)
(constructor_declaration name: (_) @descriptor.method @kind.constructor)

(block) @local

(field_declaration (variable_declaration (variable_declarator (identifier) @descriptor.term @kind.field)))
(event_field_declaration (variable_declaration (variable_declarator (identifier) @kind.event @descriptor.term)))
(property_declaration name: (identifier) @descriptor.term @kind.property)
(enum_member_declaration name: (_) @descriptor.term @kind.enummember)
(delegate_declaration name: (identifier) @descriptor.method @kind.delegate)
