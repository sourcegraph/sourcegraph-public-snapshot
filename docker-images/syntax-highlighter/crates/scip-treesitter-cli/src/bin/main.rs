use std::{
    collections::{HashMap, HashSet},
    fs::File,
    io::BufReader,
    path::Path,
};

use colored::*;
use protobuf::CodedInputStream;
use scip::types::{Index, Occurrence};
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter::types::PackedRange;
use scip_treesitter_languages::parsers::BundledParser;

pub fn write_message_to_file<P>(
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

use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
#[command(propagate_version = true)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Adds files to myapp
    Index {
        #[arg(short, long)]
        language: String,

        #[arg(short, long)]
        out: Option<String>,
        filenames: Vec<String>,
    },

    Compare {
        candidate: String,
        ground_truth: String,
        #[arg(short, long)]
        tp: bool,
    },

    ScipEvaluate {
        candidate: String,
        ground_truth: String,
        #[arg(short, long)]
        verbose: bool,
    },
}

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Name of the person to greet
    #[arg(short, long)]
    language: String,

    #[arg(short, long)]
    out: Option<String>,
    filenames: Vec<String>,
}

/*

Key idea: treat the OCCURRENCES as the data points against which we evaluate.

The target of course is compiler-accurate - for a given file it will produce a list
of occurrences without regard to the names of the symbols - TS and CA will
produce different names.

After we match the locations, we can identify the symbols in those positions -
and then compare the definition/references edges going to/from these occurrences.

This will give us a two-way evaluation, with individual precise/recall metrics
to test the breadth of our TS queries and the algorithm/heuristics for detecting references
or definitions.
*/

use protobuf::Message;

#[derive(Debug, PartialEq, Eq, Default, Hash, Clone)]
pub struct Location {
    rng: PackedRange,
    file: String,
}

pub fn main() {
    let cli = Cli::parse();
    match &cli.command {
        Commands::Compare {
            candidate,
            ground_truth,
            tp,
        } => compare_impl(candidate, ground_truth, *tp),

        Commands::ScipEvaluate {
            candidate,
            ground_truth,
            verbose,
        } => evaluate_impl(candidate, ground_truth, *verbose),

        Commands::Index {
            language,
            filenames,
            out,
        } => index_impl(language, filenames, out),
    }
}

pub fn read_from_file(file: &str) -> scip::types::Index {
    let mut candidate_idx = scip::types::Index::new();
    let candidate_f = File::open(file).unwrap();

    let mut reader = BufReader::new(candidate_f);
    let mut cis = CodedInputStream::from_buf_read(&mut reader);

    candidate_idx.merge_from(&mut cis).unwrap();
    return candidate_idx;
}

pub fn index_occurrences(idx: &Index) -> HashMap<Location, String> {
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

fn extract(lines: &HashMap<i32, &str>, occ: &Occurrence) -> Option<String> {
    let range = PackedRange::from_vec(&occ.range)?;

    let line = lines.get(&range.start_line)?;
    let label = &line[range.start_col as usize..range.end_col as usize];

    return Some(label.to_string());
}

pub fn index_impl(language: &String, filenames: &Vec<String>, out: &Option<String>) {
    let p = BundledParser::get_parser(language).unwrap();

    let directory = Path::new("./");
    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-local-nav".to_string(),
                version: "0.0.1".to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: "file://".to_string() + directory.to_str().unwrap(),
            ..Default::default()
        })
        .into(),
        ..Default::default()
    };

    for (_, filename) in filenames.iter().enumerate() {
        let contents = std::fs::read(filename).unwrap();
        let string = String::from_utf8(contents.clone()).unwrap();
        let lines: HashMap<i32, &str> = (0..).into_iter().zip(string.split("\n")).collect();
        let mut document = get_symbols(&p, &contents).unwrap();

        document.relative_path = filename.clone();
        let locals = get_locals(&p, &contents);

        match locals {
            Some(Ok(occurrences)) => {
                for occ in occurrences {
                    println!("{:?}: {occ}", extract(&lines, &occ));
                    // let mut cl_occ = occ.symbol.split_whitespace();
                    // cl_occ.next();
                    // let local_num = format!("local {}_{}", i, cl_occ.next().unwrap());

                    // let mut new_occ = occ.clone();
                    // new_occ.symbol = local_num;

                    document.occurrences.push(occ);
                }
            }
            _ => {}
        }

        // println!("{:?}", locals);
        index.documents.push(document);
    }

    let out_name = out.clone().unwrap_or("index.scip".to_string());
    let path = directory.join(out_name);

    write_message_to_file(path, index).expect("to write the file");
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

#[allow(dead_code, unused_variables)]
pub fn evaluate_impl<'a>(candidate: &String, ground_truth: &String, verbose: bool) {
    type Occ = (Location, String);
    let candidate_occs: HashMap<Location, String> = index_occurrences(&read_from_file(candidate));
    let ground_truth_occs: HashMap<Location, String> =
        index_occurrences(&read_from_file(&ground_truth));

    #[derive(Eq, Hash, PartialEq, Clone, Debug, Ord, PartialOrd)]
    struct SymbolPair {
        ground_truth: String,
        candidate: String,
    }

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
                    Some(overlap) => {
                        // overlap.total += 1;
                        overlap.common += 1
                    }
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

    for (pair, ov) in overlaps_vec {
        let similarity = ov.jaccard();
        println!(
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
                    None => println!("Couldn't find a mapping for symbol {}", occ.red()),
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

    summarise(results);
}

enum Mark {
    TruePositive(f32),
    FalsePositive(f32),
    FalseNegative(f32),
}

type ClassifiedLocation = (Location, String, Mark);
type Occ = (Location, String);

fn summarise(classified: Vec<ClassifiedLocation>) {
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

    let precision = tps / (tps + fps);

    // let recall = (true_positives.len() as f32)
    //     / (true_positives.len() as f32 + false_negatives.len() as f32);

    let recall = tps / (tps + fns);
    println!(
        "Precision: {}, Recall: {}",
        precision.to_string().bold(),
        recall.to_string().bold()
    );
    println!(
        "{}",
        "Precision = 'out of all found symbol occurrences, how many are also found by compiler?'"
            .italic()
    );
    println!("{}", "Recall = 'how close are we to finding the full set of occurrences as reported by the compiler?'".italic());
    println!("");

    println!(
        "{}: {}",
        "False negatives".red(),
        false_negatives.len().to_string().bold()
    );
    println!(
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
            // let spacing = " ".repeat(header.len());

            println!(
                "{} {}",
                header.yellow(),
                occ // ground_truth_occs.get(&rng).unwrap().bold()
            );
            // println!(
            //     "{file}: L{} C{} -- {occ}",
            //     rng.rng.start_line, rng.rng.start_col
            // );
        }
    }
}

pub fn compare_impl(candidate: &String, ground_truth: &String, tp: bool) {
    let canddate_occs = index_occurrences(&read_from_file(candidate));
    let ground_truth_occs = index_occurrences(&read_from_file(&ground_truth));

    let mut true_positives: Vec<Occ> = Vec::new();
    let mut false_positives: Vec<Occ> = Vec::new();
    let mut false_negatives: Vec<Occ> = Vec::new();

    for (rng, occ) in ground_truth_occs.clone() {
        match canddate_occs.get(&rng) {
            None => false_negatives.push((rng, occ)),
            Some(c) => true_positives.push((rng, c.to_string())),
        }
    }

    for (rng, occ) in canddate_occs {
        if !ground_truth_occs.contains_key(&rng) {
            false_positives.push((rng, occ));
        }
    }

    let precision = (true_positives.len() as f32)
        / (true_positives.len() as f32 + false_positives.len() as f32);

    let recall = (true_positives.len() as f32)
        / (true_positives.len() as f32 + false_negatives.len() as f32);
    println!(
        "Precision: {}, Recall: {}",
        precision.to_string().bold(),
        recall.to_string().bold()
    );
    println!(
        "{}",
        "Precision = 'out of all found symbol occurrences, how many are also found by compiler?'"
            .italic()
    );
    println!("{}", "Recall = 'how close are we to finding the full set of occurrences as reported by the compiler?'".italic());
    println!("");

    println!(
        "{}: {}",
        "False negatives".red(),
        false_negatives.len().to_string().bold()
    );
    println!(
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
    if tp {
        println!("");
        println!(
            "{}: {}",
            "True positives".green(),
            true_positives.len().to_string().bold()
        );

        for (rng, occ) in &true_positives {
            let file = &rng.file;
            let header = format!("{file}: L{} C{} -- ", rng.rng.start_line, rng.rng.start_col);
            // let spacing = " ".repeat(header.len());

            println!(
                "{} {} ({})",
                header.yellow(),
                occ,
                ground_truth_occs.get(&rng).unwrap().bold()
            );
            // println!(
            //     "{file}: L{} C{} -- {occ}",
            //     rng.rng.start_line, rng.rng.start_col
            // );
        }
    }
}
