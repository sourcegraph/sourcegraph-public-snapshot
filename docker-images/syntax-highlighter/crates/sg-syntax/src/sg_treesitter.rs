use std::collections::VecDeque;
use std::fmt::Write as _; // import without risk of name clashing

use protobuf::Message;
use rocket::serde::json::{serde_json::json, Value as JsonValue};
use scip::types::{Document, Occurrence, SyntaxKind};
use scip_treesitter_languages::highlights::{
    get_highlighting_configuration, get_syntax_kind_for_hl,
};
use tree_sitter::Parser;
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
    mut include_locals: bool,
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

    // TODO: Don't let this happen in the final build haha
    include_locals = true;
    if include_locals {
        let parser = scip_treesitter_languages::parsers::BundledParser::get_parser(filetype);
        if let Some(parser) = parser {
            // TODO: Could probably write this in a much better way.
            let mut local_occs = scip_syntax::get_locals(parser, code.as_bytes())
                .unwrap_or(Ok(vec![]))
                .unwrap_or(vec![]);

            // Get ranges in top-to-bottom order
            local_occs.sort_by_key(|a| (-a.range[0], -a.range[1]));

            let mut next_doc_idx = 0;
            while let Some(local) = local_occs.pop() {
                let x = match doc
                    .occurrences
                    .iter_mut()
                    .enumerate()
                    .skip(next_doc_idx)
                    .find(|(_, occ)| {
                        // TODO: Actually compare the languages
                        occ.range[0] == local.range[0] && occ.range[1] == local.range[1]
                    }) {
                    Some(found) => found,
                    None => continue,
                };

                next_doc_idx = x.0;
                let occ = x.1;

                // Update occurrence with new information from locals
                occ.symbol = local.symbol;
                occ.symbol_roles = local.symbol_roles;
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

#[derive(Debug, PartialEq, Eq)]
pub struct PackedRange {
    pub start_line: i32,
    pub start_col: i32,
    pub end_line: i32,
    pub end_col: i32,
}

impl PackedRange {
    pub fn from_vec(v: &[i32]) -> Self {
        match v.len() {
            3 => Self {
                start_line: v[0],
                start_col: v[1],
                end_line: v[0],
                end_col: v[2],
            },
            4 => Self {
                start_line: v[0],
                start_col: v[1],
                end_line: v[2],
                end_col: v[3],
            },
            _ => {
                panic!("Unexpected vector length: {:?}", v);
            }
        }
    }
}

impl PartialOrd for PackedRange {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        (self.start_line, self.end_line, self.start_col).partial_cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
    }
}

impl Ord for PackedRange {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        (self.start_line, self.end_line, self.start_col).cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
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

pub fn dump_document(doc: &Document, source: &str) -> String {
    dump_document_range(doc, source, &None)
}

pub struct FileRange {
    pub start: usize,
    pub end: usize,
}

pub fn dump_document_range(doc: &Document, source: &str, file_range: &Option<FileRange>) -> String {
    let mut occurrences = doc.occurrences.clone();
    occurrences.sort_by_key(|o| PackedRange::from_vec(&o.range));
    let mut occurrences = VecDeque::from(occurrences);

    let mut result = String::new();

    let line_iterator: Box<dyn Iterator<Item = (usize, &str)>> = match file_range {
        Some(range) => Box::new(
            source
                .lines()
                .enumerate()
                .skip(range.start - 1)
                .take(range.end - range.start + 1),
        ),
        None => Box::new(source.lines().enumerate()),
    };

    for (idx, line) in line_iterator {
        result += "  ";
        result += &line.replace('\t', " ");
        result += "\n";

        while let Some(occ) = occurrences.pop_front() {
            if occ.syntax_kind.enum_value_or_default() == SyntaxKind::UnspecifiedSyntaxKind {
                continue;
            }

            let range = PackedRange::from_vec(&occ.range);
            let is_single_line = range.start_line == range.end_line;
            let end_col = if is_single_line {
                range.end_col
            } else {
                line.len() as i32
            };

            match range.start_line.cmp(&(idx as i32)) {
                std::cmp::Ordering::Less => continue,
                std::cmp::Ordering::Greater => {
                    occurrences.push_front(occ);
                    break;
                }
                std::cmp::Ordering::Equal => {
                    let length = (end_col - range.start_col) as usize;
                    let multiline_suffix = if is_single_line {
                        "".to_string()
                    } else {
                        format!(
                            " {}:{}..{}:{}",
                            range.start_line, range.start_col, range.end_line, range.end_col
                        )
                    };
                    let symbol_suffix = if occ.symbol.is_empty() {
                        "".to_owned()
                    } else {
                        format!(" {}", occ.symbol)
                    };
                    let _ = writeln!(
                        result,
                        "//{}{} {:?}{multiline_suffix} {}",
                        " ".repeat(range.start_col as usize),
                        "^".repeat(length),
                        occ.syntax_kind,
                        symbol_suffix,
                    );
                }
            }
        }
    }

    result
}

#[cfg(test)]
mod test {
    use std::{
        fs::{read_dir, File},
        io::Read,
    };

    use super::*;
    use crate::determine_filetype;

    #[test]
    fn test_highlights_one_comment() -> Result<(), Error> {
        let src = "// Hello World";
        let document = index_language("go", src, false)?;
        insta::assert_snapshot!(dump_document(&document, src));

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
        insta::assert_snapshot!(dump_document(&document, src));

        Ok(())
    }

    #[test]
    fn test_highlight_csharp_file() -> Result<(), Error> {
        let src = "using System;";
        let document = index_language("c_sharp", src, false)?;
        insta::assert_snapshot!(dump_document(&document, src));

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

            let indexed = index_language(filetype, &contents, false);
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
                    dump_document(&document, &contents)
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
