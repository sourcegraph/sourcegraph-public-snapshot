use std::{path::Path, time::Instant};

use clap::Parser;
use scip_syntax::locals::parse_tree;
use scip_treesitter_languages::parsers::BundledParser;
use walkdir::WalkDir;

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Arguments {
    /// Root directory to run local navigation over
    root_dir: String,
}

struct ParseTiming {
    pub filepath: String,
    pub duration: std::time::Duration,
}

fn parse_files(dir: &Path) -> Vec<ParseTiming> {
    let config = scip_syntax::languages::get_local_configuration(BundledParser::Go).unwrap();
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
        let mut parser = config.get_parser();
        let tree = parser.parse(source_bytes, None).unwrap();
        parse_tree(config, &tree, source_bytes).unwrap();

        let finish = Instant::now();

        timings.push(ParseTiming {
            filepath: entry.file_stem().unwrap().to_string_lossy().to_string(),
            duration: finish - start,
        });
    }

    timings
}

fn measure_parsing() {
    let args = Arguments::parse();
    println!("Measuring parsing");
    let start = Instant::now();

    let root = Path::new(&args.root_dir);

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
