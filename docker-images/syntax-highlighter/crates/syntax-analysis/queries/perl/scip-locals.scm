(source_file) @scope
(block) @scope
(for_statement) @scope
(subroutine_declaration_statement) @scope

;; TODO: Add `state` variables once we've updated the grammar. Only
;; `our` variables are non-local, so we must not include them here
(variable_declaration "my" (_) @definition.term)
(for_statement my_var: (_) @definition.term)

(scalar) @reference.local
(array) @reference.local
(arraylen) @reference.local
(hash) @reference.local
(glob) @reference.local
