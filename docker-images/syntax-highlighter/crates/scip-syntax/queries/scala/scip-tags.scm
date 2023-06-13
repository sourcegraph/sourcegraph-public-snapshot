; TODO: Exclude functions in non-container blocks

;; Matches against `package com.example`
(compilation_unit
 (package_clause
  name: (package_identifier) @descriptor.namespace
  !body)) @scope

;; Matches against `package inner { ... }`
(compilation_unit
 (package_clause
  name: (package_identifier) @descriptor.namespace
  body: (_) @scope))

(compilation_unit
  [(val_definition (identifier) @descriptor.term)
   (var_definition (identifier) @descriptor.term)])

(template_body
  [(val_definition (identifier) @descriptor.term)
   (var_definition (identifier) @descriptor.term)])

;; Function definitions.
(function_definition
  name: (_) @descriptor.method) @scope
(function_declaration
  name: (_) @descriptor.method) @scope

(class_definition
  name: (identifier) @descriptor.type) @scope

(class_parameter name: (identifier) @descriptor.term)

(object_definition (identifier) @descriptor.type (template_body) @scope)
(trait_definition (identifier) @descriptor.type (template_body) @scope)

(type_definition
  name: (type_identifier) @descriptor.type)
