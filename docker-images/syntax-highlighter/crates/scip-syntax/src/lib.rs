use anyhow::Result;
use scip::types::Occurrence;
use scip_treesitter_languages::parsers::BundledParser;

pub mod ctags;
pub mod globals;
pub mod languages;
pub mod locals;
pub mod ts_scip;

pub fn get_globals(parser: BundledParser, source_bytes: &[u8]) -> Option<Result<Vec<Occurrence>>> {
    let mut config = languages::get_tag_configuration(parser)?;
    let tree = config.parser.parse(source_bytes, None).unwrap();
    Some(globals::parse_tree(&mut config, &tree, source_bytes))
}

pub fn get_locals(parser: BundledParser, source_bytes: &[u8]) -> Option<Result<Vec<Occurrence>>> {
    let mut config = languages::get_local_configuration(parser)?;
    let tree = config.parser.parse(source_bytes, None).unwrap();
    Some(locals::parse_tree(&mut config, &tree, source_bytes))
}
