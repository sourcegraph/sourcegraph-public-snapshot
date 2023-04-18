use anyhow::Result;
use scip::types::Occurrence;
use scip_treesitter_languages::parsers::BundledParser;
use tree_sitter::Parser;

pub mod byterange;
pub mod ctags;
pub mod globals;
pub mod languages;
pub mod locals;
pub mod ts_scip;

pub fn get_globals<'a>(
    parser: BundledParser,
    source_bytes: &'a [u8],
) -> Option<Result<(globals::Scope, usize)>> {
    let config = languages::get_tag_configuration(parser)?;
    let mut parser = Parser::new();
    parser.set_language(config.language).unwrap();
    let tree = parser.parse(source_bytes, None).unwrap();
    Some(globals::parse_tree(&config, &tree, source_bytes))
}

pub fn get_locals(parser: BundledParser, source_bytes: &[u8]) -> Option<Result<Vec<Occurrence>>> {
    let mut config = languages::get_local_configuration(parser)?;
    let tree = config.parser.parse(source_bytes, None).unwrap();
    Some(locals::parse_tree(&mut config, &tree, source_bytes))
}

#[cfg(test)]
mod test {
    use std::io::BufWriter;

    use crate::ctags::generate_tags;

    #[test]
    fn test_generate_ctags_go_globals() {
        let mut buffer = vec![0u8; 1024];
        let mut buf_writer = BufWriter::new(&mut buffer);

        generate_tags(
            &mut buf_writer,
            "go-globals.go".to_string(),
            include_bytes!("../testdata/go-globals.go"),
        );
        insta::assert_snapshot!(String::from_utf8_lossy(buf_writer.buffer()));
    }

    #[test]
    fn test_generate_ctags_empty_scope() {
        let mut buffer = vec![0u8; 1024];
        let mut buf_writer = BufWriter::new(&mut buffer);

        generate_tags(
            &mut buf_writer,
            "ctags-empty-scope.rs".to_string(),
            include_bytes!("../testdata/ctags-empty-scope.rs"),
        );
        insta::assert_snapshot!(String::from_utf8_lossy(buf_writer.buffer()));
    }
}
