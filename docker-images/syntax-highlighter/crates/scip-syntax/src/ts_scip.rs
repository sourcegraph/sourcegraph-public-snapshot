use scip::types::{descriptor::Suffix, symbol_information, Descriptor};

pub fn capture_name_to_descriptor(capture: &str, name: String) -> Descriptor {
    Descriptor {
        suffix: match capture {
            "descriptor.method" => Suffix::Method,
            "descriptor.namespace" => Suffix::Namespace,
            "descriptor.type" => Suffix::Type,
            "descriptor.term" => Suffix::Term,

            // TODO: Should consider moving to result here.
            _ => Suffix::UnspecifiedSuffix,
        }
        .into(),
        name,
        ..Default::default()
    }
}

pub fn captures_to_kind(kind: &Option<&String>) -> symbol_information::Kind {
    use symbol_information::Kind::*;

    kind.map_or(UnspecifiedKind, |kind| match kind.as_str() {
        "kind.accessor" => Accessor,
        "kind.class" => Class,
        "kind.constant" => Constant,
        "kind.constructor" => Constructor,
        "kind.delegate" => Delegate,
        "kind.enum" => Enum,
        "kind.enummember" => EnumMember,
        "kind.event" => Event,
        "kind.field" => Field,
        "kind.function" => Function,
        "kind.getter" => Getter,
        "kind.interface" => Interface,
        "kind.method" => Method,
        "kind.methodalias" => MethodAlias,
        "kind.methodspec" => MethodSpecification,
        "kind.module" => Module,
        "kind.namespace" => Namespace,
        "kind.package" => Package,
        "kind.property" => Property,
        "kind.setter" => Setter,
        "kind.singletonmethod" => SingletonMethod,
        "kind.struct" => Struct,
        "kind.typealias" => TypeAlias,
        "kind.trait" => Trait,
        // "kind.implementation" => Implementation, TODO
        _ => UnspecifiedKind,
    })
}

pub fn symbol_kind_to_ctags_kind(kind: &symbol_information::Kind) -> Option<&'static str> {
    use symbol_information::Kind::*;

    match kind {
        Accessor => Some("accessor"),
        Class => Some("class"),
        Constant => Some("constant"),
        Constructor => Some("constructor"),
        Delegate => Some("delegate"),
        Enum => Some("enum"),
        EnumMember => Some("enumMember"),
        Event => Some("event"),
        Field => Some("field"),
        Function => Some("function"),
        Getter => Some("getter"),
        Interface => Some("interface"),
        Method => Some("method"),
        MethodAlias => Some("methodAlias"),
        MethodSpecification => Some("methodSpec"),
        Module => Some("module"),
        Namespace => Some("namespace"),
        Package => Some("package"),
        Property => Some("property"),
        Setter => Some("setter"),
        SingletonMethod => Some("singletonMethod"),
        Struct => Some("struct"),
        TypeAlias => Some("typeAlias"),
        Trait => Some("trait"),
        // Implementation => Some("implementation"), TODO
        _ => None,
    }
}
