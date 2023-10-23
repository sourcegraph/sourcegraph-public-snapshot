use clap::{Parser, Subcommand};
use protobuf::{CodedInputStream, Message};
use scip::types::Document;
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter_languages::parsers::BundledParser;

use anyhow::Result;
use std::{fs::File, io::BufReader, path::Path};

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
        #[arg(long)]
        no_locals: bool,
        #[arg(long)]
        no_globals: bool,
        #[arg(long)]
        strict: bool,
        #[arg(long)]
        cwd: Option<String>,
    },
}

struct Options {
    locals: bool,
    globals: bool,
    strict: bool,
}

pub fn main() {
    let cli = Cli::parse();

    match cli.command {
        Commands::Index {
            language,
            out,
            filenames,
            no_locals,
            no_globals,
            strict,
            cwd,
        } => index_command(
            &language,
            &filenames,
            &out,
            &cwd,
            &Options {
                locals: !no_locals,
                globals: !no_globals,
                strict,
            },
        ),
    }
}

fn index_command(
    language: &String,
    filenames: &Vec<String>,
    out: &Option<String>,
    cwd: &Option<String>,
    options: &Options,
) {
    let p = BundledParser::get_parser(language).unwrap();

    let working_directory: String = cwd.clone().unwrap_or("./".to_string()); //= cwd.map(|p| Path::new(p.as_str())).unwrap_or(Path::new("./"));
    let working_path = Path::new(&working_directory);

    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-treesitter".to_string(),
                version: clap::crate_version!().to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: format!("file://{}", working_directory),
            ..Default::default()
        })
        .into(),
        ..Default::default()
    };

    for (_, filename) in filenames.iter().enumerate() {
        let contents = std::fs::read(filename).unwrap();
        eprintln!("Processing {filename}");
        match index_content(contents, &p, options) {
            Ok(mut document) => {
                document.relative_path = filename.to_string();
                index.documents.push(document);
            }
            other => {
                if options.strict {
                    other.unwrap();
                } else {
                    eprintln!("Failed to extract locals: {:?}", other)
                }
            }
        }
    }

    let out_name = out.clone().unwrap_or("index.scip".to_string());
    let path = working_path.join(out_name);

    eprintln!(
        "Writing index for {} documents into {}",
        index.documents.len(),
        path.display()
    );

    write_message_to_file(path, index).expect("to write the file");
}

fn index_content(contents: Vec<u8>, parser: &BundledParser, options: &Options) -> Result<Document> {
    let mut document: Document;

    if options.globals {
        document = get_symbols(parser, &contents).unwrap();
    } else {
        document = Document::new();
    }

    if options.locals {
        let locals = get_locals(parser, &contents);

        match locals {
            Some(Ok(occurrences)) => {
                for occ in occurrences {
                    document.occurrences.push(occ);
                }
            }
            Some(other) => {
                if options.strict {
                    other.unwrap();
                } else {
                    eprintln!("Failed to extract locals: {:?}", other)
                }
            }
            None => {}
        }
    }

    return Ok(document);
}

fn write_message_to_file<P>(
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

fn read_index_from_file(file: &str) -> scip::types::Index {
    let mut candidate_idx = scip::types::Index::new();
    let candidate_f = File::open(file).unwrap();

    let mut reader = BufReader::new(candidate_f);
    let mut cis = CodedInputStream::from_buf_read(&mut reader);

    candidate_idx.merge_from(&mut cis).unwrap();
    return candidate_idx;
}

#[cfg(test)]
mod tests {
    use crate::read_index_from_file;
    use assert_cmd::prelude::*; // Add methods on commands
    use std::process::Command;
    use std::{env::temp_dir, path::Path}; // Run programs

    #[test]
    fn e2e() {
        let mut cmd = Command::cargo_bin("scip-treesitter-cli").unwrap();
        let out_dir = temp_dir();
        let path = out_dir.join("globals.java");
        let out_path = out_dir
            .join("index-java.scip")
            .to_str()
            .unwrap()
            .to_string();

        write_file(
            &path,
            include_str!("../../scip-syntax/testdata/globals.java").to_string(),
        );

        cmd.current_dir(out_dir)
            .arg("index")
            .args(["-l", "java", "-o", &out_path])
            .arg("globals.java");

        cmd.assert().success();

        let index = read_index_from_file(&out_path);

        assert!(index.documents.len() == 1);
        assert_eq!(index.documents[0].relative_path, "globals.java");
        assert!(index.documents[0].symbols.len() > 0);
    }

    fn write_file(path: &Path, contents: String) {
        use std::io::Write;

        let res = contents.into_bytes();
        let output = std::fs::File::create(path).unwrap();
        let mut writer = std::io::BufWriter::new(output);
        writer.write_all(&res).unwrap();
    }
}
