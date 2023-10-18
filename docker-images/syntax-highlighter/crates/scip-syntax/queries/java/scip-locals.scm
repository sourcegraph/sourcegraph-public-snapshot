(class_declaration) @scope
(interface_declaration) @scope
(enum_declaration) @scope
(record_declaration) @scope
(method_declaration) @scope

(class_declaration
    name: (identifier) @definition.type
)

(interface_declaration
    name: (identifier) @definition.type
)

(enum_declaration
    name: (identifier) @definition.type
)

(record_declaration
    name: (identifier) @definition.type
)

(method_declaration
    name: (identifier) @definition.function (#set! "scope" "parent")
)

(block) @scope


(for_statement) @scope

(enum_constant
  name: (identifier) @definition.term
)

(enhanced_for_statement
    name: (identifier) @definition.term
) @scope

(lambda_expression

  parameters: [
   (identifier) @definition.term

   (inferred_parameters
        (identifier) @definition.term
    )
  ]
) @scope

(formal_parameter
    name: (identifier) @definition.term
)

(variable_declarator
    name: (identifier) @definition.term
)

(identifier) @reference
(type_identifier) @reference
