#![allow(macro_expanded_macro_exports_accessed_by_absolute_paths)]

#[macro_use]
extern crate lazy_static;
extern crate rayon;
#[macro_use]
extern crate rocket;
#[macro_use]
extern crate serde_derive;
extern crate serde_json;
extern crate syntect;

use rocket::serde::json::{json, Json, Value as JsonValue};
use std::env;
use std::panic;
use std::path::Path;
use syntect::parsing::SyntaxReference;
use syntect::{
    highlighting::ThemeSet,
    html::{highlighted_html_for_string, ClassStyle},
    parsing::SyntaxSet,
};

mod css_table;
use css_table::ClassedTableGenerator;

thread_local! {
    static SYNTAX_SET: SyntaxSet = SyntaxSet::load_defaults_newlines();
}

lazy_static! {
    static ref THEME_SET: ThemeSet = ThemeSet::load_defaults();
}

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

fn highlight(q: Query) -> JsonValue {
    SYNTAX_SET.with(|syntax_set| {
        // Determine syntax definition by extension.
        let syntax_def = match determine_language(&q, syntax_set) {
            Ok(v) => v,
            Err(e) => return e,
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

    use crate::{Query, determine_language};

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
fn rocket() -> _ {
    // Only list features if QUIET != "true"
    match env::var("QUIET") {
        Ok(v) => {
            if v != "true" {
                list_features()
            }
        }
        Err(_) => list_features(),
    };

    rocket::build()
        .mount("/", routes![index, health])
        .register("/", catchers![not_found])
}
