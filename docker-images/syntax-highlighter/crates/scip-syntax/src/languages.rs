use scip_macros::include_scip_query;
use tree_sitter::{Language, Parser, Query};

pub struct TagConfiguration {
    pub language: Language,
    pub query: Query,
    pub parser: Parser,
}

pub fn rust() -> TagConfiguration {
    let language = scip_treesitter_languages::rust();
    let query = include_scip_query!("rust", "scip-tags");

    let mut parser = Parser::new();
    parser.set_language(language).unwrap();

    TagConfiguration {
        language,
        parser,
        query: Query::new(language, query).unwrap(),
    }
}

pub fn go() -> TagConfiguration {
    let language = scip_treesitter_languages::go();
    let query = include_scip_query!("go", "scip-tags");

    let mut parser = Parser::new();
    parser.set_language(language).unwrap();

    TagConfiguration {
        language,
        parser,
        query: Query::new(language, query).unwrap(),
    }
}

pub struct LocalConfiguration {
    pub language: Language,
    pub query: Query,
    pub parser: Parser,
}

pub fn go_locals() -> LocalConfiguration {
    let language = scip_treesitter_languages::go();
    let query = include_scip_query!("go", "scip-locals");

    let mut parser = Parser::new();
    parser.set_language(language).unwrap();

    LocalConfiguration {
        language,
        parser,
        query: Query::new(language, query).unwrap(),
    }
}
