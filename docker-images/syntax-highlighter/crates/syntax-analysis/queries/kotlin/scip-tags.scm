(source_file
    (package_header
        (identifier)
        @descriptor.namespace @kind.package)) @scope

(function_declaration
  (simple_identifier) @descriptor.method @kind.method
  (function_body)? @local)

(anonymous_function) @local

(class_declaration "interface" (type_identifier) @descriptor.type @kind.interface) @scope
(class_declaration "enum" "class" (type_identifier) @descriptor.type @kind.enum) @scope

;; Exclude enums from the 'class' kind
((class_declaration
    ("enum")? @_enum "class"
    (type_identifier) @descriptor.type @kind.class)
  (#filter! @_enum "enum")) @scope

(object_declaration (type_identifier) @descriptor.type @kind.object) @scope
(companion_object (type_identifier) @descriptor.type @kind.object) @scope

(type_alias (type_identifier) @descriptor.type @kind.typealias)

(class_parameter (simple_identifier) @descriptor.term @kind.property)
(enum_entry (simple_identifier) @descriptor.term @kind.enummember)

;; In the grammar, property_modifier always represents 'const'
(property_declaration
    (modifiers (property_modifier))
    (variable_declaration (simple_identifier) @descriptor.term @kind.constant))

;; Exclude constants from the 'property' kind
((property_declaration
    (modifiers (property_modifier) @_const)?
    (variable_declaration (simple_identifier) @descriptor.term @kind.property))
  (#filter! @_const "property_modifier"))

(property_declaration
  (multi_variable_declaration (variable_declaration (simple_identifier) @descriptor.term @kind.property)))

;; Future TODOs:
;; - Should probably unescape `Escaped` simple identifiers
