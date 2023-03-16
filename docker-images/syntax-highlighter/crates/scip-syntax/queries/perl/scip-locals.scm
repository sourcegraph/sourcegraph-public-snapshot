(source_file) @scope
(block) @scope
(for_statement) @scope
(subroutine_declaration_statement) @scope

(variable_declaration
  variable: (_) @definition.term)

(assignment_expression
  left: (variable_declaration
          (_) @definition.term))

(for_statement my_var: (_) @definition.term)

;; Top Level Variables
(source_file (expression_statement (assignment_expression left: (_) @definition.term)))

;; TODO: We don't handle assignment_expressions for non-"my" variables.
;;       It would be great if we had a way to mark something like "first reference is a def"
;;       But I'm not sure how to do that yet.

(scalar) @reference
(array) @reference
(arraylen) @reference
(hash) @reference
(glob) @reference
