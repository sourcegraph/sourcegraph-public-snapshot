use std::fs;

use rustyline::config::Configurer;
use rustyline::Config;
use sg_syntax::dump_document_range;
use sg_syntax::lsif_index_with_config;
use sg_syntax::make_highlight_config;
use sg_syntax::DocumentFileRange;

fn main() {
    println!("========================================");
    println!("  Welcome to lsif-syntax-repl");
    println!("========================================");

    let contents = if let Some(path) = std::env::args().nth(1) {
        match fs::read_to_string(&path) {
            Ok(contents) => contents,
            Err(err) => {
                eprintln!("Failed to read path: {:?}. {}", path, err);
                return;
            }
        }
    } else {
        let mut rl = rustyline::Editor::<()>::new();
        match rl.readline("Contents: ") {
            Ok(contents) => contents,
            Err(err) => {
                eprintln!("Failed to read path: {err}");
                return;
            }
        }
    };

    let mut config = Config::builder();
    config.set_max_history_size(100);
    config.set_auto_add_history(true);

    let mut rl = rustyline::Editor::<()>::with_config(config.build());

    let range: Option<DocumentFileRange> = match rl.readline("Range (Optional, 1-Indexed): ") {
        Ok(line) => {
            if line.is_empty() {
                None
            } else {
                let line_number = line.parse().unwrap();
                Some(DocumentFileRange {
                    start: line_number,
                    end: line_number,
                })
            }
        }
        _ => return,
    };

    eprintln!();
    eprintln!("Usage Instructions:");
    eprintln!("- <Up> / <Down> to cycle through history");
    eprintln!();

    while let Ok(line) = rl.readline("Query >> ") {
        if line.is_empty() {
            break;
        }

        let config = match make_highlight_config("c_sharp", &line) {
            Some(config) => config,
            None => {
                eprintln!("=> Error when constructing configuration, probably invalid query.");
                continue;
            }
        };

        let document = match lsif_index_with_config(&contents, &config) {
            Ok(document) => document,
            Err(err) => {
                eprintln!("Failed to index document: {:?}", err);
                return;
            }
        };

        eprintln!("{}", dump_document_range(&document, &contents, &range));
    }
}
