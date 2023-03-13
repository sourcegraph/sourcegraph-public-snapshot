use std::{collections::VecDeque, fmt::Write};

use protobuf::Enum;
use scip::types::{Document, SymbolRole};

pub struct FileRange {
    pub start: usize,
    pub end: usize,
}

pub fn dump_document(doc: &Document, source: &str) -> String {
    dump_document_range(doc, source, &None)
}

pub fn dump_document_range(doc: &Document, source: &str, file_range: &Option<FileRange>) -> String {
    let mut occurrences = doc.occurrences.clone();
    occurrences.sort_by_key(|o| PackedRange::from_vec(&o.range));
    let mut occurrences = VecDeque::from(occurrences);

    let mut result = String::new();

    let line_iterator: Box<dyn Iterator<Item = (usize, &str)>> = match file_range {
        Some(range) => Box::new(
            source
                .lines()
                .enumerate()
                .skip(range.start - 1)
                .take(range.end - range.start + 1),
        ),
        None => Box::new(source.lines().enumerate()),
    };

    for (idx, line) in line_iterator {
        result += "  ";
        result += &line.replace('\t', " ");
        result += "\n";

        while let Some(occ) = occurrences.pop_front() {
            let range = PackedRange::from_vec(&occ.range);
            let is_single_line = range.start_line == range.end_line;
            let end_col = if is_single_line {
                range.end_col
            } else {
                line.len() as i32
            };

            match range.start_line.cmp(&(idx as i32)) {
                std::cmp::Ordering::Less => continue,
                std::cmp::Ordering::Greater => {
                    occurrences.push_front(occ);
                    break;
                }
                std::cmp::Ordering::Equal => {
                    let length = (end_col - range.start_col) as usize;
                    let multiline_suffix = if is_single_line {
                        "".to_string()
                    } else {
                        format!(
                            " {}:{}..{}:{}",
                            range.start_line, range.start_col, range.end_line, range.end_col
                        )
                    };
                    let symbol_suffix = if occ.symbol.is_empty() {
                        "".to_owned()
                    } else {
                        format!(" {}", occ.symbol)
                    };

                    let kind = if occ.symbol_roles == SymbolRole::Definition.value() {
                        "definition"
                    } else {
                        "reference"
                    };

                    // TODO: This will only work for definitions right now
                    let _ = writeln!(
                        result,
                        "//{}{} {}{multiline_suffix} {}",
                        " ".repeat(range.start_col as usize),
                        "^".repeat(length),
                        kind,
                        symbol_suffix,
                    );
                }
            }
        }
    }

    result
}

#[derive(Debug, PartialEq, Eq)]
pub struct PackedRange {
    pub start_line: i32,
    pub start_col: i32,
    pub end_line: i32,
    pub end_col: i32,
}

impl PackedRange {
    #[inline(always)]
    pub fn from_vec(v: &[i32]) -> Self {
        match v.len() {
            3 => Self {
                start_line: v[0],
                start_col: v[1],
                end_line: v[0],
                end_col: v[2],
            },
            4 => Self {
                start_line: v[0],
                start_col: v[1],
                end_line: v[2],
                end_col: v[3],
            },
            _ => {
                panic!("Unexpected vector length: {:?}", v);
            }
        }
    }
}

impl PartialOrd for PackedRange {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        (self.start_line, self.end_line, self.start_col).partial_cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
    }
}

impl Ord for PackedRange {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        (self.start_line, self.end_line, self.start_col).cmp(&(
            other.start_line,
            other.end_line,
            other.start_col,
        ))
    }
}
