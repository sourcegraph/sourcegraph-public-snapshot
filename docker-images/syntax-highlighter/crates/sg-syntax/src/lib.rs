use std::path::Path;

use rocket::serde::json::{json, Value as JsonValue};
use serde::Deserialize;
use syntect::html::{highlighted_html_for_string, ClassStyle};
use syntect::{
    highlighting::ThemeSet,
    parsing::{SyntaxReference, SyntaxSet},
};

mod sg_treesitter;
pub use sg_treesitter::dump_document;
pub use sg_treesitter::dump_document_range;
pub use sg_treesitter::index_language as lsif_index;
pub use sg_treesitter::index_language_with_config as lsif_index_with_config;
pub use sg_treesitter::lsif_highlight;
pub use sg_treesitter::make_highlight_config;
pub use sg_treesitter::FileRange as DocumentFileRange;
pub use sg_treesitter::PackedRange as LsifPackedRange;

mod sg_syntect;
use sg_syntect::ClassedTableGenerator;

thread_local! {
    pub(crate) static SYNTAX_SET: SyntaxSet = SyntaxSet::load_defaults_newlines();
}

lazy_static::lazy_static! {
    static ref THEME_SET: ThemeSet = ThemeSet::load_defaults();
}

/// Struct from: internal/gosyntect/gosyntect.go
///
/// Keep in sync with that struct.
#[derive(Deserialize)]
pub struct SourcegraphQuery {
    // Deprecated field with a default empty string value, kept for backwards
    // compatability with old clients.
    //
    // TODO: Can I just delete this because this image will only run for a particular sourcegraph
    // version... so they can't be out of sync anymore, which is pretty cool
    #[serde(default)]
    pub extension: String,

    // default empty string value for backwards compat with clients who do not specify this field.
    #[serde(default)]
    pub filepath: String,

    // The language defined by the server. Required to tree-sitter to use for the filetype name.
    // default empty string value for backwards compat with clients who do not specify this field.
    pub filetype: Option<String>,

    // If css is set, the highlighted code will be returned as a HTML table with CSS classes
    // annotating the highlighted types.
    #[serde(default)]
    pub css: bool,

    // line_length_limit is ignored if css is false
    pub line_length_limit: Option<usize>,

    // theme is ignored if css is true
    pub theme: String,

    pub code: String,
}

pub fn determine_filetype(q: &SourcegraphQuery) -> String {
    let filetype = SYNTAX_SET.with(|syntax_set| match determine_language(q, syntax_set) {
        Ok(language) => language.name.clone(),
        Err(_) => "".to_owned(),
    });

    // We normalize all the filenames here
    match filetype.as_str() {
        "C#" => "c_sharp",
        filetype => filetype,
    }
    .to_lowercase()
}

fn determine_language<'a>(
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
    let overrides = vec![Override {
        extension: "cls",
        prefix_langs: vec![("%", "TeX"), ("\\", "TeX")],
        default: "Apex",
    }];

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
    // List embedded themes.
    println!("## Embedded themes:");
    println!();
    for t in THEME_SET.themes.keys() {
        println!("- `{}`", t);
    }
    println!();

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

        if q.css {
            let output = ClassedTableGenerator::new(
                syntax_set,
                syntax_def,
                &q.code,
                q.line_length_limit,
                ClassStyle::SpacedPrefixed { prefix: "hl-" },
            )
            .generate();

            json!({
                "data": output,
                "plaintext": syntax_def.name == "Plain Text",
            })
        } else {
            // TODO(slimsag): return the theme's background color (and other info??) to caller?
            // https://github.com/trishume/syntect/blob/c8b47758a3872d478c7fc740782cd468b2c0a96b/examples/synhtml.rs#L24

            // Determine theme to use.
            //
            // TODO(slimsag): We could let the query specify the theme file's actual
            // bytes? e.g. via `load_from_reader`.
            let theme = match THEME_SET.themes.get(&q.theme) {
                Some(v) => v,
                None => return json!({"error": "invalid theme", "code": "invalid_theme"}),
            };

            json!({
                "data": highlighted_html_for_string(&q.code, syntax_set, syntax_def, theme),
                "plaintext": syntax_def.name == "Plain Text",
            })
        }
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    use syntect::parsing::SyntaxSet;

    #[test]
    fn cls_tex() {
        let syntax_set = SyntaxSet::load_defaults_newlines();
        let query = SourcegraphQuery {
            filepath: "foo.cls".to_string(),
            filetype: None,
            code: "%".to_string(),
            css: false,
            line_length_limit: None,
            extension: String::new(),
            theme: String::new(),
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
            css: false,
            line_length_limit: None,
            extension: String::new(),
            theme: String::new(),
        };
        let result = determine_language(&query, &syntax_set);
        assert_eq!(result.unwrap().name, "Apex");
    }
}
