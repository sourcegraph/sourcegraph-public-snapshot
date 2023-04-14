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
