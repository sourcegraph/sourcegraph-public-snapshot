; Scopes
;-------

[
  (let_binding)
  (let_expression)
  (class_binding)
  (class_function)
  (method_definition)
  (fun_expression)
  (object_expression)
  (for_expression)
  (match_case)
  (attribute_payload)] @scope


; Definitions
;------------

(value_pattern) @definition.term
(let_binding pattern: (_) @definition.term)
(module_binding name: (module_name) @definition.type)

; References
;-----------

(value_path . (value_name) @reference)
(module_path (module_name) @reference)
