use once_cell::sync::OnceCell;
use scip_macros::include_scip_query;
use scip_treesitter_languages::parsers::BundledParser;
use tree_sitter::{Language, Parser, Query};

pub struct TagConfiguration {
    language: Language,
    pub query: Query,
}

impl TagConfiguration {
    pub fn get_parser(&self) -> Parser {
        let mut parser = Parser::new();
        parser.set_language(self.language).expect("to get a parser");
        parser
    }
}

pub struct LocalConfiguration {
    language: Language,
    pub query: Query,
}

impl LocalConfiguration {
    pub fn get_parser(&self) -> Parser {
        let mut parser = Parser::new();
        parser.set_language(self.language).expect("to get a parser");
        parser
    }
}

mod tags {
    use super::*;

    macro_rules! create_tags_configuration {
        ($name:tt, $parser:path, $file:tt) => {
            pub fn $name() -> &'static TagConfiguration {
                static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

                INSTANCE.get_or_init(|| {
                    let language = $parser.get_language();
                    let query = include_scip_query!($file, "scip-tags");

                    TagConfiguration {
                        language,
                        query: Query::new(language, query).unwrap(),
                    }
                })
            }
        };
    }

    create_tags_configuration!(c, BundledParser::C, "c");
    create_tags_configuration!(javascript, BundledParser::Javascript, "javascript");
    create_tags_configuration!(kotlin, BundledParser::Kotlin, "kotlin");
    create_tags_configuration!(ruby, BundledParser::Ruby, "ruby");
    create_tags_configuration!(python, BundledParser::Python, "python");
    create_tags_configuration!(cpp, BundledParser::Cpp, "cpp");
    create_tags_configuration!(typescript, BundledParser::Typescript, "typescript");
    create_tags_configuration!(scala, BundledParser::Scala, "scala");
    create_tags_configuration!(c_sharp, BundledParser::C_Sharp, "c_sharp");
    create_tags_configuration!(java, BundledParser::Java, "java");
    create_tags_configuration!(rust, BundledParser::Rust, "rust");
    create_tags_configuration!(go, BundledParser::Go, "go");
    create_tags_configuration!(zig, BundledParser::Zig, "zig");

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
}

mod locals {
    use super::*;

    macro_rules! create_locals_configuration {
        ($name:tt, $parser:path, $file:tt) => {
            pub fn $name() -> &'static LocalConfiguration {
                static INSTANCE: OnceCell<LocalConfiguration> = OnceCell::new();

                INSTANCE.get_or_init(|| {
                    let language = $parser.get_language();
                    let query = include_scip_query!($file, "scip-locals");

                    LocalConfiguration {
                        language,
                        query: Query::new(language, query).unwrap(),
                    }
                })
            }
        };
    }

    create_locals_configuration!(go, BundledParser::Go, "go");
    create_locals_configuration!(perl, BundledParser::Perl, "perl");

    pub fn get_local_configuration(parser: BundledParser) -> Option<&'static LocalConfiguration> {
        match parser {
            BundledParser::Go => Some(go()),
            BundledParser::Perl => Some(perl()),
            _ => None,
        }
    }
}

pub use locals::get_local_configuration;
pub use tags::get_tag_configuration;
