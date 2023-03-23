use anyhow::Result;
use protobuf::Enum;
use rustc_hash::FxHashMap as HashMap;
use scip::{
    symbol::format_symbol,
    types::{Occurrence, Symbol},
};
use scip_treesitter::prelude::*;
use tree_sitter::Node;

use crate::languages::LocalConfiguration;

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct ByteRange {
    start: usize,
    end: usize,
}

impl ByteRange {
    pub fn contains(&self, other: &Self) -> bool {
        self.start <= other.start && self.end >= other.end
    }
}

#[derive(Debug)]
pub struct Scope<'a> {
    pub scope: Node<'a>,
    pub range: ByteRange,
    pub definitions: HashMap<&'a str, Definition<'a>>,
    pub references: HashMap<&'a str, Vec<Reference<'a>>>,
    pub children: Vec<Scope<'a>>,
}

impl<'a> Eq for Scope<'a> {}

impl<'a> PartialEq for Scope<'a> {
    fn eq(&self, other: &Self) -> bool {
        self.scope.id() == other.scope.id()
    }
}

impl<'a> PartialOrd for Scope<'a> {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        self.range.partial_cmp(&other.range)
    }
}

impl<'a> Ord for Scope<'a> {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.range.cmp(&other.range)
    }
}

impl<'a> Scope<'a> {
    pub fn new(scope: Node<'a>) -> Self {
        Self {
            scope,
            range: ByteRange {
                start: scope.start_byte(),
                end: scope.end_byte(),
            },
            definitions: HashMap::default(),
            references: HashMap::default(),
            children: vec![],
        }
    }

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

    pub fn insert_definition(&mut self, definition: Definition<'a>) {
        // TODO: Probably should assert that this the root node?
        if definition.scope_modifier == ScopeModifier::Global {
            self.definitions.insert(definition.identifier, definition);
            return;
        }

        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.range.contains(&definition.range))
        {
            child.insert_definition(definition)
        } else {
            self.definitions.insert(definition.identifier, definition);
        }
    }

    pub fn insert_reference(&mut self, reference: Reference<'a>) {
        if let Some(definition) = self.definitions.get(&reference.identifier) {
            if definition.node.id() == reference.node.id() {
                return;
            }
        }

        match self
            .children
            .binary_search_by_key(&reference.range.start, |r| r.range.start)
        {
            Ok(_) => {
                // self.children[idx].insert_reference(reference);
                todo!("I'm not sure what to do yet, think more now");
            }
            Err(idx) => match idx {
                0 => self
                    .references
                    .entry(reference.identifier)
                    .or_default()
                    .push(reference),
                idx => {
                    if self.children[idx - 1].range.contains(&reference.range) {
                        self.children[idx - 1].insert_reference(reference)
                    } else {
                        self.references
                            .entry(reference.identifier)
                            .or_default()
                            .push(reference)
                    }
                }
            },
        }
    }

    // This flattens our scope tree so that we don't have any scopes
    // remaining for when we do reference lookups that don't actually
    // contain any definitions. Those are pretty useless.
    pub fn clean_empty_scopes(&mut self) {
        self.children.iter_mut().for_each(|child| {
            child.clean_empty_scopes();
        });

        let mut empty_children = vec![];
        for (idx, child) in self.children.iter().enumerate() {
            if child.definitions.is_empty() {
                empty_children.push(idx);
            }
        }

        let mut to_clean = vec![];
        for idx in empty_children.into_iter().rev() {
            to_clean.push(self.children.remove(idx));
        }

        // Add the children to the parent scope
        for child in to_clean {
            self.children.extend(child.children);
        }

        self.children.sort_by_key(|s| s.range.start);
    }

    pub fn into_occurrences(&mut self, hint: usize) -> Vec<Occurrence> {
        let mut occs = Vec::with_capacity(hint);
        self.rec_into_occurrences(&mut 0, &mut occs);
        occs
    }

    fn rec_into_occurrences(&self, id: &mut usize, occurrences: &mut Vec<Occurrence>) {
        // TODO: I'm a little sad about this.
        //  We could probably make this a runtime option, where `self` has a `sorted` value
        //  that decides whether we need to or not. But on a huge file, this made no difference.
        let mut values = self.definitions.values().collect::<Vec<_>>();
        values.sort_by_key(|d| d.range.start);

        for definition in values {
            *id += 1;

            let symbol = format_symbol(Symbol::new_local(*id));
            let symbol_roles = scip::types::SymbolRole::Definition.value();

            occurrences.push(scip::types::Occurrence {
                range: definition.node.to_scip_range(),
                symbol: symbol.clone(),
                symbol_roles,
                // syntax_kind: todo!(),
                ..Default::default()
            });

            if let Some(references) = self.references.get(definition.identifier) {
                for reference in references {
                    occurrences.push(scip::types::Occurrence {
                        range: reference.node.to_scip_range(),
                        symbol: symbol.clone(),
                        ..Default::default()
                    });
                }
            }

            self.children
                .iter()
                .for_each(|c| c.occurrences_for_children(definition, symbol.as_str(), occurrences));
        }

        self.children
            .iter()
            .for_each(|c| c.rec_into_occurrences(id, occurrences));
    }

    fn occurrences_for_children(
        self: &Scope<'a>,
        def: &Definition<'a>,
        symbol: &str,
        occurrences: &mut Vec<Occurrence>,
    ) {
        if self.definitions.contains_key(def.identifier) {
            return;
        }

        if let Some(references) = self.references.get(def.identifier) {
            for reference in references {
                occurrences.push(scip::types::Occurrence {
                    range: reference.node.to_scip_range(),
                    symbol: symbol.to_string(),
                    ..Default::default()
                });
            }
        }

        self.children
            .iter()
            .for_each(|c| c.occurrences_for_children(def, symbol, occurrences));
    }

    #[allow(dead_code)]
    fn find_scopes_with(
        &'a self,
        scopes: &mut Vec<&Scope<'a>>,
        // predicate: impl Fn(&Scope<'a>) -> bool,
    ) {
        if self.definitions.is_empty() {
            scopes.push(self);
        }

        for child in &self.children {
            child.find_scopes_with(scopes);
        }
    }

    pub fn display_scopes(&self) {
        self.rec_display_scopes(0);
    }

    fn rec_display_scopes(&self, depth: usize) {
        let depth = depth + 1;
        for child in self.children.iter() {
            child.rec_display_scopes(depth);
        }
    }
}

#[derive(Debug, Default, PartialEq, Eq)]
pub enum ScopeModifier {
    #[default]
    Local,
    Parent,
    Global,
}

#[derive(Debug)]
pub struct Definition<'a> {
    pub group: &'a str,
    pub identifier: &'a str,
    pub node: Node<'a>,
    pub range: ByteRange,
    pub scope_modifier: ScopeModifier,
}

#[derive(Debug)]
pub struct Reference<'a> {
    pub group: &'a str,
    pub identifier: &'a str,
    pub node: Node<'a>,
    pub range: ByteRange,
}

pub fn parse_tree<'a>(
    config: &mut LocalConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
) -> Result<Vec<scip::types::Occurrence>> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.query.capture_names();

    let mut scopes = vec![];
    let mut definitions = vec![];
    let mut references = vec![];

    for m in cursor.matches(&config.query, root_node, source_bytes) {
        let mut node = None;

        let mut scope = None;
        let mut definition = None;
        let mut reference = None;
        let mut scope_modifier = None;

        for capture in m.captures {
            let capture_name = match capture_names.get(capture.index as usize) {
                Some(capture_name) => capture_name,
                None => continue,
            };

            node = Some(capture.node);

            if capture_name.starts_with("definition") {
                assert!(definition.is_none(), "only one definition per match");
                definition = Some(capture_name);

                // Handle scope modifiers
                let properties = config.query.property_settings(m.pattern_index);
                for prop in properties {
                    if &(*prop.key) == "scope" {
                        match prop.value.as_deref() {
                            Some("global") => scope_modifier = Some(ScopeModifier::Global),
                            Some("parent") => scope_modifier = Some(ScopeModifier::Parent),
                            Some("local") => scope_modifier = Some(ScopeModifier::Local),
                            // TODO: Should probably error instead
                            Some(other) => panic!("unknown scope-testing: {}", other),
                            None => {}
                        }
                    }
                }
            }

            if capture_name.starts_with("reference") {
                assert!(reference.is_none(), "only one reference per match");
                reference = Some(capture_name);
            }

            if capture_name.starts_with("scope") {
                assert!(scope.is_none(), "declare only one scope per match");
                scope = Some(capture);
            }
        }

        let node = match node {
            Some(node) => node,
            None => continue,
        };

        if let Some(group) = definition {
            let identifier = match node.utf8_text(source_bytes) {
                Ok(identifier) => identifier,
                Err(_) => continue,
            };

            let scope_modifier = scope_modifier.unwrap_or_default();
            definitions.push(Definition {
                range: ByteRange {
                    start: node.start_byte(),
                    end: node.end_byte(),
                },
                group,
                identifier,
                node,
                scope_modifier,
            });
        } else if let Some(group) = reference {
            let identifier = match node.utf8_text(source_bytes) {
                Ok(identifier) => identifier,
                Err(_) => continue,
            };

            references.push(Reference {
                range: ByteRange {
                    start: node.start_byte(),
                    end: node.end_byte(),
                },
                group,
                identifier,
                node,
            });
        } else {
            let scope = match scope {
                Some(scope) => scope,
                None => continue,
            };

            scopes.push(Scope::new(scope.node));
        }
    }

    let mut root = Scope::new(root_node);

    // Sort smallest to largest, so we can pop off the end of the list for the largest, first scope
    scopes.sort_by_key(|m| {
        (
            std::cmp::Reverse(m.range.start),
            m.range.end - m.range.start,
        )
    });

    let capacity = definitions.len() + references.len();

    // Add all the scopes to our tree
    while let Some(m) = scopes.pop() {
        root.insert_scope(m);
    }

    while let Some(m) = definitions.pop() {
        root.insert_definition(m);
    }

    root.clean_empty_scopes();

    while let Some(m) = references.pop() {
        root.insert_reference(m);
    }

    let occs = root.into_occurrences(capacity);

    Ok(occs)
}

#[cfg(test)]
mod test {
    use anyhow::Result;
    use scip::types::Document;
    use scip_treesitter::snapshot::{dump_document_with_config, EmitSymbol, SnapshotOptions};
    use scip_treesitter_languages::parsers::BundledParser;

    use super::*;
    use crate::languages::LocalConfiguration;

    fn snapshot_syntax_document(doc: &Document, source: &str) -> String {
        dump_document_with_config(
            doc,
            source,
            SnapshotOptions {
                emit_symbol: EmitSymbol::All,
                ..Default::default()
            },
        )
        .expect("dump document")
    }

    fn parse_file_for_lang(config: &mut LocalConfiguration, source_code: &str) -> Result<Document> {
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
    fn test_can_do_go() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::Go).unwrap();
        let source_code = include_str!("../testdata/locals.go");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_do_nested_locals() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::Go).unwrap();
        let source_code = include_str!("../testdata/locals-nested.go");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_do_functions() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::Go).unwrap();
        let source_code = include_str!("../testdata/funcs.go");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_do_perl() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::Perl).unwrap();
        let source_code = include_str!("../testdata/perl.pm");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_do_ocaml() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::OCaml).unwrap();
        let source_code = include_str!("../testdata/ocaml.ml");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_do_ocaml_features() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::OCaml).unwrap();
        let source_code = include_str!("../testdata/ocaml-features.ml");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }
}
