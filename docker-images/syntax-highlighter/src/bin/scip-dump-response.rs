use std::{fs, path::Path};

fn main() -> Result<(), std::io::Error> {
    println!("scip-dump-response - write a dump for a filepath");

    // read file from args
    let path = std::env::args().nth(1).expect("pass an input filepath");
    let path = Path::new(&path);
    if !path.exists() {
        return Err(std::io::Error::new(
            std::io::ErrorKind::NotFound,
            "file not found",
        ));
    }

    let output = std::env::args().nth(2).expect("pass an output path");
    let output = Path::new(&output);

    println!("  reading {:?}", path);
    let contents = fs::read_to_string(path)?;
    println!("  read {} bytes", contents.len());

    let filetype = sg_syntax::determine_filetype(&sg_syntax::SourcegraphQuery {
        extension: path
            .extension()
            .expect("extension")
            .to_str()
            .expect("valid utf-8 path")
            .to_string(),
        code: contents.clone(),
        filepath: "".to_string(),
        filetype: None,
        css: false,
        line_length_limit: None,
        theme: "".to_string(),
    });

    println!("  filetype: {:?}", filetype);

    let document = sg_syntax::treesitter_index(&filetype, &contents).expect("parse document");
    println!("  parsed document");

    scip::write_message_to_file(output, document).expect("writes document");
    println!("  wrote document to {:?}", output);

    Ok(())
}
