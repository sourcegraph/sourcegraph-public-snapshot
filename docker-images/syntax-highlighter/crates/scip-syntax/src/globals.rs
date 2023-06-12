use anyhow::Result;
use bitvec::prelude::*;
use protobuf::Enum;
use scip::types::{Descriptor, Occurrence};
use scip_treesitter::types::PackedRange;

use crate::languages::TagConfiguration;

#[derive(Debug)]
pub struct Scope {
    pub ident_range: PackedRange,
    pub scope_range: PackedRange,
    pub globals: Vec<Global>,
    pub children: Vec<Scope>,
    pub descriptors: Vec<Descriptor>,
}

#[derive(Debug)]
pub struct Global {
    pub range: PackedRange,
    pub enclosing: Option<PackedRange>,
    pub descriptors: Vec<Descriptor>,
}

impl Scope {
    pub fn insert_scope(&mut self, scope: Scope) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.scope_range.contains(&scope.scope_range))
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
            .find(|child| child.scope_range.contains(&global.range))
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
                range: self.ident_range.to_vec(),
                symbol: scip::symbol::format_symbol(scip::types::Symbol {
                    scheme: "scip-ctags".into(),
                    package: None.into(),
                    descriptors: descriptor_stack.clone(),
                    ..Default::default()
                }),
                symbol_roles: scip::types::SymbolRole::Definition.value(),
                enclosing_range: self.scope_range.to_vec(),
                ..Default::default()
            });
        }

        for global in &self.globals {
            let mut global_descriptors = descriptor_stack.clone();
            global_descriptors.extend(global.descriptors.clone());

            let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                scheme: "scip-ctags".into(),
                package: None.into(),
                descriptors: global_descriptors,
                ..Default::default()
            });

            let symbol_roles = scip::types::SymbolRole::Definition.value();
            occurrences.push(scip::types::Occurrence {
                range: global.range.to_vec(),
                symbol,
                symbol_roles,
                enclosing_range: match &global.enclosing {
                    Some(enclosing) => enclosing.to_vec(),
                    None => vec![],
                },
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

    let mut local_ranges = BitVec::<u8, Msb0>::repeat(false, source_bytes.len());

    let matches = cursor.matches(&config.query, root_node, source_bytes);
    for m in matches {
        if config.is_filtered(&m) {
            continue;
        }

        let mut node = None;
        let mut enclosing_node = None;
        let mut scope = None;
        let mut local_range = None;
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

            if capture_name.starts_with("enclosing") {
                assert!(enclosing_node.is_none(), "declare only one scope per match");
                enclosing_node = Some(capture.node);
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

                let descriptors = descriptors
                    .iter()
                    .map(|(capture, name)| {
                        crate::ts_scip::capture_name_to_descriptor(capture, name.to_string())
                    })
                    .collect();

                match scope {
                    Some(scope_ident) => scopes.push(Scope {
                        ident_range: node.into(),
                        scope_range: scope_ident.node.into(),
                        globals: vec![],
                        children: vec![],
                        descriptors,
                    }),
                    None => {
                        let (last, rest) = match descriptors.split_last() {
                            Some(res) => res,
                            None => continue,
                        };

                        match config.transform(m.pattern_index, last) {
                            Some(transformed) => {
                                for transform in transformed {
                                    globals.push(Global {
                                        range: node.into(),
                                        enclosing: enclosing_node.map(|n| n.into()),
                                        descriptors: {
                                            let mut descriptors = rest
                                                .iter()
                                                .map(|d| (*d).clone())
                                                .collect::<Vec<_>>();
                                            descriptors.push(transform);
                                            descriptors
                                        },
                                    });
                                }
                            }
                            None => globals.push(Global {
                                range: node.into(),
                                enclosing: enclosing_node.map(|n| n.into()),
                                descriptors,
                            }),
                        }
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

    let mut root = Scope {
        ident_range: root_node.into(),
        scope_range: root_node.into(),
        globals: vec![],
        children: vec![],
        descriptors: vec![],
    };

    scopes.sort_by_key(|m| {
        std::cmp::Reverse((
            m.scope_range.start_line,
            m.scope_range.end_line,
            m.scope_range.start_col,
        ))
    });

    // Add all the scopes to our tree
    while let Some(m) = scopes.pop() {
        root.insert_scope(m);
    }

    while let Some(m) = globals.pop() {
        root.insert_global(m);
    }

    Ok((root, globals.len()))
}

#[cfg(test)]
pub mod test {
    use scip::types::Document;
    use scip_treesitter::snapshot::{dump_document_with_config, SnapshotOptions};
    use scip_treesitter_languages::parsers::BundledParser;

    use super::*;

    pub fn parse_file_for_lang(config: &TagConfiguration, source_code: &str) -> Result<Document> {
        let source_bytes = source_code.as_bytes();
        let mut parser = config.get_parser();
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
    fn test_enclosing_range() -> Result<()> {
        let config =
            crate::languages::get_tag_configuration(&BundledParser::Go).expect("to have parser");
        let source_code = include_str!("../testdata/scopes_of_go.go");
        let doc = parse_file_for_lang(config, source_code)?;

        // let dumped = dump_document(&doc, source_code)?;
        let dumped = dump_document_with_config(
            &doc,
            source_code,
            SnapshotOptions {
                snapshot_range: None,
                emit_syntax: scip_treesitter::snapshot::EmitSyntax::None,
                emit_symbol: scip_treesitter::snapshot::EmitSymbol::Enclosing,
            },
        )?;

        insta::assert_snapshot!(dumped);

        Ok(())
    }
}
