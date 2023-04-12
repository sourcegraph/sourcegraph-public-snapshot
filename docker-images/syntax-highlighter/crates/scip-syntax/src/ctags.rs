use std::path::PathBuf;

use scip::types::Descriptor;
use serde::{Deserialize, Serialize};

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
struct CTagsEntry {
    /// TODO(SuperAuguste): Handle _type != tag
    name: String,
    path: PathBuf,
    pattern: String,
    language: String,
    line: i32,
    /// TODO(SuperAuguste): Use enum
    kind: String,
    /// TODO(SuperAuguste): Use enum
    scope: Option<String>,
    #[serde(rename = "scopeKind")]
    scope_kind: Option<String>,
    roles: String,
}
