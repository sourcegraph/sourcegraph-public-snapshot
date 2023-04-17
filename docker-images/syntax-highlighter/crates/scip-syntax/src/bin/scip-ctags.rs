use std::io::{stdout, BufWriter, Read, Stdout, Write};
use std::{io, path};

use protobuf::EnumOrUnknown;
use scip::types::descriptor::Suffix;
use scip_syntax::ctags::{Reply, Request};
use scip_syntax::get_globals;
use scip_syntax::globals::Scope;
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

        let mut buf_writer = BufWriter::new(stdout());

        let request = serde_json::from_str::<Request>(&line).unwrap();
        match request {
            Request::GenerateTags { filename, size } => {
                generate_tags(&mut buf_writer, filename, size)
            }
        }

        write_to_buf_writer(
            &mut buf_writer,
            &Reply::Completed {
                command: "generate-tags".to_string(),
            },
        );

        buf_writer.flush().unwrap();
    }
}

fn write_to_buf_writer<T: serde::ser::Serialize>(buf_writer: &mut BufWriter<Stdout>, val: &T) {
    buf_writer
        .write_all(serde_json::to_string(val).unwrap().as_bytes())
        .unwrap();
    buf_writer.write_all("\n".as_bytes()).unwrap();
}

fn generate_tags(buf_writer: &mut BufWriter<Stdout>, filename: String, size: usize) {
    let mut file_data = vec![0; size];
    io::stdin()
        .read_exact(&mut file_data)
        .expect("Could not read file data");

    let path = path::Path::new(&filename);

    let parser = match BundledParser::get_parser_from_extension(
        path.extension().unwrap_or_default().to_str().unwrap(),
    ) {
        None => return,
        Some(parser) => parser,
    };

    let (root_scope, _) = match match get_globals(parser, &file_data) {
        None => return,
        Some(res) => res,
    } {
        Err(_) => return,
        Ok(vals) => vals,
    };

    emit_tags_for_scope(
        buf_writer,
        path.file_name().unwrap().to_str().unwrap(),
        None,
        &root_scope,
    );
}

fn suffix_to_string(suffix: EnumOrUnknown<Suffix>) -> String {
    return match suffix.enum_value_or_default() {
        // TODO(SuperAuguste): handle more cases + we lose info here, how do handle this?
        Suffix::Namespace => "namespace",
        Suffix::Package => "package",
        Suffix::Method => "method",
        Suffix::Type => "type",
        _ => "variable",
    }
    .to_string();
}

fn emit_tags_for_scope(
    buf_writer: &mut BufWriter<Stdout>,
    path: &str,
    parent_scope_name: Option<String>,
    scope: &Scope,
) {
    let curr_scope_name = {
        let mut curr_scope_name = parent_scope_name.clone().unwrap_or("".to_string());
        for desc in &scope.descriptors {
            if curr_scope_name.len() != 0 {
                curr_scope_name.push('.')
            }
            curr_scope_name.push_str(desc.name.as_str());
        }

        if curr_scope_name.len() == 0 {
            None
        } else {
            Some(curr_scope_name)
        }
    };

    for descriptor in &scope.descriptors {
        write_to_buf_writer(
            buf_writer,
            &Reply::Tag {
                name: descriptor.name.clone(),
                path: path.to_string(),
                // TODO(SuperAuguste): Set to correct language (does this even matter?)
                language: "Go".to_string(),
                line: scope.range[0] as usize + 1,
                kind: suffix_to_string(descriptor.suffix),
                pattern: "/.*/".to_string(),
                scope: parent_scope_name.clone(),
                scope_kind: Option::None,
                signature: Option::None,
            },
        );
    }

    for subscope in &scope.children {
        emit_tags_for_scope(buf_writer, path, curr_scope_name.clone(), &subscope);
    }

    for global in &scope.globals {
        for descriptor in &global.descriptors {
            write_to_buf_writer(
                buf_writer,
                &Reply::Tag {
                    name: descriptor.name.clone(),
                    path: path.to_string(),
                    // TODO(SuperAuguste): Set to correct language (does this even matter?)
                    language: "Go".to_string(),
                    line: global.range[0] as usize + 1,
                    kind: suffix_to_string(descriptor.suffix),
                    pattern: "/.*/".to_string(),
                    scope: curr_scope_name.clone(),
                    scope_kind: Option::None,
                    signature: Option::None,
                },
            );
        }
    }
}
