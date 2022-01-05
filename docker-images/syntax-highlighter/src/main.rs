#![allow(macro_expanded_macro_exports_accessed_by_absolute_paths)]

#[macro_use]
extern crate lazy_static;
extern crate rayon;
#[macro_use]
extern crate rocket;
#[macro_use]
extern crate rocket_contrib;
#[macro_use]
extern crate serde_derive;
extern crate serde_json;
extern crate syntect;

use rocket_contrib::json::{Json, JsonValue};
use std::collections::HashMap;
use std::env;
use std::panic;
use std::path::Path;
use syntect::parsing::SyntaxReference;
use syntect::{
    highlighting::ThemeSet,
    html::{highlighted_html_for_string, ClassStyle},
    parsing::SyntaxSet,
};
use tree_sitter::Language;
use tree_sitter_highlight::HighlightConfiguration;
use tree_sitter_highlight::Highlighter;

use tree_sitter_go;

mod css_table;
use css_table::ClassedTableGenerator;

use crate::ts_render::TableHtmlRenderer;

mod ts_render;

thread_local! {
    static SYNTAX_SET: SyntaxSet = SyntaxSet::load_defaults_newlines();
}

lazy_static! {
    static ref THEME_SET: ThemeSet = ThemeSet::load_defaults();
}

// lazy_static! {
//     static ref LANGUAGES:
// }

#[derive(Deserialize)]
struct Query {
    // Deprecated field with a default empty string value, kept for backwards
    // compatability with old clients.
    #[serde(default)]
    extension: String,

    // default empty string value for backwards compat with clients who do not specify this field.
    #[serde(default)]
    filepath: String,

    // If css is set, the highlighted code will be returned as a HTML table with CSS classes
    // annotating the highlighted types.
    #[serde(default)]
    css: bool,

    // line_length_limit is ignored if css is false
    line_length_limit: Option<usize>,

    // theme is ignored if css is true
    theme: String,

    code: String,
}

#[post("/", format = "application/json", data = "<q>")]
fn index(q: Json<Query>) -> JsonValue {
    // TODO(slimsag): In an ideal world we wouldn't be relying on catch_unwind
    // and instead Syntect would return Result types when failures occur. This
    // will require some non-trivial work upstream:
    // https://github.com/trishume/syntect/issues/98
    let result = panic::catch_unwind(|| highlight(q.into_inner()));
    match result {
        Ok(v) => v,
        Err(_) => json!({"error": "panic while highlighting code", "code": "panic"}),
    }
}

lazy_static! {
    static ref CONFIGURATIONS: HashMap<&'static str, HighlightConfiguration> = {
        let highlight_names = &[
            "attribute",
            "constant",
            "comment",
            "function.builtin",
            "function",
            "include",
            "keyword",
            "operator",
            "property",
            "punctuation",
            "punctuation.bracket",
            "punctuation.delimiter",
            "string",
            "string.special",
            "tag",
            "type",
            "type.builtin",
            "variable",
            "variable.builtin",
            "variable.parameter",
        ];

        let mut m = HashMap::new();

        let mut lang = HighlightConfiguration::new(
            tree_sitter_go::language(),
            include_str!("../queries/go/highlights.scm").as_ref(),
            "",
            "",
        )
        .unwrap();
        lang.configure(highlight_names);
        m.insert("go", lang);

        let mut lang = HighlightConfiguration::new(
            tree_sitter_html::language(),
            include_str!("../queries/html/highlights.scm").as_ref(),
            include_str!("../queries/html/injections.scm").as_ref(),
            "",
        )
        .unwrap();
        lang.configure(highlight_names);
        m.insert("html", lang);

        let mut lang = HighlightConfiguration::new(
            tree_sitter_javascript::language(),
            include_str!("../queries/javascript/highlights.scm").as_ref(),
            include_str!("../queries/javascript/injections.scm").as_ref(),
            "",
        )
        .unwrap();
        lang.configure(highlight_names);
        m.insert("javascript", lang);

        m
    };
}

fn highlight(q: Query) -> JsonValue {
    SYNTAX_SET.with(|syntax_set| {
        // Determine syntax definition by extension.
        let syntax_def = match determine_language(&q, syntax_set) {
            Ok(v) => v,
            Err(e) => return e,
        };

        println!("RUNNING: {}", syntax_def.name.to_lowercase());
        match syntax_def.name.to_lowercase().as_str() {
            mut filetype @ ("go" | "javascript" | "html" | "js custom - react") => {
                if filetype == "js custom - react" {
                    filetype = "javascript"
                }

                let class_names = &[
                    "class='hl-attribute'",
                    "class='hl-constant'",
                    "class='hl-comment'",
                    "class='hl-function.builtin'",
                    "class='hl-function'",
                    "class='hl-include'",
                    "class='hl-keyword'",
                    "class='hl-operator'",
                    "class='hl-property'",
                    "class='hl-punctuation'",
                    "class='hl-punctuation.bracket'",
                    "class='hl-punctuation.delimiter'",
                    "class='hl-string'",
                    "class='hl-string.special'",
                    "class='hl-tag'",
                    "class='hl-type'",
                    "class='hl-type.builtin'",
                    "class='hl-variable'",
                    "class='hl-variable.builtin'",
                    "class='hl-variable.parameter'",
                ];

                println!("Oh no, we got some code: {}", filetype);

                let mut highlighter = Highlighter::new();

                let lang_config = &CONFIGURATIONS[filetype];

                let highlights = highlighter
                    .highlight(&lang_config, q.code.as_bytes(), None, |l| {
                        println!("Some language: {}", l);
                        Some(&CONFIGURATIONS[l])
                    })
                    .unwrap();

                let mut html_renderer = TableHtmlRenderer::new();
                html_renderer
                    .render(highlights, q.code.as_bytes(), &|highlight| {
                        // println!("Highlight from render: {:?}", highlight);
                        class_names[highlight.0].as_bytes()
                    })
                    .unwrap();

                // println!("Highlights: {:?}", String::from_utf8(html_renderer.html));

                return json!({
                    "data": String::from_utf8(html_renderer.html).unwrap(),
                    "plaintext": false,
                });
            }
            _ => {}
        };

        if q.css {
            let output = ClassedTableGenerator::new(
                &syntax_set,
                &syntax_def,
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
                "data": highlighted_html_for_string(&q.code, &syntax_set, &syntax_def, theme),
                "plaintext": syntax_def.name == "Plain Text",
            })
        }
    })
}

fn determine_language<'a>(
    q: &Query,
    syntax_set: &'a SyntaxSet,
) -> Result<&'a SyntaxReference, JsonValue> {
    if q.filepath == "" {
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

#[cfg(test)]
mod tests {
    use syntect::parsing::SyntaxSet;

    use crate::{determine_language, Query};

    #[test]
    fn cls_tex() {
        let syntax_set = SyntaxSet::load_defaults_newlines();
        let query = Query {
            filepath: "foo.cls".to_string(),
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
        let query = Query {
            filepath: "foo.cls".to_string(),
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

#[get("/health")]
fn health() -> &'static str {
    "OK"
}

#[catch(404)]
fn not_found() -> JsonValue {
    json!({"error": "resource not found", "code": "resource_not_found"})
}

fn list_features() {
    // List embedded themes.
    println!("## Embedded themes:");
    println!("");
    for t in THEME_SET.themes.keys() {
        println!("- `{}`", t);
    }
    println!("");

    // List supported file extensions.
    SYNTAX_SET.with(|syntax_set| {
        println!("## Supported file extensions:");
        println!("");
        for sd in syntax_set.syntaxes() {
            println!("- {} (`{}`)", sd.name, sd.file_extensions.join("`, `"));
        }
        println!("");
    });
}

#[launch]
fn rocket() -> rocket::Rocket {
    // Only list features if QUIET != "true"
    match env::var("QUIET") {
        Ok(v) => {
            if v != "true" {
                list_features()
            }
        }
        Err(_) => list_features(),
    };

    rocket::ignite()
        .mount("/", routes![index, health])
        .register(catchers![not_found])
}
