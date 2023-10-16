; Scopes and type declarations
(class_declaration
    name: (identifier) @definition.type
) @scope

(interface_declaration
    name: (identifier) @definition.type
) @scope
(enum_declaration) @scope

(enum_declaration
    name: (identifier) @definition.type
) @scope

(record_declaration
    name: (identifier) @definition.type
) @scope

(method_declaration
    name: (identifier) @definition.function (#set! "scope" "parent")
)

(block) @scope


(for_statement) @scope

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

(local_variable_declaration
    (variable_declarator
        name: (identifier) @definition.term
    )
)

(identifier) @reference
(type_identifier) @reference
