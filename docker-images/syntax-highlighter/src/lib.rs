use std::fmt::Display;

use anyhow::Context;
use rocket::serde::json::{json, Value as JsonValue};
use serde::Deserialize;
use syntax_analysis::highlighting::{FileInfo, HighlightingBackend, PayloadKind};
use syntect::parsing::SyntaxSet;

thread_local! {
    pub(crate) static SYNTAX_SET: SyntaxSet = SyntaxSet::load_defaults_newlines();
}

pub fn list_features() {
    // List supported file extensions.
    SYNTAX_SET.with(|syntax_set| {
        println!("## Supported file extensions:");
        println!();
        for sd in syntax_set.syntaxes() {
            println!("- {} (`{}`)", sd.name, sd.file_extensions.join("`, `"));
        }
        println!();
    });
}

/// Struct from: internal/gosyntect/gosyntect.go
///
/// Keep in sync with that struct.
#[derive(Deserialize, Default, Debug)]
pub struct SourcegraphQuery {
    // Deprecated field with a default empty string value, kept for backwards
    // compatability with old clients.
    //
    // TODO: Can I just delete this because this image will only run for a particular sourcegraph
    // version... so they can't be out of sync anymore, which is pretty cool
    #[serde(default)]
    pub extension: String,

    // Contents of the file
    pub code: String,

    // default empty string value for backwards compat with clients who do not specify this field.
    #[serde(default)]
    pub filepath: String,

    // The language defined by the server. Required to tree-sitter to use for the filetype name.
    // default empty string value for backwards compat with clients who do not specify this field.
    pub filetype: Option<String>,

    // line_length_limit is ignored if css is false
    pub line_length_limit: Option<usize>,
}

impl SourcegraphQuery {
    fn file_info(&self) -> FileInfo<'_> {
        if self.filepath.is_empty() {
            FileInfo::new_from_extension(&self.extension, &self.code, self.filetype.as_deref())
        } else {
            FileInfo::new(&self.filepath, &self.code, self.filetype.as_deref())
        }
    }
}

// NOTE: Keep in sync: internal/gosyntect/gosyntect.go
#[derive(Deserialize, Default, Debug, PartialEq, Eq)]
pub enum SyntaxEngine {
    /// Returns highlighting data only via syntect in SCIP format
    #[default]
    #[serde(rename = "syntect")]
    Syntect,

    /// Returns highlighting data only via Tree-sitter in SCIP format
    #[serde(rename = "tree-sitter")]
    TreeSitter,

    /// Returns highlighting data and optionally locals data via Tree-sitter in SCIP format
    ///
    /// This name is preserved for backwards compatibility. The web client
    /// makes use of locals data for code navigation, so it doesn't make
    /// sense to omit locals data if it is available.
    #[serde(rename = "scip-syntax")]
    ScipSyntax,
}

#[derive(Deserialize, Default, Debug)]
pub struct ScipHighlightQuery {
    // Which highlighting engine to use.
    pub engine: SyntaxEngine,

    // Contents of the file
    pub code: String,

    // filepath is only used if language is None.
    pub filepath: String,

    // The language defined by the server. Required to tree-sitter to use for the filetype name.
    // default empty string value for backwards compat with clients who do not specify this field.
    pub filetype: Option<String>,

    // line_length_limit is used to limit syntect problems when
    // parsing very long lines
    pub line_length_limit: Option<usize>,
}

impl ScipHighlightQuery {
    fn file_info(&self) -> FileInfo<'_> {
        FileInfo::new(&self.filepath, &self.code, self.filetype.as_deref())
    }
}

pub fn jsonify_err(e: impl Display) -> JsonValue {
    json!({"error": format!("{:#}", e)})
}

pub fn syntect_highlight(q: SourcegraphQuery) -> Result<JsonValue, JsonValue> {
    SYNTAX_SET.with(|syntax_set| {
        let backend = HighlightingBackend::SyntectHtml {
            syntax_set,
            line_length_limit: q.line_length_limit,
        };
        let output = backend.highlight(&q.file_info()).map_err(jsonify_err)?;

        debug_assert!(output.kind == PayloadKind::Html);
        Ok(json!({ "data": output.payload, "plaintext": &output.grammar == "Plain Text", }))
    })
}

// TODO(cleanup_lsif): Remove this when we remove /lsif endpoint
// Currently left unchanged
pub fn lsif_highlight(q: SourcegraphQuery) -> Result<JsonValue, JsonValue> {
    let output = HighlightingBackend::TreeSitter {
        include_locals: false,
    }
    .highlight(&q.file_info())
    .context("/lsif endpoint")
    .map_err(jsonify_err)?;
    debug_assert!(output.kind == PayloadKind::Base64EncodedScip);
    Ok(json!({"data": output.payload, "plaintext": false}))
}

pub fn scip_highlight(q: ScipHighlightQuery) -> Result<JsonValue, JsonValue> {
    SYNTAX_SET.with(|syntax_set| {
        let backend = match q.engine {
            SyntaxEngine::Syntect => HighlightingBackend::SyntectScip {
                syntax_set,
                line_length_limit: q.line_length_limit,
            },
            SyntaxEngine::TreeSitter | SyntaxEngine::ScipSyntax => {
                HighlightingBackend::TreeSitter {
                    include_locals: q.engine == SyntaxEngine::ScipSyntax,
                }
            }
        };
        let output = backend
            .highlight(&q.file_info())
            .context("/scip endpoint")
            .map_err(jsonify_err)?;
        debug_assert!(output.kind == PayloadKind::Base64EncodedScip);
        Ok(json!({"scip": output.payload, "plaintext": false}))
    })
}

#[cfg(test)]
mod tests {

    use crate::{scip_highlight, ScipHighlightQuery, SyntaxEngine};

    #[test]
    fn check_error() {
        let result = scip_highlight(ScipHighlightQuery {
            engine: SyntaxEngine::TreeSitter,
            code: "int a = 3;".to_string(),
            filepath: "a.c".to_string(),
            filetype: None,
            line_length_limit: None,
        });
        assert!(result.is_err());
        // The default string formatting for an error only picks the most
        // recent element. Make sure we're serializing the full context.
        insta::assert_display_snapshot!(result.unwrap_err());
    }
}
