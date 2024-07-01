use std::{
    cell::RefCell,
    collections::{hash_map::Entry, HashMap},
    env,
    fs::File,
    io::{self, prelude::*},
};

use anyhow::{anyhow, bail, Context, Result};
use camino::{Utf8Path, Utf8PathBuf};
use clap::ValueEnum;
use path_clean;
use scip::{types::Document, write_message_to_file};
use syntax_analysis::{
    globals,
    languages::{get_local_configuration, get_tag_configuration},
    locals,
};
use tree_sitter;
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

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum, Debug)]
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
    /// Data is streamed from STDIN
    Stdin,

    /// Data is read from a .tar file
    File { location: Utf8PathBuf },
}

pub enum IndexMode {
    /// Index only this list of files, without checking file extensions
    Files { list: Vec<String> },
    /// Discover all files that can be handled by the chosen language
    /// in the passed location (which has to be a directory)
    Workspace { location: Utf8PathBuf },

    /// Discover all files that can be handled by the chosen language
    /// in either a .tar file, or from STDIN to which TAR data is streamed
    TarArchive { input: TarMode },
}

fn make_absolute(cwd: &Utf8Path, path: &Utf8Path) -> Utf8PathBuf {
    if path.is_absolute() {
        path.to_owned()
    } else {
        Utf8PathBuf::from_path_buf(path_clean::clean(cwd.join(path).as_std_path()))
            .expect("cleaning a path should not change its utf8ness")
    }
}

pub fn index_command(
    language: String,
    index_mode: IndexMode,
    out: &Utf8Path,
    project_root: &Utf8Path,
    evaluate_against: Option<Utf8PathBuf>,
    options: IndexOptions,
) -> Result<()> {
    let parser_id = ParserId::from_name(&language)
        .context(format!("No parser found for language {language}"))?;

    let cwd = Utf8PathBuf::from_path_buf(
        env::current_dir().context("Failed to get the current working directory")?,
    )
    .map_err(|_| anyhow!("Non utf8 current directory"))?;
    let absolute_project_root = make_absolute(
        &cwd,
        match &index_mode {
            IndexMode::Workspace { location } => location,
            _ => project_root,
        },
    );

    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-syntax".to_string(),
                version: clap::crate_version!().to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: format!("file://{absolute_project_root}"),
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
                bar.set_message(filename.clone());
                let filepath = make_absolute(&cwd, &Utf8PathBuf::from(filename));
                let document = index_file(&filepath, parser_id, &absolute_project_root, &options)?;
                index.documents.push(document);
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
                let Some(filepath) = Utf8Path::from_path(entry.path()) else {
                    continue;
                };
                let Some(extension) = filepath.extension() else {
                    continue;
                };
                if extensions.contains(extension) {
                    bar.set_message(entry.path().display().to_string());
                    let document =
                        index_file(filepath, parser_id, &absolute_project_root, &options)?;
                    index.documents.push(document);
                    bar.tick();
                }
            }
        }
    }

    eprintln!(
        "\nWriting index for {} documents into {out}",
        index.documents.len(),
    );

    if let Some(file) = evaluate_against {
        eprintln!("Evaluating built index against {file}");

        let ground_truth = read_index_from_file(&file)?;

        let mut evaluator = Evaluator::default();
        evaluator
            .evaluate_indexes(&index, &ground_truth)?
            .write_summary(&mut std::io::stdout(), Default::default())?
    }

    write_message_to_file(out, index)
        .map_err(|err| anyhow!("{err:?}"))
        .with_context(|| format!("When writing index to {out}"))
}

fn index_file(
    filepath: &Utf8Path,
    parser_id: ParserId,
    absolute_project_root: &Utf8Path,
    options: &IndexOptions,
) -> Result<Document> {
    let contents = std::fs::read_to_string(filepath)
        .with_context(|| format!("Failed to read file at {filepath}"))?;

    let relative_path = filepath
        .strip_prefix(absolute_project_root)
        .with_context(|| {
            format!(
                "Failed to strip project root prefix: root={absolute_project_root} file={filepath}"
            )
        })?;

    match index_content(&contents, parser_id, options) {
        Ok(mut document) => {
            document.relative_path = relative_path.to_string();
            Ok(document)
        }
        Err(error) => {
            bail!("Failed to index {filepath}: {error:?}")
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
        let Ok(path) = Utf8PathBuf::from_path_buf(e.path()?.to_path_buf()) else {
            eprintln!("Failed to convert path to utf8: {:?}", e.path()?);
            continue;
        };

        if matches!(path.extension(), Some(ext) if extensions.contains(ext)) {
            match e.read_to_string(&mut contents) {
                Ok(size) => {
                    match index_content(&contents, parser, options) {
                        Ok(mut document) => {
                            document.relative_path = path.to_string();
                            documents.push(document);
                        }
                        Err(error) => {
                            if options.fail_fast {
                                anyhow::bail!("Failed to index {path}: {error:?}");
                            } else {
                                eprintln!("Failed to index {path}: {error:?}");
                            }
                        }
                    }
                    if size > 0 {
                        contents.clear();
                    }
                }
                Err(error) => {
                    if options.fail_fast {
                        anyhow::bail!("Failed to read contents of path {path}: {error:?}",)
                    } else {
                        eprintln!("Failed to read contents of path {path}: {error:?}",);
                    }
                }
            }

            progress += 1;
            spinner.set_message(format!("[{progress}]: {path}"));
            spinner.tick();
        }
    }

    Ok(documents)
}

thread_local! {
    // We only want to initialize one parser per language per thread
    static PARSERS: RefCell<HashMap<ParserId, tree_sitter::Parser>> = RefCell::new(HashMap::new());
}

fn index_content(contents: &str, parser_id: ParserId, options: &IndexOptions) -> Result<Document> {
    PARSERS.with_borrow_mut(|parsers| {
        let parser = match parsers.entry(parser_id) {
            Entry::Occupied(entry) => {
                let p = entry.into_mut();
                // tree-sitter parsing is stateful, so reset the parser state explicitly
                p.reset();
                p
            }
            Entry::Vacant(v) => v.insert(parser_id.get_parser()),
        };
        let tree = parser
            .parse(contents.as_bytes(), None)
            .ok_or(anyhow!("Failed to parse when indexing content"))?;

        let mut document = if options.analysis_mode.globals() {
            let tag_config = get_tag_configuration(parser_id)
                .ok_or_else(|| anyhow!("No tag configuration for language: {parser_id:?}"))?;
            let (mut scope, hint) = globals::parse_tree(tag_config, &tree, contents)?;
            scope.into_document(hint, vec![])
        } else {
            Document::new()
        };

        if options.analysis_mode.locals() {
            let config = get_local_configuration(parser_id)
                .ok_or_else(|| anyhow!("No local configuration for language: {parser_id:?}"))?;
            let occurrences = locals::find_locals(config, &tree, contents)?;
            document.occurrences.extend(occurrences)
        }
        Ok(document)
    })
}
