; Mark method/function parameters as local
(parameters (parameter)) @local

(namespace_declaration
    name:  (qualified_identifier (identifier)) @descriptor.namespace @kind.namespace) @scope

(class_declaration
  name: (identifier) @descriptor.type @kind.class) @scope

(interface_declaration
  name: (identifier) @descriptor.type @kind.interface) @scope

(trait_declaration
  name: (identifier) @descriptor.type @kind.trait) @scope

(function_declaration
  name: (identifier) @descriptor.method @kind.function
  body: (_) @local)

; Match interface declarations that do not have a body
(method_declaration
  name: (identifier) @descriptor.method @kind.method
  !body)

(method_declaration
  name: (identifier) @descriptor.method @kind.constructor (#eq? @descriptor.method "__construct")
  body: (_) @local)

(method_declaration
  name: (identifier) @descriptor.method @kind.method (#not-eq? @descriptor.method "__construct")
  body: (_) @local)

(property_declaration
    (property_declarator name: (variable) @descriptor.term @kind.property)
    (#transform! "[$](.*)" "$1"))

(const_declaration
    (const_declarator name: (identifier) @descriptor.term @kind.constant))

(enum_declaration
    name: (identifier) @descriptor.type @kind.enum
) @scope

(enum_class_declaration
    name: (identifier) @descriptor.type @kind.enum
) @scope

(enumerator (identifier) @descriptor.term @kind.enummember)

(alias_declaration (identifier) @descriptor.type @kind.typealias)

(type_const_declaration name: (identifier) @descriptor.type @kind.typealias)
