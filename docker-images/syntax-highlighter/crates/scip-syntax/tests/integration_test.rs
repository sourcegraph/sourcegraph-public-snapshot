use std::{
    collections::{HashMap, HashSet},
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

use anyhow::{anyhow, Context};
use assert_cmd::{cargo::cargo_bin, prelude::*};
use scip::types::Document;
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
use tar::{Builder, Header};

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
fn java_files_indexing() {
    let out_dir = tempdir();
    let setup = indexing_data();

    let mut cmd = command("index");
    let output_location = out_dir.join("index.scip");
    let paths = extract_paths(&setup);

    prepare(&out_dir, &setup);

    cmd.args(vec![
        "--language",
        "java",
        "--out",
        output_location.to_str().unwrap(),
    ])
    .current_dir(&out_dir)
    .args(paths)
    .assert()
    .success();

    let index = read_index_from_file(&output_location).unwrap();

    assert_eq!(extract_paths(&setup), extract_indexed_paths(&index));

    let index_snapshot = snapshot_from_files(&index.documents, &out_dir);

    insta::assert_snapshot!(index_snapshot);
}

#[test]
fn java_workspace_indexing() {
    let out_dir = tempdir();
    let setup = indexing_data();

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

    assert_eq!(extract_paths(&setup), extract_indexed_paths(&index));

    let index_snapshot = snapshot_from_files(&index.documents, &out_dir);

    insta::assert_snapshot!(index_snapshot);
}

#[test]
fn java_tar_file_indexing() {
    let out_dir = tempdir();
    let setup = indexing_data();
    let tar_data = create_tar(&setup);

    let data = tar_data.unwrap();

    let mut cmd = command("index");
    let tar_file = out_dir.join("test.tar");
    let output_location = out_dir.join("index.scip");

    write_file_bytes(&tar_file, &data);

    cmd.args(vec![
        "--tar",
        tar_file.to_str().unwrap(),
        "--language",
        "java",
        "--out",
        output_location.to_str().unwrap(),
    ])
    .assert()
    .success();

    let index = read_index_from_file(&output_location).unwrap();

    assert_eq!(extract_paths(&setup), extract_indexed_paths(&index));

    let index_snapshot = snapshot_from_data(&index.documents, &setup);

    insta::assert_snapshot!(index_snapshot);
}

#[test]
fn java_tar_stream_indexing() {
    let out_dir = tempdir();
    let setup = indexing_data();
    let tar_data = create_tar(&setup);

    let data = tar_data.unwrap();

    let mut cmd = command("index");
    let tar_file = out_dir.join("test.tar");
    let output_location = out_dir.join("index.scip");

    write_file_bytes(&tar_file, &data);

    let mut spawned = cmd
        .args(vec![
            "--tar",
            "-",
            "--language",
            "java",
            "--out",
            output_location.to_str().unwrap(),
        ])
        .stdin(Stdio::piped())
        .spawn()
        .unwrap();

    spawned.stdin.take().unwrap().write_all(&data).unwrap();

    let exit_status = spawned.wait().unwrap();

    assert_eq!(exit_status.code(), Some(0));

    let index = read_index_from_file(&output_location).unwrap();

    assert_eq!(extract_paths(&setup), extract_indexed_paths(&index));

    let index_snapshot = snapshot_from_data(&index.documents, &setup);

    insta::assert_snapshot!(index_snapshot);
}

fn prepare(temp: &Path, files: &HashMap<PathBuf, String>) {
    for (path, contents) in files.iter() {
        let file_path = temp.join(path);
        write_file_string(&file_path, contents);
    }
}

fn command(sub: &str) -> Command {
    let mut cmd = Command::new(BINARY_LOCATION.to_str().unwrap());

    cmd.arg(sub);

    cmd
}

fn write_file_string(path: &PathBuf, contents: &String) {
    write_file_bytes(path, contents.as_bytes());
}

fn write_file_bytes(path: &PathBuf, contents: &[u8]) {
    use std::io::Write;

    let Some(parent) = path.parent() else {
        panic!("failed to find parent dir for {:?}", path)
    };

    std::fs::create_dir_all(parent)
        .expect(format!("Failed to create all parent folders for {:?}", path).as_str());

    let output = std::fs::File::create(path)
        .expect(format!("Failed to open file {} for writing", path.to_str().unwrap()).as_str());
    let mut writer = std::io::BufWriter::new(output);
    writer.write_all(contents).unwrap();
}

fn tempdir() -> PathBuf {
    tempfile::tempdir().unwrap().into_path()
}

fn create_tar(files: &HashMap<PathBuf, String>) -> Result<Vec<u8>, std::io::Error> {
    let mut ar = Builder::new(Vec::new());

    for (path, text) in files.into_iter() {
        let mut header = Header::new_gnu();
        let bytes = text.as_bytes();

        header
            .set_path(path.to_str().unwrap())
            .expect("Failed to set path for archive entry");
        header.set_size(bytes.len() as u64);
        header.set_cksum();
        ar.append(&header, bytes).unwrap();
    }

    ar.finish().expect("Failed to close TAR archive");
    ar.into_inner()
}

fn indexing_data() -> HashMap<PathBuf, String> {
    HashMap::from([
        (
            PathBuf::from("src/main/java/globals.java"),
            include_str!("../testdata/globals.java").to_string(),
        ),
        (
            PathBuf::from("package-info.java"),
            include_str!("../testdata/package-info.java").to_string(),
        ),
    ])
}

fn extract_paths(setup: &HashMap<PathBuf, String>) -> HashSet<String> {
    setup
        .clone()
        .into_keys()
        .map(|pb| pb.to_str().unwrap().to_string())
        .collect()
}

fn extract_indexed_paths(index: &scip::types::Index) -> HashSet<String> {
    index
        .documents
        .clone()
        .into_iter()
        .map(|pb| pb.relative_path)
        .collect()
}

fn snapshot_from_files(docs: &Vec<Document>, project_root: &Path) -> String {
    let mut str = String::new();
    let mut docs = docs.clone();
    docs.sort_by_key(|doc| doc.relative_path.clone());

    for doc in docs {
        let path = project_root.join(doc.relative_path.clone());
        let contents = std::fs::read_to_string(path.clone())
            .with_context(|| anyhow!("Failed to read path {:?}", path.clone()))
            .unwrap();

        str.push_str(&format_snapshot_document(&doc, &contents));
    }

    str
}

fn format_snapshot_document(doc: &scip::types::Document, contents: &String) -> String {
    let mut str = String::new();
    str.push_str(format!("//----FILE={}\n", doc.relative_path).as_str());
    str.push_str(&snapshot_syntax_document(&doc, &contents));
    str.push_str("\n\n");

    str
}

fn snapshot_from_data(docs: &Vec<Document>, data: &HashMap<PathBuf, String>) -> String {
    let mut str = String::new();
    let mut docs = docs.clone();
    docs.sort_by_key(|doc| doc.relative_path.clone());

    for doc in docs {
        let contents = data
            .get(&PathBuf::from(&doc.relative_path))
            .context(format!("Failed to find {} in data", &doc.relative_path))
            .unwrap();

        str.push_str(&format_snapshot_document(&doc, contents));
    }

    str
}
