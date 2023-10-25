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
struct Location {
    rng: PackedRange,
    file: String,
}

#[derive(Eq, Hash, PartialEq, Clone, Debug, Ord, PartialOrd)]
struct SymbolOccurrence {
    location: Location,
    symbol: String,
}

#[derive(Eq, Hash, PartialEq, Clone, Debug)]
struct Overlap {
    total: u32,
    common: u32,
}

impl Overlap {
    fn jaccard(&self) -> f32 {
        self.common as f32 / self.total as f32
    }
}

pub fn evaluate_command<'a>(candidate: String, ground_truth: String, options: ScipEvaluateOptions) {
    let candidate_occs: HashMap<Location, String> =
        index_occurrences(&read_index_from_file(candidate.into()));

    let ground_truth_occs: HashMap<Location, String> =
        index_occurrences(&read_index_from_file(ground_truth.into()));

    let mut overlaps: HashMap<SymbolPair, Overlap> = HashMap::new();
    let mut lookup: HashMap<String, HashSet<SymbolPair>> = HashMap::new();

    for (candidate_loc, candidate_symbol_orig) in candidate_occs.clone() {
        match ground_truth_occs.get(&candidate_loc) {
            None => {}
            Some(ground_truth_symbol) => {
                let candidate_symbol = candidate_symbol_orig.clone();
                let pair = SymbolPair {
                    ground_truth: ground_truth_symbol.clone(),
                    candidate: candidate_symbol,
                };

                match lookup.get_mut(ground_truth_symbol) {
                    None => {
                        let mut set = HashSet::new();
                        set.insert(pair.clone());
                        lookup.insert(ground_truth_symbol.clone(), set);
                        ()
                    }
                    Some(s) => {
                        s.insert(pair.clone());
                        ()
                    }
                }

                let overlap = overlaps.get_mut(&pair.clone());

                match overlap {
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

    for (_, gt_symbol) in ground_truth_occs.clone() {
        for pairs in lookup.get(&gt_symbol).into_iter() {
            for pair in pairs {
                let overlap = overlaps.get_mut(&pair);

                overlap.map(|x| x.total += 1);
            }
        }
    }

    let mapping: HashMap<SymbolPair, f32> = overlaps
        .clone()
        .into_iter()
        .map(|c| (c.0, c.1.jaccard()))
        .collect();

    let mut overlaps_vec: Vec<(SymbolPair, Overlap)> = overlaps.into_iter().collect();
    overlaps_vec.sort_by_key(|k| k.0.clone());

    if options.print_mapping {
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

    for (rng, occ) in ground_truth_occs.clone() {
        match candidate_occs.get(&rng) {
            None => results.push((rng, occ, Mark::FalseNegative(1.0))),
            //results.push((rng, occ)),
            Some(c) => {
                let similarity = mapping.get(&SymbolPair {
                    ground_truth: occ.clone(),
                    candidate: c.clone(),
                });
                match similarity {
                    None => eprintln!("Couldn't find a mapping for symbol {}", occ.red()),
                    Some(v) => results.push((rng, occ, Mark::TruePositive(*v))),
                }
            } // true_positives.push((rng, c.to_string())),
        }
    }

    for (rng, occ) in candidate_occs.clone() {
        if !ground_truth_occs.contains_key(&rng) {
            // false_positives.push((rng, occ));

            results.push((rng, occ, Mark::FalsePositive(1.0)));
        }
    }

    summarise(results, options);
}

enum Mark {
    TruePositive(f32),
    FalsePositive(f32),
    FalseNegative(f32),
}

type ClassifiedLocation = (Location, String, Mark);
type Occ = (Location, String);

fn summarise(classified: Vec<ClassifiedLocation>, options: ScipEvaluateOptions) {
    let mut true_positives: Vec<Occ> = Vec::new();
    let mut false_positives: Vec<Occ> = Vec::new();
    let mut false_negatives: Vec<Occ> = Vec::new();

    let mut tps = 0 as f32;
    let mut fps = 0 as f32;
    let mut fns = 0 as f32;

    for cl in classified {
        match cl.2 {
            Mark::TruePositive(a) => {
                true_positives.push((cl.0, cl.1));
                tps += a
            }
            Mark::FalseNegative(a) => {
                false_negatives.push((cl.0, cl.1));
                fns += a
            }
            Mark::FalsePositive(a) => {
                false_positives.push((cl.0, cl.1));
                fps += a
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

        for (rng, occ) in false_negatives {
            let file = rng.file;
            println!(
                "{file}: L{} C{} -- {occ}",
                rng.rng.start_line, rng.rng.start_col
            );
        }
    }

    if options.print_false_positives {
        println!("");
        println!(
            "{}: {}",
            "False positives".red(),
            false_positives.len().to_string().bold()
        );
        println!(
            "{}",
            "How many extra occurrences we reported compared to compiler?".italic()
        );

        for (rng, occ) in false_positives {
            let file = rng.file;
            println!(
                "{file}: L{} C{} -- {occ}",
                rng.rng.start_line, rng.rng.start_col
            );
        }
    }

    if options.print_true_positives {
        if true {
            println!("");
            println!(
                "{}: {}",
                "True positives".green(),
                true_positives.len().to_string().bold()
            );

            for (rng, occ) in &true_positives {
                let file = &rng.file;
                let header = format!("{file}: L{} C{} -- ", rng.rng.start_line, rng.rng.start_col);

                println!(
                    "{} {}",
                    header.yellow(),
                    occ // ground_truth_occs.get(&rng).unwrap().bold()
                );
            }
        }
    }
}

fn index_occurrences(idx: &Index) -> HashMap<Location, String> {
    let mut mp: HashMap<Location, String> = HashMap::new();

    for doc in &idx.documents {
        for occ in &doc.occurrences {
            let rng = PackedRange::from_vec(&occ.range).unwrap();
            let mut new_sym = occ.symbol.clone();
            if occ.symbol.starts_with("local ") {
                // doc.relative_path.to_string() + " " + occ.symbol
                new_sym = format!("{} {}", doc.relative_path, occ.symbol)
            }

            let loc = Location {
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
            let bar = ProgressBar::new(list.len() as u64);

            bar.set_style(
                ProgressStyle::with_template(
                    "[{elapsed_precise}] {bar:40.cyan/blue} {pos:>7}/{len:7}\n {msg}",
                )
                .unwrap()
                .progress_chars("##-"),
            );

            for filename in list {
                let filepath = PathBuf::from(filename).canonicalize().expect("???b");
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
