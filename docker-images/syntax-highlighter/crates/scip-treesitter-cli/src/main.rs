use std::{collections::HashMap, fs::File, io::BufReader, path::Path};

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
        groundTruth: String,
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
struct Location {
    rng: PackedRange,
    file: String,
}

pub fn main() {
    let cli = Cli::parse();
    match &cli.command {
        Commands::Compare {
            candidate,
            groundTruth,
        } => {
            type Occ = (Location, String);
            let canddate_occs = index_occurrences(&read_from_file(&candidate));
            let ground_truth_occs = index_occurrences(&read_from_file(&groundTruth));

            let mut true_positives: Vec<Occ> = Vec::new();
            let mut false_positives: Vec<Occ> = Vec::new();
            let mut false_negatives: Vec<Occ> = Vec::new();

            for (rng, occ) in ground_truth_occs.clone() {
                if canddate_occs.contains_key(&rng) {
                    true_positives.push((rng, occ));
                } else {
                    false_negatives.push((rng, occ));
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
            println!("{}", "Precision = 'out of all found symbol occurrences, how many are also found by compiler?'".italic());
            println!("{}", "Recall = 'how close are we to finding the full set of occurrences as reported by the compiler?'".italic());
            println!("");
            println!("{}", "False negatives:".red());
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
            println!("{}", "False positives:".red());
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
        Commands::Index {
            language,
            filenames,
            out,
        } => {
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

            for filename in filenames {
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
    }
}

fn read_from_file(file: &str) -> scip::types::Index {
    let mut candidate_idx = scip::types::Index::new();
    let candidate_f = File::open(file).unwrap();

    let mut reader = BufReader::new(candidate_f);
    let mut cis = CodedInputStream::from_buf_read(&mut reader);

    candidate_idx.merge_from(&mut cis).unwrap();
    return candidate_idx;
}

fn index_occurrences(idx: &Index) -> HashMap<Location, String> {
    let mut mp: HashMap<Location, String> = HashMap::new();

    for doc in &idx.documents {
        for occ in &doc.occurrences {
            let rng = PackedRange::from_vec(&occ.range).unwrap();
            let loc = Location {
                rng: rng,
                file: doc.relative_path.to_string(),
            };
            mp.insert(loc, occ.symbol.to_string());
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
