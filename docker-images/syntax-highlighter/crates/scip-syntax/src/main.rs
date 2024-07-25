use std::{num::NonZeroUsize, process};

use camino::{Utf8Path, Utf8PathBuf};
use clap::{Parser, Subcommand};
use scip_syntax::index::{index_command, AnalysisFeatures, IndexMode, IndexOptions, TarMode};

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
#[command(propagate_version = true)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Parser, Clone, Debug)]
struct AnalysisFeaturesOptions {
    #[arg(long, default_value_t = false)]
    no_global_references: bool,
    #[arg(long, default_value_t = false)]
    no_locals: bool,
    #[arg(long, default_value_t = false)]
    no_global_definitions: bool,
}

#[derive(Parser, Clone, Debug)]
struct IndexCommandOptions {
    /// Which language parser to use to process the files
    #[arg(short, long)]
    language: String,

    /// Path where the SCIP index will be written
    #[arg(short, long, default_value = "./index.scip")]
    out: String,

    /// Analysis features
    #[command(flatten)]
    analysis: AnalysisFeaturesOptions,

    /// Fail on first error
    #[arg(long, default_value_t = false)]
    fail_fast: bool,

    /// Project root to write to SCIP index
    #[arg(short, long, default_value = "./")]
    project_root: String,

    /// Number of jobs to run in parallel, defaults to number of logical cores
    #[arg(short, long)]
    jobs: Option<NonZeroUsize>,

    /// Evaluate the build index against an index from a file
    #[arg(long)]
    evaluate: Option<String>,
}

#[derive(Subcommand, Debug)]
enum IndexCommand {
    /// Index a folder, automatically detecting files
    /// to be processed by the chosen language
    Workspace {
        /// Folder to index - will be chosen as project root,
        /// and files will be discovered according to
        /// configured extensions for the selected language
        /// Has to be absolute path.
        dir: String,

        #[command(flatten)]
        options: IndexCommandOptions,
    },

    /// Index a list of files
    Files {
        /// List of files to analyse
        filenames: Vec<String>,

        #[command(flatten)]
        options: IndexCommandOptions,
    },

    /// Index a .tar archive, either from a file or streaming from STDIN
    Tar {
        /// Either a path to .tar file, or "-" to read .tar data from STDIN
        tar: String,

        #[command(flatten)]
        options: IndexCommandOptions,
    },
}
#[derive(Parser, Debug)]
struct IndexCommandParser {
    #[structopt(subcommand)]
    index_command: IndexCommand,
}

#[derive(Subcommand)]
enum Commands {
    /// Index source files using Tree Sitter parser for a given language
    /// and produce a SCIP file
    #[clap(name = "index")]
    Index(IndexCommandParser),

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
        #[arg(long, default_value_t = false, long = "no-color")]
        disable_colors: bool,
    },
}

pub fn main() -> anyhow::Result<()> {
    // Exits with a code zero if the environment variable SANITY_CHECK equals
    // to "true". This enables testing that the current program is in a runnable
    // state against the platform it's being executed on.
    //
    // See https://github.com/GoogleContainerTools/container-structure-test
    match std::env::var("SANITY_CHECK") {
        Ok(v) if v == "true" => {
            println!("Sanity check passed, exiting without error");
            std::process::exit(0)
        }
        _ => {}
    };

    let cli = Cli::parse();

    match cli.command {
        Commands::Index(index1) => {
            let result = match index1.index_command {
                IndexCommand::Files { filenames, options } => {
                    if filenames.is_empty() {
                        eprintln!("List of files cannot be empty");
                        process::exit(1)
                    }
                    run_index_command(options, IndexMode::Files { list: filenames })
                }
                IndexCommand::Workspace { dir, options } => run_index_command(
                    options,
                    IndexMode::Workspace {
                        location: dir.into(),
                    },
                ),

                IndexCommand::Tar { tar, options } => {
                    if tar == "-" {
                        run_index_command(
                            options,
                            IndexMode::TarArchive {
                                input: scip_syntax::index::TarMode::Stdin,
                            },
                        )
                    } else {
                        run_index_command(
                            options,
                            IndexMode::TarArchive {
                                input: TarMode::File {
                                    location: Utf8PathBuf::from(tar),
                                },
                            },
                        )
                    }
                }
            };

            result?
        }

        Commands::ScipEvaluate {
            candidate,
            ground_truth,
            print_mapping,
            print_true_positives,
            print_false_positives,
            print_false_negatives,
            disable_colors,
        } => scip_syntax::evaluate::evaluate_command(
            Utf8Path::new(&candidate),
            Utf8Path::new(&ground_truth),
            scip_syntax::evaluate::EvaluationOutputOptions {
                print_mapping,
                print_true_positives,
                print_false_positives,
                print_false_negatives,
                disable_colors,
            },
        )?,
    }
    Ok(())
}

fn run_index_command(options: IndexCommandOptions, mode: IndexMode) -> anyhow::Result<()> {
    index_command(
        options.language,
        mode,
        Utf8Path::new(&options.out),
        Utf8Path::new(&options.project_root),
        options.evaluate.map(Utf8PathBuf::from),
        options.jobs,
        IndexOptions {
            analysis_features: AnalysisFeatures {
                locals: !options.analysis.no_locals,
                global_references: !options.analysis.no_global_references,
                global_definitions: !options.analysis.no_global_definitions,
            },
            fail_fast: options.fail_fast,
        },
    )
}
