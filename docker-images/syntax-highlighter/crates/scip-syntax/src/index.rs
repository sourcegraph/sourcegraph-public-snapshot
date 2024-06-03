use std::path::{Path, PathBuf};

use anyhow::{anyhow, Context, Result};
use clap::ValueEnum;
use scip::{types::Document, write_message_to_file};
use syntax_analysis::{get_globals, get_locals};
use tree_sitter_all_languages::ParserId;

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
) -> Result<()> {
    let p = ParserId::from_name(&language).unwrap();
    let project_root = {
        match index_mode {
            IndexMode::Files { .. } => project_root,
            IndexMode::Workspace { ref location } => location.clone(),
        }
    };

    let canonical_project_root = project_root.canonicalize().with_context(|| {
        format!(
            "Failed to canonicalize project root: {}",
            project_root.display()
        )
    })?;

    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-syntax".to_string(),
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

    let mut index_file = |filepath: &Path| -> Result<()> {
        let contents = std::fs::read_to_string(filepath)
            .with_context(|| format!("Failed to read file at {}", filepath.display()))?;
        let filepath = filepath
            .canonicalize()
            .with_context(|| format!("Failed to canonicalize file path: {}", filepath.display()))?;
        let relative_path = filepath
            .strip_prefix(canonical_project_root.clone())
            .with_context(|| {
                format!(
                    "Failed to strip project root prefix: root={} file={}",
                    canonical_project_root.display(),
                    filepath.display()
                )
            })?;

        match index_content(&contents, p, &options) {
            Ok(mut document) => {
                document.relative_path = relative_path.display().to_string();
                index.documents.push(document);
                Ok(())
            }
            Err(error) => {
                if options.fail_fast {
                    Err(anyhow!(
                        "Failed to index {}: {:?}",
                        filepath.display(),
                        error
                    ))
                } else {
                    eprintln!("Failed to index {}: {:?}", filepath.display(), error);
                    Ok(())
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
                index_file(&filepath)?;
                bar.inc(1);
            }

            bar.finish();
        }
        IndexMode::Workspace { location } => {
            let extensions = ParserId::language_extensions(&p);
            let bar = create_spinner();

            for entry in walkdir::WalkDir::new(location) {
                let Ok(entry) = entry else { continue };
                if entry.file_type().is_dir() {
                    continue;
                }
                let Some(extension) = entry.path().extension().and_then(|p| p.to_str()) else {
                    continue;
                };
                if extensions.contains(extension) {
                    bar.set_message(entry.path().display().to_string());
                    index_file(&entry.into_path())?;
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

        let ground_truth = read_index_from_file(&file)?;

        let mut evaluator = Evaluator::default();
        evaluator
            .evaluate_indexes(&index, &ground_truth)?
            .write_summary(&mut std::io::stdout(), Default::default())?
    }

    write_message_to_file(out.clone(), index)
        .map_err(|err| anyhow!("{err:?}"))
        .with_context(|| format!("When writing index to {}", out.display()))
}

fn index_content(contents: &str, parser: ParserId, options: &IndexOptions) -> Result<Document> {
    let mut document: Document;

    if options.analysis_mode.globals() {
        let (mut scope, hint) = get_globals(parser, contents)?;
        document = scope.into_document(hint, vec![]);
    } else {
        document = Document::new();
    }

    if options.analysis_mode.locals() {
        let occurrences = get_locals(parser, contents)?;
        document.occurrences.extend(occurrences)
    }

    Ok(document)
}
