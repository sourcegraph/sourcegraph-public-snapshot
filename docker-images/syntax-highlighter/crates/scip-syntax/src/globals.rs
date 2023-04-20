use std::collections::HashMap;

use anyhow::Result;
use protobuf::Enum;
use scip::types::{Descriptor, Occurrence};

use crate::{byterange::ByteRange, languages::TagConfiguration};

#[derive(Debug)]
pub struct Scope {
    pub range: [i32; 3],
    pub byte_range: ByteRange,
    pub globals: Vec<Global>,
    pub children: Vec<Scope>,
    pub descriptors: Vec<Descriptor>,
}

#[derive(Debug)]
pub struct Global {
    pub range: [i32; 3],
    pub byte_range: ByteRange,
    pub descriptors: Vec<Descriptor>,
}

impl Scope {
    pub fn insert_scope(&mut self, scope: Scope) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.byte_range.contains(&scope.byte_range))
        {
            child.insert_scope(scope);
        } else {
            self.children.push(scope);
        }
    }

    pub fn insert_global(&mut self, global: Global) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.byte_range.contains(&global.byte_range))
        {
            child.insert_global(global)
        } else {
            self.globals.push(global);
        }
    }

    pub fn into_occurrences(
        &mut self,
        hint: usize,
        base_descriptors: Vec<Descriptor>,
    ) -> Vec<Occurrence> {
        let mut descriptor_stack = base_descriptors;
        let mut occs = Vec::with_capacity(hint);
        self.rec_into_occurrences(true, &mut occs, &mut descriptor_stack);
        occs
    }

    fn rec_into_occurrences(
        &self,
        is_root: bool,
        occurrences: &mut Vec<Occurrence>,
        descriptor_stack: &mut Vec<Descriptor>,
    ) {
        descriptor_stack.extend(self.descriptors.clone());

        if !is_root {
            occurrences.push(scip::types::Occurrence {
                range: Vec::from(self.range),
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

        for global in &self.globals {
            let mut global_descriptors = descriptor_stack.clone();
            global_descriptors.extend(global.descriptors.clone());

            let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                scheme: "scip-ctags".into(),
                // TODO: Package?
                package: None.into(),
                descriptors: global_descriptors,
                ..Default::default()
            });

            let symbol_roles = scip::types::SymbolRole::Definition.value();
            occurrences.push(scip::types::Occurrence {
                range: Vec::from(global.range),
                symbol,
                symbol_roles,
                // TODO:
                // syntax_kind: todo!(),
                ..Default::default()
            });
        }

        self.children
            .iter()
            .for_each(|c| c.rec_into_occurrences(false, occurrences, descriptor_stack));

        self.descriptors.iter().for_each(|_| {
            descriptor_stack.pop();
        });
    }
}

pub fn parse_tree<'a>(
    config: &TagConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
) -> Result<(Scope, usize)> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.query.capture_names();

    let mut scopes = vec![];
    let mut globals = vec![];

    // NOTE: I considered a lot of different approaches for eliminating locals:
    // - bitset of size source_bytes with bits in @local range set to denote invalidity (space cost too high)
    // - binary search tree of some sort (maybe? not too confident in implementing this)
    // final solution is just iterating over children of a local scope and invalidating
    // them in an id map (which shouldn't be too expensive as ts treecursors are pretty fast)
    let mut local_nodes = HashMap::new();
    let mut tree_cursor = root_node.walk();

    let matches = cursor.matches(&config.query, root_node, source_bytes);

    for m in matches {
        // eprintln!("\n==== NEW MATCH ====");

        let mut node = None;
        let mut scope = None;
        let mut descriptors = vec![];

        for capture in m.captures {
            let capture_name = capture_names
                .get(capture.index as usize)
                .expect("capture indexes should always work");

            if capture_name.starts_with("descriptor") {
                descriptors.push((capture_name, capture.node.utf8_text(source_bytes)?));
                node = Some(capture.node);
            }

            if capture_name.starts_with("scope") {
                assert!(scope.is_none(), "declare only one scope per match");
                scope = Some(capture);
            }

            if capture_name.starts_with("local") {
                for child in capture.node.children(&mut tree_cursor) {
                    local_nodes.insert(child.id(), ());
                }
            }

            // eprintln!(
            //     "{}: {}",
            //     capture_name,
            //     capture.node.utf8_text(source_bytes).unwrap()
            // );
        }

        let node = node.expect("there must always be at least one descriptor");

        if local_nodes.contains_key(&node.id()) {
            continue;
        }

        let descriptors = descriptors
            .iter()
            .map(|(capture, name)| {
                crate::ts_scip::capture_name_to_descriptor(capture, name.to_string())
            })
            .collect();

        // dbg!(node);

        match scope {
            Some(scope_ident) => scopes.push(Scope {
                range: [
                    node.start_position().row as i32,
                    node.start_position().column as i32,
                    node.end_position().column as i32,
                ],
                byte_range: ByteRange {
                    start: scope_ident.node.start_byte(),
                    end: scope_ident.node.end_byte(),
                },
                globals: vec![],
                children: vec![],
                descriptors,
            }),
            None => globals.push(Global {
                range: [
                    node.start_position().row as i32,
                    node.start_position().column as i32,
                    node.end_position().column as i32,
                ],
                byte_range: ByteRange {
                    start: node.start_byte(),
                    end: node.end_byte(),
                },
                descriptors,
            }),
        }
    }

    // dbg!(&matched);

    let mut root = Scope {
        range: [
            root_node.start_position().row as i32,
            root_node.start_position().column as i32,
            root_node.end_position().column as i32,
        ],
        byte_range: ByteRange {
            start: root_node.start_byte(),
            end: root_node.end_byte(),
        },
        globals: vec![],
        children: vec![],
        descriptors: vec![],
    };

    scopes.sort_by_key(|m| {
        (
            std::cmp::Reverse(m.byte_range.start),
            m.byte_range.end - m.byte_range.start,
        )
    });

    // Add all the scopes to our tree
    while let Some(m) = scopes.pop() {
        root.insert_scope(m);
    }

    while let Some(m) = globals.pop() {
        root.insert_global(m);
    }
    // dbg!(&root);

    Ok((root, globals.len()))
}

#[cfg(test)]
mod test {
    use scip::types::Document;
    use scip_treesitter::snapshot::dump_document;
    use tree_sitter::Parser;

    use super::*;

    fn parse_file_for_lang(config: &TagConfiguration, source_code: &str) -> Result<Document> {
        let source_bytes = source_code.as_bytes();

        let mut parser = Parser::new();
        parser.set_language(config.language).unwrap();
        let tree = parser.parse(source_bytes, None).unwrap();

        let mut occ = parse_tree(config, &tree, source_bytes)?;
        let mut doc = Document::new();
        doc.occurrences = occ.0.into_occurrences(occ.1, vec![]);
        doc.symbols = doc
            .occurrences
            .iter()
            .map(|o| scip::types::SymbolInformation {
                symbol: o.symbol.clone(),
                ..Default::default()
            })
            .collect();

        Ok(doc)
    }

    #[test]
    fn test_can_parse_rust_tree() -> Result<()> {
        let config = crate::languages::rust();
        let source_code = include_str!("../testdata/scopes.rs");
        let doc = parse_file_for_lang(config, source_code)?;

        let dumped = dump_document(&doc, source_code)?;
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_parse_go_tree() -> Result<()> {
        let config = crate::languages::go();
        let source_code = include_str!("../testdata/example.go");
        let doc = parse_file_for_lang(config, source_code)?;
        // dbg!(doc);

        let dumped = dump_document(&doc, source_code)?;
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_parse_go_internal_tree() -> Result<()> {
        let config = crate::languages::go();
        let source_code = include_str!("../testdata/internal_go.go");
        let doc = parse_file_for_lang(config, source_code)?;
        // dbg!(doc);

        let dumped = dump_document(&doc, source_code)?;
        insta::assert_snapshot!(dumped);

        Ok(())
    }
}
