use clap::{Parser, Subcommand, ValueEnum};
use indicatif::{ProgressBar, ProgressStyle};
use protobuf::{CodedInputStream, Message};
use scip::types::{Document, Index};
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter::types::PackedRange;
use scip_treesitter_languages::parsers::BundledParser;

use anyhow::Result;
use colored::*;
use std::{
    collections::{HashMap, HashSet},
    fs::File,
    io::BufReader,
    path::PathBuf,
};
use walkdir::DirEntry;

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
        return self == AnalysisMode::Locals || self == AnalysisMode::Full;
    }
    fn globals(self) -> bool {
        return self == AnalysisMode::Globals || self == AnalysisMode::Full;
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

struct IndexOptions {
    analysis_mode: AnalysisMode,
    /// When true, fail on first encountered error
    /// Otherwise errors are logged but they don't
    /// interrupt the process
    fail_fast: bool,
}

pub struct ScipEvaluateOptions {
    print_mapping: bool,
    print_true_positives: bool,
    print_false_positives: bool,
    print_false_negatives: bool,
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
        } => evaluate_command(
            candidate,
            ground_truth,
            ScipEvaluateOptions {
                print_mapping,
                print_true_positives,
                print_false_positives,
                print_false_negatives,
            },
        ),
    }
}

#[derive(Eq, Hash, PartialEq, Clone, Debug, Ord, PartialOrd)]
struct SymbolPair {
    ground_truth: String,
    candidate: String,
}

#[derive(Debug, PartialEq, Eq, Default, Hash, Clone, PartialOrd, Ord)]
struct LocationInFile {
    rng: PackedRange,
    file: String,
}

#[derive(Eq, Hash, PartialEq, Clone, Debug, Ord, PartialOrd)]
struct SymbolOccurrence {
    location: LocationInFile,
    symbol: String,
}

impl SymbolOccurrence {
    fn range(&self) -> &PackedRange {
        &self.location.rng
    }
}

#[derive(Eq, Hash, PartialEq, Clone, Debug)]
struct Overlap {
    /// Total number of occurrence of a symbol from ground truth
    total: u32,
    /// Number of common occurrences between a ground truth symbol and a candidate
    common: u32,
}

impl Overlap {
    fn jaccard(&self) -> f32 {
        self.common as f32 / self.total as f32
    }
}

pub fn evaluate_command<'a>(candidate: String, ground_truth: String, options: ScipEvaluateOptions) {
    let bar = create_spinner();
    bar.set_message("Indexing candidate symbols by location");
    let candidate_occurrences: HashMap<LocationInFile, String> =
        index_occurrences(&read_index_from_file(candidate.into()));
    bar.tick();

    bar.set_message("Indexing ground truth symbols by location");
    let ground_truth_occurrences: HashMap<LocationInFile, String> =
        index_occurrences(&read_index_from_file(ground_truth.into()));
    bar.tick();

    // For each symbol pair we maintain an Overlap instance
    let mut overlaps: HashMap<SymbolPair, Overlap> = HashMap::new();

    // Lookup table where key is ground truth symbol, and the value
    // is all the symbol pairs.
    // Each symbol from ground truth dataset can be mapped to any number of
    // symbols from the candidate set
    let mut lookup: HashMap<String, HashSet<SymbolPair>> = HashMap::new();

    bar.set_message("Analysing occurrences in candidate SCIP");
    for (candidate_loc, candidate_symbol) in candidate_occurrences.clone() {
        // At given location from the candidate dataset, see
        // if ground truth dataset contains any symbol at same location
        match ground_truth_occurrences.get(&candidate_loc) {
            // If ground truth dataset doesn't have any symbol at this location,
            // we treat it as a false positive, to be handled later
            None => {}
            Some(ground_truth_symbol) => {
                let pair = SymbolPair {
                    ground_truth: ground_truth_symbol.clone(),
                    candidate: candidate_symbol,
                };

                // See if we already have a lookup entry for this ground truth symbol
                match lookup.get_mut(ground_truth_symbol) {
                    None => {
                        // If this is the first time we're seeing this symbol,
                        // create a lookup entry and put a single pair in there
                        lookup.insert(ground_truth_symbol.clone(), HashSet::from([pair.clone()]));
                    }
                    Some(s) => {
                        // Otherwise, add the symbol pair to the set (it might already be there)
                        s.insert(pair.clone());
                    }
                }

                // As we are currently iterating over candidate occurrences,
                // we don't know how many occurrences there are in total in the
                // ground truth dataset.
                // That's why here we only manage the number of occurrences
                // the datasets have in common
                match overlaps.get_mut(&pair) {
                    None => {
                        overlaps.insert(
                            pair,
                            Overlap {
                                total: 0,
                                common: 1,
                            },
                        );
                    }
                    Some(overlap) => overlap.common += 1,
                }
            }
        }
    }
    bar.tick();

    bar.set_message("Computing overlap with ground truth SCIP occurrences");
    for (_, ground_truth_symbol) in &ground_truth_occurrences {
        // now that we're iterating over all the ground truth occurrences,
        // we can update the `total` counter for each symbol pair
        // associated with that ground truth symbol
        lookup
            .get(ground_truth_symbol)
            .into_iter()
            .for_each(|pairs| {
                for pair in pairs {
                    overlaps.get_mut(&pair).map(|x| x.total += 1);
                }
            });
    }
    bar.tick();

    // We have produced the final counts for all symbol pairs -
    // it's time to produce final weights
    let symbol_pair_weight: HashMap<SymbolPair, f32> = overlaps
        .clone()
        .into_iter()
        .map(|(symbol_pair, overlap)| (symbol_pair, overlap.jaccard()))
        .collect();

    if options.print_mapping {
        let mut overlaps_vec: Vec<(SymbolPair, Overlap)> = overlaps.into_iter().collect();
        overlaps_vec.sort_by_key(|(symbol_pair, _)| symbol_pair.clone());

        for (pair, ov) in overlaps_vec {
            eprintln!(
                "{} -- {}    {} (ambiguity: {})",
                ov.jaccard().to_string().bold(),
                pair.ground_truth.green(),
                pair.candidate.red(),
                lookup
                    .get(&pair.ground_truth)
                    .map(|s| s.len() - 1)
                    .unwrap_or(0)
            )
        }
    }

    let mut results: Vec<ClassifiedLocation> = Vec::new();

    bar.set_message("Classifying occurrences into false negatives and true positives");
    // Now that the mapping is fully built, iterate over all ground truth occurrences
    // and see if the canidate contains it.
    // By iterating over ground_truth_occurrences we can only identify false negatives
    // and true positives.
    // False negatives are counted later
    for (ground_truth_location, ground_truth_symbol) in ground_truth_occurrences.clone() {
        match candidate_occurrences.get(&ground_truth_location) {
            // if this location is not marked in the candidate dataset,
            // this is a false negative - with full weight 1.0, as
            // there's no ambiguity to speak of - this location *should* contain
            // some symbol but doesn't
            None => results.push(ClassifiedLocation {
                location: ground_truth_location,
                symbol: ground_truth_symbol,
                mark: Mark::FalseNegative { weight: 1.0 },
            }),
            Some(candidate_symbol) => {
                let pair = SymbolPair {
                    ground_truth: ground_truth_symbol.clone(),
                    candidate: candidate_symbol.clone(),
                };

                match symbol_pair_weight.get(&pair) {
                    // At this location we found both the ground truth
                    // and candidate occurrence. We want to reward it - but with a
                    // weight indicating how precisely we matched candidate symbols to
                    // ground truth symbols - this weight comes from mapping we constructed earlier.
                    Some(weight) => results.push(ClassifiedLocation {
                        location: ground_truth_location,
                        symbol: ground_truth_symbol,
                        mark: Mark::TruePositive { weight: *weight },
                    }),
                    // This is an impossible situation by construction
                    None => panic!(
                        "Couldn't find a mapping for symbol {}",
                        ground_truth_symbol.red()
                    ),
                }
            }
        }
    }

    bar.set_message("Identifying false positives");
    for (candidate_location, candidate_symbol) in &candidate_occurrences {
        // If there are occurrences present in candidate dataset, but
        // not present in the ground truth, we treat it as a false positive
        // and penalise it will full strength.
        //
        // Technically this may be a mistake, in case the indexer
        // that produces ground truth has bugs in it.
        // But for simplicity we assume that scip-* indexers
        // are "perfect"
        if !ground_truth_occurrences.contains_key(candidate_location) {
            results.push(ClassifiedLocation {
                location: candidate_location.clone(),
                symbol: candidate_symbol.clone(),
                mark: Mark::FalsePositive { weight: 1.0 },
            });
        }
    }

    bar.finish_and_clear();

    summarise(results, options);
}

#[derive(Debug, PartialEq)]
enum Mark {
    TruePositive { weight: f32 },
    FalsePositive { weight: f32 },
    FalseNegative { weight: f32 },
}

#[derive(Debug, PartialEq)]
struct ClassifiedLocation {
    location: LocationInFile,
    symbol: String,
    mark: Mark,
}

fn summarise(classified: Vec<ClassifiedLocation>, options: ScipEvaluateOptions) {
    let mut true_positives: Vec<SymbolOccurrence> = Vec::new();
    let mut false_positives: Vec<SymbolOccurrence> = Vec::new();
    let mut false_negatives: Vec<SymbolOccurrence> = Vec::new();

    let mut tps = 0 as f32;
    let mut fps = 0 as f32;
    let mut fns = 0 as f32;

    for classified_location in classified {
        let symbol_occurrence = SymbolOccurrence {
            location: classified_location.location,
            symbol: classified_location.symbol,
        };
        match classified_location.mark {
            Mark::TruePositive { weight } => {
                true_positives.push(symbol_occurrence);
                tps += weight
            }
            Mark::FalseNegative { weight } => {
                false_negatives.push(symbol_occurrence);
                fns += weight
            }
            Mark::FalsePositive { weight } => {
                false_positives.push(symbol_occurrence);
                fps += weight
            }
        }
    }

    let precision = 100.0 * tps / (tps + fps);
    let recall = 100.0 * tps / (tps + fns);

    println!("{{\"precision_percent\": {precision:.2}, \"recall_percent\": {recall:.2}, \"true_positives\": {tps:.0}, \"false_positives\": {fps:.0}, \"false_negatives\": {fns:.0} }}");

    if options.print_false_negatives {
        eprintln!("");

        eprintln!(
            "{}: {}",
            "False negatives".red(),
            false_negatives.len().to_string().bold()
        );
        eprintln!(
            "{}",
            "How many actual occurrences we DIDN'T find compared to compiler?".italic()
        );

        for symbol_occurrence in false_negatives {
            eprintln!(
                "{}: L{} C{} -- {}",
                symbol_occurrence.location.file,
                symbol_occurrence.range().start_line,
                symbol_occurrence.range().start_col,
                symbol_occurrence.symbol
            );
        }
    }

    if options.print_false_positives {
        eprintln!("");
        eprintln!(
            "{}: {}",
            "False positives".red(),
            false_positives.len().to_string().bold()
        );
        eprintln!(
            "{}",
            "How many extra occurrences we reported compared to compiler?".italic()
        );

        for symbol_occurrence in false_positives {
            eprintln!(
                "{}: L{} C{} -- {}",
                symbol_occurrence.location.file,
                symbol_occurrence.range().start_line,
                symbol_occurrence.range().start_col,
                symbol_occurrence.symbol
            );
        }
    }

    if options.print_true_positives {
        if true {
            eprintln!("");
            eprintln!(
                "{}: {}",
                "True positives".green(),
                true_positives.len().to_string().bold()
            );

            for symbol_occurrence in &true_positives {
                let file = &symbol_occurrence.location.file;
                let rng = symbol_occurrence.range();
                let header = format!("{file}: L{} C{} -- ", rng.start_line, rng.start_col);

                eprintln!("{} {}", header.yellow(), symbol_occurrence.symbol);
            }
        }
    }
}

fn index_occurrences(idx: &Index) -> HashMap<LocationInFile, String> {
    let mut mp: HashMap<LocationInFile, String> = HashMap::new();

    for doc in &idx.documents {
        for occ in &doc.occurrences {
            let rng = PackedRange::from_vec(&occ.range).unwrap();
            let mut new_sym = occ.symbol.clone();
            let sanitised_relative_path: String = doc
                .relative_path
                .chars()
                .filter(|c| c.is_alphanumeric() || *c == '-' || *c == '+' || *c == '$')
                .collect();

            if occ.symbol.starts_with("local ") {
                new_sym = format!(
                    "local {}-{}",
                    sanitised_relative_path,
                    occ.symbol.strip_prefix("local ").unwrap()
                )
            }

            let loc = LocationInFile {
                rng,
                file: doc.relative_path.to_string(),
            };
            mp.insert(loc, new_sym.to_string());
        }
    }

    return mp;
}

enum IndexMode {
    Files { list: Vec<String> },
    Workspace { location: PathBuf },
}

fn index_command(
    language: String,
    index_mode: IndexMode,
    out: PathBuf,
    project_root: PathBuf,
    options: IndexOptions,
) {
    let p = BundledParser::get_parser(&language).unwrap();
    let canonical_project_root = {
        match index_mode {
            IndexMode::Files { .. } => project_root
                .canonicalize()
                .expect("Failed to canonicalize project root"),

            IndexMode::Workspace { ref location } => location
                .clone()
                .canonicalize()
                .expect("Failed to canonicalize project root"),
        }
    };

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
                let filepath = PathBuf::from(filename)
                    .canonicalize()
                    .unwrap();
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

    eprintln!("");

    eprintln!(
        "Writing index for {} documents into {}",
        index.documents.len(),
        out.display()
    );

    write_message_to_file(out, index).expect("to write the file");
}

fn create_progress_bar(len: u64) -> ProgressBar {
    let bar = ProgressBar::new(len);

    bar.set_style(
        ProgressStyle::with_template(
            "[{elapsed_precise}] {bar:40.cyan/blue} {pos:>7}/{len:7}\n {msg}",
        )
        .unwrap()
        .progress_chars("##-"),
    );

    return bar;
}

fn create_spinner() -> ProgressBar {
    let bar = ProgressBar::new_spinner();

    bar.set_style(
        ProgressStyle::with_template("{spinner:.blue} {msg}")
            .unwrap()
            .tick_strings(&[
                "▹▹▹▹▹",
                "▸▹▹▹▹",
                "▹▸▹▹▹",
                "▹▹▸▹▹",
                "▹▹▹▸▹",
                "▹▹▹▹▸",
                "▪▪▪▪▪",
            ]),
    );

    return bar;
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

    return Ok(document);
}

fn write_message_to_file<P>(
    path: P,
    msg: impl protobuf::Message,
) -> Result<(), Box<dyn std::error::Error>>
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

fn read_index_from_file(file: PathBuf) -> scip::types::Index {
    let mut candidate_idx = scip::types::Index::new();
    let candidate_f = File::open(file).unwrap();

    let mut reader = BufReader::new(candidate_f);
    let mut cis = CodedInputStream::from_buf_read(&mut reader);

    candidate_idx.merge_from(&mut cis).unwrap();
    return candidate_idx;
}

#[cfg(test)]
mod tests {
    use crate::read_index_from_file;
    use assert_cmd::cargo::cargo_bin;
    use assert_cmd::prelude::*;
    use std::collections::HashMap;
    use std::process::Command;
    use std::{env::temp_dir, path::PathBuf};

    lazy_static::lazy_static! {
        static ref BINARY_LOCATION: PathBuf = {
            let mut c: PathBuf;
            match std::env::var("SCIP_CLI_LOCATION") {
                Ok(va) => {
                    c = {
                        std::env::current_dir().unwrap().join(va)
                    }
                }
                _ => c = cargo_bin("scip-treesitter-cli")
            }

            c
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

    fn prepare(temp: &PathBuf, files: &HashMap<PathBuf, String>) {
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
