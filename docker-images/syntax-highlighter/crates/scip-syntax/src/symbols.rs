use anyhow::Result;
use protobuf::Enum;
use scip::types::{symbol_information, Descriptor, Document, Occurrence, SymbolInformation};
use scip_treesitter::types::PackedRange;

use crate::languages::TagConfiguration;

#[derive(Debug)]
pub struct Scope {
    pub ident_range: PackedRange,
    pub scope_range: PackedRange,
    pub globals: Vec<Global>,
    pub referrences: Vec<Reference>,
    pub children: Vec<Scope>,
    pub descriptors: Vec<Descriptor>,
    pub kind: symbol_information::Kind,
}

#[derive(Debug)]
pub struct Global {
    pub range: PackedRange,
    pub enclosing: Option<PackedRange>,
    pub descriptors: Vec<Descriptor>,
    pub kind: symbol_information::Kind,
}

#[derive(Debug)]
pub struct Reference {
    pub range: PackedRange,
    pub descriptors: Vec<Descriptor>,
    pub kind: symbol_information::Kind,
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

    pub fn insert_reference(&mut self, reference: Reference) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.scope_range.contains(&reference.range))
        {
            child.insert_reference(reference)
        } else {
            self.referrences.push(reference);
        }
    }

    pub fn into_document(&mut self, hint: usize, base_descriptors: Vec<Descriptor>) -> Document {
        let mut descriptor_stack = base_descriptors;

        let mut occurrences = Vec::with_capacity(hint);
        let mut symbols = Vec::with_capacity(hint);
        self.traverse(true, &mut occurrences, &mut descriptor_stack, &mut symbols);

        Document {
            occurrences,
            symbols,
            ..Default::default()
        }
    }

    fn traverse(
        &self,
        is_root: bool,
        occurrences: &mut Vec<Occurrence>,
        descriptor_stack: &mut Vec<Descriptor>,
        symbols: &mut Vec<SymbolInformation>,
    ) {
        descriptor_stack.extend(self.descriptors.clone());

        if !is_root {
            let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                scheme: "scip-ctags".into(),
                package: None.into(),
                descriptors: descriptor_stack.clone(),
                ..Default::default()
            });

            occurrences.push(scip::types::Occurrence {
                range: self.ident_range.to_vec(),
                symbol: symbol.clone(),
                symbol_roles: scip::types::SymbolRole::Definition.value(),
                enclosing_range: self.scope_range.to_vec(),
                ..Default::default()
            });

            symbols.push(SymbolInformation {
                symbol,
                kind: self.kind.into(),
                ..Default::default()
            })
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
                symbol: symbol.clone(),
                symbol_roles,
                enclosing_range: match &global.enclosing {
                    Some(enclosing) => enclosing.to_vec(),
                    None => vec![],
                },
                ..Default::default()
            });

            symbols.push(SymbolInformation {
                symbol,
                kind: global.kind.into(),
                ..Default::default()
            });
        }

        for reference in &self.referrences {
            let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                scheme: "scip-ctags".into(),
                package: None.into(),
                descriptors: reference.descriptors.clone(),
                ..Default::default()
            });

            occurrences.push(scip::types::Occurrence {
                range: reference.range.to_vec(),
                symbol: symbol.clone(),
                ..Default::default()
            });
        }

        self.children
            .iter()
            .for_each(|c| c.traverse(false, occurrences, descriptor_stack, symbols));

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
    let capture_names = config.sym_query.capture_names();

    let mut scopes = vec![];
    let mut globals = vec![];
    let mut references = vec![];

    let matches = cursor.matches(&config.sym_query, root_node, source_bytes);
    for m in matches {
        if config.is_filtered(&m) {
            continue;
        }

        let mut node = None;
        let mut enclosing_node = None;
        let mut scope = None;
        let mut descriptors = vec![];
        let mut reference = None;
        let mut kind = None;

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

            if capture_name.starts_with("kind") {
                assert!(kind.is_none(), "declare only one kind per match");
                kind = Some(capture_name)
            }

            if capture_name.starts_with("reference") {
                assert!(reference.is_none(), "can only have one reference per match");
                reference = Some(node)
            }
        }

        match node {
            Some(node) => {
                let kind = crate::ts_scip::captures_to_kind(&kind);

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
                        referrences: vec![],
                        descriptors,
                        kind,
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
                                        kind,
                                    });
                                }
                            }
                            None => {
                                if reference.is_some() {
                                    references.push(Reference {
                                        range: node.into(),
                                        descriptors,
                                        kind,
                                    })
                                } else {
                                    globals.push(Global {
                                        range: node.into(),
                                        enclosing: enclosing_node.map(|n| n.into()),
                                        descriptors,
                                        kind,
                                    })
                                }
                            }
                        }
                    }
                }
            }
            None => continue,
        }
    }

    let mut root = Scope {
        ident_range: root_node.into(),
        scope_range: root_node.into(),
        globals: vec![],
        children: vec![],
        referrences: vec![],
        descriptors: vec![],
        kind: symbol_information::Kind::UnspecifiedKind,
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

    while let Some(reference) = references.pop() {
        root.insert_reference(reference)
    }

    Ok((root, globals.len()))
}

#[cfg(test)]
mod test {
    use scip_treesitter::snapshot::dump_document;
    use scip_treesitter_languages::parsers::BundledParser;

    use super::*;

    fn parse_file_for_lang(config: &TagConfiguration, source_code: &str) -> Result<Document> {
        let source_bytes = source_code.as_bytes();
        let mut parser = config.get_parser();
        let tree = parser.parse(source_bytes, None).unwrap();

        let (mut scope, hint) = parse_tree(config, &tree, source_bytes)?;
        Ok(scope.into_document(hint, vec![]))
    }

    #[test]
    fn generates_some_go_symbols() -> Result<()> {
        let config = crate::languages::get_tag_configuration(BundledParser::Go).unwrap();
        let source_code = include_str!("../testdata/symbols.go");
        let document = parse_file_for_lang(config, source_code)?;
        let dumped = dump_document(&document, source_code)?;
        insta::assert_snapshot!("generates_go_symbols", dumped);
        Ok(())
    }
}
