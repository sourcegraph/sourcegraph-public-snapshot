use std::fs;

use sg_syntax::{dump_document, lsif_index};

fn main() {
    if let Some(path) = std::env::args().nth(1) {
        let contents = match fs::read_to_string(&path) {
            Ok(contents) => contents,
            Err(err) => {
                eprintln!("Failed to read path: {:?}. {}", path, err);
                return;
            }
        };

        // let language = determine_language();
        let document = match lsif_index("go", &contents) {
            Ok(document) => document,
            Err(err) => {
                eprintln!("Failed to index document: {:?}", err);
                return;
            }
        };

        println!("\n\n{}", dump_document(&document, &contents));
    } else {
        panic!("Must pass a filepath");
    }
}
