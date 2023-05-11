use std::ops::Range;

use anyhow::Result;
use bitvec::prelude::*;
use protobuf::Enum;
use scip::types::{Descriptor, Occurrence};
use scip_treesitter::types::PackedRange;

use crate::languages::TagConfiguration;

#[derive(Debug)]
pub struct MemoryBundle {
    pub scopes: Vec<Scope>,
    pub globals: Vec<Global>,
    pub descriptors: Vec<Descriptor>,

    pub children: Vec<Child>,
    pub scope_stack: Vec<u32>,
}

#[derive(Debug)]
pub struct Child {
    index: u32,
    prev: u32,
}

#[derive(Debug)]
pub struct Scope {
    pub ident_range: PackedRange,
    pub scope_range: PackedRange,
    pub globals: u32,
    pub children: u32,
    pub descriptors: Range<u32>,
}

#[derive(Debug)]
pub struct Global {
    pub range: PackedRange,
    pub descriptors: Range<usize>,
}

pub struct ScopeChildrenIterator<'a> {
    current: u32,
    children: &'a Vec<Child>,
}

impl<'a> Iterator for ScopeChildrenIterator<'a> {
    // We can refer to this type using Self::Item
    type Item = usize;

    fn next(&mut self) -> Option<Self::Item> {
        // current is dead (but don't kill populated root children)
        if self.current == u32::MAX {
            return None;
        }

        let current_child = &self.children[self.current as usize];
        let result = current_child.index;
        self.current = current_child.prev;

        return Some(result as usize);
    }
}

pub fn iterate_children_indices<'a>(
    children: &'a Vec<Child>,
    children_or_globals: u32,
) -> ScopeChildrenIterator {
    return ScopeChildrenIterator::<'a> {
        current: children_or_globals,
        children,
    };
}

impl MemoryBundle {
    pub fn into_occurrences(&mut self, base_descriptors: Vec<Descriptor>) -> Vec<Occurrence> {
        let mut descriptor_stack = base_descriptors;
        let mut occs = Vec::with_capacity(self.globals.len());
        self.rec_into_occurrences(0, true, &mut occs, &mut descriptor_stack);
        occs
    }

    fn rec_into_occurrences(
        &self,
        scope_idx: usize,
        is_root: bool,
        occurrences: &mut Vec<Occurrence>,
        descriptor_stack: &mut Vec<Descriptor>,
    ) {
        let scope = &self.scopes[scope_idx];

        descriptor_stack.extend(
            (&self.descriptors[scope.descriptors.start as usize..scope.descriptors.end as usize])
                .to_vec(),
        );

        if !is_root {
            occurrences.push(scip::types::Occurrence {
                range: scope.ident_range.to_vec(),
                symbol: scip::symbol::format_symbol(scip::types::Symbol {
                    scheme: "scip-ctags".into(),
                    // TODO: Package?
                    package: None.into(),
                    descriptors: descriptor_stack.clone(),
                    ..Default::default()
                }),
                symbol_roles: scip::types::SymbolRole::Definition.value(),
                // TODO:
                // syntax_kind: todo!(),
                ..Default::default()
            });
        }

        for global_idx in iterate_children_indices(&self.children, scope.globals) {
            let global = &self.globals[global_idx];
            let mut global_descriptors = descriptor_stack.clone();
            global_descriptors.extend(
                (&self.descriptors
                    [global.descriptors.start as usize..global.descriptors.end as usize])
                    .to_vec(),
            );

            let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                scheme: "scip-ctags".into(),
                // TODO: Package?
                package: None.into(),
                descriptors: global_descriptors,
                ..Default::default()
            });

            let symbol_roles = scip::types::SymbolRole::Definition.value();
            occurrences.push(scip::types::Occurrence {
                range: global.range.to_vec(),
                symbol,
                symbol_roles,
                // TODO:
                // syntax_kind: todo!(),
                ..Default::default()
            });
        }

        // self.children
        //     .iter()
        //     .for_each(|c| self.rec_into_occurrences(false, occurrences, descriptor_stack));
        iterate_children_indices(&self.children, scope.children)
            .for_each(|c| self.rec_into_occurrences(c, false, occurrences, descriptor_stack));

        self.descriptors.iter().for_each(|_| {
            descriptor_stack.pop();
        });
    }
}

pub fn parse_tree<'a>(
    config: &TagConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
    bundle: &'a mut MemoryBundle,
) -> Result<()> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.query.capture_names();

    // Reusing vecs allows us to reduce the number of alloc/realloc syscalls we send out
    // This leads to a small (few ms) improvement on small files which adds up
    // TODO: Apply similar optimization (arena would work) on tag emission

    let scopes = &mut bundle.scopes;
    let globals = &mut bundle.globals;
    let descriptors = &mut bundle.descriptors;

    let children = &mut bundle.children;
    let scope_stack = &mut bundle.scope_stack;

    scopes.clear();
    globals.clear();
    descriptors.clear();

    children.clear();
    scope_stack.clear();

    scopes.push(Scope {
        ident_range: root_node.into(),
        scope_range: root_node.into(),
        globals: u32::MAX,
        children: u32::MAX,
        descriptors: 0..0,
    });

    scope_stack.push(0);

    let mut local_ranges = BitVec::<u8, Msb0>::repeat(false, source_bytes.len());

    let matches = cursor.matches(&config.query, root_node, source_bytes);

    // According to some quick testing, tree-sitter returns query results
    // in order, so we can do everything in one loop rather than aggregating first
    // and then sorting/processing

    for m in matches {
        let mut node = None;
        let mut scope = None;
        let mut local_range = None;

        let start = descriptors.len();

        for capture in m.captures {
            let capture_name = capture_names
                .get(capture.index as usize)
                .expect("capture indexes should always work");

            if capture_name.starts_with("descriptor") {
                descriptors.push(crate::ts_scip::capture_name_to_descriptor(
                    capture_name,
                    capture.node.utf8_text(source_bytes)?.to_string(),
                ));
                node = Some(capture.node);
            }

            if capture_name.starts_with("scope") {
                assert!(scope.is_none(), "declare only one scope per match");
                scope = Some(capture);
            }

            if capture_name.starts_with("local") {
                local_range = Some(capture.node.byte_range());
            }
        }

        match node {
            Some(node) => {
                if local_ranges[node.start_byte()] {
                    continue;
                }

                {
                    let mut i: i64 = (scope_stack.len() - 1) as i64;
                    while i >= 0 {
                        if !scopes[scope_stack[i as usize] as usize]
                            .scope_range
                            .contains(&node.into())
                        {
                            scope_stack.pop();
                            i -= 1;
                        } else {
                            break;
                        }

                        i -= 1;
                    }
                }

                let current_scope = *scope_stack.last().unwrap() as usize;

                match scope {
                    Some(scope_ident) => {
                        scopes.push(Scope {
                            ident_range: node.into(),
                            scope_range: scope_ident.node.into(),
                            globals: u32::MAX,
                            children: u32::MAX,
                            descriptors: start as u32..descriptors.len() as u32,
                        });

                        children.push(Child {
                            index: (scopes.len() - 1) as u32,
                            prev: scopes[current_scope].children,
                        });

                        scopes[current_scope].children = (children.len() - 1) as u32;
                        scope_stack.push((scopes.len() - 1) as u32);
                    }
                    None => {
                        globals.push(Global {
                            range: node.into(),
                            descriptors: start..descriptors.len(),
                        });

                        children.push(Child {
                            index: (globals.len() - 1) as u32,
                            prev: scopes[current_scope].globals,
                        });

                        scopes[current_scope].globals = (children.len() - 1) as u32;
                    }
                }
            }
            None => {
                if local_range.is_none() {
                    panic!("there must always be at least one descriptor (except for @local)");
                }
            }
        }

        if let Some(local_range) = local_range {
            local_ranges.get_mut(local_range).unwrap().fill(true);
        }
    }

    Ok(())
}

// #[cfg(test)]
// mod test {
//     use scip::types::Document;
//     use scip_treesitter::snapshot::dump_document;
//     use tree_sitter::Parser;

//     use super::*;

//     fn parse_file_for_lang(config: &TagConfiguration, source_code: &str) -> Result<Document> {
//         let source_bytes = source_code.as_bytes();

//         let mut parser = Parser::new();
//         parser.set_language(config.language).unwrap();
//         let tree = parser.parse(source_bytes, None).unwrap();

//         let mut bundle = MemoryBundle {
//             scopes: vec![],
//             globals: vec![],
//         };

//         let mut occ = parse_tree(config, &tree, source_bytes, &mut bundle)?;
//         let mut doc = Document::new();
//         doc.occurrences = occ.0.into_occurrences(occ.1, vec![]);
//         doc.symbols = doc
//             .occurrences
//             .iter()
//             .map(|o| scip::types::SymbolInformation {
//                 symbol: o.symbol.clone(),
//                 ..Default::default()
//             })
//             .collect();

//         Ok(doc)
//     }

//     #[test]
//     fn test_can_parse_rust_tree() -> Result<()> {
//         let config = crate::languages::rust();
//         let source_code = include_str!("../testdata/scopes.rs");
//         let doc = parse_file_for_lang(config, source_code)?;

//         let dumped = dump_document(&doc, source_code)?;
//         insta::assert_snapshot!(dumped);

//         Ok(())
//     }

//     #[test]
//     fn test_can_parse_go_tree() -> Result<()> {
//         let config = crate::languages::go();
//         let source_code = include_str!("../testdata/example.go");
//         let doc = parse_file_for_lang(config, source_code)?;
//         // dbg!(doc);

//         let dumped = dump_document(&doc, source_code)?;
//         insta::assert_snapshot!(dumped);

//         Ok(())
//     }

//     #[test]
//     fn test_can_parse_go_internal_tree() -> Result<()> {
//         let config = crate::languages::go();
//         let source_code = include_str!("../testdata/internal_go.go");
//         let doc = parse_file_for_lang(config, source_code)?;
//         // dbg!(doc);

//         let dumped = dump_document(&doc, source_code)?;
//         insta::assert_snapshot!(dumped);

//         Ok(())
//     }
// }
