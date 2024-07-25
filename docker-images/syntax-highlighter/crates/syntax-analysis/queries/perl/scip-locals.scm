(source_file) @scope
(block) @scope
(for_statement) @scope
(subroutine_declaration_statement) @scope

;; TODO: Add `state` variables once we've updated the grammar. Only
;; `our` variables are non-local, so we must not include them here
(variable_declaration "my" (_) @definition.term)
(for_statement my_var: (_) @definition.term)

((scalar) @reference (#set! "kind" "local"))
((array) @reference (#set! "kind" "local"))
((arraylen) @reference (#set! "kind" "local"))
((hash) @reference (#set! "kind" "local"))
((glob) @reference (#set! "kind" "local"))
