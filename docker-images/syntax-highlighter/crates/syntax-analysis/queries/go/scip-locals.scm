(func_literal) @scope
(function_declaration) @scope
(method_declaration) @scope
(expression_switch_statement) @scope
(expression_case) @scope
;; if_statement needs to open an extra scope, because you can define
;; variables inside an if in go:
;; if num := 10; num > 3 {
;;     ..
;; } else {
;;     ..
;; }
(if_statement) @scope
(for_statement) @scope
(block) @scope

(source_file
 (var_declaration
  (var_spec name: (identifier) @occurrence.skip)))

(var_spec
 name: (identifier) @definition.var)

(short_var_declaration
 left: (expression_list (identifier) @definition.term))

(source_file
 (const_declaration
  (const_spec
   name: (identifier) @occurrence.skip)))

(const_declaration
 (const_spec
  name: (identifier) @definition.var))

(parameter_declaration name: (identifier) @definition.term)
(variadic_parameter_declaration (identifier) @definition.var)

;; This syntax is only allowed to define top-level functions,
;; which we consider as non-locals, so we don't want to track these
;; here.
;;
;; (function_declaration
;;  name: ((identifier) @definition.function)
;;  (#set! "hoist" "function"))

;; import (
;;   f "fmt"
;;   ^- This is the spot that gets matched
;; )
;;
(import_spec_list
  (import_spec
    name: (package_identifier) @definition.namespace))

(for_statement
 (range_clause
   left: (expression_list
           (identifier) @definition.var)))

(identifier) @reference
(type_identifier) @reference
(field_identifier) @reference
