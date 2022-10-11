use paste::paste;
use protobuf::Message;
use std::collections::{HashMap, VecDeque};
use std::fmt::{Write as _}; // import without risk of name clashing
use tree_sitter_highlight::Error;
use tree_sitter_highlight::{Highlight, HighlightEvent};

use rocket::serde::json::serde_json::json;
use rocket::serde::json::Value as JsonValue;
use tree_sitter_highlight::{HighlightConfiguration, Highlighter as TSHighlighter};

use crate::SourcegraphQuery;
use scip::types::{Document, Occurrence, SyntaxKind};
use sg_macros::include_project_file_optional;

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
    ("attribute",               SyntaxKind::UnspecifiedSyntaxKind),
    ("boolean",                 SyntaxKind::BooleanLiteral),
    ("comment",                 SyntaxKind::Comment),
    ("conditional",             SyntaxKind::IdentifierKeyword),
    ("constant",                SyntaxKind::IdentifierConstant),
    ("constant.builtin",        SyntaxKind::IdentifierBuiltin),
    ("constant.null",           SyntaxKind::IdentifierNull),
    ("float",                   SyntaxKind::NumericLiteral),
    ("function",                SyntaxKind::IdentifierFunctionDefinition),
    ("function.builtin",        SyntaxKind::IdentifierBuiltin),
    ("identifier",              SyntaxKind::Identifier),
    ("identifier.function",     SyntaxKind::IdentifierFunctionDefinition),
    ("include",                 SyntaxKind::IdentifierKeyword),
    ("keyword",                 SyntaxKind::IdentifierKeyword),
    ("keyword.function",        SyntaxKind::IdentifierKeyword),
    ("keyword.return",          SyntaxKind::IdentifierKeyword),
    ("method",                  SyntaxKind::IdentifierFunction),
    ("number",                  SyntaxKind::NumericLiteral),
    ("operator",                SyntaxKind::IdentifierOperator),
    ("property",                SyntaxKind::Identifier),
    ("punctuation",             SyntaxKind::UnspecifiedSyntaxKind),
    ("punctuation.bracket",     SyntaxKind::UnspecifiedSyntaxKind),
    ("punctuation.delimiter",   SyntaxKind::PunctuationDelimiter),
    ("string",                  SyntaxKind::StringLiteral),
    ("string.special",          SyntaxKind::StringLiteral),
    ("tag",                     SyntaxKind::UnspecifiedSyntaxKind),
    ("type",                    SyntaxKind::IdentifierType),
    ("type.builtin",            SyntaxKind::IdentifierType),
    ("variable",                SyntaxKind::Identifier),
    ("variable.builtin",        SyntaxKind::UnspecifiedSyntaxKind),
    ("variable.parameter",      SyntaxKind::IdentifierParameter),
    ("variable.module",         SyntaxKind::IdentifierModule),
];

/// Maps a highlight to a syntax kind.
/// This only works if you've correctly used the highlight_names from MATCHES_TO_SYNTAX_KINDS
fn get_syntax_kind_for_hl(hl: Highlight) -> SyntaxKind {
    MATCHES_TO_SYNTAX_KINDS[hl.0].1
}

/// Add a language highlight configuration to the CONFIGURATIONS global.
///
/// This makes it so you don't have to understand how configurations are added,
/// just add the name of filetype that you want.
macro_rules! create_configurations {
    ( $($name: tt),* ) => {{
        let mut m = HashMap::new();
        let highlight_names = MATCHES_TO_SYNTAX_KINDS.iter().map(|hl| hl.0).collect::<Vec<&str>>();

        $(
            {
                // Create HighlightConfiguration language
                let mut lang = HighlightConfiguration::new(
                    paste! { [<tree_sitter_ $name>]::language() },
                    include_project_file_optional!("queries/", $name, "/highlights.scm"),
                    include_project_file_optional!("queries/", $name, "/injections.scm"),
                    include_project_file_optional!("queries/", $name, "/locals.scm"),
                ).expect(stringify!("parser for '{}' must be compiled", $name));

                // Associate highlights with configuration
                lang.configure(&highlight_names);

                // Insert into configurations, so we only create once at startup.
                m.insert(stringify!($name), lang);
            }
        )*

        m
    }}
}

lazy_static::lazy_static! {
    static ref CONFIGURATIONS: HashMap<&'static str, HighlightConfiguration> = {
        create_configurations!( go, sql, c_sharp, jsonnet )
    };
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

    match index_language(&filetype, &q.code) {
        Ok(document) => {
            let encoded = document.write_to_bytes().map_err(jsonify_err)?;

            Ok(json!({"data": base64::encode(&encoded), "plaintext": false}))
        }
        Err(Error::InvalidLanguage) => Err(json!({
            "error": format!("{} is not a valid filetype for treesitter", filetype)
        })),
        Err(err) => Err(jsonify_err(err)),
    }
}

pub fn index_language(filetype: &str, code: &str) -> Result<Document, Error> {
    match CONFIGURATIONS.get(filetype) {
        Some(lang_config) => index_language_with_config(code, lang_config),
        None => Err(Error::InvalidLanguage),
    }
}

pub fn make_highlight_config(name: &str, highlights: &str) -> Option<HighlightConfiguration> {
    let config = CONFIGURATIONS.get(name)?;

    // Create HighlightConfiguration language
    let mut lang = match HighlightConfiguration::new(config.language, highlights, "", "") {
        Ok(lang) => lang,
        Err(_) => return None,
    };

    // Associate highlights with configuration
    let highlight_names = MATCHES_TO_SYNTAX_KINDS
        .iter()
        .map(|hl| hl.0)
        .collect::<Vec<&str>>();
    lang.configure(&highlight_names);

    Some(lang)
}

pub fn index_language_with_config(
    code: &str,
    lang_config: &HighlightConfiguration,
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
        CONFIGURATIONS.get(l)
    })?;

    let mut emitter = ScipEmitter::new();
    emitter.render(highlights, &code, &get_syntax_kind_for_hl)
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

    use crate::determine_filetype;

    use super::*;

    #[test]
    fn test_highlights_one_comment() -> Result<(), Error> {
        let src = "// Hello World";
        let document = index_language("go", src)?;
        insta::assert_snapshot!(dump_document(&document, src));

        Ok(())
    }

    #[test]
    fn test_highlights_simple_main() -> Result<(), Error> {
        let src = r#"package main
import "fmt"

func main() {
	fmt.Println("Hello, world", 5)
}
"#;

        let document = index_language("go", src)?;
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

        let document = index_language("go", src)?;
        insta::assert_snapshot!(dump_document(&document, src));

        Ok(())
    }

    #[test]
    fn test_highlight_csharp_file() -> Result<(), Error> {
        let src = "using System;";
        let document = index_language("c_sharp", src)?;
        insta::assert_snapshot!(dump_document(&document, src));

        Ok(())
    }

    #[test]
    fn test_all_files() -> Result<(), std::io::Error> {
        let dir = read_dir("./src/snapshots/files/")?;
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

            let document = index_language(filetype, &contents).unwrap();
            insta::assert_snapshot!(
                filepath
                    .to_str()
                    .unwrap()
                    .replace("/src/snapshots/files", ""),
                dump_document(&document, &contents)
            );
        }

        Ok(())
    }
}
