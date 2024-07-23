use std::{
    cell::RefCell,
    collections::{hash_map::Entry, HashMap},
    env,
    fs::File,
    io::{self, Read},
    num::NonZeroUsize,
    sync::{
        atomic::{AtomicU32, Ordering},
        mpsc::{sync_channel, Receiver},
    },
    thread,
};

use anyhow::{anyhow, Context, Result};
use camino::{Utf8Path, Utf8PathBuf};
use path_clean;
use rayon::{prelude::*, ThreadPoolBuilder};
use scip::{types::Document, write_message_to_file};
use syntax_analysis::{
    globals,
    languages::{get_local_configuration, get_tag_configuration},
    locals::{self, LocalResolutionOptions},
    SCIP_SYNTAX_SCHEME,
};
use tree_sitter;
use tree_sitter_all_languages::ParserId;

use crate::{evaluate::Evaluator, io::read_index_from_file, progress::create_spinner};

#[derive(Debug, Copy, Clone)]
pub struct IndexOptions {
    pub analysis_features: AnalysisFeatures,
    /// When true, fail on first encountered error
    /// Otherwise errors are logged but they don't
    /// interrupt the process
    pub fail_fast: bool,
}

#[derive(Debug, Copy, Clone, PartialEq, Eq)]
pub struct AnalysisFeatures {
    pub locals: bool,
    pub global_references: bool,
    pub global_definitions: bool,
}

impl Default for AnalysisFeatures {
    fn default() -> Self {
        AnalysisFeatures {
            locals: true,
            global_references: true,
            global_definitions: true,
        }
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

#[derive(Debug)]
struct IndexJob {
    path_to_index_root: Utf8PathBuf,
    // NOTE(Christoph): For both 'workspace' and 'files' mode we would be able to
    // read the files in the worker threads.
    // When benchmarking on my M2, this ended being slower than reading
    // the files in the job producers.
    // As we only care about tar mode in production, I've decided all producers need
    // to read the files.
    contents: String,
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
    worker_count: Option<NonZeroUsize>,
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

    let job_queue = match index_mode {
        IndexMode::Files { list } => files_producer(&cwd, &absolute_project_root, list),
        IndexMode::Workspace { location } => {
            workspace_producer(&absolute_project_root, location, parser_id)
        }
        IndexMode::TarArchive { input } => match input {
            TarMode::File { location } => tar_producer(File::open(location)?, parser_id),
            TarMode::Stdin => tar_producer(io::stdin(), parser_id),
        },
    };
    let documents = if let Some(worker_count) = worker_count {
        let pool = ThreadPoolBuilder::new()
            .num_threads(worker_count.into())
            .build()
            .context("failed to initialize ThreadPool")?;
        pool.install(|| process_jobs(job_queue, parser_id, options))
    } else {
        process_jobs(job_queue, parser_id, options)
    };
    index.documents.extend(documents?);

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

fn files_producer(
    cwd: &Utf8Path,
    absolute_project_root: &Utf8Path,
    list: Vec<String>,
) -> Receiver<IndexJob> {
    let (tx, rx) = sync_channel(rayon::current_num_threads() * 2);
    let cwd = cwd.to_path_buf();
    let absolute_project_root = absolute_project_root.to_path_buf();
    thread::spawn(move || {
        list.into_iter()
            .filter_map(|filename| {
                let path = make_absolute(&cwd, &Utf8PathBuf::from(filename));
                let relative_path = path
                    .strip_prefix(&absolute_project_root)
                    .with_context(|| {
                        format!(
                    "Failed to strip project root prefix: root={absolute_project_root} file={path}"
                )
                    })
                    .ok()?;
                let contents = std::fs::read_to_string(&path)
                    .with_context(|| format!("Failed to read file at {path}"))
                    .ok()?;
                Some(IndexJob {
                    path_to_index_root: relative_path.to_path_buf(),
                    contents,
                })
            })
            .for_each(|x| tx.send(x).unwrap());
    });
    rx
}

fn workspace_producer(
    absolute_project_root: &Utf8Path,
    location: Utf8PathBuf,
    parser_id: ParserId,
) -> Receiver<IndexJob> {
    let (tx, rx) = sync_channel(rayon::current_num_threads() * 2);
    let absolute_project_root = absolute_project_root.to_path_buf();
    thread::spawn(move || {
        let extensions = ParserId::language_extensions(&parser_id);
        walkdir::WalkDir::new(location).into_iter().filter_map(move |entry| {
              // TODO: Skipping any entry we can't read (mostly for non-utf-8 reasons)
              // Do we want to log these cases?
              let entry = entry.ok()?;
              let path = Utf8Path::from_path(entry.path())?;
              if !path.is_file() || !extensions.contains(path.extension()?) {
                  return None;
              }
              let relative_path = path
                  .strip_prefix(&absolute_project_root)
                  .with_context(|| {
                      format!(
                          "Failed to strip project root prefix: root={absolute_project_root} file={path}"
                      )
                  }).ok()?;
              let contents = std::fs::read_to_string(path)
                  .with_context(|| format!("Failed to read file at {path}"))
                  .ok()?;
                Some(IndexJob {
                    path_to_index_root: relative_path.to_path_buf(),
                    contents,
                })
      }).for_each(|x| tx.send(x).unwrap());
    });
    rx
}

fn tar_entry_path<R: Read>(entry: &tar::Entry<'_, R>) -> Result<Utf8PathBuf> {
    let path = entry.path()?;
    Utf8PathBuf::from_path_buf(path.to_path_buf()).map_err(|_| anyhow!("Non utf-8 path"))
}

fn tar_entry_contents<R: Read>(mut entry: tar::Entry<'_, R>) -> Result<String> {
    let mut contents = String::with_capacity(entry.size() as usize);
    entry.read_to_string(&mut contents)?;
    Ok(contents)
}

fn tar_producer<R: Read + Send + 'static>(reader: R, parser_id: ParserId) -> Receiver<IndexJob> {
    let (tx, rx) = sync_channel(rayon::current_num_threads() * 2);
    thread::spawn(move || -> Result<()> {
        let extensions = ParserId::language_extensions(&parser_id);
        let mut ar: tar::Archive<_> = tar::Archive::new(reader);
        // The only way .entries() fails is if the tar archive is not at position 0.
        // As we just created it that cannot happen.
        let entries = ar.entries().expect("Failed to read tar entries");
        entries
            .filter_map(|entry| {
                // TODO: Skipping any entry we can't read (mostly for non-utf-8 reasons)
                // Do we want to log these cases?
                let entry = entry.ok()?;
                let path = tar_entry_path(&entry).ok()?;
                if !extensions.contains(path.extension()?) {
                    return None;
                }
                let contents = tar_entry_contents(entry).ok()?;
                Some(IndexJob {
                    path_to_index_root: path,
                    contents,
                })
            })
            .for_each(|x| tx.send(x).unwrap());
        Ok(())
    });
    rx
}

fn process_jobs(
    rx: Receiver<IndexJob>,
    parser_id: ParserId,
    options: IndexOptions,
) -> Result<Vec<Document>> {
    let spinner = create_spinner();
    let progress: AtomicU32 = AtomicU32::new(1);
    rx.into_iter()
        .par_bridge()
        .filter_map(
            |IndexJob {
                 path_to_index_root,
                 contents,
             }| {
                let progress = progress.fetch_add(1, Ordering::Relaxed);
                spinner.set_message(format!("[{progress}]: {path_to_index_root}"));
                match index_content(&contents, parser_id, options) {
                    Ok(mut document) => {
                        spinner.tick();
                        document.relative_path = path_to_index_root.to_string();
                        Some(Ok(document))
                    }
                    Err(error) => {
                        if options.fail_fast {
                            Some(Err(anyhow!(
                                "failed to index {path_to_index_root}: {error:?}"
                            )))
                        } else {
                            eprintln!("failed to index {path_to_index_root}: {error:?}");
                            None
                        }
                    }
                }
            },
        )
        .collect()
}

thread_local! {
    // We only want to initialize one parser per language per thread
    static PARSERS: RefCell<HashMap<ParserId, tree_sitter::Parser>> = RefCell::new(HashMap::new());
}

fn index_content(contents: &str, parser_id: ParserId, options: IndexOptions) -> Result<Document> {
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

        let mut document = if options.analysis_features.global_definitions {
            let tag_config = get_tag_configuration(parser_id)
                .ok_or_else(|| anyhow!("No tag configuration for language: {parser_id:?}"))?;
            let (mut scope, hint) = globals::parse_tree(tag_config, &tree, contents)?;
            scope.into_document(hint, SCIP_SYNTAX_SCHEME, vec![])
        } else {
            Document::new()
        };

        if options.analysis_features.locals {
            let config = get_local_configuration(parser_id)
                .ok_or_else(|| anyhow!("No local configuration for language: {parser_id:?}"))?;
            let options = LocalResolutionOptions {
                emit_global_references: options.analysis_features.global_references,
            };
            let occurrences = locals::find_locals(config, &tree, contents, options)?;
            document.occurrences.extend(occurrences)
        }
        Ok(document)
    })
}
