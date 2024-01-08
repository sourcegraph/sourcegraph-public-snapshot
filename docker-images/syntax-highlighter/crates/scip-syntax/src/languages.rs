use std::collections::HashMap;

use once_cell::sync::OnceCell;
use regex::Regex;
use scip::types::Descriptor;
use scip_macros::include_scip_query;
use scip_treesitter_languages::parsers::BundledParser;
use tree_sitter::{Language, Parser, Query};

#[derive(Debug)]
pub struct Transform {
    pattern: Regex,
    replace: String,
}

#[derive(Debug)]
pub struct NodeFilter {
    capture: u32,
    names: Vec<String>,
}

pub struct TagConfiguration {
    language: Language,
    query_text: String,
    pub tag_query: Query,
    pub sym_query: Query,

    // Handles #transform! predicates in queries
    transforms: HashMap<usize, Vec<Transform>>,

    // Handles #filter! predicates in queries
    filters: HashMap<usize, Vec<NodeFilter>>,
}

impl TagConfiguration {
    fn new(language: Language, tag_query: &str, sym_query: Option<String>) -> Self {
        let first_line = tag_query.lines().next();
        let query_text = match first_line {
            Some(line) if line.starts_with(";;include") => {
                let (_, included_lang) =
                    line.split_once(";;include").expect("must have ;; include");
                let included_lang = included_lang.trim();

                let parser = BundledParser::get_parser(included_lang).expect("valid language");
                let configuration = get_tag_configuration(parser).expect("valid config");

                format!("{}\n{}", configuration.query_text, tag_query)
            }
            _ => tag_query.to_string(),
        };

        let query = Query::new(language, &query_text).expect("to parse query");
        let sym_query = match sym_query {
            Some(text) => {
                let text = format!("{query_text}\n{text}");
                Query::new(language, &text).expect("to parse symbol query")
            }
            None => Query::new(language, &query_text).expect("to parse query"),
        };

        let mut transforms = HashMap::new();
        let mut filters = HashMap::new();

        for index in 0..query.pattern_count() {
            let predicate = query.general_predicates(index);

            if !predicate.is_empty() {
                // Collect #transform! predicates
                transforms.insert(
                    index,
                    predicate
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
                        .collect::<Vec<_>>(),
                );

                // Collect #filter! predicates
                filters.insert(
                    index,
                    predicate
                        .iter()
                        .filter_map(|pred| match pred.operator.as_ref() {
                            "filter!" => {
                                let args = &pred.args;

                                if args.len() < 2 {
                                    panic!("should have at least two arguments for filter");
                                }

                                let capture = match &args[0] {
                                    tree_sitter::QueryPredicateArg::Capture(capture) => *capture,
                                    _ => panic!("filter! first arg should be a  capture"),
                                };

                                let names = args[1..]
                                    .iter()
                                    .map(|arg| {
                                        let name = match arg {
                                            tree_sitter::QueryPredicateArg::String(name) => name,
                                            _ => panic!("filter! 1.. args should be string names"),
                                        };

                                        name.to_string()
                                    })
                                    .collect();

                                Some(NodeFilter { capture, names })
                            }
                            _ => None,
                        })
                        .collect::<Vec<_>>(),
                );
            }
        }

        Self {
            language,
            query_text,
            tag_query: query,
            sym_query,
            transforms,
            filters,
        }
    }

    pub fn get_parser(&self) -> Parser {
        let mut parser = Parser::new();
        parser.set_language(self.language).expect("to get a parser");
        parser
    }

    pub fn transform(&self, index: usize, captured: &Descriptor) -> Option<Vec<Descriptor>> {
        match self.transforms.get(&index) {
            Some(transforms) if !transforms.is_empty() => Some(
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
                    .collect::<Vec<_>>(),
            ),
            _ => None,
        }
    }

    pub fn is_filtered(&self, m: &tree_sitter::QueryMatch) -> bool {
        match self.filters.get(&m.pattern_index) {
            Some(filters) if !filters.is_empty() => filters.iter().any(|filter| {
                m.nodes_for_capture_index(filter.capture)
                    .any(|node| filter.names.iter().any(|name| name == node.kind()))
            }),
            _ => false,
        }
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
                    TagConfiguration::new(language, query, None)
                })
            }
        };
        ($name:tt, $parser:path, $file:tt, $symbol_file:tt) => {
            pub fn $name() -> &'static TagConfiguration {
                static INSTANCE: OnceCell<TagConfiguration> = OnceCell::new();

                INSTANCE.get_or_init(|| {
                    let language = $parser.get_language();
                    let query = include_scip_query!($file, "scip-tags");
                    let sym_query = include_scip_query!($file, "scip-references").to_string();
                    TagConfiguration::new(language, query, Some(sym_query))
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
    create_tags_configuration!(go, BundledParser::Go, "go", "go");
    create_tags_configuration!(zig, BundledParser::Zig, "zig");

    pub fn get_tag_configuration(parser: BundledParser) -> Option<&'static TagConfiguration> {
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
    create_locals_configuration!(matlab, BundledParser::Matlab, "matlab");
    create_locals_configuration!(java, BundledParser::Java, "java");

    pub fn get_local_configuration(parser: BundledParser) -> Option<&'static LocalConfiguration> {
        match parser {
            BundledParser::Go => Some(go()),
            BundledParser::Perl => Some(perl()),
            BundledParser::Matlab => Some(matlab()),
            BundledParser::Java => Some(java()),
            _ => None,
        }
    }
}

pub use locals::get_local_configuration;
pub use tags::get_tag_configuration;
