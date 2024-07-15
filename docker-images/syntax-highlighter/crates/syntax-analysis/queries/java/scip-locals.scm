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

(field_declaration
 (variable_declarator
  name: (identifier) @occurrence.skip))

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

(field_access field: (identifier) @occurrence.skip)

; REFERENCES

; new MyType(...)
;     ^^^^^^
(object_creation_expression
    type: [
           ; This can be a reference to a local (method's type parameter)
           ; or a global
           (type_identifier) @reference.either

           ; For nested classes used in `new` expressions (e.g. `new TodoApp.Item`)
           ; we emit references to TodoApp, Item, and TodoApp.Item - the latter
           ; to bump up the fuzzy matching against this exact form
           (scoped_type_identifier
               (type_identifier)* @reference.either
            ) @reference.ether
    ]
)

; hello(...)
; ^^^^^
; As we don't support local methods yet, we unequivocally mark this reference
; as global
(method_invocation
  name: (identifier) @reference.global
)

; MyType variable = ...
; ^^^^^^
(local_variable_declaration
  type: (type_identifier) @reference
    (#not-eq? @reference "var")
)

; class Binary<N extends Number> {...
;                        ^^^^^^
(type_bound
  (type_identifier)* @reference
)

; for (MyType variable: variables) {...
;      ^^^^^^
(enhanced_for_statement
  type: (type_identifier) @reference
)

; public class test<T extends Exception> {
; 	private void provideFieldValue()
; 			throws T, NoSuchFieldException, IllegalAccessException {}
;                  ^  ^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^
; }
(throws (type_identifier)* @reference)

; Person::getName
; ^^^^^^  ^^^^^^^
(method_reference (identifier)* @reference.global)

; all other references we assume to be local only
(identifier) @reference.local
(type_identifier) @reference.local
