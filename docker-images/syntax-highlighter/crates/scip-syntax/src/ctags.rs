use std::{io::BufWriter, io::Write, ops::Not, path};

use protobuf::EnumOrUnknown;
use scip::types::{descriptor::Suffix, Descriptor};
use scip_treesitter_languages::parsers::BundledParser;
use serde::{Deserialize, Serialize};

use crate::{get_globals, globals::Scope};

#[derive(Debug)]
pub enum TagKind {
    Function,
    Class,
}

#[derive(Debug)]
pub struct TagEntry {
    pub descriptors: Vec<Descriptor>,
    pub kind: TagKind,
    pub parent: Option<Box<TagEntry>>,

    pub line: usize,
    // pub column: usize,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(tag = "command")]
pub enum Request {
    #[serde(rename = "generate-tags")]
    GenerateTags {
        // command == generate-tags
        filename: String,
        size: usize,
    },
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(tag = "_type")]
pub enum Reply {
    #[serde(rename = "program")]
    Program { name: String, version: String },
    #[serde(rename = "completed")]
    Completed { command: String },
    #[serde(rename = "error")]
    Error { message: String, fatal: bool },
    #[serde(rename = "tag")]
    Tag {
        name: String,
        path: String,
        language: String,
        /// Starts at 1
        line: usize,
        kind: String,
        pattern: String,
        scope: Option<String>,
        #[serde(rename = "scopeKind")]
        scope_kind: Option<String>,
        signature: Option<String>,
        // TODO(SuperAuguste): Any other properties required? Roles? Access?
    },
}

pub fn write_to_buf_writer<T: serde::ser::Serialize, W: std::io::Write>(
    buf_writer: &mut BufWriter<W>,
    val: &T,
) {
    buf_writer
        .write_all(serde_json::to_string(val).unwrap().as_bytes())
        .unwrap();
    buf_writer.write_all("\n".as_bytes()).unwrap();
}

pub fn generate_tags<W: std::io::Write>(
    buf_writer: &mut BufWriter<W>,
    filename: String,
    file_data: &[u8],
) {
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
        vec![],
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

fn emit_tags_for_scope<W: std::io::Write>(
    buf_writer: &mut BufWriter<W>,
    path: &str,
    parent_scopes: Vec<String>,
    scope: &Scope,
) {
    let curr_scopes = {
        let mut curr_scopes = parent_scopes.clone();
        for desc in &scope.descriptors {
            curr_scopes.push(desc.name.clone());
        }
        curr_scopes
    };

    if scope.descriptors.len() > 0 {
        write_to_buf_writer(
            buf_writer,
            &Reply::Tag {
                name: scope
                    .descriptors
                    .iter()
                    .map(|d| d.name.clone())
                    .collect::<Vec<String>>()
                    .join("."),
                path: path.to_string(),
                // TODO(SuperAuguste): Set to correct language (does this even matter?)
                language: "Go".to_string(),
                line: scope.range[0] as usize + 1,
                kind: suffix_to_string(scope.descriptors.last().unwrap().suffix),
                pattern: "/.*/".to_string(),
                scope: parent_scopes
                    .is_empty()
                    .not()
                    .then(|| parent_scopes.join(".")),
                scope_kind: Option::None,
                signature: Option::None,
            },
        );
    }

    for subscope in &scope.children {
        emit_tags_for_scope(buf_writer, path, curr_scopes.clone(), &subscope);
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

        write_to_buf_writer(
            buf_writer,
            &Reply::Tag {
                name: global.descriptors.last().unwrap().name.clone(),
                path: path.to_string(),
                // TODO(SuperAuguste): Set to correct language (does this even matter?)
                language: "Go".to_string(),
                line: global.range[0] as usize + 1,
                kind: suffix_to_string(global.descriptors.last().unwrap().suffix),
                pattern: "/.*/".to_string(),
                scope: scope_name.is_empty().not().then(|| scope_name.join(".")),
                scope_kind: Option::None,
                signature: Option::None,
            },
        );
    }
}
