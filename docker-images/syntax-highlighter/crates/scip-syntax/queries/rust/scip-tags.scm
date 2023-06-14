;; TODO: Could do @scope.ignore to ignore this as a definition

(mod_item
 name: (_) @descriptor.namespace) @scope

(trait_item
 name: (_) @descriptor.type) @scope

(impl_item
 trait: [(generic_type type: (type_identifier) @descriptor.type)
         (type_identifier) @descriptor.type]?

 type: [(generic_type type: (type_identifier) @descriptor.type)
        (type_identifier) @descriptor.type]) @scope

;; TODO: @local to stop traversal
(function_signature_item
 name: (identifier) @descriptor.method)

(function_item
 name: (identifier) @descriptor.method body: (_) @local)

(struct_item
 name: (type_identifier) @descriptor.type) @scope

(field_declaration
  name: (_) @descriptor.term) @enclosing

(enum_item name: (_) @descriptor.type) @scope
(enum_variant name: (_) @descriptor.term)
