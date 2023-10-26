use std::collections::{HashMap, HashSet};

use colored::Colorize;
use scip::types::Index;
use scip_treesitter::types::PackedRange;

use crate::{io::read_index_from_file, progress::*};

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

pub struct ScipEvaluateOptions {
    pub print_mapping: bool,
    pub print_true_positives: bool,
    pub print_false_positives: bool,
    pub print_false_negatives: bool,
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
