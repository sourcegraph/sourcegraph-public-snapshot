(class_declaration) @scope
(interface_declaration) @scope
(enum_declaration) @scope
(record_declaration) @scope
(method_declaration) @scope
(constructor_declaration) @scope


; NOTE: The definitions below are commented out
; as they overlap with global symbol indexing
; marking type declarations as locals causes
; various confusions, for example around constructors
;
; They are kept here for reference and to avoid re-introducing them

; (class_declaration
;     name: (identifier) @definition.type
; )

; (interface_declaration
;     name: (identifier) @definition.type
; )

; (enum_declaration
;     name: (identifier) @definition.type
; )

; (record_declaration
;     name: (identifier) @definition.type
; )

; (method_declaration
;     name: (identifier) @definition.function (#set! "scope" "parent")
; )

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

(record_pattern_component
  (identifier) @definition.term
)

; In Java grammar, declarations of type parameters are marked as type_parameter,
; and call-site generic parameters are marked as type_arguments
(type_parameter
    (type_identifier) @definition.type
)

(identifier) @reference
(type_identifier) @reference
