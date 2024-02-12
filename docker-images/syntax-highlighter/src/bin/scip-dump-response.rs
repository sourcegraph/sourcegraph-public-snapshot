use std::{fs, path::Path};

use clap::Parser;
use syntax_analysis::highlighting::FileInfo;
use syntect::parsing::SyntaxSet;

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

    let file_info = FileInfo::new(path.to_string_lossy().as_ref(), &contents, None);

    let language = file_info
        .determine_language(&SyntaxSet::load_defaults_newlines())
        .expect("failed to determine language");
    println!("  language: {:?}", language.to_string());

    let document = language
        .highlight_document(&contents, args.include_locals)
        .unwrap_or_else(|e| {
            panic!(
                "failed to run Tree-sitter for file '{}': {:?}",
                path.to_string_lossy(),
                e
            )
        });
    println!("  highlighted document");

    scip::write_message_to_file(output, document).expect("writes document");
    println!("  wrote document to {:?}", output);

    Ok(())
}
