use clap::{Parser, Subcommand, ValueEnum};
use indicatif::{ProgressBar, ProgressStyle};
use protobuf::Message;
use scip::types::Document;
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter_languages::parsers::BundledParser;

use anyhow::Result;
use std::path::PathBuf;

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
        self == AnalysisMode::Locals || self == AnalysisMode::Full
    }
    fn globals(self) -> bool {
        self == AnalysisMode::Globals || self == AnalysisMode::Full
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
        fail_fast: bool,

        /// Project root to write to SCIP index
        #[arg(short, long, default_value = "./")]
        project_root: String,
    },
}

struct Options {
    analysis_mode: AnalysisMode,
    /// When true, fail on first encountered error
    /// Otherwise errors are logged but they don't
    /// interrupt the process
    fail_fast: bool,
}

pub fn main() {
    let cli = Cli::parse();

    match cli.command {
        Commands::Index {
            language,
            out,
            filenames,
            mode,
            fail_fast,
            project_root,
        } => index_command(
            language,
            filenames,
            PathBuf::from(out),
            PathBuf::from(project_root),
            Options {
                analysis_mode: mode,
                fail_fast,
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
                name: "scip-treesitter-cli".to_string(),
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

    let bar = ProgressBar::new(filenames.len() as u64);

    bar.set_style(
        ProgressStyle::with_template(
            "[{elapsed_precise}] {bar:40.cyan/blue} {pos:>7}/{len:7}\n {msg}",
        )
        .unwrap()
        .progress_chars("##-"),
    );

    for (_, filename) in filenames.iter().enumerate() {
        let contents = std::fs::read(filename).unwrap();
        bar.set_message(filename.clone());
        bar.inc(1);
        match index_content(contents, p, &options) {
            Ok(mut document) => {
                document.relative_path = filename.to_string();
                index.documents.push(document);
            }
            Err(error) => {
                if options.fail_fast {
                    panic!("Failed to index {filename}: {:?}", error);
                } else {
                    eprintln!("Failed to index {filename}: {:?}", error)
                }
            }
        }
    }

    bar.finish();

    eprintln!();

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
            Some(Err(e)) => return Err(e),
            None => {}
        }
    }

    Ok(document)
}

fn write_message_to_file<P>(path: P, msg: impl Message) -> Result<(), Box<dyn std::error::Error>>
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

#[cfg(test)]
mod tests {
    use assert_cmd::cargo::cargo_bin;
    use assert_cmd::prelude::*;
    use protobuf::{CodedInputStream, Message};
    use std::collections::HashMap;
    use std::fs::File;
    use std::io::BufReader;
    use std::path::Path;
    use std::process::Command;
    use std::{env::temp_dir, path::PathBuf};

    lazy_static::lazy_static! {
        static ref BINARY_LOCATION: PathBuf = {
            match std::env::var("SCIP_CLI_LOCATION") {
                Ok(va) => std::env::current_dir().unwrap().join(va),
                _ => cargo_bin("scip-treesitter-cli"),
            }
        };
    }

    use scip_treesitter::snapshot::{dump_document_with_config, EmitSymbol, SnapshotOptions};

    fn read_index_from_file(file: PathBuf) -> scip::types::Index {
        let mut candidate_idx = scip::types::Index::new();
        let candidate_f = File::open(file).unwrap();

        let mut reader = BufReader::new(candidate_f);
        let mut cis = CodedInputStream::from_buf_read(&mut reader);

        candidate_idx.merge_from(&mut cis).unwrap();
        candidate_idx
    }

    fn snapshot_syntax_document(doc: &scip::types::Document, source: &str) -> String {
        dump_document_with_config(
            doc,
            source,
            SnapshotOptions {
                emit_symbol: EmitSymbol::All,
                ..Default::default()
            },
        )
        .expect("dump document")
    }

    #[test]
    fn java_e2e() {
        let out_dir = temp_dir();
        let setup = HashMap::from([(
            PathBuf::from("globals.java"),
            include_str!("../testdata/globals.java").to_string(),
        )]);

        run_index(&out_dir, &setup, vec!["--language", "java"]);

        let index = read_index_from_file(out_dir.join("index.scip"));

        for doc in &index.documents {
            let path = &doc.relative_path;
            let dumped =
                snapshot_syntax_document(doc, setup.get(&PathBuf::from(&path)).expect("??"));

            insta::assert_snapshot!(path.clone(), dumped);
        }
    }

    fn prepare(temp: &Path, files: &HashMap<PathBuf, String>) {
        for (path, contents) in files.iter() {
            let file_path = temp.join(path);
            write_file(&file_path, contents);
        }
    }

    fn run_index(location: &PathBuf, files: &HashMap<PathBuf, String>, extra_arguments: Vec<&str>) {
        prepare(location, files);

        let mut base_args = vec!["index"];
        base_args.extend(extra_arguments);

        let mut cmd = Command::new(BINARY_LOCATION.to_str().unwrap());

        cmd.args(base_args);

        for (path, _) in files.iter() {
            cmd.arg(path.to_str().unwrap());
        }

        cmd.current_dir(location);

        cmd.assert().success();
    }

    fn write_file(path: &PathBuf, contents: &String) {
        use std::io::Write;

        let output = std::fs::File::create(path).unwrap();
        let mut writer = std::io::BufWriter::new(output);
        writer.write_all(contents.as_bytes()).unwrap();
    }
}
