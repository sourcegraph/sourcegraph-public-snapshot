(source_file (package_clause (package_identifier) @descriptor.namespace) @scope)

(function_declaration
 name: (identifier) @descriptor.method)

(method_declaration
 receiver: (parameter_list
            (parameter_declaration
             type: (pointer_type
                     (type_identifier) @descriptor.type)))
 name: (field_identifier) @descriptor.method)

(method_declaration
  receiver: (parameter_list
               (parameter_declaration type: (type_identifier) @descriptor.type))
  name: (field_identifier) @descriptor.method)

(type_declaration (type_spec name: (type_identifier) @descriptor.type))
