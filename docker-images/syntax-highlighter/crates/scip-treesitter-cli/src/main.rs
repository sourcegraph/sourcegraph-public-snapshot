use clap::{Parser, Subcommand};
use scip_treesitter_cli::index::{index_command, AnalysisMode, IndexMode, IndexOptions};
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
        /// Which language parser to use to process the files
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
        /// SCIP file to evaluate (refered to as "candidate")
        #[arg(long)]
        candidate: String,

        /// SCIP file to be used as the source of truth (referred to as "ground truth")
        #[arg(long)]
        ground_truth: String,

        /// Print to stdout the mapping between candidate symbols and groun truth symbols
        #[arg(long)]
        print_mapping: bool,

        /// Print all occurrences in candidate SCIP that are matching occurrences in ground truth SCIP
        #[arg(long)]
        print_true_positives: bool,

        /// Print all occurrences in candidate SCIP that don't match any occurrences in ground truth SCIP
        #[arg(long)]
        print_false_positives: bool,

        /// Print all occurrences in ground truth SCIP that don't match any occurrences in candidate SCIP
        #[arg(long)]
        print_false_negatives: bool,

        /// Disable color output
        #[arg(long, default_value_t = false)]
        no_color: bool,
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
            no_color,
        } => scip_treesitter_cli::evaluate::evaluate_command(
            PathBuf::from(candidate),
            PathBuf::from(ground_truth),
            scip_treesitter_cli::evaluate::EvaluationOutputOptions {
                print_mapping,
                print_true_positives,
                print_false_positives,
                print_false_negatives,
                disable_colors: no_color,
            },
        ),
    }
}
