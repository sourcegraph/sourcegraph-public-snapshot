; Make use of @local

(function_definition body: (_) @local)

(parameter_list) @local

(function_declarator declarator: (_) @descriptor.method @kind.function)

(pointer_declarator declarator: ((identifier) @descriptor.term @kind.variable))

(init_declarator
    declarator: (identifier) @descriptor.term @kind.variable
    value: (_))

(array_declarator
    declarator: (identifier) @descriptor.term @kind.variable
    size: (_))

(declaration
  (type_qualifier)?
  declarator: ((identifier) @descriptor.term @kind.variable))

;; Enums
(enum_specifier
  name: (_)?
  body: (enumerator_list
          (enumerator name: (_) @descriptor.term @kind.enummember))) ;; <-- capture enum kind

(enum_specifier
  name: (type_identifier) @descriptor.type @kind.enum ;; <-- capture enum kind
  body: (_))

;; fields (inside structs, or unions, ...)
(field_declaration_list
          (field_declaration declarator: [
            (pointer_declarator (field_identifier) @descriptor.term @kind.field)
            (field_identifier) @descriptor.term @kind.field]))

;; structs
(struct_specifier
  name: (type_identifier) @descriptor.type @kind.struct ;; <-- capture struct kidn
  body: (_)) @scope

;; typedefs
(type_definition
  type: (_)
  declarator: (type_identifier) @descriptor.term @kind.typealias)

;; macros
(preproc_def
  name: (identifier) @descriptor.term @kind.macro)

;; union
(union_specifier
  name: (type_identifier) @descriptor.type @kind.union
  body: (_)) @scope
