use std::{path::Path, time::Instant};

use scip_syntax::locals::parse_tree;
use walkdir::WalkDir;

struct ParseTiming {
    pub filepath: String,
    pub duration: std::time::Duration,
}

fn parse_files(dir: &Path) -> Vec<ParseTiming> {
    let mut config = scip_syntax::languages::go_locals();
    let extension = "go";

    let mut timings = vec![];

    for entry in WalkDir::new(dir) {
        let entry = entry.unwrap();
        let entry = entry.path();

        match entry.extension() {
            Some(ext) if extension == ext => {}
            _ => continue,
        }

        let start = Instant::now();

        let source = std::fs::read_to_string(entry).unwrap();
        let source_bytes = source.as_bytes();
        let tree = config.parser.parse(source_bytes, None).unwrap();
        parse_tree(&mut config, &tree, source_bytes).unwrap();

        let finish = Instant::now();

        timings.push(ParseTiming {
            filepath: entry.file_stem().unwrap().to_string_lossy().to_string(),
            duration: finish - start,
        });
    }

    timings
}

fn measure_parsing() {
    println!("Measuring parsing");
    let start = Instant::now();

    let root = Path::new(
        // "/home/tjdevries/sourcegraph/sourcegraph.git/main/",
        "/home/tjdevries/sourcegraph/sourcegraph.git/main/internal/database/mocks_temp.go",
        // "/home/tjdevries/sourcegraph/scip-semantic/testdata/locals-nested.go",
        // "/home/tjdevries/sourcegraph/scip-semantic/testdata/funcs.go",
        // "/home/tjdevries/sourcegraph/scip-semantic/testdata/multi-scopes.go",
    );
    let mut timings = parse_files(root);
    timings.sort_by(|a, b| a.duration.cmp(&b.duration));
    println!("Slowest files:");
    for timing in timings.iter().rev().take(10) {
        println!("{}: {:?}", timing.filepath, timing.duration);
    }

    let finish = Instant::now();

    println!("Done {:?}", finish - start);
}

fn main() {
    // TODO: parameterize
    let measure = "parsing";

    match measure {
        "parsing" => measure_parsing(),
        _ => panic!("Unknown measure: {}", measure),
    }
}
