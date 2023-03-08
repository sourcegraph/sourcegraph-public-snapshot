(block) @scope

(variable_declaration
  variable: (_) @definition.term)

(assignment_expression
  left: (variable_declaration
          (_) @definition.term))

(scalar) @reference
(array) @reference
(arraylen) @reference
(hash) @reference
(glob) @reference
