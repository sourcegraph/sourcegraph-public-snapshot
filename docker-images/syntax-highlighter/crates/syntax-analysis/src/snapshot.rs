use std::{collections::VecDeque, fmt::Write};

use anyhow::Result;
use protobuf::Enum;
use scip::types::{
    symbol_information, Document, Occurrence, SymbolInformation, SymbolRole, SyntaxKind,
};

use crate::range::Range;

#[derive(Debug, Clone, Default)]
pub struct SnapshotRange {
    pub start: usize,
    pub end: usize,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub enum EmitSyntax {
    #[default]
    None,
    Highlighted,
    All,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub enum EmitSymbol {
    None,
    Definitions,
    References,
    Enclosing,
    Unqualified,
    #[default]
    All,
}

#[derive(Debug, Clone, Default)]
pub struct SnapshotOptions {
    pub snapshot_range: Option<SnapshotRange>,

    pub emit_syntax: EmitSyntax,
    pub emit_symbol: EmitSymbol,
}

pub fn dump_document(doc: &Document, source: &str) -> Result<String> {
    dump_document_with_config(doc, source, SnapshotOptions::default())
}

pub fn dump_document_with_config(
    doc: &Document,
    source: &str,
    opts: SnapshotOptions,
) -> Result<String> {
    let mut occurrences = doc.occurrences.clone();
    occurrences.sort_by_key(|o| Range::from_vec(&o.range).unwrap_or_default());
    let mut occurrences = VecDeque::from(occurrences);

    let mut result = String::new();

    let line_iterator: Box<dyn Iterator<Item = (usize, &str)>> = match &opts.snapshot_range {
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
            // EmitSymbol::Enclosing means only do the enclosing range,
            // rather than the range of the symbol itself.
            let range = match &opts.emit_symbol {
                EmitSymbol::Enclosing => &occ.enclosing_range,
                _ => &occ.range,
            };

            let range = match Range::from_vec(range) {
                Some(range) => range,
                None => continue,
            };

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
                        // TODO: I might want to add (...) around the range
                        format!(
                            " {}:{}..{}:{}",
                            range.start_line, range.start_col, range.end_line, range.end_col
                        )
                    };

                    let syntax =
                        format_syntax(&occ.syntax_kind.enum_value_or_default(), &opts.emit_syntax);
                    let symbol = format_symbol(&occ, &opts.emit_symbol, &doc.symbols);

                    if syntax.is_some() || symbol.is_some() {
                        let syntax = syntax.unwrap_or_default();
                        let symbol = symbol.unwrap_or_default();

                        write!(
                            result,
                            "//{}{}{syntax}{multiline_suffix}{symbol}",
                            " ".repeat(range.start_col as usize),
                            "^".repeat(length)
                        )?;
                        result += "\n";
                    }

                    // write!(result, "\n")?;
                }
            }
        }
    }

    Ok(result)
}

fn format_syntax(kind: &SyntaxKind, emit_syntax: &EmitSyntax) -> Option<String> {
    match emit_syntax {
        EmitSyntax::None => None,
        EmitSyntax::Highlighted if kind == &SyntaxKind::UnspecifiedSyntaxKind => None,
        _ => Some(format!(" {:?}", kind)),
    }
}

fn format_symbol(
    occ: &Occurrence,
    emit_symbol: &EmitSymbol,
    symbols: &[SymbolInformation],
) -> Option<String> {
    if occ.symbol.is_empty() {
        return None;
    }

    let is_definition = occ.symbol_roles == SymbolRole::Definition.value();

    let symbol = match scip::symbol::parse_symbol(&occ.symbol) {
        Ok(symbol) => scip::symbol::format_symbol_with(
            symbol,
            scip::symbol::SymbolFormatOptions {
                include_scheme: true,
                include_package_manager: false,
                include_package_name: false,
                include_package_version: false,
                include_descriptor: true,
            },
        ),
        Err(_) => occ.symbol.clone(),
    };

    match emit_symbol {
        EmitSymbol::None => None,
        EmitSymbol::Definitions if !is_definition => None,
        EmitSymbol::References if is_definition => None,
        EmitSymbol::Unqualified => Some(format!(" {}", symbol)),
        _ => {
            let kind = if is_definition {
                "definition"
            } else {
                "reference"
            };

            let mut kind_info = String::new();
            if let Some(info) = symbols.iter().find(|sym| sym.symbol == occ.symbol) {
                let symbol_kind = info.kind.enum_value().expect("to be a valid kind");
                if symbol_kind != symbol_information::Kind::UnspecifiedKind {
                    kind_info = format!("({:?})", symbol_kind);
                }
            }

            Some(format!(" {kind}{kind_info} {symbol}"))
        }
    }
}
