use std::path::Path;

use protobuf::Message;
use rocket::serde::json::{json, Value as JsonValue};
use serde::Deserialize;
use sg_treesitter::jsonify_err;
use syntect::{
    html::ClassStyle,
    parsing::{SyntaxReference, SyntaxSet},
};

mod sg_treesitter;
pub use sg_treesitter::{
    index_language as treesitter_index, index_language_with_config as treesitter_index_with_config,
    lsif_highlight,
};

mod sg_syntect;
use sg_syntect::ClassedTableGenerator;
use tree_sitter_highlight::Error;

use crate::sg_treesitter::treesitter_language;

mod sg_sciptect;

thread_local! {
    pub(crate) static SYNTAX_SET: SyntaxSet = SyntaxSet::load_defaults_newlines();
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

// NOTE: Keep in sync: internal/gosyntect/gosyntect.go
#[derive(Deserialize, Default, Debug, PartialEq, Eq)]
pub enum SyntaxEngine {
    #[default]
    #[serde(rename = "syntect")]
    Syntect,

    #[serde(rename = "tree-sitter")]
    TreeSitter,

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

pub fn determine_filetype(q: &SourcegraphQuery) -> String {
    let filetype = SYNTAX_SET.with(|syntax_set| match determine_language(q, syntax_set) {
        Ok(language) => language.name.clone(),
        Err(_) => "".to_owned(),
    });

    if filetype.is_empty() || filetype.to_lowercase() == "plain text" {
        #[allow(clippy::single_match)]
        match q.extension.as_str() {
            "ncl" => return "nickel".to_string(),
            _ => {}
        };
    }

    // Normalize all the filenames here
    match filetype.as_str() {
        "Rust Enhanced" => "rust",
        "C++" => "cpp",
        "C#" => "c_sharp",
        "JS Custom - React" => "javascript",
        "TypeScriptReact" => {
            if q.filepath.ends_with(".tsx") {
                "tsx"
            } else {
                "typescript"
            }
        }
        filetype => filetype,
    }
    .to_lowercase()
}

pub fn determine_language<'a>(
    q: &SourcegraphQuery,
    syntax_set: &'a SyntaxSet,
) -> Result<&'a SyntaxReference, JsonValue> {
    // If filetype is passed, we should choose that if possible.
    if let Some(filetype) = &q.filetype {
        // This is `find_syntax_by_name` except that it doesn't care about
        // case sensitivity or anything like that.
        //
        // This makes it just a lost simpler to move between frontend and backend.
        // At some point, we need a definitive list for this.
        if let Some(language) = syntax_set
            .syntaxes()
            .iter()
            .rev()
            .find(|&s| filetype == &s.name.to_lowercase())
        {
            return Ok(language);
        }
    }

    if q.filepath.is_empty() {
        // Legacy codepath, kept for backwards-compatability with old clients.
        return match syntax_set.find_syntax_by_extension(&q.extension) {
            Some(v) => Ok(v),
            // Fall back: Determine syntax definition by first line.
            None => match syntax_set.find_syntax_by_first_line(&q.code) {
                Some(v) => Ok(v),
                None => Err(json!({"error": "invalid extension"})),
            },
        };
    }

    // Split the input path ("foo/myfile.go") into file name
    // ("myfile.go") and extension ("go").
    let path = Path::new(&q.filepath);
    let file_name = path.file_name().and_then(|n| n.to_str()).unwrap_or("");
    let extension = path.extension().and_then(|x| x.to_str()).unwrap_or("");

    // Override syntect's language detection for conflicting file extensions because
    // it's impossible to express this logic in a syntax definition.
    struct Override {
        extension: &'static str,
        prefix_langs: Vec<(&'static str, &'static str)>,
        default: &'static str,
    }
    let overrides = vec![
        Override {
            extension: "cls",
            prefix_langs: vec![("%", "TeX"), ("\\", "TeX")],
            default: "Apex",
        },
        Override {
            extension: "xlsg",
            prefix_langs: vec![],
            default: "xlsg",
        },
    ];

    if let Some(Override {
        prefix_langs,
        default,
        ..
    }) = overrides.iter().find(|o| o.extension == extension)
    {
        let name = match prefix_langs
            .iter()
            .find(|(prefix, _)| q.code.starts_with(prefix))
        {
            Some((_, lang)) => lang,
            None => default,
        };
        return Ok(syntax_set
            .find_syntax_by_name(name)
            .unwrap_or_else(|| syntax_set.find_syntax_plain_text()));
    }

    Ok(syntax_set
        // First try to find a syntax whose "extension" matches our file
        // name. This is done due to some syntaxes matching an "extension"
        // that is actually a whole file name (e.g. "Dockerfile" or "CMakeLists.txt")
        // see https://github.com/trishume/syntect/pull/170
        .find_syntax_by_extension(file_name)
        .or_else(|| syntax_set.find_syntax_by_extension(extension))
        .or_else(|| syntax_set.find_syntax_by_first_line(&q.code))
        .unwrap_or_else(|| syntax_set.find_syntax_plain_text()))
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

pub fn syntect_highlight(q: SourcegraphQuery) -> JsonValue {
    SYNTAX_SET.with(|syntax_set| {
        // Determine syntax definition by extension.
        let syntax_def = match determine_language(&q, syntax_set) {
            Ok(v) => v,
            Err(e) => return e,
        };

        let output = ClassedTableGenerator::new(
            syntax_set,
            syntax_def,
            &q.code,
            q.line_length_limit,
            ClassStyle::SpacedPrefixed { prefix: "hl-" },
        )
        .generate();

        json!({ "data": output, "plaintext": syntax_def.name == "Plain Text", })
    })
}

pub fn scip_highlight(q: ScipHighlightQuery) -> Result<JsonValue, JsonValue> {
    match q.engine {
        SyntaxEngine::Syntect => SYNTAX_SET.with(|ss| {
            let sg_query = SourcegraphQuery {
                extension: "".to_string(),
                filepath: q.filepath.clone(),
                filetype: q.filetype.clone(),
                line_length_limit: None,
                code: q.code.clone(),
            };

            let language = determine_language(&sg_query, ss).map_err(jsonify_err)?;
            let document = sg_sciptect::DocumentGenerator::new(
                ss,
                language,
                q.code.as_str(),
                q.line_length_limit,
            )
            .generate();
            let encoded = document.write_to_bytes().map_err(jsonify_err)?;
            Ok(json!({"scip": base64::encode(encoded), "plaintext": false}))
        }),
        SyntaxEngine::TreeSitter | SyntaxEngine::ScipSyntax => {
            let language = q
                .filetype
                .ok_or_else(|| json!({"error": "Must pass a language for /scip" }))?
                .to_lowercase();

            let include_locals = q.engine == SyntaxEngine::ScipSyntax;

            match treesitter_index(treesitter_language(&language), &q.code, include_locals) {
                Ok(document) => {
                    let encoded = document.write_to_bytes().map_err(jsonify_err)?;

                    Ok(json!({"scip": base64::encode(encoded), "plaintext": false}))
                }
                Err(Error::InvalidLanguage) => Err(json!({
                    "error": format!("{} is not a valid filetype for treesitter", language)
                })),
                Err(err) => Err(jsonify_err(err)),
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use syntect::parsing::SyntaxSet;

    use super::*;

    #[test]
    fn cls_tex() {
        let syntax_set = SyntaxSet::load_defaults_newlines();
        let query = SourcegraphQuery {
            filepath: "foo.cls".to_string(),
            filetype: None,
            code: "%".to_string(),
            line_length_limit: None,
            extension: String::new(),
        };
        let result = determine_language(&query, &syntax_set);
        assert_eq!(result.unwrap().name, "TeX");
    }

    #[test]
    fn cls_apex() {
        let syntax_set = SyntaxSet::load_defaults_newlines();
        let query = SourcegraphQuery {
            filepath: "foo.cls".to_string(),
            filetype: None,
            code: "/**".to_string(),
            line_length_limit: None,
            extension: String::new(),
        };
        let result = determine_language(&query, &syntax_set);
        assert_eq!(result.unwrap().name, "Apex");
    }
}
