mod evaluate;
mod index;
mod io;
mod progress;

use clap::{Parser, Subcommand};
use crate::{
    evaluate::ScipEvaluateOptions,
    index::{index_command, AnalysisMode, IndexMode, IndexOptions},
};
use std::path::PathBuf;

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
#[command(propagate_version = true)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
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

        /// Folder to index - will be chosen as project root,
        /// and files will be discovered according to
        /// configured extensions for the selected language
        #[arg(long)]
        workspace: Option<String>,

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

        /// Evaluate the build index against an index from a file
        #[arg(long)]
        evaluate: Option<String>,
    },

    /// Fuzzily evaluate candidate SCIP index against known ground truth
    ScipEvaluate {
        #[arg(long)]
        candidate: String,
        #[arg(long)]
        ground_truth: String,
        #[arg(long)]
        print_mapping: bool,
        #[arg(long)]
        print_true_positives: bool,
        #[arg(long)]
        print_false_positives: bool,
        #[arg(long)]
        print_false_negatives: bool,
    },
}

pub fn main() {
    let cli = Cli::parse();

    match cli.command {
        Commands::Index {
            language,
            out,
            filenames,
            workspace,
            mode,
            fail_fast,
            project_root,
            evaluate,
        } => {
            let index_mode = {
                match workspace {
                    None => IndexMode::Files { list: filenames },
                    Some(location) => {
                        if !filenames.is_empty() {
                            panic!("--workspace option cannot be combined with a list of files");
                        } else {
                            IndexMode::Workspace {
                                location: location.into(),
                            }
                        }
                    }
                }
            };

            index_command(
                language,
                index_mode,
                PathBuf::from(out),
                PathBuf::from(project_root),
                evaluate.map(PathBuf::from),
                IndexOptions {
                    analysis_mode: mode,
                    fail_fast,
                },
            )
        }

        Commands::ScipEvaluate {
            candidate,
            ground_truth,
            print_mapping,
            print_true_positives,
            print_false_positives,
            print_false_negatives,
        } => crate::evaluate::evaluate_command(
            PathBuf::from(candidate),
            PathBuf::from(ground_truth),
            ScipEvaluateOptions {
                print_mapping,
                print_true_positives,
                print_false_positives,
                print_false_negatives,
            },
        ),
    }
}

#[cfg(test)]
mod tests {
    use crate::io::read_index_from_file;
    use assert_cmd::cargo::cargo_bin;
    use assert_cmd::prelude::*;
    use std::collections::HashMap;
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
    fn java_e2e_indexing() {
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
