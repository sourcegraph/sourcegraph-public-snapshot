(namespace_import (identifier) @descriptor.term @kind.variable)
(named_imports
  (import_specifier alias: (_) @descriptor.term @kind.variable))
(named_imports
  (import_specifier name: (_) @descriptor.term @kind.variable !alias))

;; Function / Generator declaration.
;;   Don't think there is any reason to expose anything from within the body of the functions
(function_declaration (identifier) @descriptor.method @kind.function body: (_) @local)
(generator_function_declaration  (identifier) @descriptor.method @kind.function body: (_) @local)

(lexical_declaration (variable_declarator name: (identifier) @descriptor.term @kind.variable)) @scope
(variable_declaration (variable_declarator name: (identifier) @descriptor.term @kind.variable)) @scope

;; {{{ Handle multiple scenarios of literal objects at top level
;; var X = { key: value }
;;           ^^^ X.key
;;
;;   First query makes sure to make a method
;;   Second query collects the rest of the options as a term
;;     (best effort method detection)
;;
(object
  (pair
    key: (property_identifier) @descriptor.method @kind.function
    value: [(function) (arrow_function)]))

((object
   (pair
     key: (property_identifier) @descriptor.term @kind.property
     value: (_) @_value_type))
 (#filter! @_value_type "function" "arrow_function"))
;; }}}

;; class X { ... }
(class_declaration
  name: (_) @descriptor.type @kind.class
  body: (_) @scope)

(class_declaration
 (class_body
  [(method_definition
     name: (_) @descriptor.method @kind.method (#not-eq? @descriptor.method "constructor")
     body: (_) @local)]))

(class_declaration
 (class_body
  [(method_definition
     name: (_) @descriptor.method @kind.constructor (#eq? @descriptor.method "constructor")
     body: (_) @local)]))

[(if_statement) (while_statement) (for_statement) (do_statement) (call_expression)] @local
