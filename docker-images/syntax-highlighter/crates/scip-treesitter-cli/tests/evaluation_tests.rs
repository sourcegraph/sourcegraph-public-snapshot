use std::{env::temp_dir, path::PathBuf};

use scip_treesitter_cli::{
    evaluate::Evaluator,
    index::{index_command, AnalysisMode, IndexMode, IndexOptions},
};

lazy_static::lazy_static! {
    static ref BASE: PathBuf = {
        match std::env::var("CARGO_MANIFEST_DIR") {
            Ok(va) => PathBuf::from(va),
            _ => todo!("This needs to be fixed to work with Bazel")
        }
    };
}

#[test]
fn java_e2e_indexing() {
    let dir = BASE.join("testdata/java");

    let out_dir = temp_dir();

    let candidate = out_dir.join("index-tree-sitter.scip");
    let ground_truth = dir.join("index.scip");

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
    );

    let mut str = vec![];

    Evaluator::default()
        .evaluate_files(candidate, ground_truth)
        .unwrap()
        .write_summary(
            &mut str,
            scip_treesitter_cli::evaluate::EvaluationOutputOptions {
                print_false_negatives: true,
                print_true_positives: true,
                print_false_positives: true,
                print_mapping: true,
                disable_colors: true,
                ..Default::default()
            },
        )
        .unwrap();

    insta::assert_snapshot!("java_evaluation", String::from_utf8(str).unwrap());
}
