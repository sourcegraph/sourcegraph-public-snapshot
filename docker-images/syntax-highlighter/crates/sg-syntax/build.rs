use anyhow::Result;
use std::fs;
use std::path::{Path, PathBuf};

fn collect_tree_sitter_dirs(ignore: &[String]) -> Result<Vec<String>> {
    let mut dirs = Vec::new();
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("languages");

    for entry in fs::read_dir(path)? {
        let entry = entry?;
        let path = entry.path();

        if !entry.file_type()?.is_dir() {
            continue;
        }

        let dir = path.file_name().unwrap().to_str().unwrap().to_string();

        // filter ignores
        if ignore.contains(&dir) {
            continue;
        }
        dirs.push(dir)
    }

    Ok(dirs)
}

fn scanner_file(src_path: &Path) -> Option<PathBuf> {
    let mut scanner_path = src_path.join("scanner.c");

    if scanner_path.exists() {
        Some(scanner_path)
    } else {
        // TODO: This may be a problem if someone uses c++ here?... :'(
        // If I am unfortunately cursed with having to build a C++ thing, as defined by Cargo Book
        scanner_path.set_extension("cc");
        if scanner_path.exists() {
            Some(scanner_path)
        } else {
            None
        }
    }
}

fn build_library(src_path: &Path, language: &str) -> Result<()> {
    let header_path = src_path;
    let parser_path = src_path.join("parser.c");
    let scanner_path = scanner_file(src_path);

    println!("cargo:rerun-if-changed={}", parser_path.display());
    if let Some(scanner) = &scanner_path {
        println!("cargo:rerun-if-changed={}", scanner.display());
    }

    let mut config = cc::Build::new();
    config.opt_level(2);
    config.debug(false);
    config.warnings(false);
    config.shared_flag(false);

    config.include(header_path);

    config.file(parser_path);
    if let Some(scanner) = &scanner_path {
        config.file(scanner);
    }

    config.compile(language);

    Ok(())
}

fn build_dir(dir: &str, language: &str) {
    println!("Build language {}", language);
    if PathBuf::from("languages")
        .join(dir)
        .read_dir()
        .unwrap()
        .next()
        .is_none()
    {
        eprintln!(
            "The directory {} is empty, you probably need to use 'git submodule update --init --recursive'?",
            dir
        );
        std::process::exit(1);
    }

    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("languages")
        .join(dir)
        .join("src");

    build_library(&path, language).unwrap();
}

fn main() {
    let ignore = vec![
        "tree-sitter-typescript".to_string(),
        "tree-sitter-ocaml".to_string(),
    ];

    let dirs = collect_tree_sitter_dirs(&ignore).expect("to have found found the tree-sitter dirs");

    for dir in dirs {
        let language = &dir.strip_prefix("tree-sitter-").unwrap();
        build_dir(&dir, language);
    }

    // TODO: Some languages have to be handled special (when we add them in later);
    //
    // build_dir("tree-sitter-typescript/tsx", "tsx");
    // build_dir("tree-sitter-typescript/typescript", "typescript");
    // build_dir("tree-sitter-ocaml/ocaml", "ocaml");
    // build_dir("tree-sitter-ocaml/interface", "ocaml-interface")
}
