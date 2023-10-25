use clap::{Parser, Subcommand, ValueEnum};
use protobuf::{CodedInputStream, Message};
use scip::types::Document;
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter_languages::parsers::BundledParser;

use anyhow::Result;
use std::{fs::File, io::BufReader, path::PathBuf};

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
#[command(propagate_version = true)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum)]
enum AnalysisMode {
    /// Only extract occurrences of local definitions
    Locals,
    /// Only extract globally-accessible symbols
    Globals,
    /// Locals + Globals, extract everything
    Full,
}

impl AnalysisMode {
    fn locals(self) -> bool {
        return self == AnalysisMode::Locals || self == AnalysisMode::Full;
    }
    fn globals(self) -> bool {
        return self == AnalysisMode::Globals || self == AnalysisMode::Full;
    }
}

#[derive(Subcommand)]
enum Commands {
    /// Index source files using Tree Sitter parser for a given language
    /// and produce a SCIP file
    Index {
        ///
        #[arg(short, long)]
        language: String,

        /// Path where the SCIP index will be written
        #[arg(short, long, default_value = "./index.scip")]
        out: String,

        /// List of files to analyse
        filenames: Vec<String>,

        /// Analysis mode
        #[arg(short, long, default_value = "full")]
        mode: AnalysisMode,

        /// Fail on first error
        #[arg(long, default_value_t = false)]
        strict: bool,

        /// Project root to write to SCIP index
        #[arg(short, long, default_value = "./")]
        project_root: String,
    },
}

struct Options {
    analysis_mode: AnalysisMode,
    strict: bool,
}

pub fn main() {
    let cli = Cli::parse();

    match cli.command {
        Commands::Index {
            language,
            out,
            filenames,
            mode,
            strict,
            project_root,
        } => index_command(
            language,
            filenames,
            PathBuf::from(out),
            PathBuf::from(project_root),
            Options {
                analysis_mode: mode,
                strict,
            },
        ),
    }
}

fn index_command(
    language: String,
    filenames: Vec<String>,
    out: PathBuf,
    project_root: PathBuf,
    options: Options,
) {
    let p = BundledParser::get_parser(&language).unwrap();

    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-treesitter".to_string(),
                version: clap::crate_version!().to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: format!(
                "file://{}",
                project_root
                    .canonicalize()
                    .expect("Failed to canonicalize project root")
                    .display()
            ),
            ..Default::default()
        })
        .into(),
        ..Default::default()
    };

    for (_, filename) in filenames.iter().enumerate() {
        let contents = std::fs::read(filename).unwrap();
        eprintln!("Processing {filename}");
        match index_content(contents, p, &options) {
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

    eprintln!(
        "Writing index for {} documents into {}",
        index.documents.len(),
        out.display()
    );

    write_message_to_file(out, index).expect("to write the file");
}

fn index_content(contents: Vec<u8>, parser: BundledParser, options: &Options) -> Result<Document> {
    let mut document: Document;

    if options.analysis_mode.globals() {
        document = get_symbols(parser, &contents).unwrap();
    } else {
        document = Document::new();
    }

    if options.analysis_mode.locals() {
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
    use assert_cmd::prelude::*;
    use std::process::Command;
    use std::{env::temp_dir, path::Path};

    #[test]
    fn e2e() {
        let mut cmd: Command;
        match std::env::var("SCIP_CLI_LOCATION") {
            Ok(va) => cmd = {
                let cwd = std::env::current_dir().unwrap().join(va);
                Command::new(cwd)
            },
            _ => cmd = Command::cargo_bin("scip-treesitter-cli").unwrap(),
        }

        println!("{:?}", std::env::var("SCIP_CLI_LOCATION"));

        let out_dir = temp_dir();
        let path = out_dir.join("globals.java");
        let out_path = out_dir
            .join("index-java.scip")
            .to_str()
            .unwrap()
            .to_string();

        write_file(&path, include_str!("../testdata/globals.java").to_string());


        cmd.args(["index", "-l", "java", "-o", &out_path, "globals.java"])
            .current_dir(out_dir)
            .assert()
            .success();

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
