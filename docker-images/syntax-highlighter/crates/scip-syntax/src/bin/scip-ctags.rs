use std::io::Read;
use std::{io, path};

use scip_syntax::ctags::{Reply, Request};
use scip_syntax::get_globals;
use scip_treesitter_languages::parsers::BundledParser;

fn main() {
    println!(
        "{}\n",
        serde_json::to_string(&Reply::Program {
            name: "SCIP Ctags".to_string(),
            version: "5.9.0".to_string(),
        })
        .unwrap()
    );

    loop {
        let mut line = String::new();
        std::io::stdin()
            .read_line(&mut line)
            .expect("Could not read line");

        if line.len() == 0 {
            break;
        }

        let request = serde_json::from_str::<Request>(&line).unwrap();
        match request {
            Request::GenerateTags { filename, size } => {
                let mut file_data = vec![0; size];
                io::stdin()
                    .read_exact(&mut file_data)
                    .expect("Could not read file data");

                let path = path::Path::new(&filename);

                match path.extension().unwrap_or_default().to_str().unwrap() {
                    "go" => {
                        let maybe_globals = get_globals(BundledParser::Go, &file_data);
                        match maybe_globals {
                            Some(result) => match result {
                                Ok(occurrences) => {
                                    for occurrence in occurrences {
                                        println!(
                                            "{}\n",
                                            serde_json::to_string(&Reply::Tag {
                                                name: occurrence.symbol,
                                                path: path
                                                    .file_name()
                                                    .unwrap()
                                                    .to_string_lossy()
                                                    .to_string(),
                                                language: "Go".to_string(),
                                                line: occurrence.range[0] as usize + 1,
                                                kind: "variable".to_string(),
                                                pattern: "/.*/".to_string(),
                                                scope: Option::None,
                                                scope_kind: Option::None,
                                                signature: Option::None,
                                            })
                                            .unwrap()
                                        );
                                    }
                                }
                                Err(_) => {}
                            },
                            None => {}
                        }
                    }
                    _ => {}
                }

                println!(
                    "{}\n",
                    serde_json::to_string(&Reply::Completed {
                        command: "generate-tags".to_string(),
                    })
                    .unwrap()
                );
            }
        }
    }
}
