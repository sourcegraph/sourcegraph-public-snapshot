use std::{env, path::Path};

use scip_syntax::get_symbols;
use scip_treesitter_languages::parsers::BundledParser;
use sg_syntax::treesitter_index;

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
        let contents = std::fs::read(filename).unwrap().to_vec();
        let result = get_symbols(&p, &contents).unwrap();
        index.documents.push(result);
    }

    let out_name = args.out.unwrap_or("index.scip".to_string());

    write_message_to_file(directory.join(out_name), index).expect("to write the file");
}
