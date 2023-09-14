;; (source
  ;; (scope: Something)
  ;; (scope: Another
    ;; call Something))

(func_literal) @scope
(function_declaration) @scope
(method_declaration) @scope
(expression_switch_statement) @scope
(if_statement) @scope
(for_statement) @scope
(block) @scope

(short_var_declaration
  left: (expression_list (identifier) @definition.term))

;; TODO: We should talk about these: they could be params instead
(parameter_declaration name: (identifier) @definition.term)
(variadic_parameter_declaration (identifier) @definition.var)

(function_declaration
    name: ((identifier) @definition.function
           (#set! "scope" "global")))

((method_declaration name: (field_identifier) @definition.method))

;; import (
;;   f "fmt"
;;   ^- This is the spot that gets matched
;; )
;;
(import_spec_list
  (import_spec
    name: (package_identifier) @definition.namespace))

(var_spec
  name: (identifier) @definition.var)

(for_statement
 (range_clause
   left: (expression_list
           (identifier) @definition.var)))

(const_declaration
 (const_spec
  name: (identifier) @definition.var))

(type_declaration
  (type_spec
    name: (type_identifier) @definition.type))

;; reference
(identifier) @reference
(type_identifier) @reference
(field_identifier) @reference

; ;; Call references
; ((call_expression
;    function: (identifier) @reference)
;  (set! reference.kind "call"))
;
; ((call_expression
;     function: (selector_expression
;                 field: (field_identifier) @reference))
;  (set! reference.kind "call"))
;
;
; ((call_expression
;     function: (parenthesized_expression
;                 (identifier) @reference))
;  (set! reference.kind "call"))
;
; ((call_expression
;    function: (parenthesized_expression
;                (selector_expression
;                  field: (field_identifier) @reference)))
;  (set! reference.kind "call"))
;

;; TODO: These may not make much sense to have for locals... {{{
((package_identifier) @reference
  (set! reference.kind "namespace"))

(package_clause
   (package_identifier) @definition.namespace)
;; }}}
