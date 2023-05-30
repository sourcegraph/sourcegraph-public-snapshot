use scip::types::{descriptor::Suffix, Descriptor};

pub fn capture_name_to_descriptor(capture: &str, name: String) -> Descriptor {
    Descriptor {
        suffix: match capture {
            "descriptor.method" => Suffix::Method,
            "descriptor.namespace" => Suffix::Namespace,
            "descriptor.type" => Suffix::Type,

            // TODO: Should consider moving to result here.
            _ => Suffix::UnspecifiedSuffix,
        }
        .into(),
        name,
        ..Default::default()
    }
}
