use std::collections::HashMap;

use once_cell::sync::OnceCell;
use regex::Regex;
use scip::types::Descriptor;
use scip_macros::include_scip_query;
use scip_treesitter_languages::parsers::BundledParser;
use tree_sitter::{Language, Parser, Query};

pub struct Transform {
    pattern: Regex,
    replace: String,
}

pub struct TagConfiguration {
    language: Language,
    pub query: Query,

    // Handles #transform! predicates in queries
    transforms: HashMap<usize, Vec<Transform>>,
}

impl TagConfiguration {
    fn new(language: Language, query: Query) -> Self {
        let mut transforms = HashMap::new();

        for index in 0..query.pattern_count() {
            let predicate = query.general_predicates(index);

            if !predicate.is_empty() {
                let pattern_transforms = predicate
                    .iter()
                    .filter_map(|pred| match pred.operator.as_ref() {
                        "transform!" => {
                            let args = &pred.args;
                            if args.len() != 2 {
                                panic!("bad transform!??!");
                            }

                            let pattern = {
                                let replace_str = match &args[0] {
                                    tree_sitter::QueryPredicateArg::String(str) => str,
                                    _ => panic!("pattern for #transform! should be a string"),
                                };

                                Regex::new(replace_str)
                                    .expect("pattern for #transform! should be a valid regex")
                            };

                            let replace = match &args[1] {
                                tree_sitter::QueryPredicateArg::String(str) => str.to_string(),
                                _ => panic!("replace to #transform! should be a string"),
                            };

                            Some(Transform { pattern, replace })
                        }
                        _ => None,
                    })
                    .collect::<Vec<_>>();

                transforms.insert(index, pattern_transforms);
            }
        }

        Self {
            language,
            query,
            transforms,
        }
    }

    pub fn get_parser(&self) -> Parser {
        let mut parser = Parser::new();
        parser.set_language(self.language).expect("to get a parser");
        parser
    }

    pub fn transform(&self, index: usize, captured: &Descriptor) -> Option<Vec<Descriptor>> {
        self.transforms.get(&index).map(|transforms| {
            transforms
                .iter()
                .map(|t| Descriptor {
                    name: t
                        .pattern
                        .replace_all(&captured.name, &t.replace)
                        .to_string(),
                    suffix: captured.suffix,
                    disambiguator: captured.disambiguator.clone(),
                    ..Default::default()
                })
                .collect::<Vec<_>>()
        })
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
                    let query = Query::new(language, query).expect("to parse query");

                    TagConfiguration::new(language, query)
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
