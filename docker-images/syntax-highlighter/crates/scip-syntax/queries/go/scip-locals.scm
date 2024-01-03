(func_literal) @scope
(function_declaration) @scope
(method_declaration) @scope
(expression_switch_statement) @scope
;: See https://gobyexample.com/if-else for why if_statements need an
;; extra scope other than the blocks they open
(if_statement) @scope
(for_statement) @scope
(block) @scope

(short_var_declaration
 left: (expression_list (identifier) @definition.term))

;; TODO: We should talk about these: they could be params instead
(parameter_declaration name: (identifier) @definition.term)
(variadic_parameter_declaration (identifier) @definition.var)

;; This syntax is only allowed to define top-level functions,
;; which we consider as non-locals, so we don't want to track these
;; here.
;;
;; (function_declaration
;;  name: ((identifier) @definition.function)
;;  (#set! "hoist" "function"))

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

(identifier) @reference
(type_identifier) @reference
(field_identifier) @reference
