use std::{borrow::Cow, fmt};

use super::{parse, Descriptor, NonLocalSymbol, Package, Scheme, Symbol};

impl fmt::Display for Symbol<'_> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Symbol::NonLocal(non_local) => non_local.fmt(f),
            Symbol::Local { local_id } => write!(f, "local {}", local_id),
        }
    }
}

impl fmt::Display for NonLocalSymbol<'_> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} {} ", self.scheme, self.package)?;
        for descriptor in &self.descriptors {
            write!(f, "{}", descriptor)?;
        }
        Ok(())
    }
}

impl fmt::Display for Scheme<'_> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", escape_space_terminated(&self.0))
    }
}

impl fmt::Display for Package<'_> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(
            f,
            "{} {} {}",
            escape_space_terminated(&self.manager),
            escape_space_terminated(&self.package_name),
            escape_space_terminated(&self.version),
        )
    }
}

impl fmt::Display for Descriptor<'_> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Descriptor::Namespace(name) => write!(f, "{}/", escape_name(name)),
            Descriptor::Type(name) => write!(f, "{}#", escape_name(name)),
            Descriptor::Term(name) => write!(f, "{}.", escape_name(name)),
            Descriptor::Meta(name) => write!(f, "{}:", escape_name(name)),
            Descriptor::Macro(name) => write!(f, "{}!", escape_name(name)),
            Descriptor::Method {
                name,
                disambiguator,
            } => write!(
                f,
                "{}({}).",
                escape_name(name),
                disambiguator.unwrap_or_default()
            ),
            Descriptor::TypeParameter(name) => write!(f, "[{}]", escape_name(name)),
            Descriptor::Parameter(name) => write!(f, "({})", escape_name(name)),
        }
    }
}

fn escape_name<'a>(name: &'a Cow<'a, str>) -> Cow<'a, str> {
    if name.chars().all(parse::is_simple_identifier_char) {
        name.as_ref().into()
    } else {
        format!("`{}`", name.replace('`', "``")).into()
    }
}

fn escape_space_terminated<'a>(s: &'a Cow<'a, str>) -> Cow<'a, str> {
    if s.contains(' ') {
        s.replace(' ', "  ").into()
    } else {
        s.as_ref().into()
    }
}
