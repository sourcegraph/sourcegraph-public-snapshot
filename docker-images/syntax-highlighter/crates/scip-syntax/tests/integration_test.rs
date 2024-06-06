use std::{
    collections::{HashMap, HashSet},
    path::{Path, PathBuf},
    process::Command,
};

use assert_cmd::{cargo::cargo_bin, prelude::*};
use scip_syntax::{
    evaluate::Evaluator,
    index::{index_command, AnalysisMode, IndexMode, IndexOptions},
    io::read_index_from_file,
};

lazy_static::lazy_static! {
    static ref BINARY_LOCATION: PathBuf = {
        match std::env::var("SCIP_SYNTAX_PATH") {
            Ok(va) => std::env::current_dir().unwrap().join(va),
            _ => cargo_bin("scip-syntax"),
        }
    };

    static ref BASE: PathBuf = {
        match std::env::var("CARGO_MANIFEST_DIR") {
            Ok(va) => std::env::current_dir().unwrap().join(va),
            _ => std::env::current_dir().unwrap()            }
    };

    static ref JAVA_SCIP_INDEX: PathBuf = {
        match std::env::var("JAVA_SCIP_INDEX") {
            Ok(va) => std::env::current_dir().unwrap().join(va),
            _ => BASE.join("testdata/java/index.scip")
        }
    };


}

use syntax_analysis::snapshot::{dump_document_with_config, EmitSymbol, SnapshotOptions};

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
fn java_e2e_evaluation() {
    let dir = BASE.join("testdata/java");

    let out_dir = tempdir();

    let candidate = out_dir.join("index-tree-sitter.scip");

    index_command(
        "java".to_string(),
        IndexMode::Workspace {
            location: dir.clone(),
        },
        candidate.clone(),
        dir.clone(),
        None,
        IndexOptions {
            analysis_mode: AnalysisMode::Full,
            fail_fast: true,
        },
    )
    .unwrap();

    let mut str = vec![];

    Evaluator::default()
        .evaluate_files(candidate, JAVA_SCIP_INDEX.to_path_buf())
        .unwrap()
        .write_summary(
            &mut str,
            scip_syntax::evaluate::EvaluationOutputOptions {
                print_false_negatives: true,
                print_true_positives: true,
                print_false_positives: true,
                print_mapping: true,
                disable_colors: true,
            },
        )
        .unwrap();

    insta::assert_snapshot!("java_evaluation", String::from_utf8(str).unwrap());
}

#[test]
fn java_e2e_indexing() {
    let out_dir = tempdir();
    let setup = HashMap::from([(
        PathBuf::from("globals.java"),
        include_str!("../testdata/globals.java").to_string(),
    )]);

    let mut cmd = command("index");
    let output_location = out_dir.join("index.scip");
    let paths: Vec<String> = setup
        .clone()
        .into_keys()
        .map(|pb| pb.to_str().unwrap().to_string())
        .collect();

    prepare(&out_dir, &setup);

    cmd.args(vec![
        "--language",
        "java",
        "--out",
        output_location.to_str().unwrap(),
    ])
    .current_dir(out_dir)
    .args(paths)
    .assert()
    .success();

    let index = read_index_from_file(&output_location).unwrap();
    let mut indexed_paths: HashSet<String> = HashSet::new();

    for doc in &index.documents {
        let path = &doc.relative_path;
        indexed_paths.insert(path.to_string());
        let dumped = snapshot_syntax_document(
            doc,
            setup.get(&PathBuf::from(&path)).expect(
                format!(
                    "Unexpected relative path {} found in the index. Valid paths are: {:?}",
                    path,
                    setup.keys()
                )
                .as_str(),
            ),
        );

        insta::assert_snapshot!(format!("files_indexing-{}", path.clone()), &dumped);
    }

    assert_eq!(
        indexed_paths,
        setup
            .into_keys()
            .map(|p| p.to_str().unwrap().to_string())
            .collect()
    );
}

#[test]
fn java_workspace_indexing() {
    let out_dir = tempdir();
    let setup = HashMap::from([
        (
            PathBuf::from("src/main/java/globals.java"),
            include_str!("../testdata/globals.java").to_string(),
        ),
        (
            PathBuf::from("package-info.java"),
            include_str!("../testdata/package-info.java").to_string(),
        ),
    ]);

    let mut cmd = command("index");
    let output_location = out_dir.join("index.scip");

    prepare(&out_dir, &setup);

    cmd.args(vec![
        "--workspace",
        out_dir.to_str().unwrap(),
        "--language",
        "java",
        "--out",
        output_location.to_str().unwrap(),
    ])
    .assert()
    .success();

    let index = read_index_from_file(&output_location).unwrap();

    let mut indexed_paths: HashSet<String> = HashSet::new();

    for doc in &index.documents {
        let path = &doc.relative_path;
        indexed_paths.insert(path.to_string());
        let dumped = snapshot_syntax_document(
            doc,
            setup.get(&PathBuf::from(&path)).expect(
                format!(
                    "Unexpected relative path {} found in the index. Valid paths are: {:?}",
                    path,
                    setup.keys()
                )
                .as_str(),
            ),
        );

        insta::assert_snapshot!(format!("workspace_indexing-{}", path.clone()), &dumped);
    }

    println!("{:?}", index.documents);

    assert_eq!(
        indexed_paths,
        setup
            .into_keys()
            .map(|p| p.to_str().unwrap().to_string())
            .collect()
    );
}

fn prepare(temp: &Path, files: &HashMap<PathBuf, String>) {
    for (path, contents) in files.iter() {
        let file_path = temp.join(path);
        write_file(&file_path, contents);
    }
}

fn command(sub: &str) -> Command {
    let mut cmd = Command::new(BINARY_LOCATION.to_str().unwrap());

    cmd.arg(sub);

    cmd
}

fn write_file(path: &PathBuf, contents: &String) {
    use std::io::Write;

    let Some(parent) = path.parent() else {
        panic!("failed to find parent dir for {:?}", path)
    };

    std::fs::create_dir_all(parent)
        .expect(format!("Failed to create all parent folders for {:?}", path).as_str());

    let output = std::fs::File::create(path)
        .expect(format!("Failed to open file {} for writing", path.to_str().unwrap()).as_str());
    let mut writer = std::io::BufWriter::new(output);
    writer.write_all(contents.as_bytes()).unwrap();
}

fn tempdir() -> PathBuf {
    tempfile::tempdir().unwrap().into_path()
}
