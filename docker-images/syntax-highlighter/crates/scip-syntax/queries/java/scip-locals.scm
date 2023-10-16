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


(for_statement) @scope

(lambda_expression

  parameters: [
   (identifier) @definition.term

   (inferred_parameters
        (identifier) @definition.term
    )
  ]
) @scope

(method_declaration
    name: (identifier) @definition.function
)

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
