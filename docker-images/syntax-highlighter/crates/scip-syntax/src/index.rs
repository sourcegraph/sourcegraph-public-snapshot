use std::{
    fs::File,
    path::{Path, PathBuf},
};

use anyhow::{anyhow, bail, Context, Result};
use clap::ValueEnum;
use scip::{types::Document, write_message_to_file};
use std::io::{self, prelude::*};
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

pub enum TarMode {
    Stdin,
    File { location: PathBuf },
}

pub enum IndexMode {
    /// Index only this list of files, without checking file extensions
    Files {
        list: Vec<String>,
    },
    /// Discover all files that can be handled by the chosen language
    /// in the passed location (which has to be a directory)
    Workspace {
        location: PathBuf,
    },

    TarArchive {
        input: TarMode,
    },
}

pub fn index_command(
    language: String,
    index_mode: IndexMode,
    out: PathBuf,
    project_root: PathBuf,
    evaluate_against: Option<PathBuf>,
    options: IndexOptions,
) -> Result<()> {
    let parser_id = ParserId::from_name(&language)
        .context(format!("No parser found for language {language}"))?;

    let project_root = {
        match index_mode {
            IndexMode::Files { .. } => project_root,
            IndexMode::TarArchive { .. } => project_root,
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

    let extensions = ParserId::language_extensions(&parser_id);

    match index_mode {
        IndexMode::Files { list } => {
            let bar = create_progress_bar(list.len() as u64);
            for filename in list {
                let filepath = PathBuf::from(filename).canonicalize()?;
                bar.set_message(filepath.display().to_string());
                index.documents.push(index_file(
                    &filepath,
                    parser_id,
                    &canonical_project_root,
                    &options,
                )?);
                bar.inc(1);
            }

            bar.finish();
        }
        IndexMode::TarArchive { input } => match input {
            TarMode::File { location } => {
                let mut ar = tar::Archive::new(File::open(location)?);
                let entries = ar.entries()?;
                let documents = index_tar_entries(entries, parser_id, &options)?;
                index.documents.extend(documents);
            }
            TarMode::Stdin => {
                let stdin = io::stdin();
                let mut ar: tar::Archive<_> = tar::Archive::new(stdin);
                let entries = ar.entries()?;
                let documents = index_tar_entries(entries, parser_id, &options)?;
                index.documents.extend(documents);
            }
        },
        IndexMode::Workspace { location } => {
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
                    index.documents.push(index_file(
                        &entry.into_path(),
                        parser_id,
                        &canonical_project_root,
                        &options,
                    )?);
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

fn index_file(
    filepath: &Path,
    parser_id: ParserId,
    canonical_project_root: &PathBuf,
    options: &IndexOptions,
) -> Result<Document> {
    let contents = std::fs::read_to_string(filepath)
        .with_context(|| format!("Failed to read file at {}", filepath.display()))?;

    // TODO(Anton&Christoph): revise this logic. currently uncommented version is the only one that
    // passes tests on MacOS
    //let filepath = if filepath.is_absolute() {
    //    filepath.to_owned()
    //} else {
    //    filepath
    //        .canonicalize()
    //        .with_context(|| format!("Failed to canonicalize file path: {}", filepath.display()))?
    //};
    //
    let filepath = filepath
        .canonicalize()
        .with_context(|| format!("Failed to canonicalize file path: {}", filepath.display()))?;
    // end TODO

    let relative_path = filepath
        .strip_prefix(canonical_project_root.clone())
        .with_context(|| {
            format!(
                "Failed to strip project root prefix: root={} file={}",
                canonical_project_root.display(),
                filepath.display()
            )
        })?;

    match index_content(&contents, parser_id, &options) {
        Ok(mut document) => {
            document.relative_path = relative_path.display().to_string();
            Ok(document)
        }
        Err(error) => {
            bail!("Failed to index {}: {:?}", filepath.display(), error)
        }
    }
}

fn index_tar_entries<R: Read>(
    entries: tar::Entries<'_, R>,
    parser: ParserId,
    options: &IndexOptions,
) -> anyhow::Result<Vec<Document>> {
    let extensions = ParserId::language_extensions(&parser);
    let mut contents = String::new();
    let mut documents: Vec<Document> = vec![];
    let mut progress = 0;
    let spinner = create_spinner();
    for entry in entries {
        let mut e = entry?;
        let path = PathBuf::from(e.path()?);

        if matches!(path.extension().and_then(|e| e.to_str()), Some(ext) if extensions.contains(ext))
        {
            match e.read_to_string(&mut contents) {
                Ok(size) => {
                    match index_content(&contents, parser, &options) {
                        Ok(mut document) => {
                            document.relative_path = path.display().to_string();
                            documents.push(document);
                        }
                        Err(error) => {
                            if options.fail_fast {
                                anyhow::bail!("Failed to index {}: {:?}", path.display(), error);
                            } else {
                                eprintln!("Failed to index {}: {:?}", path.display(), error);
                            }
                        }
                    }
                    if size > 0 {
                        contents.clear();
                    }
                }
                Err(error) => {
                    if options.fail_fast {
                        anyhow::bail!(
                            "Failed to read contents of path {}: {:?}",
                            path.display(),
                            error
                        )
                    } else {
                        eprintln!(
                            "Failed to read contents of path {}: {:?}",
                            path.display(),
                            error
                        );
                    }
                }
            }

            progress += 1;
            spinner.set_message(format!("[{}]: {}", progress, path.display().to_string()));
            spinner.tick();
        }
    }

    Ok(documents)
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
