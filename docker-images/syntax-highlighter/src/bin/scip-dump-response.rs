use std::{fs, path::Path};

use clap::Parser;

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Arguments {
    /// Path to the input file
    input: String,

    /// Path to the output file
    output: String,

    /// Include locals, default false
    #[arg(long)]
    include_locals: bool,
}

fn main() -> Result<(), std::io::Error> {
    println!("scip-dump-response - write a dump for a filepath");

    let args = Arguments::parse();

    // read file from args
    let path = Path::new(&args.input);
    if !path.exists() {
        return Err(std::io::Error::new(
            std::io::ErrorKind::NotFound,
            "file not found",
        ));
    }

    let output = Path::new(&args.output);
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
        line_length_limit: None,
    });

    println!("  filetype: {:?}", filetype);

    let document = sg_syntax::treesitter_index(&filetype, &contents, args.include_locals)
        .expect("parse document");
    println!("  parsed document");

    scip::write_message_to_file(output, document).expect("writes document");
    println!("  wrote document to {:?}", output);

    Ok(())
}
