use std::fs;

use scip_treesitter::snapshot::dump_document;
use sg_syntax::treesitter_index;

fn main() {
    if let Some(path) = std::env::args().nth(1) {
        let contents = match fs::read_to_string(&path) {
            Ok(contents) => contents,
            Err(err) => {
                eprintln!("Failed to read path: {:?}. {}", path, err);
                return;
            }
        };

        let document = match treesitter_index("go", &contents, false) {
            Ok(document) => document,
            Err(err) => {
                eprintln!("Failed to index document: {:?}", err);
                return;
            }
        };

        println!(
            "\n\n{}",
            dump_document(&document, &contents).expect("to dump document")
        );
    } else {
        panic!("Must pass a filepath");
    }
}
