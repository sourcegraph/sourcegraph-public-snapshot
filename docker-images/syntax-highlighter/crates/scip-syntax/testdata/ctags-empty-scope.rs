use std::{path::Path, time::Instant};

use clap::Parser;
use scip_syntax::locals::parse_tree;
use scip_treesitter_languages::parsers::BundledParser;
use walkdir::WalkDir;

static LANGUAGE: &str = "Rust";
const THRESHOLD: i32 = 10;

trait Yay {}

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Arguments {
    /// Root directory to run local navigation over
    root_dir: String,
}

impl Arguments {
    fn parse() {}
}

impl Yay for Arguments {
    fn pog() {}
}

struct ParseTiming {
    pub filepath: String,
    pub duration: std::time::Duration,
}

fn parse_files(dir: &Path) -> Vec<ParseTiming> {
    // TODO
}

fn measure_parsing() {
    // TODO
}

fn main() {
    // TODO
}
