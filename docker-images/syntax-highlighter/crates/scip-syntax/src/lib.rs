use anyhow::Result;
use scip::types::Occurrence;
use scip_treesitter_languages::parsers::BundledParser;

pub mod ctags;
pub mod globals;
pub mod languages;
pub mod locals;
pub mod symbols;
pub mod ts_scip;

pub fn get_symbols(parser: BundledParser, source_bytes: &[u8]) -> Result<scip::types::Document> {
    let config = match crate::languages::get_tag_configuration(parser) {
        Some(config) => config,
        None => return Err(anyhow::anyhow!("Missing config for language")),
    };
    let mut parser = config.get_parser();
    let tree = parser.parse(source_bytes, None).expect("to parse tree");
    let (mut symbol_scope, hint) = symbols::parse_tree(config, &tree, source_bytes)?;
    let document = symbol_scope.into_document(hint, vec![]);
    Ok(document)
}

pub fn get_globals(
    parser: BundledParser,
    source_bytes: &[u8],
) -> Option<Result<(globals::Scope, usize)>> {
    let config = languages::get_tag_configuration(parser)?;
    let mut parser = config.get_parser();
    let tree = parser.parse(source_bytes, None).unwrap();
    Some(globals::parse_tree(config, &tree, source_bytes))
}

pub fn get_locals(parser: BundledParser, source_bytes: &[u8]) -> Option<Result<Vec<Occurrence>>> {
    let config = languages::get_local_configuration(parser)?;
    let mut parser = config.get_parser();
    let tree = parser.parse(source_bytes, None).unwrap();
    Some(locals::parse_tree(config, &tree, source_bytes))
}

#[cfg(test)]
mod test {
    use std::{io::BufWriter, path::Path};

    use scip_treesitter::snapshot::dump_document;
    use scip_treesitter_languages::parsers::BundledParser;

    use crate::ctags::generate_tags;

    macro_rules! generate_tags_and_snapshot {
        (Scip, $scip_name:tt, $filename:tt) => {
            #[test]
            fn $scip_name() {
                let filename = $filename;
                let dumped_name = format!("scip_snapshot_{filename}");

                let source_code = include_str!(concat!("../testdata/", $filename));

                let extension = Path::new(&filename)
                    .extension()
                    .expect("to have extension")
                    .to_str()
                    .expect("to have valid utf8 string");
                let parser =
                    BundledParser::get_parser_from_extension(extension).expect("to have parser");
                let config =
                    crate::languages::get_tag_configuration(parser).expect("to have rust parser");
                let doc = crate::globals::test::parse_file_for_lang(config, &source_code)
                    .expect("to parse document");
                let dumped = dump_document(&doc, &source_code).expect("to dumb document");
                insta::assert_snapshot!(dumped_name, dumped);
            }
        };
        (Tags, $tags_name:tt, $filename:tt) => {
            #[test]
            fn $tags_name() {
                let filename = $filename;

                let mut buffer = vec![0u8; 1024];
                let mut buf_writer = BufWriter::new(&mut buffer);

                let ctags_name = format!("tags_snapshot_{filename}");
                let contents = include_str!(concat!("../testdata/", $filename));

                generate_tags(&mut buf_writer, filename.to_string(), contents.as_bytes());
                insta::assert_snapshot!(ctags_name, String::from_utf8_lossy(buf_writer.buffer()));
            }
        };
        (All, $tags_name:tt, $scip_name:tt, $filename:tt) => {
            generate_tags_and_snapshot!(Tags, $tags_name, $filename);
            generate_tags_and_snapshot!(Scip, $scip_name, $filename);
        };
    }

    // A few smoke tests to make sure that we're generating ctags files correctly
    generate_tags_and_snapshot!(All, test_tags_go, test_scip_go, "go-globals.go");
    generate_tags_and_snapshot!(All, test_tags_rust, test_scip_rust, "ctags-empty-scope.rs");

    // But most tests should go here and just generate scip snapshots
    generate_tags_and_snapshot!(Scip, test_scip_zig, "globals.zig");
    generate_tags_and_snapshot!(All, test_tags_python, test_scip_python, "globals.py");
    generate_tags_and_snapshot!(Scip, test_scip_python_comp, "python-repo-comp.py");
    generate_tags_and_snapshot!(All, test_tags_ruby, test_scip_ruby, "ruby-globals.rb");
    generate_tags_and_snapshot!(Scip, test_scip_java, "globals.java");
    generate_tags_and_snapshot!(Scip, test_scip_typescript, "globals.ts");
    generate_tags_and_snapshot!(All, test_tags_csharp, test_scip_csharp, "globals.cs");
    generate_tags_and_snapshot!(Scip, test_scip_scala, "globals.scala");
    generate_tags_and_snapshot!(All, test_tags_kotlin, test_scip_kotlin, "globals.kt");

    generate_tags_and_snapshot!(Scip, test_scip_go_internal, "internal_go.go");
    generate_tags_and_snapshot!(Scip, test_scip_go_example, "example.go");

    generate_tags_and_snapshot!(Scip, test_scip_rust_scopes, "scopes.rs");

    generate_tags_and_snapshot!(Scip, test_scip_javascript, "globals.js");
    generate_tags_and_snapshot!(Scip, test_scip_javascript_object, "javascript-object.js");

    generate_tags_and_snapshot!(All, test_tags_c_example, test_scip_c_example, "example.c");

    // Test to make sure that kinds are the override behavior
    generate_tags_and_snapshot!(All, test_tags_go_diff, test_scip_go_diff, "go-diff.go");
    generate_tags_and_snapshot!(
        All,
        test_tags_go_constant,
        test_scip_tags_go_constant,
        "go-const.go"
    );
}
