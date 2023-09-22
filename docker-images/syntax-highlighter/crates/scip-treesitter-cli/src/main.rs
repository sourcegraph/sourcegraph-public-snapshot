use std::{collections::HashMap, env, path::Path};

use scip::types::Occurrence;
use scip_syntax::{get_locals, get_symbols};
use scip_treesitter::types::PackedRange;
use scip_treesitter_languages::parsers::BundledParser;

pub fn write_message_to_file<P>(
    path: P,
    msg: impl protobuf::Message,
) -> Result<(), Box<dyn std::error::Error>>
where
    P: AsRef<std::path::Path>,
{
    use std::io::Write;

    let res = msg.write_to_bytes()?;
    let output = std::fs::File::create(path)?;
    let mut writer = std::io::BufWriter::new(output);
    writer.write_all(&res)?;

    Ok(())
}

use clap::Parser;

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Name of the person to greet
    #[arg(short, long)]
    language: String,

    #[arg(short, long)]
    out: Option<String>,
    filenames: Vec<String>,
}

pub fn main() {
    let args = Args::parse();

    let language = &args.language;
    let filenames = &args.filenames;
    let p = BundledParser::get_parser(language).unwrap();

    let directory = Path::new("./");
    let mut index = scip::types::Index {
        metadata: Some(scip::types::Metadata {
            tool_info: Some(scip::types::ToolInfo {
                name: "scip-local-nav".to_string(),
                version: "0.0.1".to_string(),
                arguments: vec![],
                ..Default::default()
            })
            .into(),
            project_root: "file://".to_string() + directory.to_str().unwrap(),
            ..Default::default()
        })
        .into(),
        ..Default::default()
    };

    for filename in filenames {
        let contents = std::fs::read(filename).unwrap();
        let string = String::from_utf8(contents.clone()).unwrap();
        let lines: HashMap<i32, &str> = (0..).into_iter().zip(string.split("\n")).collect();
        let mut document = get_symbols(&p, &contents).unwrap();
        let locals = get_locals(&p, &contents);

        match locals {
            Some(Ok(occurrences)) => {
                for occ in occurrences {

                    println!("{:?}", extract(&lines, &occ));
                    document.occurrences.push(occ);
                    // let range = PackedRange::from_vec(&occ.range).unwrap();

                    // let line = lines.get(&range.start_line).unwrap();
                    // let label = &line[range.start_col as usize..range.end_col as usize];

                    // println!("{line}: {occ} {label}");
                    // document.occurrences.push(occ);
                }
            }
            _ => {}
        }

        // println!("{:?}", locals);
        index.documents.push(document);
    }

    let out_name = args.out.unwrap_or("index.scip".to_string());

    write_message_to_file(directory.join(out_name), index).expect("to write the file");
}

fn extract(lines: &HashMap<i32, &str>, occ: &Occurrence) -> Option<String> {
    let range = PackedRange::from_vec(&occ.range)?;

    let line = lines.get(&range.start_line)?;
    let label = &line[range.start_col as usize..range.end_col as usize];

    return Some(label.to_string());
}
