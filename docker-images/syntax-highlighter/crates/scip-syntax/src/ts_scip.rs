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
        "kind.constant" => Constant,
        "kind.field" => Field,
        "kind.function" => Function,
        "kind.interface" => Interface,
        "kind.method" => Method,
        "kind.package" => Package,
        "kind.struct" => Struct,
        "kind.typealias" => TypeAlias,
        _ => UnspecifiedKind,
    })
}

pub fn symbol_kind_to_ctags_kind(kind: &symbol_information::Kind) -> Option<&'static str> {
    use symbol_information::Kind::*;

    match kind {
        Constant => Some("constant"),
        Field => Some("field"),
        Function => Some("function"),
        Interface => Some("interface"),
        Method => Some("method"),
        Package => Some("package"),
        Struct => Some("struct"),
        TypeAlias => Some("typealias"),
        _ => None,
    }
}
