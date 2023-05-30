;; TODO: Could do @scope.ignore to ignore this as a definition

(mod_item
 name: (_) @descriptor.namespace) @scope

(trait_item
 name: (_) @descriptor.type) @scope

(impl_item
 trait: (_) @descriptor.type
 type: (_) @descriptor.type) @scope

;; TODO: @local to stop traversal
(function_signature_item
 name: (identifier) @descriptor.method)

;; TODO: @local to stop traversal
(function_item
 name: (identifier) @descriptor.method)

(struct_item
 name: (type_identifier) @descriptor.type) @scope
