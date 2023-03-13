use anyhow::Result;
use protobuf::Enum;
use scip::types::Descriptor;
use scip_treesitter::prelude::*;
use tree_sitter::Node;

use crate::languages::TagConfiguration;

#[derive(Debug)]
pub struct Root<'a> {
    pub root: Node<'a>,
    pub children: Vec<Matched<'a>>,
}

// #[derive(Debug)]
// pub struct Namespace {}

pub struct Scope<'a> {
    pub definer: Node<'a>,
    pub scope: Node<'a>,
    pub descriptors: Vec<Descriptor>,
    pub children: Vec<Matched<'a>>,
}

impl<'a> std::fmt::Debug for Scope<'a> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let descriptors = dbg_format_descriptors(&self.descriptors);

        write!(
            f,
            "({}, {}, {:?}) -> {:?}",
            self.scope.kind(),
            self.scope.start_position(),
            descriptors,
            self.children
        )
    }
}

pub struct Global<'a> {
    pub node: Node<'a>,
    pub descriptors: Vec<Descriptor>,
}

impl<'a> std::fmt::Debug for Global<'a> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let descriptors = dbg_format_descriptors(&self.descriptors);

        write!(
            f,
            "({}, {}, {:?})",
            self.node.kind(),
            self.node.start_position(),
            descriptors
        )
    }
}

#[derive(Debug)]
// TODO: Root as it's own type?
pub enum Matched<'a> {
    /// The root node of a file
    Root(Root<'a>),

    /// Does not generate a definition, simply a place ot add new descriptors
    /// TODO: Haven't done this one for real yet
    // Namespace(Namespace),

    /// Generates a new definition, and is itself a place to add additional descriptors
    Scope(Scope<'a>),

    /// Generates a new definition, but does not generate a new scope
    Global(Global<'a>),
}

impl<'a> ContainsNode for Matched<'a> {
    fn contains_node(&self, node: &Node) -> bool {
        self.node().contains_node(node)
    }
}

impl<'a> Matched<'a> {
    pub fn node(&self) -> &Node<'a> {
        match self {
            Matched::Root(m) => &m.root,
            Matched::Scope(m) => &m.scope,
            Matched::Global(m) => &m.node,
        }
    }

    // pub fn children(&self) -> &Vec<Matched<'a>> {
    //     match self {
    //         Matched::Root(m) => &m.children,
    //         Matched::Scope(s) => &s.children,
    //         Matched::Global(_) => todo!(),
    //     }
    // }

    pub fn insert(&mut self, m: Matched<'a>) {
        match self {
            Matched::Root(root) => {
                if let Some(child) = root
                    .children
                    .iter_mut()
                    .find(|child| child.contains_node(m.node()))
                {
                    child.insert(m);
                } else {
                    root.children.push(m);
                }
            }
            Matched::Scope(scope) => {
                if let Some(child) = scope
                    .children
                    .iter_mut()
                    .find(|child| child.contains_node(m.node()))
                {
                    child.insert(m);
                } else {
                    scope.children.push(m);
                }
            }
            Matched::Global(_) => unreachable!(),
        }
    }

    pub fn into_occurences(&self) -> Vec<scip::types::Occurrence> {
        self.rec_into_occurrences(&[])
    }

    // TODO: Could we use a dequeue for this to pop on and off quickly?
    // TODO: Could we use a way to format the symbol w/out all the preamble?
    //  Perhaps just a "format_descriptors" function in the lib, that I didn't expose beforehand
    fn rec_into_occurrences(&self, descriptors: &[Descriptor]) -> Vec<scip::types::Occurrence> {
        match self {
            Matched::Root(root) => {
                assert!(descriptors.is_empty(), "root should not have descriptors");
                root.children
                    .iter()
                    .flat_map(|c| c.rec_into_occurrences(descriptors))
                    .collect()
            }
            Matched::Scope(scope) => {
                let mut these_descriptors = descriptors.to_vec();
                these_descriptors.extend(scope.descriptors.iter().cloned());

                let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                    scheme: "scip-ctags".into(),
                    // TODO: Package?
                    package: None.into(),
                    descriptors: these_descriptors,
                    ..Default::default()
                });

                let symbol_roles = scip::types::SymbolRole::Definition.value();
                let mut children = vec![scip::types::Occurrence {
                    range: vec![
                        scope.definer.start_position().row as i32,
                        scope.definer.start_position().column as i32,
                        scope.definer.end_position().column as i32,
                    ],
                    symbol,
                    symbol_roles,
                    // TODO:
                    // syntax_kind: todo!(),
                    ..Default::default()
                }];

                children.extend(scope.children.iter().flat_map(|c| {
                    let mut descriptors = descriptors.to_vec();
                    descriptors.extend(scope.descriptors.iter().cloned());
                    c.rec_into_occurrences(&descriptors)
                }));

                children
            }
            Matched::Global(global) => {
                let mut these_descriptors = descriptors.to_vec();
                these_descriptors.extend(global.descriptors.iter().cloned());

                let symbol = scip::symbol::format_symbol(scip::types::Symbol {
                    scheme: "scip-ctags".into(),
                    // TODO: Package?
                    package: None.into(),
                    descriptors: these_descriptors,
                    ..Default::default()
                });

                let symbol_roles = scip::types::SymbolRole::Definition.value();
                vec![scip::types::Occurrence {
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
                }]
            }
        }
    }
}

pub fn parse_tree<'a>(
    config: &mut TagConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
) -> Result<Vec<scip::types::Occurrence>> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.query.capture_names();

    let mut matched = vec![];
    for m in cursor.matches(&config.query, root_node, source_bytes) {
        println!("\n==== NEW MATCH ====");

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

            println!(
                "{}: {}",
                capture_name,
                capture.node.utf8_text(source_bytes).unwrap()
            );
        }

        let descriptors = descriptors
            .into_iter()
            .map(|(capture, name)| {
                crate::ts_scip::capture_name_to_descriptor(capture, name.to_string())
            })
            .collect::<Vec<_>>();

        let node = node.expect("there must always be at least one descriptor");
        dbg!(node);

        matched.push(match scope {
            Some(scope) => Matched::Scope(Scope {
                definer: node,
                scope: scope.node,
                descriptors,
                children: vec![],
            }),
            None => Matched::Global(Global { node, descriptors }),
        })
    }

    dbg!(&matched);

    let mut root = Matched::Root(Root {
        root: root_node,
        children: vec![],
    });

    matched.sort_by_key(|m| {
        let node = m.node();
        node.end_byte() - node.start_byte()
    });

    while let Some(m) = matched.pop() {
        root.insert(m);
    }
    dbg!(&root);

    let tags = root.into_occurences();
    Ok(dbg!(tags))
}

fn dbg_format_descriptors(descriptors: &[Descriptor]) -> Vec<String> {
    descriptors
        .iter()
        .map(|d| format!("{} ({:?})", d.name, d.suffix))
        .collect::<Vec<_>>()
}

#[cfg(test)]
mod test {
    use scip::types::Document;

    use super::*;
    use crate::snapshot::dump_document;

    fn parse_file_for_lang(config: &mut TagConfiguration, source_code: &str) -> Result<Document> {
        let source_bytes = source_code.as_bytes();
        let tree = config.parser.parse(source_bytes, None).unwrap();

        let occ = parse_tree(config, &tree, source_bytes)?;
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

        let dumped = dump_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_parse_go_tree() -> Result<()> {
        let mut config = crate::languages::go();
        let source_code = include_str!("../testdata/example.go");
        let doc = dbg!(parse_file_for_lang(&mut config, source_code)?);

        let dumped = dump_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }
}
