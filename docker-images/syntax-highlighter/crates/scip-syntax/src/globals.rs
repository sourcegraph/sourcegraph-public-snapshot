use std::collections::HashMap;

use anyhow::Result;
use protobuf::Enum;
use scip::types::{descriptor::Suffix, Descriptor, Occurrence};
use scip_treesitter::prelude::*;
use tree_sitter::Node;

use crate::{byterange::ByteRange, languages::TagConfiguration, locals::ScopeModifier};

#[derive(Debug)]
pub struct Scope<'a> {
    pub scope: Node<'a>,
    pub range: ByteRange,
    pub globals: Vec<Global<'a>>,
    pub children: Vec<Scope<'a>>,
    pub descriptors: Vec<Descriptor>,
}

#[derive(Debug)]
pub struct Global<'a> {
    pub node: Node<'a>,
    pub range: ByteRange,
    pub descriptors: Vec<Descriptor>,
}

impl<'a> Scope<'a> {
    pub fn insert_scope(&mut self, scope: Scope<'a>) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.range.contains(&scope.range))
        {
            child.insert_scope(scope);
        } else {
            self.children.push(scope);
        }
    }

    pub fn insert_global(&mut self, global: Global<'a>) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.range.contains(&global.range))
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
                range: vec![
                    self.scope.start_position().row as i32,
                    self.scope.start_position().column as i32,
                    self.scope.end_position().column as i32,
                ],
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
                range: vec![
                    global.node.start_position().row as i32,
                    global.node.start_position().column as i32,
                    global.node.end_position().column as i32,
                ],
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
    config: &mut TagConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
    base_descriptors: Vec<Descriptor>,
) -> Result<Vec<scip::types::Occurrence>> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.query.capture_names();

    let mut scopes = vec![];
    let mut globals = vec![];

    for m in cursor.matches(&config.query, root_node, source_bytes) {
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

            // eprintln!(
            //     "{}: {}",
            //     capture_name,
            //     capture.node.utf8_text(source_bytes).unwrap()
            // );
        }

        let descriptors = descriptors
            .iter()
            .map(|(capture, name)| {
                crate::ts_scip::capture_name_to_descriptor(capture, name.to_string())
            })
            .collect();

        let node = node.expect("there must always be at least one descriptor");
        // dbg!(node);

        match scope {
            Some(scope_ident) => scopes.push(Scope {
                scope: node,
                range: ByteRange {
                    start: scope_ident.node.start_byte(),
                    end: scope_ident.node.end_byte(),
                },
                globals: vec![],
                children: vec![],
                descriptors,
            }),
            None => globals.push(Global {
                node,
                range: ByteRange {
                    start: node.start_byte(),
                    end: node.end_byte(),
                },
                descriptors,
            }),
        }
    }

    // dbg!(&matched);

    let mut root = Scope {
        scope: root_node,
        range: ByteRange {
            start: root_node.start_byte(),
            end: root_node.end_byte(),
        },
        globals: vec![],
        children: vec![],
        descriptors: vec![],
    };

    scopes.sort_by_key(|m| {
        (
            std::cmp::Reverse(m.range.start),
            m.range.end - m.range.start,
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

    let tags = root.into_occurrences(globals.len(), base_descriptors);
    // dbg!(tags);
    Ok(tags)
}

#[cfg(test)]
mod test {
    use scip::types::Document;
    use scip_treesitter::snapshot::dump_document;

    use super::*;

    fn parse_file_for_lang(config: &mut TagConfiguration, source_code: &str) -> Result<Document> {
        let source_bytes = source_code.as_bytes();
        let tree = config.parser.parse(source_bytes, None).unwrap();

        let occ = parse_tree(config, &tree, source_bytes, vec![])?;
        let mut doc = Document::new();
        doc.occurrences = occ;
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
        let mut config = crate::languages::rust();
        let source_code = include_str!("../testdata/scopes.rs");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = dump_document(&doc, source_code)?;
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_parse_go_tree() -> Result<()> {
        let mut config = crate::languages::go();
        let source_code = include_str!("../testdata/example.go");
        let doc = parse_file_for_lang(&mut config, source_code)?;
        // dbg!(doc);

        let dumped = dump_document(&doc, source_code)?;
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_parse_go_internal_tree() -> Result<()> {
        let mut config = crate::languages::go();
        let source_code = include_str!("../testdata/internal_go.go");
        let doc = parse_file_for_lang(&mut config, source_code)?;
        // dbg!(doc);

        let dumped = dump_document(&doc, source_code)?;
        insta::assert_snapshot!(dumped);

        Ok(())
    }
}
