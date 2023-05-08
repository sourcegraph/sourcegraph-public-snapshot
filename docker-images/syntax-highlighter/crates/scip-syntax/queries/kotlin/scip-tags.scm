(source_file
    (package_header
        (identifier)
        @descriptor.namespace
    )
) @scope

(function_declaration (simple_identifier) @descriptor.method) @local
(class_declaration (type_identifier) @descriptor.type) @local
(object_declaration (type_identifier) @descriptor.type) @local
(property_declaration (variable_declaration (simple_identifier) @descriptor.term))
