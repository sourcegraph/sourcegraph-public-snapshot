use once_cell::sync::OnceCell;
use scip_macros::include_scip_query;
use scip_treesitter_languages::parsers::BundledParser;
use tree_sitter::{Language, Parser, Query};

pub struct TagConfiguration {
    pub language: Language,
    pub query: Query,
}

pub fn c() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::C.get_language();
        let query = include_scip_query!("c", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn javascript() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Javascript.get_language();
        let query = include_scip_query!("javascript", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn kotlin() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Kotlin.get_language();
        let query = include_scip_query!("kotlin", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn ruby() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Ruby.get_language();
        let query = include_scip_query!("ruby", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn python() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Python.get_language();
        let query = include_scip_query!("python", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn cpp() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Cpp.get_language();
        let query = include_scip_query!("cpp", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn typescript() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Typescript.get_language();
        let query = include_scip_query!("typescript", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn scala() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Scala.get_language();
        let query = include_scip_query!("scala", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn c_sharp() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::C_Sharp.get_language();
        let query = include_scip_query!("c_sharp", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn java() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Java.get_language();
        let query = include_scip_query!("java", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn rust() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Rust.get_language();
        let query = include_scip_query!("rust", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn go() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Go.get_language();
        let query = include_scip_query!("go", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub fn zig() -> &'static TagConfiguration {
    static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

    INSTANCE.get_or_init(|| {
        let language = BundledParser::Zig.get_language();
        let query = include_scip_query!("zig", "scip-tags");

        let mut parser = Parser::new();
        parser.set_language(language).unwrap();

        TagConfiguration {
            language,
            query: Query::new(language, query).unwrap(),
        }
    })
}

pub struct LocalConfiguration {
    pub language: Language,
    pub query: Query,
    pub parser: Parser,
}

fn go_locals() -> LocalConfiguration {
    let language = BundledParser::Go.get_language();
    let query = include_scip_query!("go", "scip-locals");

    let mut parser = Parser::new();
    parser.set_language(language).unwrap();

    LocalConfiguration {
        language,
        parser,
        query: Query::new(language, query).unwrap(),
    }
}

fn perl_locals() -> LocalConfiguration {
    let language = BundledParser::Perl.get_language();
    let query = include_scip_query!("perl", "scip-locals");

    let mut parser = Parser::new();
    parser.set_language(language).unwrap();

    LocalConfiguration {
        language,
        parser,
        query: Query::new(language, query).unwrap(),
    }
}

pub fn get_tag_configuration(parser: &BundledParser) -> Option<&'static TagConfiguration> {
    match parser {
        BundledParser::C => Some(c()),
        BundledParser::Javascript => Some(javascript()),
        BundledParser::Kotlin => Some(kotlin()),
        BundledParser::Ruby => Some(ruby()),
        BundledParser::Python => Some(python()),
        BundledParser::Cpp => Some(cpp()),
        BundledParser::Typescript => Some(typescript()),
        BundledParser::Scala => Some(scala()),
        BundledParser::C_Sharp => Some(c_sharp()),
        BundledParser::Java => Some(java()),
        BundledParser::Rust => Some(rust()),
        BundledParser::Go => Some(go()),
        BundledParser::Zig => Some(zig()),
        _ => None,
    }
}

pub fn get_local_configuration(parser: BundledParser) -> Option<LocalConfiguration> {
    match parser {
        BundledParser::Go => Some(go_locals()),
        BundledParser::Perl => Some(perl_locals()),
        _ => None,
    }
}
