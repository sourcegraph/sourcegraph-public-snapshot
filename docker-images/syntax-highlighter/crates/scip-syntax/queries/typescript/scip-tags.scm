;;include javascript

(module name: (string (string_fragment) @descriptor.namespace) body: (_) @scope)

(interface_declaration name: (_) @descriptor.type body: (_) @scope)
(interface_declaration
    (object_type
        [
            (method_signature (property_identifier) @descriptor.method)
            (property_signature (property_identifier) @descriptor.term)]))

(class_declaration
    (class_body
        [(public_field_definition name: (_) @descriptor.term)]))


(enum_declaration name: (_) @descriptor.type body: (_) @scope)
(enum_declaration
    (enum_body
        (property_identifier) @descriptor.term))
