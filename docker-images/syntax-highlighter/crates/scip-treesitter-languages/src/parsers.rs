use tree_sitter::Language;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum BundledParser {
    C,
    Cpp,
    #[allow(non_camel_case_types)]
    C_Sharp,
    Go,
    Java,
    Javascript,
    Jsonnet,
    Nickel,
    Perl,
    Pod,
    Python,
    Ruby,
    Rust,
    Scala,
    Sql,
    Xlsg,
    Zig,

    // These two are special cases
    Typescript,
    Tsx,
}

impl BundledParser {
    pub fn get_language(&self) -> Language {
        match self {
            BundledParser::C => tree_sitter_c::language(),
            BundledParser::Cpp => tree_sitter_cpp::language(),
            BundledParser::C_Sharp => tree_sitter_c_sharp::language(),
            BundledParser::Go => tree_sitter_go::language(),
            BundledParser::Java => tree_sitter_java::language(),
            BundledParser::Javascript => tree_sitter_javascript::language(),
            BundledParser::Jsonnet => tree_sitter_jsonnet::language(),
            BundledParser::Nickel => tree_sitter_nickel::language(),
            BundledParser::Perl => tree_sitter_perl::language(),
            BundledParser::Pod => tree_sitter_pod::language(),
            BundledParser::Python => tree_sitter_python::language(),
            BundledParser::Ruby => tree_sitter_ruby::language(),
            BundledParser::Rust => tree_sitter_rust::language(),
            BundledParser::Scala => tree_sitter_scala::language(),
            BundledParser::Sql => tree_sitter_sql::language(),
            BundledParser::Typescript => tree_sitter_typescript::language_typescript(),
            BundledParser::Tsx => tree_sitter_typescript::language_tsx(),
            BundledParser::Xlsg => tree_sitter_xlsg::language(),
            BundledParser::Zig => tree_sitter_zig::language(),
        }
    }

    pub fn get_parser(name: &str) -> Option<Self> {
        match name {
            "c" => Some(BundledParser::C),
            "cpp" => Some(BundledParser::Cpp),
            "c_sharp" => Some(BundledParser::C_Sharp),
            "go" => Some(BundledParser::Go),
            "java" => Some(BundledParser::Java),
            "javascript" => Some(BundledParser::Javascript),
            "jsonnet" => Some(BundledParser::Jsonnet),
            "nickel" => Some(BundledParser::Nickel),
            "perl" => Some(BundledParser::Perl),
            "pod" => Some(BundledParser::Pod),
            "python" => Some(BundledParser::Python),
            "ruby" => Some(BundledParser::Ruby),
            "rust" => Some(BundledParser::Rust),
            "scala" => Some(BundledParser::Scala),
            "sql" => Some(BundledParser::Sql),
            "typescript" => Some(BundledParser::Typescript),
            "tsx" => Some(BundledParser::Tsx),
            "xlsg" => Some(BundledParser::Xlsg),
            "zig" => Some(BundledParser::Zig),
            _ => None,
        }
    }
}
