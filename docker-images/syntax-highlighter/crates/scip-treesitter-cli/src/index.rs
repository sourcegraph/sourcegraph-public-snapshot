use std::path::PathBuf;

use anyhow::Result;
use clap::ValueEnum;
use scip::{types::Document, write_message_to_file};
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter_languages::parsers::BundledParser;
use walkdir::DirEntry;

use crate::{
    evaluate::Evaluator,
    io::read_index_from_file,
    progress::{create_progress_bar, create_spinner},
};

pub struct IndexOptions {
    pub analysis_mode: AnalysisMode,
    /// When true, fail on first encountered error
    /// Otherwise errors are logged but they don't
    /// interrupt the process
    pub fail_fast: bool,
}

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum)]
pub enum AnalysisMode {
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

pub enum IndexMode {
    /// Index only this list of files, without checking file extensions
    Files { list: Vec<String> },
    /// Discover all files that can be handled by the chosen language
    /// in the passed location (which has to be a directory)
    Workspace { location: PathBuf },
}

pub fn index_command(
    language: String,
    index_mode: IndexMode,
    out: PathBuf,
    project_root: PathBuf,
    evaluate_against: Option<PathBuf>,
    options: IndexOptions,
) {
    let p = BundledParser::get_parser(&language).unwrap();
    let project_root = {
        match index_mode {
            IndexMode::Files { .. } => project_root,
            IndexMode::Workspace { ref location } => location.clone(),
        }
    };

    let canonical_project_root = project_root
        .canonicalize()
        .expect("Failed to canonicalize project root");

    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-treesitter-cli".to_string(),
                version: clap::crate_version!().to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: format!("file://{}", canonical_project_root.display()),
            ..Default::default()
        })
        .into(),
        ..Default::default()
    };

    let mut index_file = |filepath: &PathBuf| {
        let contents = std::fs::read(filepath).unwrap();
        let relative_path = filepath
            .strip_prefix(canonical_project_root.clone())
            .expect("Failed to strip project root prefix");

        match index_content(contents, p, &options) {
            Ok(mut document) => {
                document.relative_path = relative_path.display().to_string();
                index.documents.push(document);
            }
            Err(error) => {
                if options.fail_fast {
                    panic!("Failed to index {}: {:?}", filepath.display(), error);
                } else {
                    eprintln!("Failed to index {}: {:?}", filepath.display(), error)
                }
            }
        }
    };

    match index_mode {
        IndexMode::Files { list } => {
            let bar = create_progress_bar(list.len() as u64);
            for filename in list {
                let filepath = PathBuf::from(filename).canonicalize().unwrap();
                bar.set_message(filepath.display().to_string());
                index_file(&filepath);
                bar.inc(1);
            }

            bar.finish();
        }
        IndexMode::Workspace { location } => {
            let extensions = BundledParser::get_language_extensions(&p);
            let is_valid = |entry: &DirEntry| {
                entry.file_type().is_dir()
                    || entry
                        .file_name()
                        .to_str()
                        .map(|s| {
                            s.split('.')
                                .last()
                                .filter(|ext| extensions.contains(ext))
                                .is_some()
                        })
                        .unwrap_or(false)
            };

            let bar = create_spinner();

            for entry in walkdir::WalkDir::new(location)
                .into_iter()
                .filter_entry(|e| is_valid(e))
            {
                let entry = entry.unwrap();
                if !entry.file_type().is_dir() {
                    bar.set_message(entry.path().display().to_string());
                    index_file(&entry.into_path());
                    bar.tick();
                }
            }
        }
    }

    eprintln!();

    eprintln!(
        "Writing index for {} documents into {}",
        index.documents.len(),
        out.display()
    );

    if let Some(file) = evaluate_against {
        eprintln!("Evaluating built index against {}", file.display());

        let ground_truth = read_index_from_file(file);

        let mut evaluator = Evaluator::default();
        evaluator
            .evaluate_indexes(&index, &ground_truth, Default::default())
            .unwrap()
            .print_summary();
    }

    write_message_to_file(out, index).expect("to write the file");
}

fn index_content(
    contents: Vec<u8>,
    parser: BundledParser,
    options: &IndexOptions,
) -> Result<Document> {
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
