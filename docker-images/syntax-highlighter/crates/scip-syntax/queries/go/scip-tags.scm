(source_file (package_clause (package_identifier) @descriptor.namespace @kind.package)) @scope

(function_declaration
 name: (identifier) @descriptor.method @kind.function) @enclosing

;; Function bodies are local
(function_declaration body: (block) @local)

(method_declaration
 receiver: (parameter_list
            (parameter_declaration
             type: [(pointer_type (type_identifier) @descriptor.type)
                    (type_identifier) @descriptor.type]))
 name: (field_identifier) @descriptor.method @enclosing
 body: (_) @local)

(type_declaration (type_spec name: (type_identifier) @descriptor.type)) @scope

;; For fields, we have nested struct definitions.
;;   To get the scope properly
((field_declaration_list
   (field_declaration
     name: (_) @descriptor.term
     type: (_) @_type) @enclosing)
 (#filter! @_type "interface_type" "struct_type"))

(field_declaration_list
  (field_declaration
    name: (_) @descriptor.type
    type: [(interface_type) (struct_type)] @scope))

(const_spec name: (_) @descriptor.term) @kind.constant @enclosing
(import_spec name: (_) @descriptor.term) @enclosing
(method_spec name: (_) @descriptor.method) @enclosing
(var_spec name: (_) @descriptor.term) @enclosing
