use clap::{Parser, Subcommand};
use scip_treesitter_cli::{
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
        } => scip_treesitter_cli::evaluate::evaluate_command(
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
