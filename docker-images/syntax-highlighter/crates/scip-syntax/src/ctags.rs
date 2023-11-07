use std::{
    collections::HashMap,
    io::{BufRead, BufReader, BufWriter, Read, Write},
    ops::Not,
    path,
};

use anyhow::{Context, Result};
use itertools::intersperse;
use scip::types::{descriptor::Suffix, symbol_information, Descriptor};
use scip_treesitter_languages::parsers::BundledParser;
use serde::{Deserialize, Serialize};

use crate::{get_globals, globals::Scope};

#[derive(Serialize, Deserialize, Debug)]
#[serde(tag = "command", rename_all = "kebab-case")]
pub enum Request {
    GenerateTags { filename: String, size: usize },
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(tag = "_type", rename_all = "kebab-case")]
pub enum Reply<'a> {
    Program {
        name: String,
        version: String,
    },
    Completed {
        command: String,
    },
    Error {
        message: String,
        fatal: bool,
    },
    Tag {
        name: String,
        path: &'a str,
        language: &'a str,
        /// Starts at 1
        line: usize,
        kind: &'a str,
        scope: Option<&'a str>,
        // Can't find any uses of these. If someone reports a bug, we can support this
        // scope_kind: Option<String>,
        // signature: Option<String>,
    },
}

impl<'a> Reply<'a> {
    pub fn write<W: std::io::Write>(self, writer: &mut W) {
        writer
            .write_all(serde_json::to_string(&self).unwrap().as_bytes())
            .unwrap();
        writer.write_all("\n".as_bytes()).unwrap();
    }

    pub fn write_tag<W: std::io::Write>(
        writer: &mut W,
        scope: &Scope,
        path: &'a str,
        language: &'a str,
        tag_scope: Option<&'a str>,
        scope_deduplicator: &mut HashMap<String, ()>,
    ) {
        let descriptors = &scope.descriptors;
        let names = descriptors.iter().map(|d| d.name.as_str());
        let name = intersperse(names, ".").collect::<String>();

        let mut dedup = match tag_scope {
            Some(ts) => vec![ts],
            None => vec![],
        };
        dedup.push(&name);
        let dedup = dedup.join(".");
        if scope_deduplicator.contains_key(&dedup) {
            return;
        }
        scope_deduplicator.insert(dedup, ());

        let tag = Self::Tag {
            name,
            path,
            language,
            line: scope.scope_range.start_line as usize + 1,
            kind: descriptors_to_kind(&scope.descriptors, &scope.kind),
            scope: tag_scope,
        };

        tag.write(writer);
    }
}

fn descriptors_to_kind(
    descriptors: &[Descriptor],
    symbol_kind: &symbol_information::Kind,
) -> &'static str {
    // Override using kind when we have more information
    if let Some(kind) = crate::ts_scip::symbol_kind_to_ctags_kind(symbol_kind) {
        return kind;
    }

    match descriptors
        .last()
        .unwrap_or_default()
        .suffix
        .enum_value_or_default()
    {
        Suffix::Namespace => "namespace",
        Suffix::Package => "package",
        Suffix::Method => "method",
        Suffix::Type => "type",
        _ => "variable",
    }
}

fn emit_tags_for_scope<W: std::io::Write>(
    buf_writer: &mut BufWriter<W>,
    path: &str,
    parent_scopes: Vec<String>,
    scope: &Scope,
    language: &str,
    scope_deduplicator: &mut HashMap<String, ()>,
) {
    let curr_scopes = {
        let mut curr_scopes = parent_scopes.clone();
        for desc in &scope.descriptors {
            curr_scopes.push(desc.name.clone());
        }
        curr_scopes
    };

    if !scope.descriptors.is_empty() {
        let tag_scope = parent_scopes
            .is_empty()
            .not()
            .then(|| parent_scopes.join("."));
        let tag_scope = tag_scope.as_deref();

        Reply::write_tag(
            &mut *buf_writer,
            scope,
            path,
            language,
            tag_scope,
            scope_deduplicator,
        );
    }

    for subscope in &scope.children {
        emit_tags_for_scope(
            buf_writer,
            path,
            curr_scopes.clone(),
            subscope,
            language,
            scope_deduplicator,
        );
    }

    for global in &scope.globals {
        let mut scope_name = curr_scopes.clone();
        scope_name.extend(
            global
                .descriptors
                .iter()
                .take(global.descriptors.len() - 1)
                .map(|d| d.name.clone()),
        );

        Reply::Tag {
            name: global.descriptors.last().unwrap().name.clone(),
            path,
            language,
            line: global.range.start_line as usize + 1,
            kind: descriptors_to_kind(&global.descriptors, &global.kind),
            scope: scope_name
                .is_empty()
                .not()
                .then(|| scope_name.join("."))
                .as_deref(),
        }
        .write(buf_writer);
    }
}

pub fn generate_tags<W: std::io::Write>(
    buf_writer: &mut BufWriter<W>,
    filename: String,
    file_data: &[u8],
) -> Option<()> {
    let path = path::Path::new(&filename);
    let extension = path.extension()?.to_str()?;
    let filepath = path.file_name()?.to_str()?;

    let parser = BundledParser::get_parser_from_extension(extension)?;
    let (root_scope, _) = match get_globals(parser, file_data)? {
        Ok(vals) => vals,
        Err(err) => {
            // TODO: Not sure I want to keep this or not
            #[cfg(debug_assertions)]
            if true {
                panic!("Could not parse file: {}", err);
            }

            let _ = err;

            return None;
        }
    };

    let mut scope_deduplicator = HashMap::new();
    emit_tags_for_scope(
        buf_writer,
        filepath,
        vec![],
        &root_scope,
        // I don't believe the language name is actually used anywhere but we'll
        // keep it to be compliant with the ctags spec
        parser.get_language_name(),
        &mut scope_deduplicator,
    );
    Some(())
}

pub fn ctags_runner<R: Read, W: Write>(
    input: &mut BufReader<R>,
    output: &mut std::io::BufWriter<W>,
) -> Result<()> {
    Reply::Program {
        name: "SCIP Ctags".to_string(),
        version: "5.9.0".to_string(),
    }
    .write(output);
    output.flush().unwrap();

    loop {
        let mut line = String::new();
        input.read_line(&mut line)?;

        if line.is_empty() {
            break;
        }

        let request = serde_json::from_str::<Request>(&line);
        let request = match request {
            Ok(request) => request,
            Err(_) => {
                eprintln!("Could not parse request: {}", line);
                continue;
            }
        };

        match request {
            Request::GenerateTags { filename, size } => {
                let mut file_data = vec![0; size];
                input
                    .read_exact(&mut file_data)
                    .expect("Could not fill file data exactly");

                generate_tags(output, filename, &file_data);
            }
        }

        Reply::Completed {
            command: "generate-tags".to_string(),
        }
        .write(output);

        output.flush().unwrap();
    }

    Ok(())
}

pub fn helper_execute_one_file(name: &str, contents: &str) -> Result<String> {
    let command = format!(
        r#"
{{ "command":"generate-tags","filename":"{}","size":{} }}
{}
"#,
        name,
        contents.len(),
        contents
    )
    .trim()
    .to_string();

    let mut input = BufReader::new(command.as_bytes());
    let mut output = BufWriter::new(Vec::new());
    ctags_runner(&mut input, &mut output)?;

    String::from_utf8(output.get_ref().to_vec()).context("Could not parse output")
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_ctags_runner_basic() -> Result<()> {
        let file = r#"
fn main() {
    println!("Hello, world!");
}

fn something() -> bool { true }
fn other() -> bool { false }
"#
        .trim();

        let output = helper_execute_one_file("main.rs", file)?;
        insta::assert_snapshot!(output);

        Ok(())
    }
}
