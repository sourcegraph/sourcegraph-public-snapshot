(compilation_unit) @scope
(class_definition) @scope
(object_definition) @scope
(trait_definition) @scope


(function_declaration
      name: (identifier) @definition.function)

(function_definition
      name: (identifier) @definition.function)

(parameter
  name: (identifier) @definition.term)

(binding
  name: (identifier) @definition.term)

(val_definition
  pattern: (identifier) @definition.term)

(var_definition
  pattern: (identifier) @definition.term)

(identifier) @reference
(type_identifier) @reference


(lambda_expression
      parameters: (identifier) @definition.function)

; (type_definition
;   name: (stable_type_identifier) @definition.type)
