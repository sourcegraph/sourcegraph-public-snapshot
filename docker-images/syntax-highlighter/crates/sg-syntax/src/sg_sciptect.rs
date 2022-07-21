#![allow(dead_code)]
use std::collections::{HashMap, HashSet};

use protobuf::EnumOrUnknown;
use scip::types::{Document, Occurrence, SyntaxKind};
use syntect::{
    parsing::{BasicScopeStackOp, ParseState, ScopeStack, SyntaxReference, SyntaxSet, SCOPE_REPO},
    util::LinesWithEndings,
};

/// The RangeGenerator generate a Document with occurrences set to the corresponding syntax kinds
///
/// If max_line_len is not None, any lines with length greater than the
/// provided number will not be highlighted.
pub struct DocumentGenerator<'a> {
    syntax_set: &'a SyntaxSet,
    parse_state: ParseState,
    stack: ScopeStack,
    code: &'a str,
    max_line_len: Option<usize>,
}

#[derive(Debug)]
struct ScopeStart {
    start_row: i32,
    start_col: i32,
    kind: Option<SyntaxKind>,
}

impl ScopeStart {
    fn some(start_row: usize, start_col: usize, kind: SyntaxKind) -> Self {
        Self {
            start_row: start_row as i32,
            start_col: start_col as i32,
            kind: Some(kind),
        }
    }

    fn none() -> Self {
        Self {
            start_row: 0,
            start_col: 0,
            kind: None,
        }
    }
}

impl<'a> DocumentGenerator<'a> {
    pub fn new(
        ss: &'a SyntaxSet,
        sr: &SyntaxReference,
        code: &'a str,
        max_line_len: Option<usize>,
    ) -> Self {
        Self {
            code,
            syntax_set: ss,
            parse_state: ParseState::new(sr),
            stack: ScopeStack::new(),
            max_line_len,
        }
    }

    // generate takes ownership of self so that it can't be re-used
    pub fn generate(mut self) -> Document {
        let mut scope_mapping = HashMap::new();
        scope_mapping.insert("keyword".to_string(), SyntaxKind::IdentifierKeyword);
        scope_mapping.insert("variable".to_string(), SyntaxKind::Identifier);
        scope_mapping.insert("punctuation".to_string(), SyntaxKind::PunctuationBracket);

        let mut ignore_mapping = HashSet::new();
        ignore_mapping.insert("source");

        let mut document = Document::default();

        let mut last_seen = (0, 0);
        let mut occ_stack: Vec<ScopeStart> = vec![];
        for (row, line) in LinesWithEndings::from(self.code).enumerate() {
            if self.max_line_len.map_or(false, |n| line.len() > n) {
                // self.write_escaped_html(line);
                continue;
            }

            let ops = self.parse_state.parse_line(line, self.syntax_set);
            for &(line_offset, ref op) in ops.as_slice() {
                // I think i is the byte offset, test with some multi-byte stuffs
                println!("offset: {} w/ op: {:?}", line_offset, op);

                let mut stack = self.stack.clone();
                stack.apply_with_hook(op, |basic_op, _| match basic_op {
                    BasicScopeStackOp::Push(scope) => {
                        if last_seen.0 != row || last_seen.1 != line_offset {
                            if let Some(scope) = occ_stack.last() {
                                if scope.kind != None {
                                    dbg!(&occ_stack);
                                    panic!("yayayayayayaayay");
                                }
                            }
                        }

                        let repo = SCOPE_REPO.lock().unwrap();

                        let atom = scope.atom_at(0 as usize);
                        let atom_s = repo.atom_str(atom);
                        if ignore_mapping.contains(atom_s) {
                            println!("Ignoring Atom: {:?}", atom_s);
                            occ_stack.push(ScopeStart::none());
                            return;
                        }

                        occ_stack.push(ScopeStart::some(
                            row,
                            line_offset,
                            *scope_mapping
                                .get(atom_s)
                                .unwrap_or(&SyntaxKind::UnspecifiedSyntaxKind),
                        ));
                    }
                    BasicScopeStackOp::Pop => {
                        let scope_start = occ_stack.pop().unwrap();
                        match scope_start.kind {
                            Some(kind) => {
                                // TODO: line_offset -> column (not bytes)
                                let mut occ = Occurrence::default();
                                occ.syntax_kind = EnumOrUnknown::new(kind);
                                occ.range = vec![
                                    scope_start.start_row,
                                    scope_start.start_col,
                                    row as i32,
                                    line_offset as i32,
                                ];
                                document.occurrences.push(occ);
                                // if span_empty {
                                //     // self.html.truncate(span_start);
                                // } else {
                                //     self.close_scope();
                                // }
                                // span_empty = false;
                            }
                            None => {}
                        }
                    }
                });
                self.stack = stack;
                last_seen = (row, line_offset);
            }
        }

        while let Some(scope_start) = occ_stack.pop() {
            if scope_start.kind.is_some() {
                panic!("{:?}", scope_start);
                // let mut occ = Occurrence::default();
                // occ.range = vec![scope_start.start_row, scope_start.start_col, 9999, 9999];
                // document.occurrences.push(occ);
            }
        }

        // close_table(&mut self.html);
        // self.html

        document
    }
}

#[cfg(test)]
mod test {
    use std::{
        fs::{read_dir, File},
        io::Read,
    };

    use crate::{determine_language, dump_document};

    use super::*;

    #[test]
    fn test_generates_empty_file() {
        let syntax_set = SyntaxSet::load_defaults_newlines();
        let mut q = crate::SourcegraphQuery::default();
        q.filetype = Some("go".to_string());
        q.code = "".to_string();

        let syntax_def = determine_language(&q, &syntax_set).unwrap();
        let output = DocumentGenerator::new(&syntax_set, syntax_def, &q.code, q.line_length_limit)
            .generate();

        assert_eq!(Document::default(), output);
    }

    #[test]
    fn test_generates_go_package() {
        let syntax_set = SyntaxSet::load_defaults_newlines();
        let mut q = crate::SourcegraphQuery::default();
        q.filetype = Some("go".to_string());
        q.code = "package main".to_string();

        let syntax_def = determine_language(&q, &syntax_set).unwrap();
        let output = DocumentGenerator::new(&syntax_set, syntax_def, &q.code, q.line_length_limit)
            .generate();

        assert_eq!(Document::default().occurrences, output.occurrences);
        assert_eq!(false, true);
    }

    #[test]
    fn test_all_files() -> Result<(), std::io::Error> {
        let dir = read_dir("./src/snapshots/syntect_files/")?;
        for entry in dir {
            let entry = entry?;
            let filepath = entry.path();
            let mut file = File::open(&filepath)?;
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;

            // let filetype = &determine_filetype(&SourcegraphQuery {
            //     extension: filepath.extension().unwrap().to_str().unwrap().to_string(),
            //     filepath: filepath.to_str().unwrap().to_string(),
            //     filetype: None,
            //     css: false,
            //     line_length_limit: None,
            //     theme: "".to_string(),
            //     code: contents.clone(),
            // });

            let mut q = crate::SourcegraphQuery::default();
            q.filetype = Some("go".to_string());
            q.code = contents.clone();
            let syntax_set = SyntaxSet::load_defaults_newlines();
            let syntax_def = determine_language(&q, &syntax_set).unwrap();

            let document =
                DocumentGenerator::new(&syntax_set, syntax_def, &q.code, q.line_length_limit)
                    .generate();

            insta::assert_snapshot!(
                filepath
                    .to_str()
                    .unwrap()
                    .replace("/src/snapshots/syntect_files", ""),
                dump_document(&document, &contents)
            );
        }

        Ok(())
    }
}
