use anyhow::Result;
use scip_treesitter::types::PackedRange;
use scip_treesitter_languages::parsers::BundledParser;

use crate::languages::{get_tag_configuration, TagConfiguration};

#[derive(Debug)]
pub struct Scope {
    pub start_byte: usize,
    pub end_byte: usize,
    pub packed: PackedRange,
    pub children: Vec<Scope>,
}

impl Scope {
    pub fn insert_scope(&mut self, scope: Scope) {
        if let Some(child) = self
            .children
            .iter_mut()
            .find(|child| child.packed.contains(&scope.packed))
        {
            child.insert_scope(scope);
        } else {
            self.children.push(scope);
        }
    }
}

pub struct Chunk {
    pub text: String,
    pub start_byte: usize,
    pub end_byte: usize,
}

impl std::fmt::Debug for Chunk {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        if self.text.len() < 1000 {
            std::fmt::write(
                f,
                format_args!(
                    "Chunk<{:?},{:?}>({:?})",
                    self.start_byte, self.end_byte, &self.text
                ),
            )
        } else {
            std::fmt::write(
                f,
                format_args!(
                    "Chunk<{:?},{:?}>({:?}...{:?})",
                    self.start_byte,
                    self.end_byte,
                    &self.text[0..25],
                    &self.text[self.text.len() - 20..self.text.len()]
                ),
            )
        }
    }
}

pub fn parse_tree<'a>(
    config: &TagConfiguration,
    tree: &'a tree_sitter::Tree,
    source_bytes: &'a [u8],
) -> Result<Scope> {
    let mut cursor = tree_sitter::QueryCursor::new();

    let root_node = tree.root_node();
    let capture_names = config.chunk_query.capture_names();

    let mut scopes = vec![];

    let matches = cursor.matches(&config.chunk_query, root_node, source_bytes);
    for m in matches {
        let mut start_byte = usize::MAX;
        let mut end_byte = 0;

        let mut start_node = None;
        let mut end_node = None;

        for capture in m.captures {
            let capture_name = capture_names
                .get(capture.index as usize)
                .expect("capture indexes should always work");

            if capture_name.starts_with("chunk") {
                let range = capture.node.range();
                if start_byte > range.start_byte {
                    start_byte = range.start_byte;
                    start_node = Some(capture.node);
                }

                if end_byte < range.end_byte {
                    end_byte = range.end_byte;
                    end_node = Some(capture.node);
                }
            }
        }

        let start_node: PackedRange = start_node.expect("start_node").into();
        let finish_node: PackedRange = end_node.expect("finish_node").into();

        scopes.push(Scope {
            start_byte,
            end_byte,
            packed: PackedRange {
                start_line: start_node.start_line,
                start_col: start_node.start_col,
                end_line: finish_node.end_line,
                end_col: finish_node.end_col,
            },
            children: vec![],
        });
    }

    let range = root_node.range();
    let mut root = Scope {
        start_byte: range.start_byte,
        end_byte: range.end_byte,
        packed: root_node.into(),
        children: vec![],
    };

    scopes.sort_by_key(|m| {
        std::cmp::Reverse((m.packed.start_line, m.packed.end_line, m.packed.start_col))
    });

    // Add all the scopes to our tree
    while let Some(m) = scopes.pop() {
        root.insert_scope(m);
    }

    Ok(root)
}

const CHARS_PER_TOKEN: usize = 4;
const TOKENS_PER_EMBEDDING: usize = 120;
const MAX_CHARS: usize = CHARS_PER_TOKEN * TOKENS_PER_EMBEDDING;

pub fn next_chunk(
    chunks: &mut Vec<Chunk>,
    source_bytes: &[u8],
    start_byte: usize,
    end_byte: usize,
) -> usize {
    if start_byte == end_byte {
        return start_byte;
    }

    let bytes = &source_bytes[start_byte..end_byte];
    let text = String::from_utf8(bytes.to_vec()).unwrap();

    if text.trim().is_empty() {
        return end_byte;
    }

    if let Some(chunk) = chunks.last_mut() {
        if chunk.text.len() + text.len() < MAX_CHARS {
            chunk.text.push_str(&text);
            chunk.end_byte = end_byte;
            return end_byte;
        }
    }

    chunks.push(Chunk {
        text,
        start_byte,
        end_byte,
    });

    return end_byte;
}

pub fn do_the_chunkin(
    chunks: &mut Vec<Chunk>,
    source_bytes: &[u8],
    scope: &Scope,
    mut start_byte: usize,
) -> usize {
    if scope.end_byte - scope.start_byte > MAX_CHARS || scope.end_byte - start_byte > MAX_CHARS {
        for child in &scope.children {
            start_byte = do_the_chunkin(chunks, source_bytes, child, start_byte);
        }

        return next_chunk(chunks, source_bytes, start_byte, scope.end_byte);
    }

    start_byte = next_chunk(chunks, source_bytes, start_byte, scope.start_byte);
    for child in &scope.children {
        if child.end_byte - start_byte > MAX_CHARS {
            for child in &scope.children {
                start_byte = do_the_chunkin(chunks, source_bytes, child, start_byte);
            }
        }

        if child.end_byte - start_byte > MAX_CHARS {
            start_byte = next_chunk(chunks, source_bytes, start_byte, child.end_byte)
        }
    }

    return next_chunk(chunks, source_bytes, start_byte, scope.end_byte);
}

pub fn chunk_file_for_lang(config: &TagConfiguration, source_code: &str) -> Result<Vec<Chunk>> {
    if source_code.len() < MAX_CHARS {
        return Ok(vec![Chunk {
            text: source_code.to_string(),
            start_byte: 0,
            end_byte: source_code.len(),
        }]);
    }

    let source_bytes = source_code.as_bytes();
    let mut parser = config.get_parser();
    let tree = parser.parse(source_bytes, None).unwrap();

    let scope = parse_tree(config, &tree, source_bytes)?;

    let mut chunks = vec![];
    let mut start_byte = 0;
    for child in &scope.children {
        if child.end_byte - child.start_byte > MAX_CHARS {
            if start_byte != child.start_byte {
                start_byte = next_chunk(&mut chunks, source_bytes, start_byte, child.start_byte)
            }

            for child in &scope.children {
                start_byte = do_the_chunkin(&mut chunks, source_bytes, child, start_byte);
            }

            start_byte = next_chunk(&mut chunks, source_bytes, start_byte, child.end_byte)
        }

        if child.end_byte - start_byte > MAX_CHARS {
            // let bytes = &source_bytes[start_byte..child.start_byte];
            // chunks.push(Chunk {
            //     text: String::from_utf8(bytes.to_vec()).unwrap(),
            //     start_byte: child.start_byte,
            //     end_byte: child.end_byte,
            // });
            // start_byte = next_chunk(&mut chunks, source_bytes, start_byte, child.end_byte);
            panic!("OH NO");
        }
    }

    Ok(chunks)
}

pub fn doit() -> Result<Vec<Chunk>> {
    let config = get_tag_configuration(&BundledParser::Go).expect("go");
    let source_code = include_str!("../testdata/go-globals.go");
    chunk_file_for_lang(config, source_code)
}
