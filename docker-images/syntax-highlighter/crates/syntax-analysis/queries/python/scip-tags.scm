;; Allow (if_statement) since many conditional imports are created via if statements
[(while_statement) (for_statement) (with_statement)] @local

; Previously, ctags did not generate these.
; As of June 13, 2023: I will leave them out. We can come back to this later if need be.
;   I suspect generally speaking you do NOT want to reexport these cause it
;   will generate tons of symbols that are effectively nonsense.
; (import_statement name: (_) @descriptor.term)
; (import_from_statement name: (_) @descriptor.term)

(class_definition name: (_) @descriptor.type @kind.class body: (_) ) @scope
(class_definition
  body: (block
          [(function_definition
             name: (_) @descriptor.method @kind.method
             body: (_) @local)
           (decorated_definition
             definition: (function_definition
                          name: (_) @descriptor.method @kind.method
                          body: (_) @local))]))

(module (function_definition name: (_) @descriptor.method @kind.function body: (_) @local))
(module (decorated_definition (function_definition name: (_) @descriptor.method @kind.function body: (_) @local)))

;; foo = 1
(expression_statement (assignment left: (identifier) @descriptor.term @kind.variable))

;; foo, bar, baz = 1, 2, 3
(expression_statement (assignment left: (pattern_list (identifier) @descriptor.term @kind.variable)))
