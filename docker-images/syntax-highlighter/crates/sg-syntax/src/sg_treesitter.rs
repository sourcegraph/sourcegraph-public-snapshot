use anyhow::Result;
use protobuf::Message;
use rocket::serde::json::{serde_json::json, Value as JsonValue};
use scip::types::{Document, Occurrence, SyntaxKind};
use scip_treesitter::types::PackedRange;
use scip_treesitter_languages::highlights::{
    get_highlighting_configuration, get_syntax_kind_for_hl,
};
use tree_sitter_highlight::{
    Error, Highlight, HighlightConfiguration, HighlightEvent, Highlighter as TSHighlighter,
};

use crate::SourcegraphQuery;

// Handle special cases where syntect language names don't match treesitter names.
pub fn treesitter_language(syntect_language: &str) -> &str {
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
        let parser = scip_treesitter_languages::parsers::BundledParser::get_parser(filetype);
        if let Some(parser) = parser {
            // TODO: Could probably write this in a much better way.
            let mut local_occs = scip_syntax::get_locals(parser, code.as_bytes())
                .unwrap_or(Ok(vec![]))
                .unwrap_or_default();

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
pub struct ScipEmitter {}

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

    use scip_treesitter::snapshot::dump_document_with_config;

    use super::*;
    use crate::determine_filetype;

    fn snapshot_treesitter_syntax_kinds(doc: &Document, source: &str) -> String {
        dump_document_with_config(
            doc,
            source,
            scip_treesitter::snapshot::SnapshotOptions {
                emit_syntax: scip_treesitter::snapshot::EmitSyntax::Highlighted,
                emit_symbol: scip_treesitter::snapshot::EmitSymbol::None,
                ..Default::default()
            },
        )
        .expect("dump document")
    }

    fn snapshot_treesitter_syntax_and_symbols(doc: &Document, source: &str) -> String {
        dump_document_with_config(
            doc,
            source,
            scip_treesitter::snapshot::SnapshotOptions {
                emit_syntax: scip_treesitter::snapshot::EmitSyntax::Highlighted,
                emit_symbol: scip_treesitter::snapshot::EmitSymbol::All,
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
        let input_dir = crate_root.join("src").join("snapshots").join("files");
        let dir = read_dir(&input_dir).unwrap();

        let mut failed_tests = vec![];
        for entry in dir {
            let entry = entry?;
            let filepath = entry.path();
            let mut file = File::open(&filepath)?;
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;

            let filetype = &determine_filetype(&SourcegraphQuery {
                extension: filepath.extension().unwrap().to_str().unwrap().to_string(),
                filepath: filepath.to_str().unwrap().to_string(),
                filetype: None,
                css: false,
                line_length_limit: None,
                theme: "".to_string(),
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
            .join("snapshots")
            .join("files-with-locals");
        let dir = read_dir(&input_dir).unwrap();

        let mut failed_tests = vec![];
        for entry in dir {
            let entry = entry?;
            let filepath = entry.path();
            let mut file = File::open(&filepath)?;
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;

            let filetype = &determine_filetype(&SourcegraphQuery {
                extension: filepath.extension().unwrap().to_str().unwrap().to_string(),
                filepath: filepath.to_str().unwrap().to_string(),
                filetype: None,
                css: false,
                line_length_limit: None,
                theme: "".to_string(),
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

        if !failed_tests.is_empty() {
            return Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("{} tests failed", failed_tests.len()),
            ));
        }

        Ok(())
    }
}
