(source_file
    (package_header
        (identifier)
        @descriptor.namespace)) @scope

(function_declaration
  (simple_identifier) @descriptor.method
  (function_body) @local)

(anonymous_function (_ (type_identifier) @descriptor.type . (type_identifier) @descriptor.method)) @local
(class_declaration (type_identifier) @descriptor.type) @scope
(object_declaration (type_identifier) @descriptor.type) @scope
(class_parameter (simple_identifier) @descriptor.term)
(enum_entry (simple_identifier) @descriptor.term)
(property_declaration (variable_declaration (simple_identifier) @descriptor.term))

(multi_variable_declaration (variable_declaration (simple_identifier) @descriptor.term))

;; Future TODOs:
;; - Should probably unescape `Escaped` simple identifiers
