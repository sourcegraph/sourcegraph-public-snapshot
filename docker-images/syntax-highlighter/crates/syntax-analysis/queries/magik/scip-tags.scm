(source_file
    (package (identifier) @descriptor.namespace @kind.package)
) @scope

; This matches a global top-level assignments
(fragment "_constant" (identifier) @kind.constant @descriptor.term)
(fragment "_global" (identifier) @kind.variable @descriptor.term)

(invoke
  receiver: (variable) @name
  (symbol) @descriptor.type @kind.struct
  (#eq? @name "def_slotted_exemplar")) @scope

(method
    exemplarname: (_) @descriptor.type
    name: (_) @descriptor.method @kind.method)

[
    (procedure)
    (block)
    (iterator)
    (while)
    (try)
    (loop)
    (if)
] @local
