; TODO: Exclude functions in non-container blocks

(compilation_unit
    (package_clause
        (package_identifier)
        @descriptor.namespace
    )
) @scope

(compilation_unit (val_definition (identifier) @descriptor.term))
(compilation_unit (var_definition (identifier) @descriptor.term))
(template_body (val_definition (identifier) @descriptor.term))
(template_body (var_definition (identifier) @descriptor.term))

(function_definition (identifier) @descriptor.method)

(class_definition (identifier) @descriptor.type (template_body) @scope)
(object_definition (identifier) @descriptor.type (template_body) @scope)
(trait_definition (identifier) @descriptor.type (template_body) @scope)
