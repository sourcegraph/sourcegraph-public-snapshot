use anyhow::Result;
use paste::paste;
use protobuf::Message;
use rocket::serde::json::{serde_json::json, Value as JsonValue};
use scip::types::{Document, Occurrence, SyntaxKind};
use std::collections::HashMap;
use tree_sitter_all_languages::BundledParser;
use tree_sitter_highlight::{
    Error, Highlight, HighlightConfiguration, HighlightEvent, Highlighter as TSHighlighter,
};

use crate::highlighting::SourcegraphQuery;
use crate::range::PackedRange;

macro_rules! include_scip_query {
    ($lang: expr, $query: literal) => {
        include_str!(concat!(
            env!("CARGO_MANIFEST_DIR"),
            "/queries/",
            $lang,
            "/",
            $query,
            ".scm"
        ))
    };
}
pub(crate) use include_scip_query;

#[rustfmt::skip]
// Table of (@CaptureGroup, SyntaxKind) mapping.
//
// Any capture defined in a query will be mapped to the following SyntaxKind via the highlighter.
//
// To extend what types of captures are included, simply add a line below that takes a particular
// match group that you're interested in and map it to a new SyntaxKind.
//
// We can also define our own new capture types that we want to use and add to queries to provide
// particular highlights if necessary.
//
// (I can also add per-language mappings for these if we want, but you could also just do that with
//  unique match groups. For example `@rust-bracket`, or similar. That doesn't need any
//  particularly new rust code to be written. You can just modify queries for that)
const MATCHES_TO_SYNTAX_KINDS: &[(&str, SyntaxKind)] = &[
    ("boolean",                 SyntaxKind::BooleanLiteral),
    ("character",               SyntaxKind::CharacterLiteral),
    ("comment",                 SyntaxKind::Comment),
    ("conditional",             SyntaxKind::IdentifierKeyword),
    ("constant",                SyntaxKind::IdentifierConstant),
    ("identifier.constant",     SyntaxKind::IdentifierConstant),
    ("constant.builtin",        SyntaxKind::IdentifierBuiltin),
    ("constant.null",           SyntaxKind::IdentifierNull),
    ("float",                   SyntaxKind::NumericLiteral),
    ("function",                SyntaxKind::IdentifierFunction),
    ("method",                  SyntaxKind::IdentifierFunction),
    ("identifier.function",     SyntaxKind::IdentifierFunction),
    ("function.builtin",        SyntaxKind::IdentifierBuiltin),
    ("identifier.builtin",      SyntaxKind::IdentifierBuiltin),
    ("identifier",              SyntaxKind::Identifier),
    ("identifier.attribute",    SyntaxKind::IdentifierAttribute),
    ("tag.attribute",           SyntaxKind::TagAttribute),
    ("include",                 SyntaxKind::IdentifierKeyword),
    ("keyword",                 SyntaxKind::IdentifierKeyword),
    ("keyword.function",        SyntaxKind::IdentifierKeyword),
    ("keyword.return",          SyntaxKind::IdentifierKeyword),
    ("number",                  SyntaxKind::NumericLiteral),
    ("operator",                SyntaxKind::IdentifierOperator),
    ("identifier.operator",     SyntaxKind::IdentifierOperator),
    ("property",                SyntaxKind::Identifier),
    ("punctuation",             SyntaxKind::UnspecifiedSyntaxKind),
    ("punctuation.bracket",     SyntaxKind::UnspecifiedSyntaxKind),
    ("punctuation.delimiter",   SyntaxKind::PunctuationDelimiter),
    ("string",                  SyntaxKind::StringLiteral),
    ("string.special",          SyntaxKind::StringLiteral),
    ("string.escape",           SyntaxKind::StringLiteralEscape),
    ("tag",                     SyntaxKind::UnspecifiedSyntaxKind),
    ("type",                    SyntaxKind::IdentifierType),
    ("identifier.type",         SyntaxKind::IdentifierType),
    ("type.builtin",            SyntaxKind::IdentifierBuiltinType),
    ("regex.delimiter",         SyntaxKind::RegexDelimiter),
    ("regex.join",              SyntaxKind::RegexJoin),
    ("regex.escape",            SyntaxKind::RegexEscape),
    ("regex.repeated",          SyntaxKind::RegexRepeated),
    ("regex.wildcard",          SyntaxKind::RegexWildcard),
    ("identifier",              SyntaxKind::Identifier),
    ("variable",                SyntaxKind::Identifier),
    ("identifier.builtin",      SyntaxKind::IdentifierBuiltin),
    ("variable.builtin",        SyntaxKind::IdentifierBuiltin),
    ("identifier.parameter",    SyntaxKind::IdentifierParameter),
    ("variable.parameter",      SyntaxKind::IdentifierParameter),
    ("identifier.module",       SyntaxKind::IdentifierModule),
    ("variable.module",         SyntaxKind::IdentifierModule),
];

/// Add a language highlight configuration to the CONFIGURATIONS global.
///
/// This makes it so you don't have to understand how configurations are added,
/// just add the name of filetype that you want.
macro_rules! create_configurations {
    ( $(($name:tt,$dirname:literal)),* ) => {{
        use tree_sitter_all_languages::BundledParser;

        let mut m = HashMap::new();
        let highlight_names = MATCHES_TO_SYNTAX_KINDS.iter().map(|hl| hl.0).collect::<Vec<&str>>();

        $(
            {
                // Create HighlightConfiguration language
                let mut lang = HighlightConfiguration::new(
                    paste! { BundledParser::$name.get_language() },
                    include_scip_query!($dirname, "highlights"),
                    include_scip_query!($dirname, "injections"),
                    include_scip_query!($dirname, "locals"),
                ).expect(stringify!("parser for '{}' must be compiled", $name));

                // Associate highlights with configuration
                lang.configure(&highlight_names);

                // Insert into configurations, so we only create once at startup.
                m.insert(BundledParser::$name, lang);
            }
        )*

        {
            let highlights = vec![
                include_scip_query!("typescript", "highlights"),
                include_scip_query!("javascript", "highlights"),
            ];
            let mut lang = HighlightConfiguration::new(
                BundledParser::Typescript.get_language(),
                &highlights.join("\n"),
                include_scip_query!("typescript", "injections"),
                include_scip_query!("typescript", "locals"),
            ).expect("parser for 'typescript' must be compiled");
            lang.configure(&highlight_names);
            m.insert(BundledParser::Typescript, lang);
        }
        {
            let highlights = vec![
                include_scip_query!("tsx", "highlights"),
                include_scip_query!("typescript", "highlights"),
                include_scip_query!("javascript", "highlights"),
            ];
            let mut lang = HighlightConfiguration::new(
                BundledParser::Tsx.get_language(),
                &highlights.join("\n"),
                include_scip_query!("tsx", "injections"),
                include_scip_query!("tsx", "locals"),
            ).expect("parser for 'tsx' must be compiled");
            lang.configure(&highlight_names);
            m.insert(BundledParser::Tsx, lang);
        }

        m
    }}
}

lazy_static::lazy_static! {
    pub static ref CONFIGURATIONS: HashMap<BundledParser, HighlightConfiguration> = {
        // NOTE: typescript/tsx crates are included, even though not listed below.

        // You can add any new crate::parsers::Parser variants here.
        create_configurations!(
            (C, "c"),
            (Cpp, "cpp"),
            (C_Sharp, "c_sharp"),
            (Go, "go"),
            (Java, "java"),
            (Javascript, "javascript"),
            (Jsonnet, "jsonnet"),
            (Kotlin, "kotlin"),
            (Matlab, "matlab"),
            (Nickel, "nickel"),
            (Perl, "perl"),
            (Pod, "pod"),
            (Python, "python"),
            (Ruby, "ruby"),
            (Rust, "rust"),
            (Scala, "scala"),
            (Sql, "sql"),
            (Xlsg, "xlsg"),
            (Zig, "zig")
        )
    };
}

fn get_highlighting_configuration(filetype: &str) -> Option<&'static HighlightConfiguration> {
    BundledParser::get_parser(filetype).and_then(|parser| CONFIGURATIONS.get(&parser))
}

fn get_syntax_kind_for_hl(hl: Highlight) -> SyntaxKind {
    MATCHES_TO_SYNTAX_KINDS[hl.0].1
}

// Handle special cases where syntect language names don't match treesitter names.
pub(crate) fn treesitter_language(syntect_language: &str) -> &str {
    match syntect_language {
        "c++" => "cpp",
        _ => syntect_language,
    }
}

pub fn jsonify_err(e: impl ToString) -> JsonValue {
    json!({"error": e.to_string()})
}

// TODO(cleanup_lsif): Remove this when we remove /lsif endpoint
// Currently left unchanged
pub fn lsif_highlight(q: SourcegraphQuery) -> Result<JsonValue, JsonValue> {
    let filetype = q
        .filetype
        .ok_or_else(|| json!({"error": "Must pass a filetype for /lsif" }))?
        .to_lowercase();

    match index_language(&filetype, &q.code, false) {
        Ok(document) => {
            let encoded = document.write_to_bytes().map_err(jsonify_err)?;

            Ok(json!({"data": base64::encode(encoded), "plaintext": false}))
        }
        Err(Error::InvalidLanguage) => Err(json!({
            "error": format!("{} is not a valid filetype for treesitter", filetype)
        })),
        Err(err) => Err(jsonify_err(err)),
    }
}

pub fn index_language(filetype: &str, code: &str, include_locals: bool) -> Result<Document, Error> {
    match get_highlighting_configuration(filetype) {
        Some(lang_config) => {
            index_language_with_config(filetype, code, lang_config, include_locals)
        }
        None => Err(Error::InvalidLanguage),
    }
}

pub fn index_language_with_config(
    filetype: &str,
    code: &str,
    lang_config: &HighlightConfiguration,
    include_locals: bool,
) -> Result<Document, Error> {
    // Normalize string to be always only \n endings.
    //  We don't care that the byte offsets are "incorrect" now for this
    //  because we are using a line,col based approach
    let code = code.replace("\r\n", "\n");

    // TODO: We should automatically apply no highlights when we are
    // in an injected piece of code.
    //
    // Unfortunately, that information isn't currently available when
    // we are iterating in the higlighter.
    let mut highlighter = TSHighlighter::new();
    let highlights = highlighter.highlight(lang_config, code.as_bytes(), None, |l| {
        get_highlighting_configuration(l)
    })?;

    let mut emitter = ScipEmitter::new();
    let mut doc = emitter.render(highlights, &code, &get_syntax_kind_for_hl)?;
    doc.occurrences.sort_by_key(|a| (a.range[0], a.range[1]));

    if include_locals {
        let parser = tree_sitter_all_languages::BundledParser::get_parser(filetype);
        if let Some(parser) = parser {
            // TODO: Could probably write this in a much better way.
            let mut local_occs = crate::get_locals(parser, code.as_bytes()).unwrap_or_default();

            // Get ranges in reverse order, because we're going to pop off the back of the list.
            //  (that's why we're sorting the opposite way of the document occurrences above).
            local_occs.sort_by_key(|a| (-a.range[0], -a.range[1]));

            let mut next_doc_idx = 0;
            while let Some(local) = local_occs.pop() {
                // We *should* be able to assume that all these ranges are valid ranges
                // but for now we'll skip if they aren't.
                //
                // We can add some observability stuff to this later, and/or make
                // certain builds fail or something to test this out better (but
                // not have syntax highlighting completely fall apart from one
                // bad range)
                let local_range = match PackedRange::from_vec(&local.range) {
                    Some(range) => range,
                    None => continue,
                };

                let (matching_idx, matching_occ) = match doc
                    .occurrences
                    .iter_mut()
                    .enumerate()
                    .skip(next_doc_idx)
                    .find(|(_, occ)| local_range.eq_vec(&occ.range))
                {
                    Some(found) => found,
                    None => continue,
                };

                next_doc_idx = matching_idx;

                // Update occurrence with new information from locals
                matching_occ.symbol = local.symbol;
                matching_occ.symbol_roles = local.symbol_roles;
            }
        }
    }

    Ok(doc)
}

struct OffsetManager {
    source: String,
    offsets: Vec<usize>,
}

impl OffsetManager {
    fn new(s: &str) -> Result<Self, Error> {
        if s.is_empty() {
            // TODO: Make an error here
            // Error(
        }

        let source = s.to_string();

        let mut offsets = Vec::new();
        let mut pos = 0;
        for line in s.lines() {
            offsets.push(pos);
            // pos += line.chars().count() + 1;
            //
            // NOTE: This intentionally in bytes. The correct stuff is done in
            // self.line_and_col later
            pos += line.len() + 1;
        }

        Ok(Self { source, offsets })
    }

    fn line_and_col(&self, offset_byte: usize) -> (usize, usize) {
        // let offset_char = self.source.bytes
        let mut line = 0;
        for window in self.offsets.windows(2) {
            let curr = window[0];
            let next = window[1];
            if next > offset_byte {
                return (
                    line,
                    // Return the number of characters between the locations (which is the column)
                    self.source[curr..offset_byte].chars().count(),
                );
            }

            line += 1;
        }

        (
            line,
            // Return the number of characters between the locations (which is the column)
            self.source[*self.offsets.last().unwrap()..offset_byte]
                .chars()
                .count(),
        )
    }

    // range takes in start and end offsets and returns start/end line/column.
    fn range(&self, start_byte: usize, end_byte: usize) -> Vec<i32> {
        let start_pos = self.line_and_col(start_byte);
        let end_pos = self.line_and_col(end_byte);

        if start_pos.0 == end_pos.0 {
            vec![start_pos.0 as i32, start_pos.1 as i32, end_pos.1 as i32]
        } else {
            vec![
                start_pos.0 as i32,
                start_pos.1 as i32,
                end_pos.0 as i32,
                end_pos.1 as i32,
            ]
        }
    }
}

/// Converts a general-purpose syntax highlighting iterator into a sequence of lines of HTML.
pub(crate) struct ScipEmitter {}

/// Our version of `tree_sitter_highlight::HtmlRenderer`, which emits stuff as a table.
///
/// You can see the original version in the tree_sitter_highlight crate.
impl ScipEmitter {
    pub fn new() -> Self {
        ScipEmitter {}
    }

    pub fn render<F>(
        &mut self,
        highlighter: impl Iterator<Item = Result<HighlightEvent, Error>>,
        source: &str,
        _attribute_callback: &F,
    ) -> Result<Document, Error>
    where
        F: Fn(Highlight) -> SyntaxKind,
    {
        let mut doc = Document::new();

        let line_manager = OffsetManager::new(source)?;

        let mut highlights = vec![];
        for event in highlighter {
            match event? {
                HighlightEvent::HighlightStart(s) => highlights.push(s),
                HighlightEvent::HighlightEnd => {
                    highlights.pop();
                }

                // No highlights matched
                HighlightEvent::Source { .. } if highlights.is_empty() => {}

                // When a `start`->`end` has some highlights
                HighlightEvent::Source {
                    start: start_byte,
                    end: end_byte,
                } => {
                    let mut occurrence = Occurrence::new();
                    occurrence.range = line_manager.range(start_byte, end_byte);
                    occurrence.syntax_kind =
                        get_syntax_kind_for_hl(*highlights.last().unwrap()).into();

                    doc.occurrences.push(occurrence);
                }
            }
        }

        Ok(doc)
    }
}

#[cfg(test)]
mod test {
    use std::{
        fs::{read_dir, File},
        io::Read,
    };

    use super::*;
    use crate::highlighting::determine_filetype;
    use crate::snapshot::{self, dump_document_with_config};

    fn snapshot_treesitter_syntax_kinds(doc: &Document, source: &str) -> String {
        dump_document_with_config(
            doc,
            source,
            snapshot::SnapshotOptions {
                emit_syntax: snapshot::EmitSyntax::Highlighted,
                emit_symbol: snapshot::EmitSymbol::None,
                ..Default::default()
            },
        )
        .expect("dump document")
    }

    fn snapshot_treesitter_syntax_and_symbols(doc: &Document, source: &str) -> String {
        dump_document_with_config(
            doc,
            source,
            snapshot::SnapshotOptions {
                emit_syntax: snapshot::EmitSyntax::Highlighted,
                emit_symbol: snapshot::EmitSymbol::All,
                ..Default::default()
            },
        )
        .expect("dump document")
    }

    #[test]
    fn test_highlights_one_comment() -> Result<(), Error> {
        let src = "// Hello World";
        let document = index_language("go", src, false)?;
        insta::assert_snapshot!(snapshot_treesitter_syntax_kinds(&document, src));

        Ok(())
    }

    #[test]
    fn test_highlights_a_sql_query_within_go() -> Result<(), Error> {
        let src = r#"package main

const MySqlQuery = `
SELECT * FROM my_table
`
"#;

        let document = index_language("go", src, false)?;
        insta::assert_snapshot!(snapshot_treesitter_syntax_kinds(&document, src));

        Ok(())
    }

    #[test]
    fn test_highlight_csharp_file() -> Result<(), Error> {
        let src = "using System;";
        let document = index_language("c_sharp", src, false)?;
        insta::assert_snapshot!(snapshot_treesitter_syntax_kinds(&document, src));

        Ok(())
    }

    #[test]
    fn test_all_files() -> Result<(), std::io::Error> {
        let crate_root: std::path::PathBuf = std::env::var("CARGO_MANIFEST_DIR").unwrap().into();
        let input_dir = crate_root
            .join("src")
            .join("highlighting")
            .join("snapshots")
            .join("files");
        let dir = read_dir(&input_dir).unwrap();

        let mut failed_tests = vec![];
        let mut count = 0;
        for entry in dir {
            count += 1;
            let entry = entry?;
            let filepath = entry.path();
            let mut file = File::open(&filepath)?;
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;

            let filetype = &determine_filetype(&SourcegraphQuery {
                extension: filepath.extension().unwrap().to_str().unwrap().to_string(),
                filepath: filepath.to_str().unwrap().to_string(),
                filetype: None,
                line_length_limit: None,
                code: contents.clone(),
            });

            let indexed = index_language(filetype, &contents, true);
            if indexed.is_err() {
                // assert failure
                panic!("unknown filetype {:?}", filetype);
            }
            let document = indexed.unwrap();

            // TODO: I'm not sure if there's a better way to run the snapshots without
            // panicing and then catching, but this will do for now.
            match std::panic::catch_unwind(|| {
                insta::assert_snapshot!(
                    filepath.strip_prefix(&input_dir).unwrap().to_str().unwrap(),
                    snapshot_treesitter_syntax_kinds(&document, &contents)
                );
            }) {
                Ok(_) => println!("{}: OK", filepath.to_str().unwrap()),
                Err(err) => failed_tests.push(err),
            }
        }
        assert_ne!(count, 0, "Found at least one file");

        if !failed_tests.is_empty() {
            return Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("{} tests failed", failed_tests.len()),
            ));
        }

        Ok(())
    }

    #[test]
    fn test_files_with_locals() -> Result<(), std::io::Error> {
        let crate_root: std::path::PathBuf = std::env::var("CARGO_MANIFEST_DIR").unwrap().into();
        let input_dir = crate_root
            .join("src")
            .join("highlighting")
            .join("snapshots")
            .join("files-with-locals");
        let dir = read_dir(&input_dir).unwrap();

        let mut failed_tests = vec![];
        let mut count = 0;
        for entry in dir {
            count += 1;
            let entry = entry?;
            let filepath = entry.path();
            let mut file = File::open(&filepath)?;
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;

            let filetype = &determine_filetype(&SourcegraphQuery {
                extension: filepath.extension().unwrap().to_str().unwrap().to_string(),
                filepath: filepath.to_str().unwrap().to_string(),
                filetype: None,
                line_length_limit: None,
                code: contents.clone(),
            });

            let indexed = index_language(filetype, &contents, true);
            if indexed.is_err() {
                // assert failure
                panic!("unknown filetype {:?}", filetype);
            }
            let document = indexed.unwrap();

            // TODO: I'm not sure if there's a better way to run the snapshots without
            // panicing and then catching, but this will do for now.
            match std::panic::catch_unwind(|| {
                insta::assert_snapshot!(
                    filepath.strip_prefix(&input_dir).unwrap().to_str().unwrap(),
                    snapshot_treesitter_syntax_and_symbols(&document, &contents)
                );
            }) {
                Ok(_) => println!("{}: OK", filepath.to_str().unwrap()),
                Err(err) => failed_tests.push(err),
            }
        }
        assert_ne!(count, 0, "Found at least one file");

        if !failed_tests.is_empty() {
            return Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("{} tests failed", failed_tests.len()),
            ));
        }

        Ok(())
    }
}
