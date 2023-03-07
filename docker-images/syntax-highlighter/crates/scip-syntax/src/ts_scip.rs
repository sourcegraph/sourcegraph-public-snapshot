use scip::types::{descriptor::Suffix, Descriptor};

pub fn capture_name_to_descriptor(capture: &str, name: String) -> Descriptor {
    Descriptor {
        suffix: match capture {
            "descriptor.method" => Suffix::Method,
            "descriptor.namespace" => Suffix::Namespace,
            "descriptor.type" => Suffix::Type,
            _ => unimplemented!("Missing {}", name),
        }
        .into(),
        name,
        ..Default::default()
    }
}
