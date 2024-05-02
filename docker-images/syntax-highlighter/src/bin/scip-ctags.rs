use std::io::{BufReader, BufWriter, Write};

use syntax_analysis::ctags::{ctags_runner, Reply};

fn main() {
    // Exits with a code zero if the environment variable SANITY_CHECK equals
    // to "true". This enables testing that the current program is in a runnable
    // state against the platform it's being executed on.
    //
    // See https://github.com/GoogleContainerTools/container-structure-test
    match std::env::var("SANITY_CHECK") {
        Ok(v) if v == "true" => {
            println!("Sanity check passed, exiting without error");
            std::process::exit(0)
        }
        _ => {}
    };

    let mut stdin = BufReader::new(std::io::stdin());
    let mut stdout = BufWriter::new(std::io::stdout());

    if let Err(err) = ctags_runner(&mut stdin, &mut stdout) {
        // If there's a top-level error, treat it as fatal. If an error is recoverable,
        // we would've already handled it and included it in the ctags response.
        Reply::Error {
            message: err.to_string(),
            fatal: true,
        }
        .write(&mut stdout);
        stdout.flush().unwrap()
    }
}
