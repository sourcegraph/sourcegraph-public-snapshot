;; TODO: Could do @scope.ignore to ignore this as a definition

(mod_item
  name: (_) @descriptor.namespace @kind.namespace) @scope

(trait_item
  name: (_) @descriptor.type @kind.trait) @scope

(impl_item
  trait: (generic_type type: (type_identifier) @descriptor.type @kind.trait)) @scope

(impl_item
  trait: (type_identifier) @descriptor.type @kind.trait) @scope

(impl_item
  type: (generic_type type: (type_identifier) @descriptor.type @kind.struct)
  body: (_) @scope)

(impl_item
  type: (type_identifier) @descriptor.type @kind.struct
  body: (_) @scope)

;; TODO: @local to stop traversal
(function_signature_item
  name: (identifier) @descriptor.method @kind.function)

(function_item
  name: (identifier) @descriptor.method @kind.function body: (_) @local)

(struct_item
  name: (type_identifier) @descriptor.type @kind.struct) @scope

(field_declaration
  name: (_) @descriptor.term @kind.field) @enclosing

(enum_item name: (_) @descriptor.type @kind.enum) @scope
(enum_variant name: (_) @descriptor.term @kind.enummember)

(const_item name: (_) @descriptor.term @kind.constant)
(static_item name: (_) @descriptor.term @kind.variable)
