; Make use of @local

(translation_unit (declaration (init_declarator declarator: (_) @descriptor.term)))

(enum_specifier name: (_) @descriptor.type body: (enumerator_list (enumerator name: (_) @descriptor.term)) @descriptor.scope)

(field_declaration declarator: [
    (pointer_declarator (field_identifier) @descriptor.term)
    (field_identifier) @descriptor.term
])
(function_definition (function_declarator declarator: (_) @descriptor.method))
