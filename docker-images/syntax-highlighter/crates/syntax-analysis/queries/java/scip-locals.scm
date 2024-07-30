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

(variable_declarator
    name: (identifier) @definition.term
)

(record_pattern_component
  (identifier) @definition.term
)


(formal_parameter
    name: (identifier) @definition.term
)

; In Java grammar, declarations of type parameters are marked as type_parameter,
; and call-site generic parameters are marked as type_arguments
(type_parameter
    (type_identifier) @definition.type
)


; SKIPPED DECLARATIONS
; We mark the declarations as skipped occurrences
; to avoid marking them as references

(record_declaration
 (formal_parameters
  (formal_parameter
   name: (identifier) @occurrence.skip)))

(method_declaration
  name: (identifier) @occurrence.skip)

(field_declaration
 (variable_declarator
  name: (identifier) @occurrence.skip))

(class_declaration
  name: (identifier) @occurrence.skip
)

(interface_declaration
  name: (identifier) @occurrence.skip
)

(enum_declaration
  name: (identifier) @occurrence.skip
)

(enum_constant
  name: (identifier) @occurrence.skip
)

(record_declaration
  name: (identifier) @occurrence.skip
)

(package_declaration [
          (scoped_identifier) @occurrence.skip
          (identifier) @occurrence.skip
      ]
)

(package_declaration
    (scoped_identifier
      name: (_) @occurrence.skip))

; REFERENCES


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
    (#eq? @occurrence.skip "var")
)

; Person::getName
; ^^^^^^  ^^^^^^^
(method_reference (identifier)* @reference (#set! "kind" "global.method"))

; type references are generally global
((type_identifier) @reference (#set! "kind" "type"))
; all other references we assume to be local only
(identifier) @reference
