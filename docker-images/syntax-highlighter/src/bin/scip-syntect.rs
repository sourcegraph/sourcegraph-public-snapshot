use std::fs;

use sg_syntax::{determine_language, SourcegraphQuery};
use syntect::{
    html::{ClassStyle, ClassedHTMLGenerator},
    parsing::SyntaxSet,
};

fn main() -> Result<(), std::io::Error> {
    println!("scip-syntect tester");
    let (path, contents) = if let Some(path) = std::env::args().nth(1) {
        match fs::read_to_string(&path) {
            Ok(contents) => (path, contents),
            Err(err) => {
                eprintln!("Failed to read path: {:?}. {}", path, err);
                return Ok(());
            }
        }
    } else {
        eprintln!("Must pass a path as the argument");
        return Ok(());
    };

    let q = SourcegraphQuery {
        filepath: path,
        code: contents.clone(),
        ..Default::default()
    };

    let syntax_set = SyntaxSet::load_defaults_newlines();
    let syntax_def = determine_language(&q, &syntax_set).unwrap();

    let mut html_generator =
        ClassedHTMLGenerator::new_with_class_style(syntax_def, &syntax_set, ClassStyle::Spaced);
    for line in contents.lines() {
        html_generator.parse_html_for_line_which_includes_newline(line);
    }
    let html = html_generator.finalize();
    println!("{}", html);

    Ok(())
}
