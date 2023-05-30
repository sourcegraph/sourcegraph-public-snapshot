use std::io::{BufReader, BufWriter};

use scip_syntax::ctags::ctags_runner;

fn main() {
    let mut stdin = BufReader::new(std::io::stdin());
    let mut stdout = BufWriter::new(std::io::stdout());

    if let Err(err) = ctags_runner(&mut stdin, &mut stdout) {
        eprintln!("Error while executing: {}", err);
    }
}
