;; References
(function_declaration
  result: (type_identifier) @descriptor.type) @reference

(parameter_declaration
  type: (qualified_type
          package: (package_identifier) @descriptor.namespace
          name: (type_identifier) @descriptor.type)) @reference

(parameter_declaration
  type: (type_identifier) @descriptor.type) @reference

(parameter_declaration
  type: (pointer_type (type_identifier) @descriptor.type)) @reference

(call_expression
 function: (selector_expression
             operand: (_) @descriptor.namespace
             field: (_) @descriptor.method)) @reference


(call_expression function: (identifier) @descriptor.method) @reference
