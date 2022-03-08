; Scopes

[
  (block)
  (method_declaration)
  (function_declaration)
  (func_literal)
  (source_file)
  (if_statement)
  (for_statement)
  (expression_switch_statement)
] @local.scope

; Definitions

(parameter_declaration (identifier) @local.definition)
(variadic_parameter_declaration (identifier) @local.definition)

(short_var_declaration
  left: (expression_list
          (identifier) @local.definition))

(var_spec
  name: (identifier) @local.definition)

(for_statement
 (range_clause
   left: (expression_list
           (identifier) @local.definition)))

(const_declaration
 (const_spec
  name: (identifier) @local.definition))

; References

(identifier) @local.reference

; helix had this, but I don't think it makes sense for us.
;   This makes sense from a code intelligence perspective,
;   but not from a syntax highlighting perspective
; (field_identifier) @local.reference
