use std::collections::HashMap;
use std::path::Path;
use std::process::Command;
use std::{env::temp_dir, path::PathBuf};

use assert_cmd::cargo::cargo_bin;
use assert_cmd::prelude::*;

use scip_treesitter_cli::io::read_index_from_file;

lazy_static::lazy_static! {
    static ref BINARY_LOCATION: PathBuf = {
        match std::env::var("SCIP_CLI_LOCATION") {
            Ok(va) => std::env::current_dir().unwrap().join(va),
            _ => cargo_bin("scip-treesitter-cli"),
        }
    };
}

use scip_treesitter::snapshot::{dump_document_with_config, EmitSymbol, SnapshotOptions};

fn snapshot_syntax_document(doc: &scip::types::Document, source: &str) -> String {
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

#[test]
fn java_e2e_indexing() {
    let out_dir = temp_dir();
    let setup = HashMap::from([(
        PathBuf::from("globals.java"),
        include_str!("../testdata/globals.java").to_string(),
    )]);

    run_index(&out_dir, &setup, vec!["--language", "java"]);

    let index = read_index_from_file(out_dir.join("index.scip"));

    for doc in &index.documents {
        let path = &doc.relative_path;
        let dumped = snapshot_syntax_document(doc, setup.get(&PathBuf::from(&path)).expect("??"));

        insta::assert_snapshot!(path.clone(), dumped);
    }
}

fn prepare(temp: &Path, files: &HashMap<PathBuf, String>) {
    for (path, contents) in files.iter() {
        let file_path = temp.join(path);
        write_file(&file_path, contents);
    }
}

fn run_index(location: &PathBuf, files: &HashMap<PathBuf, String>, extra_arguments: Vec<&str>) {
    prepare(location, files);

    let mut base_args = vec!["index"];
    base_args.extend(extra_arguments);

    let mut cmd = Command::new(BINARY_LOCATION.to_str().unwrap());

    cmd.args(base_args);

    for (path, _) in files.iter() {
        cmd.arg(path.to_str().unwrap());
    }

    cmd.current_dir(location);

    cmd.assert().success();
}

fn write_file(path: &PathBuf, contents: &String) {
    use std::io::Write;

    let output = std::fs::File::create(path).unwrap();
    let mut writer = std::io::BufWriter::new(output);
    writer.write_all(contents.as_bytes()).unwrap();
}
