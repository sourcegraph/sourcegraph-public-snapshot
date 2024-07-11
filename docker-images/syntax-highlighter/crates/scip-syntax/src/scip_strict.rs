use std::borrow::Cow;

mod context_error;
mod format;
mod parse;

#[derive(Debug, PartialEq, Eq, Hash)]
pub enum Symbol<'a> {
    Local { local_id: &'a str },
    NonLocal(NonLocalSymbol<'a>),
}

impl Symbol<'_> {
    pub fn parse(raw: &str) -> Result<Symbol, String> {
        parse::parse_symbol(raw)
    }

    pub fn is_local(&self) -> bool {
        matches!(self, Symbol::Local { .. })
    }
}

#[derive(Debug, PartialEq, Eq, Hash)]
pub struct NonLocalSymbol<'a> {
    pub scheme: Scheme<'a>,
    pub package: Package<'a>,
    pub descriptors: Vec<Descriptor<'a>>,
}

#[derive(Debug, PartialEq, Eq, Hash, Default)]
pub struct Scheme<'a>(Cow<'a, str>);

impl Scheme<'_> {
    pub fn new(s: &str) -> Scheme {
        Scheme(s.into())
    }
}

#[derive(Debug, PartialEq, Eq, Hash)]
pub struct Package<'a> {
    manager: Cow<'a, str>,
    package_name: Cow<'a, str>,
    version: Cow<'a, str>,
}

impl Default for Package<'_> {
    fn default() -> Self {
        Self::new(None, None, None)
    }
}

impl Package<'_> {
    pub fn new<'a>(
        manager: Option<&'a str>,
        package_name: Option<&'a str>,
        version: Option<&'a str>,
    ) -> Package<'a> {
        Package {
            manager: manager.unwrap_or(".").into(),
            package_name: package_name.unwrap_or(".").into(),
            version: version.unwrap_or(".").into(),
        }
    }
    pub fn manager(&self) -> Option<&str> {
        let manager = self.manager.as_ref();
        if manager == "." {
            None
        } else {
            Some(manager)
        }
    }
    pub fn package_name(&self) -> Option<&str> {
        let package_name = self.package_name.as_ref();
        if package_name == "." {
            None
        } else {
            Some(package_name)
        }
    }
    pub fn version(&self) -> Option<&str> {
        let version = self.version.as_ref();
        if version == "." {
            None
        } else {
            Some(version)
        }
    }
}

#[derive(Debug, PartialEq, Eq, Hash)]
pub enum Descriptor<'a> {
    Namespace(Cow<'a, str>),
    Type(Cow<'a, str>),
    Term(Cow<'a, str>),
    Meta(Cow<'a, str>),
    Macro(Cow<'a, str>),
    Method {
        name: Cow<'a, str>,
        disambiguator: Option<&'a str>,
    },
    TypeParameter(Cow<'a, str>),
    Parameter(Cow<'a, str>),
}
