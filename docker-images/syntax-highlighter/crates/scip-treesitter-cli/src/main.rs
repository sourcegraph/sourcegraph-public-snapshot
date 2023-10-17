use clap::{Parser, Subcommand};
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter_languages::parsers::BundledParser;

use std::path::Path;

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
#[command(propagate_version = true)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Adds files to myapp
    Index {
        #[arg(short, long)]
        language: String,

        #[arg(short, long)]
        out: Option<String>,
        filenames: Vec<String>,
    },
}

pub fn main() {
    let cli = Cli::parse();

    match cli.command {
        Commands::Index {
            language,
            out,
            filenames,
        } => index_impl(&language, &filenames, &out),
    }
}

fn index_impl(language: &String, filenames: &Vec<String>, out: &Option<String>) {
    let p = BundledParser::get_parser(language).unwrap();

    let working_directory = Path::new("./");
    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-treesitter".to_string(),
                version: clap::crate_version!().to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: format!("file://{}", working_directory.to_str().unwrap()),
            ..Default::default()
        })
        .into(),
        ..Default::default()
    };

    for (_, filename) in filenames.iter().enumerate() {
        let contents = std::fs::read(filename).unwrap();
        let mut document = get_symbols(&p, &contents).unwrap();

        document.relative_path = filename.clone();
        let locals = get_locals(&p, &contents);

        match locals {
            Some(Ok(occurrences)) => {
                for occ in occurrences {
                    document.occurrences.push(occ);
                }
            }
            Some(Err(msg)) => {
                println!("Error extracting locals: {}", msg);
            }
            None => {}
        }

        index.documents.push(document);
    }

    let out_name = out.clone().unwrap_or("index.scip".to_string());
    let path = working_directory.join(out_name);

    println!(
        "Writing index for {} documents into {}",
        index.documents.len(),
        path.display()
    );

    write_message_to_file(path, index).expect("to write the file");
}

pub fn write_message_to_file<P>(
    path: P,
    msg: impl protobuf::Message,
) -> Result<(), Box<dyn std::error::Error>>
where
    P: AsRef<std::path::Path>,
{
    use std::io::Write;

    let res = msg.write_to_bytes()?;
    let output = std::fs::File::create(path)?;
    let mut writer = std::io::BufWriter::new(output);
    writer.write_all(&res)?;

    Ok(())
}
