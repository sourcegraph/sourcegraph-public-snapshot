;;include javascript

(module name: (string (string_fragment) @descriptor.namespace @kind.module) body: (_) @scope)

(interface_declaration name: (_) @descriptor.type @kind.interface body: (_) @scope)
(interface_declaration
    (object_type
        [
            (method_signature (property_identifier) @descriptor.method @kind.method)
            (property_signature (property_identifier) @descriptor.term @kind.property)]))

(class_declaration
    (class_body
        [(public_field_definition name: (_) @descriptor.term @kind.property)]))


(enum_declaration name: (_) @descriptor.type @kind.enum body: (_) @scope)
(enum_declaration
    (enum_body
        (property_identifier) @descriptor.term @kind.property))
