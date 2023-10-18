use anyhow::Result;
use protobuf::Enum;
use rustc_hash::FxHashMap as HashMap;
use scip::{
    symbol::format_symbol,
    types::{Occurrence, Symbol},
};
use scip_treesitter::{prelude::*, types::PackedRange};
use tree_sitter::Node;

use crate::languages::LocalConfiguration;

#[derive(Debug)]
pub struct Scope<'a> {
    pub scope: Node<'a>,
    pub range: PackedRange,
    pub lvalues: HashMap<&'a str, LValue<'a>>,
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
            range: scope.into(),
            lvalues: HashMap::default(),
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

    pub fn insert_lvalue(&mut self, lvalue: LValue<'a>) {
        // TODO: Probably should assert that this the root node?
        if lvalue.scope_modifier == ScopeModifier::Global {
            self.lvalues.insert(lvalue.identifier, lvalue);
            return;
        }

        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.range.contains(&lvalue.range))
        {
            child.insert_lvalue(lvalue)
        } else {
            self.lvalues.insert(lvalue.identifier, lvalue);
        }
    }

    pub fn insert_reference(&mut self, reference: Reference<'a>) {
        if let Some(lvalue) = self.lvalues.get(&reference.identifier) {
            if lvalue.node.id() == reference.node.id() {
                return;
            }
        }

        match self.children.binary_search_by_key(
            &(reference.range.start_line, reference.range.start_col),
            |r| (r.range.start_line, r.range.start_col),
        ) {
            Ok(idx) => {
                let child = &self.children[idx];
                if child.range.end_line == reference.range.end_line
                    && child.range.end_col == reference.range.end_col
                {
                    eprintln!(
                        "Two or more scopes with identical ranges ({:#?}) detected while performing heuristic local code navigation indexing. This is likely an issue with a tree-sitter query. This will be ignored.", reference.range
                    );
                    return;
                }

                self.children[idx].insert_reference(reference);
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
            if child.lvalues.is_empty() {
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

        self.children
            .sort_by_key(|s| (s.range.start_line, s.range.end_line, s.range.start_col));
    }

    pub fn into_occurrences(&mut self, hint: usize) -> Vec<Occurrence> {
        let mut occs = Vec::with_capacity(hint);
        let mut declarations_above = vec![];

        self.rec_into_occurrences(&mut 0, &mut occs, &mut declarations_above);
        occs
    }

    fn rec_into_occurrences(
        &self,
        id: &mut usize,
        occurrences: &mut Vec<Occurrence>,
        declarations_above: &mut Vec<HashMap<&'a str, usize>>,
    ) {
        let mut our_declarations_above = HashMap::<&str, usize>::default();

        // TODO: I'm a little sad about this.
        //  We could probably make this a runtime option, where `self` has a `sorted` value
        //  that decides whether we need to or not. But on a huge file, this made no difference.
        let mut values = self.lvalues.values().collect::<Vec<_>>();
        values.sort_by_key(|d| &d.range);

        for lvalue in values {
            *id += 1;

            let symbol = match lvalue.reassignment_behavior {
                ReassignmentBehavior::NewestIsDefinition => {
                    let symbol = format_symbol(Symbol::new_local(*id));
                    our_declarations_above.insert(lvalue.identifier, *id);
                    let symbol_roles = scip::types::SymbolRole::Definition.value();

                    occurrences.push(scip::types::Occurrence {
                        range: lvalue.node.to_scip_range(),
                        symbol: symbol.clone(),
                        symbol_roles,
                        ..Default::default()
                    });

                    symbol
                }
                ReassignmentBehavior::OldestIsDefinition => {
                    if let Some(above) = declarations_above
                        .into_iter()
                        .rev()
                        .find(|x| x.contains_key(lvalue.identifier))
                    {
                        let symbol = format_symbol(Symbol::new_local(
                            *above.get(lvalue.identifier).unwrap(),
                        ));

                        occurrences.push(scip::types::Occurrence {
                            range: lvalue.node.to_scip_range(),
                            symbol: symbol.clone(),
                            ..Default::default()
                        });

                        continue;
                    } else {
                        let symbol = format_symbol(Symbol::new_local(*id));
                        our_declarations_above.insert(lvalue.identifier, *id);
                        let symbol_roles = scip::types::SymbolRole::Definition.value();

                        occurrences.push(scip::types::Occurrence {
                            range: lvalue.node.to_scip_range(),
                            symbol: symbol.clone(),
                            symbol_roles,
                            ..Default::default()
                        });

                        symbol
                    }
                }
            };

            if let Some(references) = self.references.get(lvalue.identifier) {
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
                .for_each(|c| c.occurrences_for_children(lvalue, symbol.as_str(), occurrences));
        }

        declarations_above.push(our_declarations_above);
        self.children
            .iter()
            .for_each(|c| c.rec_into_occurrences(id, occurrences, declarations_above));
        declarations_above.pop();
    }

    fn occurrences_for_children(
        self: &Scope<'a>,
        def: &LValue<'a>,
        symbol: &str,
        occurrences: &mut Vec<Occurrence>,
    ) {
        if let Some(def) = self.lvalues.get(def.identifier) {
            match def.reassignment_behavior {
                ReassignmentBehavior::NewestIsDefinition => return,
                ReassignmentBehavior::OldestIsDefinition => {}
            }
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
    fn find_scopes_with(&'a self, scopes: &mut Vec<&Scope<'a>>) {
        if self.lvalues.is_empty() {
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

/// Define how strong a definition is, useful for languages that use
/// the same syntax for defining a variable and setting it, like Python.
#[derive(Debug, Default, PartialEq, Eq)]
pub enum ReassignmentBehavior {
    /// a = 10
    /// ^ local 1
    /// a = 10
    /// ^ local 2
    #[default]
    NewestIsDefinition,
    /// a = 10
    /// ^ local 1
    /// a = 10
    /// ^ local 1 (reference)
    OldestIsDefinition,
}

#[derive(Debug)]
pub struct LValue<'a> {
    pub group: &'a str,
    pub identifier: &'a str,
    pub node: Node<'a>,
    pub range: PackedRange,
    pub scope_modifier: ScopeModifier,
    pub reassignment_behavior: ReassignmentBehavior,
}

#[derive(Debug)]
pub struct Reference<'a> {
    pub group: &'a str,
    pub identifier: &'a str,
    pub node: Node<'a>,
    pub range: PackedRange,
}

pub fn parse_tree<'a>(
    config: &LocalConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
) -> Result<Vec<scip::types::Occurrence>> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.query.capture_names();

    let mut scopes = vec![];
    let mut lvalues = vec![];
    let mut references = vec![];

    for m in cursor.matches(&config.query, root_node, source_bytes) {
        let mut node = None;

        let mut scope = None;
        let mut lvalue = None;
        let mut reference = None;
        let mut scope_modifier = None;
        let mut reassignment_behavior = None;

        for capture in m.captures {
            let capture_name = match capture_names.get(capture.index as usize) {
                Some(capture_name) => capture_name,
                None => continue,
            };

            node = Some(capture.node);

            // TODO: Change all captures to lvalue in later PR
            // I don't want to do this now as @definition.[...]
            // is the standard capture we use throughout our codebase
            // beyond just locals, so I want to keep things consistent
            if capture_name.starts_with("definition") {
                assert!(lvalue.is_none(), "only one definition per match");
                lvalue = Some(capture_name);

                // Handle modifiers
                let properties = config.query.property_settings(m.pattern_index);
                for prop in properties {
                    if &(*prop.key) == "scope" {
                        match prop.value.as_deref() {
                            Some("global") => scope_modifier = Some(ScopeModifier::Global),
                            Some("parent") => scope_modifier = Some(ScopeModifier::Parent),
                            Some("local") => scope_modifier = Some(ScopeModifier::Local),
                            Some(_) | None => unreachable!(),
                        }
                    } else if &(*prop.key) == "reassignment_behavior" {
                        match prop.value.as_deref() {
                            Some("newest_is_definition") => {
                                reassignment_behavior =
                                    Some(ReassignmentBehavior::NewestIsDefinition)
                            }
                            Some("oldest_is_definition") => {
                                reassignment_behavior =
                                    Some(ReassignmentBehavior::OldestIsDefinition)
                            }
                            Some(_) | None => unreachable!(),
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

        if let Some(group) = lvalue {
            let identifier = match node.utf8_text(source_bytes) {
                Ok(identifier) => identifier,
                Err(_) => continue,
            };

            let scope_modifier = scope_modifier.unwrap_or_default();
            let reassignment_behavior = reassignment_behavior.unwrap_or_default();

            lvalues.push(LValue {
                range: node.into(),
                group,
                identifier,
                node,
                scope_modifier,
                reassignment_behavior,
            });
        } else if let Some(group) = reference {
            let identifier = match node.utf8_text(source_bytes) {
                Ok(identifier) => identifier,
                Err(_) => continue,
            };

            references.push(Reference {
                range: node.into(),
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
            std::cmp::Reverse(m.range.start_line),
            m.range.end_line - m.range.start_line,
            m.range.end_col - m.range.start_col,
        )
    });

    let capacity = lvalues.len() + references.len();

    // Add all the scopes to our tree
    while let Some(m) = scopes.pop() {
        root.insert_scope(m);
    }

    while let Some(m) = lvalues.pop() {
        root.insert_lvalue(m);
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

    fn parse_file_for_lang(config: &LocalConfiguration, source_code: &str) -> Result<Document> {
        let source_bytes = source_code.as_bytes();
        let mut parser = config.get_parser();
        let tree = parser.parse(source_bytes, None).unwrap();

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
    fn test_can_do_matlab() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::Matlab).unwrap();
        let source_code = include_str!("../testdata/locals.m");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }

    #[test]
    fn test_can_do_java() -> Result<()> {
        let mut config = crate::languages::get_local_configuration(BundledParser::Java).unwrap();
        let source_code = include_str!("../testdata/locals.java");
        let doc = parse_file_for_lang(&mut config, source_code)?;

        let dumped = snapshot_syntax_document(&doc, source_code);
        insta::assert_snapshot!(dumped);

        Ok(())
    }
}
