(class_declaration) @scope
(interface_declaration) @scope
(enum_declaration) @scope
(record_declaration) @scope
(method_declaration) @scope
(constructor_declaration) @scope
(lambda_expression) @scope
(enhanced_for_statement) @scope
(for_statement) @scope
(block) @scope


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
;     name: (identifier) @definition.function
; )

; (enum_constant
;   name: (identifier) @definition.term
; )

(enhanced_for_statement
    name: (identifier) @definition.term)

(lambda_expression

  parameters: [
   (identifier) @definition.term

   (inferred_parameters
        (identifier) @definition.term
    )
  ]
)

(record_declaration
 (formal_parameters
  (formal_parameter
   name: (identifier) @occurrence.skip)))

(formal_parameter
    name: (identifier) @definition.term
)

(method_declaration
  name: (identifier) @occurrence.skip)

(field_declaration
 (variable_declarator
  name: (identifier) @occurrence.skip))

(variable_declarator
    name: (identifier) @definition.term
)

(record_pattern_component
  (identifier) @definition.term
)

(class_declaration
  name: (identifier) @occurrence.skip
)

(interface_declaration
  name: (identifier) @occurrence.skip
)


; In Java grammar, declarations of type parameters are marked as type_parameter,
; and call-site generic parameters are marked as type_arguments
(type_parameter
    (type_identifier) @definition.type
)

; REFERENCES

(package_declaration (identifier) @occurrence.skip)

; import java.util.HashSet
;        ^^^^^^^^^ namespace
;                  ^^^^^^^ type (could also be a constant, but type is more common)
(import_declaration
  (scoped_identifier
    scope: ((_) @reference (#set! "kind" "global.namespace"))
    ))

(import_declaration
  (scoped_identifier
    name: (_) @reference (#set! "kind" "global.type")
  ))

(field_access object: (identifier) @reference)
(field_access field: (identifier) @reference (#set! "kind" "global.term"))

; hello(...)
; ^^^^^
; As we don't support local methods yet, we unequivocally mark this reference
; as global
(method_invocation
  name: (identifier) @reference (#set! "kind" "global.method")
)

; @Autowired(1, MyClass.class)
;  ^^^^^^^^^
; class MyClass()
(marker_annotation
    name: (identifier) @reference (#set! "kind" "global.type")
)

; MyType variable = ...
; ^^^^^^
(local_variable_declaration
  type: (type_identifier) @occurrence.skip
    (#eq? @reference "var")
)

; class Binary<N extends Number> {...
;                        ^^^^^^
(type_bound
  ((type_identifier)* @reference (#set! "kind" "type"))
)




; Person::getName
; ^^^^^^  ^^^^^^^
(method_reference (identifier)* @reference (#set! "kind" "global.method"))

; type references are generally global
((type_identifier) @reference (#set! "kind" "type"))
; all other references we assume to be local only
(identifier) @reference (#set! "kind" "local")
