#![allow(clippy::type_complexity)]
#![allow(clippy::needless_lifetimes)]

use std::{
    collections::{HashMap, HashSet},
    marker::PhantomData,
    path::PathBuf,
};

use anyhow::*;
use colored::{Color, ColoredString, Colorize};
use scip::types::Index;
use scip_treesitter::types::PackedRange;
use serde::Serializer;
use string_interner::{symbol::SymbolU32, StringInterner, Symbol};

use crate::{io::read_index_from_file, progress::*};

pub fn evaluate_command(
    candidate: PathBuf,
    ground_truth: PathBuf,
    evaluation_output_options: EvaluationOutputOptions,
) {
    Evaluator::default()
        .evaluate_files(candidate, ground_truth)
        .unwrap()
        .write_summary(&mut std::io::stdout(), evaluation_output_options)
        .unwrap()
}

fn validate_index(idx: &Index) -> Result<()> {
    let mut occs = 0;

    for doc in &idx.documents {
        occs += doc.occurrences.len();
    }

    if occs == 0 {
        Err(anyhow!(
            "Index contains no occurrences and cannot be used for evaluation"
        ))
    } else {
        Ok(())
    }
}
// These unfortunately don't help the typesafety and are only here to aid readability
// TODO: newtype https://doc.rust-lang.org/book/ch19-03-advanced-traits.html#using-the-newtype-pattern-to-implement-external-traits-on-external-types

#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash, Ord, PartialOrd)]
struct PathId {
    value: SymbolU32,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)]
struct SymbolId<T> {
    value: SymbolU32,
    _marker: PhantomData<T>,
}

impl<T> SymbolId<T> {
    fn into_any(self) -> SymbolId<Any> {
        SymbolId {
            value: self.value,
            _marker: PhantomData,
        }
    }
}

/// Phantom marker for SymbolId
#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)] // https://github.com/rust-lang/rust/issues/26925
struct GroundTruth;

/// Phantom marker for SymbolId
#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)] // https://github.com/rust-lang/rust/issues/26925
struct Candidate;

/// Phantom marker for SymbolId
#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash)] // https://github.com/rust-lang/rust/issues/26925
struct Any;

#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash)]
struct SymbolPair {
    ground_truth: SymbolId<GroundTruth>,
    candidate: SymbolId<Candidate>,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash)]
struct LocationInFile {
    rng: PackedRange,
    path_id: PathId,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq, Hash)]
pub struct SymbolOccurrence {
    location: LocationInFile,
    symbol: SymbolId<Any>,
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
    symbol: SymbolId<Any>,
    mark: Mark,
}

#[derive(Clone, Copy, Default, Debug)]
pub struct EvaluationOutputOptions {
    pub print_mapping: bool,
    pub print_true_positives: bool,
    pub print_false_positives: bool,
    pub print_false_negatives: bool,
    pub disable_colors: bool,
}

fn round_serialize<S>(x: &f32, s: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    s.serialize_str(format!("{:.1}", x).as_str())
}

#[derive(serde::Serialize, Debug)]
pub struct EvaluationSummary {
    #[serde(serialize_with = "round_serialize")]
    pub precision_percent: f32,
    #[serde(serialize_with = "round_serialize")]
    pub recall_percent: f32,
    #[serde(serialize_with = "round_serialize")]
    pub true_positives: f32,
    #[serde(serialize_with = "round_serialize")]
    pub false_positives: f32,
    #[serde(serialize_with = "round_serialize")]
    pub false_negatives: f32,
}

#[derive(Debug)]
pub struct EvaluationResult<'e> {
    // evaluator: &'e Evaluator,
    path_formatter: &'e mut PathFormatter,
    symbol_formatter: &'e mut SymbolFormatter,
    summary: EvaluationSummary,
    symbol_mapping: SymbolMapping,
    true_positives: Vec<SymbolOccurrence>,
    false_positives: Vec<SymbolOccurrence>,
    false_negatives: Vec<SymbolOccurrence>,
    // What options were used for this evaluation
}

#[derive(Debug)]
struct SymbolMapping {
    candidate_sets: HashMap<SymbolId<Candidate>, HashMap<SymbolId<GroundTruth>, Overlap>>,
    symbol_pair_weight: HashMap<SymbolPair, f32>,
}

impl<'e> EvaluationResult<'e> {
    fn new<'a>(
        evaluator: &'a mut Evaluator,
        classified: Vec<ClassifiedLocation>,
        symbol_mapping: SymbolMapping,
    ) -> EvaluationResult<'a> {
        let mut true_positives_occurrences: Vec<SymbolOccurrence> = Vec::new();
        let mut false_positives_occurrences: Vec<SymbolOccurrence> = Vec::new();
        let mut false_negatives_occurrences: Vec<SymbolOccurrence> = Vec::new();

        let mut true_positives = 0 as f32;
        let mut false_positives = 0 as f32;
        let mut false_negatives = 0 as f32;

        for classified_location in classified {
            let symbol_occurrence = SymbolOccurrence {
                location: classified_location.location,
                symbol: classified_location.symbol,
            };
            match classified_location.mark {
                Mark::TruePositive { weight } => {
                    true_positives_occurrences.push(symbol_occurrence);
                    true_positives += weight
                }
                Mark::FalseNegative { weight } => {
                    false_negatives_occurrences.push(symbol_occurrence);
                    false_negatives += weight
                }
                Mark::FalsePositive { weight } => {
                    false_positives_occurrences.push(symbol_occurrence);
                    false_positives += weight
                }
            }
        }

        let precision = 100.0 * true_positives / (true_positives + false_positives);
        let recall = 100.0 * true_positives / (true_positives + false_negatives);

        let sorter = |occ: &SymbolOccurrence| {
            (
                occ.location.path_id,
                occ.location.rng.start_line,
                occ.location.rng.start_col,
            )
        };

        true_positives_occurrences.sort_by_key(sorter);
        false_negatives_occurrences.sort_by_key(sorter);
        false_positives_occurrences.sort_by_key(sorter);

        EvaluationResult {
            symbol_formatter: &mut evaluator.symbol_formatter,
            path_formatter: &mut evaluator.path_formatter,
            summary: EvaluationSummary {
                precision_percent: precision,
                recall_percent: recall,
                true_positives,
                false_positives,
                false_negatives,
            },
            true_positives: true_positives_occurrences,
            false_positives: false_positives_occurrences,
            false_negatives: false_negatives_occurrences,
            symbol_mapping,
        }
    }
}

impl<'e> EvaluationResult<'e> {
    fn write_occ<W: std::io::Write>(
        &self,
        label: &str,
        occ: &SymbolOccurrence,
        output: &mut W,
    ) -> anyhow::Result<()> {
        write!(
            output,
            "[{label}] {}: L{} C{} -- {}\n",
            self.path_formatter.display_path(occ.location.path_id),
            occ.range().start_line,
            occ.range().start_col,
            self.symbol_formatter.display_symbol(occ.symbol)
        )?;

        Ok(())
    }

    fn render(&self, str: ColoredString, options: EvaluationOutputOptions) -> ColoredString {
        if options.disable_colors {
            str.normal()
        } else {
            str
        }
    }

    fn write_mapping<W: std::io::Write>(
        &mut self,
        output: &mut W,
        options: EvaluationOutputOptions,
    ) -> anyhow::Result<()> {
        let mut candidate_mapping_vec: Vec<(
            SymbolId<Candidate>,
            HashMap<SymbolId<GroundTruth>, Overlap>,
        )> = self
            .symbol_mapping
            .candidate_sets
            .clone()
            .into_iter()
            .collect();

        candidate_mapping_vec.sort_by_key(|(sym, _)| *sym);

        for (candidate_symbol, alternatives) in candidate_mapping_vec.into_iter() {
            let candidate = self
                .symbol_formatter
                .try_strip_package_details(candidate_symbol);
            let mut alternatives_vec: Vec<(SymbolId<GroundTruth>, Overlap)> =
                alternatives.into_iter().collect();
            alternatives_vec.sort_by_key(|(sym, _)| *sym);

            write!(
                output,
                "{}\n",
                self.render(
                    self.symbol_formatter.display_symbol(candidate).red(),
                    options
                )
            )?;

            for (ground_truth_symbol, overlap) in &alternatives_vec {
                let ground_truth = self
                    .symbol_formatter
                    .try_strip_package_details(*ground_truth_symbol);
                let adjusted_weight = self
                    .symbol_mapping
                    .symbol_pair_weight
                    .get(&SymbolPair {
                        candidate: candidate_symbol,
                        ground_truth: *ground_truth_symbol,
                    })
                    .unwrap();

                write!(
                    output,
                    "   {:.2} {} [{}/{} occurrences]\n",
                    adjusted_weight,
                    self.render(
                        self.symbol_formatter.display_symbol(ground_truth).green(),
                        options
                    ),
                    overlap.common,
                    overlap.total
                )?;
            }

            write!(output, "\n")?;
        }

        Ok(())
    }

    pub fn write_summary<W: std::io::Write>(
        &mut self,
        output: &mut W,
        options: EvaluationOutputOptions,
    ) -> anyhow::Result<()> {
        write!(output, "{}\n", serde_json::to_string(&self.summary)?)?;

        if options.print_false_negatives && self.false_negatives.len() > 0 {
            write!(output, "\n")?;

            write!(
                output,
                "{}: {}\n",
                self.render("False negatives (FN)".red(), options),
                self.render(self.false_negatives.len().to_string().bold(), options)
            )?;

            for symbol_occurrence in &self.false_negatives {
                self.write_occ("FN", symbol_occurrence, output)?;
            }
        }

        if options.print_false_positives && self.false_positives.len() > 0 {
            write!(output, "\n")?;
            write!(
                output,
                "{}: {}\n",
                self.render("False positives".red(), options),
                self.render(self.false_positives.len().to_string().bold(), options)
            )?;

            write!(
                output,
                "{}\n",
                "How many extra occurrences we reported compared to compiler?".italic()
            )?;

            for symbol_occurrence in &self.false_positives {
                self.write_occ("FP", symbol_occurrence, output)?;
            }
        }

        if options.print_true_positives {
            write!(output, "\n")?;
            write!(
                output,
                "{}: {}\n",
                self.render("True positives".green(), options),
                self.render(self.true_positives.len().to_string().bold(), options)
            )?;

            for symbol_occurrence in &self.true_positives {
                self.write_occ("TP", symbol_occurrence, output)?;
            }
        }

        if options.print_mapping {
            write!(output, "\nSymbol mapping\n")?;
            self.write_mapping(output, options)?;
        }

        Ok(())
    }
}

#[derive(Default, Debug)]
pub struct Evaluator {
    symbol_formatter: SymbolFormatter,
    path_formatter: PathFormatter,
}

// Public API
impl Evaluator {
    pub fn evaluate_files<'e>(
        &'e mut self,
        candidate: PathBuf,
        ground_truth: PathBuf,
    ) -> Result<EvaluationResult<'e>> {
        self.evaluate_indexes(
            &read_index_from_file(candidate),
            &read_index_from_file(ground_truth),
        )
    }

    pub fn evaluate_indexes<'e>(
        &'e mut self,
        candidate: &Index,
        ground_truth: &Index,
    ) -> Result<EvaluationResult<'e>> {
        validate_index(candidate)?;
        validate_index(ground_truth)?;

        let bar = create_spinner();
        bar.set_message("Indexing candidate symbols by location");
        let candidate_occurrences: HashMap<LocationInFile, SymbolId<Candidate>> =
            self.index_occurrences(candidate);
        bar.tick();

        bar.set_message("Indexing ground truth symbols by location");
        let ground_truth_occurrences: HashMap<LocationInFile, SymbolId<GroundTruth>> =
            self.index_occurrences(ground_truth);
        bar.tick();

        // For each symbol pair we maintain an Overlap instance
        let mut overlaps: HashMap<SymbolPair, Overlap> = HashMap::new();

        // Lookup table where key is ground truth symbol, and the value
        // is all the symbol pairs.
        // Each symbol from ground truth dataset can be mapped to any number of
        // symbols from the candidate set
        let mut ground_truth_alternatives: HashMap<SymbolId<GroundTruth>, HashSet<SymbolPair>> =
            HashMap::new();

        bar.set_message("Analysing occurrences in candidate SCIP");
        for (&candidate_loc, &candidate_symbol) in candidate_occurrences.iter() {
            // At given location from the candidate SCIP, see
            // if ground truth dataset contains any symbol at same location
            match ground_truth_occurrences.get(&candidate_loc) {
                // If ground truth dataset doesn't have any symbol at this location,
                // we treat it as a false positive, to be handled later
                None => {}
                Some(&ground_truth_symbol) => {
                    let pair = SymbolPair {
                        ground_truth: ground_truth_symbol,
                        candidate: candidate_symbol,
                    };

                    // See if we already have a lookup entry for this ground truth symbol
                    match ground_truth_alternatives.get_mut(&ground_truth_symbol) {
                        None => {
                            // If this is the first time we're seeing this symbol,
                            // create a lookup entry and put a single pair in there
                            ground_truth_alternatives
                                .insert(ground_truth_symbol, HashSet::from([pair]));
                        }
                        Some(s) => {
                            // Otherwise, add the symbol pair to the set (it might already be there)
                            s.insert(pair);
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
        for ground_truth_symbol in ground_truth_occurrences.values() {
            // now that we're iterating over all the ground truth occurrences,
            // we can update the `total` counter for each symbol pair
            // associated with that ground truth symbol
            ground_truth_alternatives
                .get(ground_truth_symbol)
                .into_iter()
                .for_each(|pairs| {
                    for pair in pairs {
                        if let Some(overlap) = overlaps.get_mut(pair) {
                            overlap.total += 1
                        }
                    }
                });
        }
        bar.tick();

        // For each candidate symbol we collect all possible ground truth symbols
        // it can be mapped to
        let candidate_mapping: HashMap<
            SymbolId<Candidate>,
            HashMap<SymbolId<GroundTruth>, Overlap>,
        > = {
            let mut result: HashMap<SymbolId<Candidate>, HashMap<SymbolId<GroundTruth>, Overlap>> =
                HashMap::new();

            for (symbol_pair, overlap) in overlaps {
                match result.get_mut(&symbol_pair.candidate) {
                    None => {
                        result.insert(
                            symbol_pair.candidate,
                            HashMap::from([(symbol_pair.ground_truth, overlap)]),
                        );
                    }
                    Some(map) => {
                        map.insert(symbol_pair.ground_truth, overlap);
                    }
                }
            }

            result
        };

        // We have produced the final counts for all symbol pairs -
        // it's time to produce final weights
        let symbol_pair_weight: HashMap<SymbolPair, f32> = {
            let mut result: HashMap<SymbolPair, f32> = HashMap::new();

            for (&candidate_symbol, alternatives) in &candidate_mapping {
                let total_weight: f32 = alternatives.values().map(|i| i.jaccard()).sum();

                for (&ground_truth_symbol, overlap) in alternatives {
                    let weight = overlap.jaccard();

                    let adjusted_weight = weight / total_weight;

                    result.insert(
                        SymbolPair {
                            candidate: candidate_symbol,
                            ground_truth: ground_truth_symbol,
                        },
                        adjusted_weight,
                    );
                }
            }

            result
        };

        // if options.print_mapping {
        //     let mut candidate_mapping_vec: Vec<(
        //         SymbolId<Candidate>,
        //         HashMap<SymbolId<GroundTruth>, Overlap>,
        //     )> = candidate_mapping.into_iter().collect();

        //     candidate_mapping_vec.sort_by_key(|(sym, _)| *sym);

        //     for (candidate_symbol, alternatives) in candidate_mapping_vec.into_iter() {
        //         let candidate = self
        //             .symbol_formatter
        //             .try_strip_package_details(candidate_symbol);
        //         let mut alternatives_vec: Vec<(SymbolId<GroundTruth>, Overlap)> =
        //             alternatives.into_iter().collect();
        //         alternatives_vec.sort_by_key(|(sym, _)| *sym);

        //         eprintln!("{}", self.symbol_formatter.display_symbol(candidate).red());

        //         for (ground_truth_symbol, overlap) in &alternatives_vec {
        //             let ground_truth = self
        //                 .symbol_formatter
        //                 .try_strip_package_details(*ground_truth_symbol);
        //             let adjusted_weight = symbol_pair_weight
        //                 .get(&SymbolPair {
        //                     candidate: candidate_symbol,
        //                     ground_truth: *ground_truth_symbol,
        //                 })
        //                 .unwrap();

        //             eprintln!(
        //                 "   {:.2} {} [{}/{} occurrences]",
        //                 adjusted_weight,
        //                 self.symbol_formatter.display_symbol(ground_truth).green(),
        //                 overlap.common,
        //                 overlap.total
        //             );
        //         }

        //         eprintln!();
        //     }
        // }

        let mut classified_locations: Vec<ClassifiedLocation> = Vec::new();

        bar.set_message("Classifying occurrences into false negatives and true positives");
        // Now that the mapping is fully built, iterate over all ground truth occurrences
        // and see if the canidate contains it.
        // By iterating over ground_truth_occurrences we can only identify false negatives
        // and true positives.
        // False negatives are counted later
        for (&ground_truth_location, &ground_truth_symbol) in ground_truth_occurrences.iter() {
            match candidate_occurrences.get(&ground_truth_location) {
                // if this location is not marked in the candidate dataset,
                // this is a false negative - with full weight 1.0, as
                // there's no ambiguity to speak of - this location *should* contain
                // some symbol but doesn't
                None => classified_locations.push(ClassifiedLocation {
                    location: ground_truth_location,
                    symbol: ground_truth_symbol.into_any(),
                    mark: Mark::FalseNegative { weight: 1.0 },
                }),
                Some(candidate_symbol) => {
                    let pair = SymbolPair {
                        ground_truth: ground_truth_symbol,
                        candidate: *candidate_symbol,
                    };

                    match symbol_pair_weight.get(&pair) {
                        // At this location we found both the ground truth
                        // and candidate occurrence. We want to reward it - but with a
                        // weight indicating how precisely we matched candidate symbols to
                        // ground truth symbols - this weight comes from mapping we constructed earlier.
                        Some(weight) => classified_locations.push(ClassifiedLocation {
                            location: ground_truth_location,
                            symbol: ground_truth_symbol.into_any(),
                            mark: Mark::TruePositive { weight: *weight },
                        }),
                        // This is an impossible situation by construction
                        None => panic!(
                            "Couldn't find a mapping for symbol {}",
                            self.symbol_formatter
                                .display_symbol(ground_truth_symbol)
                                .red()
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
                classified_locations.push(ClassifiedLocation {
                    location: *candidate_location,
                    symbol: candidate_symbol.into_any(),
                    mark: Mark::FalsePositive { weight: 1.0 },
                });
            }
        }

        bar.finish_and_clear();

        Ok(EvaluationResult::new(
            self,
            classified_locations,
            SymbolMapping {
                candidate_sets: candidate_mapping,
                symbol_pair_weight,
            },
        ))
    }
}

#[derive(Default, Debug)]
struct PathFormatter {
    interner: StringInterner,
}

impl PathFormatter {
    fn make_path_id(&mut self, s: &str) -> PathId {
        PathId {
            value: self.interner.get_or_intern(s),
        }
    }

    fn display_path(&self, s: PathId) -> &str {
        self.interner.resolve(s.value).unwrap()
    }
}

#[derive(Default, Debug)]
struct SymbolFormatter {
    interner: StringInterner,
}

impl SymbolFormatter {
    fn make_symbol_id<T>(&mut self, s: &str) -> SymbolId<T> {
        SymbolId {
            value: self.interner.get_or_intern(s),
            _marker: PhantomData,
        }
    }

    fn display_symbol<T>(&self, s: SymbolId<T>) -> &str {
        self.interner.resolve(s.value).unwrap()
    }

    fn try_strip_package_details<T: Copy>(&mut self, sym: SymbolId<T>) -> SymbolId<T> {
        let s = self.display_symbol(sym);
        if s.as_bytes().iter().filter(|&c| *c == b' ').count() != 5 {
            return sym;
        }
        let parts: Vec<&str> = s.splitn(5, ' ').collect();
        let scheme = parts[0];
        let _manager = parts[1];
        let _package_name = parts[2];
        let _version = parts[3];
        let descriptor = parts[4];
        self.make_symbol_id(&format!("{scheme} . . . {descriptor}"))
    }
}

impl Evaluator {
    fn index_occurrences<T>(&mut self, index: &Index) -> HashMap<LocationInFile, SymbolId<T>> {
        let mut out: HashMap<LocationInFile, SymbolId<T>> = HashMap::new();

        for doc in &index.documents {
            let path_id = self.path_formatter.make_path_id(&doc.relative_path);
            out.reserve(doc.occurrences.len());
            for occ in &doc.occurrences {
                let rng = PackedRange::from_vec(&occ.range).unwrap();
                let sym_id: SymbolId<T>;
                if let Some(prefix) = occ.symbol.strip_prefix("local ") {
                    sym_id = self.symbol_formatter.make_symbol_id(&format!(
                        "local doc-{}-{}",
                        path_id.value.to_usize(),
                        prefix
                    ));
                } else {
                    sym_id = self.symbol_formatter.make_symbol_id(&occ.symbol);
                }
                let loc = LocationInFile { rng, path_id };
                out.insert(loc, sym_id);
            }
        }

        out
    }
}

#[cfg(test)]
mod tests {
    use scip::types::*;

    use crate::evaluate::Evaluator;

    fn occurrence(n: i32, symbol: &str) -> Occurrence {
        let mut occ = Occurrence::new();
        occ.range = vec![n, 5, 10];
        occ.symbol = symbol.to_string();
        occ
    }

    fn document(path: &str, occs: Vec<Occurrence>) -> Document {
        let mut doc = Document::new();
        doc.relative_path = path.to_string();
        doc.occurrences.extend(occs);
        doc
    }

    fn index(documents: Vec<Document>) -> Index {
        let mut idx = Index::new();
        idx.documents.extend(documents);
        idx
    }

    #[test]
    fn evaluation_fundamentals() {
        let doc1 = document(
            "document1.java",
            vec![occurrence(1, "sym1"), occurrence(2, "sym2")],
        );

        let doc2 = document(
            "document2.java",
            vec![occurrence(1, "sym1"), occurrence(2, "sym2")],
        );

        let ground_truth = index(vec![doc1, doc2]);

        let mut evaluator = Evaluator::default();

        // Evaluating index against itself should yield 100% precision and 100% recall
        {
            let evaluate_with_self = evaluator
                .evaluate_indexes(&ground_truth, &ground_truth)
                .unwrap();
            assert_eq!(evaluate_with_self.summary.precision_percent, 100.0);
            assert_eq!(evaluate_with_self.summary.recall_percent, 100.0);
            assert_eq!(evaluate_with_self.summary.true_positives, 4.0);
            assert_eq!(evaluate_with_self.summary.false_positives, 0.0);
            assert_eq!(evaluate_with_self.summary.false_negatives, 0.0);
        }

        // This has no overlap at all with the ground truth
        // Should yield 0% precision, 0% recall
        {
            let evaluate_disjoint = evaluator
                .evaluate_indexes(
                    &index(vec![
                        document(
                            "bla/document1.java",
                            vec![occurrence(1, "sym1"), occurrence(2, "sym2")],
                        ),
                        document(
                            "bla/document2.java",
                            vec![occurrence(1, "sym1"), occurrence(2, "sym2")],
                        ),
                    ]),
                    &ground_truth,
                )
                .unwrap();
            assert_eq!(evaluate_disjoint.summary.precision_percent, 0.0);
            assert_eq!(evaluate_disjoint.summary.recall_percent, 0.0);
            assert_eq!(evaluate_disjoint.summary.true_positives, 0.0);
            assert_eq!(evaluate_disjoint.summary.false_positives, 4.0);
            assert_eq!(evaluate_disjoint.summary.false_negatives, 4.0);
        }

        let empty_index = index(vec![document("bla.java", vec![])]);

        // Evaluating empty index is an error
        {
            let evaluate_empty = evaluator.evaluate_indexes(&empty_index, &ground_truth);
            assert!(evaluate_empty.is_err());
        }

        // Evaluating against an empty index is an error
        {
            let evaluate_against_empty = evaluator.evaluate_indexes(&ground_truth, &empty_index);
            assert!(evaluate_against_empty.is_err());
        }
    }
}
