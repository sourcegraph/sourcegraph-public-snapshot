(block) @scope
(for_statement) @scope
(subroutine_declaration_statement) @scope

(variable_declaration
  variable: (_) @definition.term)

(assignment_expression
  left: (variable_declaration
          (_) @definition.term))

(for_statement my_var: (_) @definition.term)

(scalar) @reference
(array) @reference
(arraylen) @reference
(hash) @reference
(glob) @reference
