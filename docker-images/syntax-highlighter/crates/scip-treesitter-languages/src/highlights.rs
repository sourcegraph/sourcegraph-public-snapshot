use once_cell::sync::OnceCell;
use paste::paste;
use scip::types::SyntaxKind;
use scip_macros::include_scip_query;
use tree_sitter_highlight::{Highlight, HighlightConfiguration};

use crate::parsers::BundledParser;

#[rustfmt::skip]
// Table of (@CaptureGroup, SyntaxKind) mapping.
//
// Any capture defined in a query will be mapped to the following SyntaxKind via the highlighter.
//
// To extend what types of captures are included, simply add a line below that takes a particular
// match group that you're interested in and map it to a new SyntaxKind.
//
// We can also define our own new capture types that we want to use and add to queries to provide
// particular highlights if necessary.
//
// (I can also add per-language mappings for these if we want, but you could also just do that with
//  unique match groups. For example `@rust-bracket`, or similar. That doesn't need any
//  particularly new rust code to be written. You can just modify queries for that)
const MATCHES_TO_SYNTAX_KINDS: &[(&str, SyntaxKind)] = &[
    ("boolean",                 SyntaxKind::BooleanLiteral),
    ("character",               SyntaxKind::CharacterLiteral),
    ("comment",                 SyntaxKind::Comment),
    ("conditional",             SyntaxKind::IdentifierKeyword),
    ("constant",                SyntaxKind::IdentifierConstant),
    ("identifier.constant",     SyntaxKind::IdentifierConstant),
    ("constant.builtin",        SyntaxKind::IdentifierBuiltin),
    ("constant.null",           SyntaxKind::IdentifierNull),
    ("float",                   SyntaxKind::NumericLiteral),
    ("function",                SyntaxKind::IdentifierFunction),
    ("method",                  SyntaxKind::IdentifierFunction),
    ("identifier.function",     SyntaxKind::IdentifierFunction),
    ("function.builtin",        SyntaxKind::IdentifierBuiltin),
    ("identifier.builtin",      SyntaxKind::IdentifierBuiltin),
    ("identifier",              SyntaxKind::Identifier),
    ("identifier.attribute",    SyntaxKind::IdentifierAttribute),
    ("tag.attribute",           SyntaxKind::TagAttribute),
    ("include",                 SyntaxKind::IdentifierKeyword),
    ("keyword",                 SyntaxKind::IdentifierKeyword),
    ("keyword.function",        SyntaxKind::IdentifierKeyword),
    ("keyword.return",          SyntaxKind::IdentifierKeyword),
    ("number",                  SyntaxKind::NumericLiteral),
    ("operator",                SyntaxKind::IdentifierOperator),
    ("identifier.operator",     SyntaxKind::IdentifierOperator),
    ("property",                SyntaxKind::Identifier),
    ("punctuation",             SyntaxKind::UnspecifiedSyntaxKind),
    ("punctuation.bracket",     SyntaxKind::UnspecifiedSyntaxKind),
    ("punctuation.delimiter",   SyntaxKind::PunctuationDelimiter),
    ("string",                  SyntaxKind::StringLiteral),
    ("string.special",          SyntaxKind::StringLiteral),
    ("string.escape",           SyntaxKind::StringLiteralEscape),
    ("tag",                     SyntaxKind::UnspecifiedSyntaxKind),
    ("type",                    SyntaxKind::IdentifierType),
    ("identifier.type",         SyntaxKind::IdentifierType),
    ("type.builtin",            SyntaxKind::IdentifierBuiltinType),
    ("regex.delimiter",         SyntaxKind::RegexDelimiter),
    ("regex.join",              SyntaxKind::RegexJoin),
    ("regex.escape",            SyntaxKind::RegexEscape),
    ("regex.repeated",          SyntaxKind::RegexRepeated),
    ("regex.wildcard",          SyntaxKind::RegexWildcard),
    ("identifier",              SyntaxKind::Identifier),
    ("variable",                SyntaxKind::Identifier),
    ("identifier.builtin",      SyntaxKind::IdentifierBuiltin),
    ("variable.builtin",        SyntaxKind::IdentifierBuiltin),
    ("identifier.parameter",    SyntaxKind::IdentifierParameter),
    ("variable.parameter",      SyntaxKind::IdentifierParameter),
    ("identifier.module",       SyntaxKind::IdentifierModule),
    ("variable.module",         SyntaxKind::IdentifierModule),
];

fn get_highlight_names() -> Vec<&'static str> {
    MATCHES_TO_SYNTAX_KINDS
        .iter()
        .map(|hl| hl.0)
        .collect::<Vec<&str>>()
}

macro_rules! make_configuration {
    ($fn_name:ident, $name:ty) => {
        fn $fn_name() -> &'static HighlightConfiguration {
            static INSTANCE: OnceCell<HighlightConfiguration> = OnceCell::new();
            INSTANCE.get_or_init(|| {
                let highlight_names = get_highlight_names();
                // Create HighlightConfiguration language
                let mut lang = HighlightConfiguration::new(
                    paste! { BundledParser::$name.get_language() },
                    include_scip_query!($name, "highlights"),
                    include_scip_query!($name, "injections"),
                    include_scip_query!($name, "locals"),
                )
                .expect(stringify!("parser for '{}' must be compiled", $name));

                // Associate highlights with configuration
                lang.configure(&highlight_names);
                lang
            })
        }
    };
}

// Add a language highlight configuration to the CONFIGURATIONS global.
make_configuration!(get_configuration_c, C);
make_configuration!(get_configuration_cpp, Cpp);
make_configuration!(get_configuration_c_sharp, C_Sharp);
make_configuration!(get_configuration_go, Go);
make_configuration!(get_configuration_java, Java);
make_configuration!(get_configuration_javascript, Javascript);
make_configuration!(get_configuration_jsonnet, Jsonnet);
make_configuration!(get_configuration_kotlin, Kotlin);
make_configuration!(get_configuration_nickel, Nickel);
make_configuration!(get_configuration_perl, Perl);
make_configuration!(get_configuration_pod, Pod);
make_configuration!(get_configuration_python, Python);
make_configuration!(get_configuration_ruby, Ruby);
make_configuration!(get_configuration_rust, Rust);
make_configuration!(get_configuration_scala, Scala);
make_configuration!(get_configuration_sql, Sql);
make_configuration!(get_configuration_xlsg, Xlsg);
make_configuration!(get_configuration_zig, Zig);

// For languages that have special one-off cases for their highlighters and queries,
// you can manually specify them here.
fn get_configuration_typescript() -> &'static HighlightConfiguration {
    static INSTANCE: OnceCell<HighlightConfiguration> = OnceCell::new();
    INSTANCE.get_or_init(|| {
        let highlight_names = get_highlight_names();

        let highlights = vec![
            include_scip_query!("typescript", "highlights"),
            include_scip_query!("javascript", "highlights"),
        ];
        let mut lang = HighlightConfiguration::new(
            BundledParser::Typescript.get_language(),
            &highlights.join("\n"),
            include_scip_query!("typescript", "injections"),
            include_scip_query!("typescript", "locals"),
        )
        .expect("parser for 'typescript' must be compiled");

        lang.configure(&highlight_names);
        lang
    })
}

fn get_configuration_tsx() -> &'static HighlightConfiguration {
    static INSTANCE: OnceCell<HighlightConfiguration> = OnceCell::new();
    INSTANCE.get_or_init(|| {
        let highlight_names = get_highlight_names();

        let highlights = vec![
            include_scip_query!("tsx", "highlights"),
            include_scip_query!("typescript", "highlights"),
            include_scip_query!("javascript", "highlights"),
        ];
        let mut lang = HighlightConfiguration::new(
            BundledParser::Tsx.get_language(),
            &highlights.join("\n"),
            include_scip_query!("tsx", "injections"),
            include_scip_query!("tsx", "locals"),
        )
        .expect("parser for 'tsx' must be compiled");

        lang.configure(&highlight_names);
        lang
    })
}

pub fn get_highlighting_configuration(filetype: &str) -> Option<&'static HighlightConfiguration> {
    BundledParser::get_parser(filetype).map(|parser| match parser {
        BundledParser::C => get_configuration_c(),
        BundledParser::Cpp => get_configuration_cpp(),
        BundledParser::C_Sharp => get_configuration_c_sharp(),
        BundledParser::Go => get_configuration_go(),
        BundledParser::Java => get_configuration_java(),
        BundledParser::Javascript => get_configuration_javascript(),
        BundledParser::Jsonnet => get_configuration_jsonnet(),
        BundledParser::Kotlin => get_configuration_kotlin(),
        BundledParser::Nickel => get_configuration_nickel(),
        BundledParser::Perl => get_configuration_perl(),
        BundledParser::Pod => get_configuration_pod(),
        BundledParser::Python => get_configuration_python(),
        BundledParser::Ruby => get_configuration_ruby(),
        BundledParser::Rust => get_configuration_rust(),
        BundledParser::Scala => get_configuration_scala(),
        BundledParser::Sql => get_configuration_sql(),
        BundledParser::Xlsg => get_configuration_xlsg(),
        BundledParser::Typescript => get_configuration_typescript(),
        BundledParser::Tsx => get_configuration_tsx(),
        BundledParser::Zig => get_configuration_zig(),
    })
}

pub fn get_syntax_kind_for_hl(hl: Highlight) -> SyntaxKind {
    MATCHES_TO_SYNTAX_KINDS[hl.0].1
}
