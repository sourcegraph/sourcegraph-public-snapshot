// use anyhow::Result;

// Some existing symbol structs
//
// zoekt
// type Symbol struct {
//  Sym        string
//  Kind       string
//  Parent     string
//  ParentKind string
// }
//
// go-ctags
// type Entry struct {
//  Name       string
//  Path       string
//  Line       int
//  Kind       string
//  Language   string
//  Parent     string
//  ParentKind string
//  Pattern    string
//  Signature  string
//
//  FileLimited bool
// }

// LSP Symbol Kinds, could be fine for us to use now
// export namespace SymbolKind {
//  export const File = 1;
//  export const Module = 2;
//  export const Namespace = 3;
//  export const Package = 4;
//  export const Class = 5;
//  export const Method = 6;
//  export const Property = 7;
//  export const Field = 8;
//  export const Constructor = 9;
//  export const Enum = 10;
//  export const Interface = 11;
//  export const Function = 12;
//  export const Variable = 13;
//  export const Constant = 14;
//  export const String = 15;
//  export const Number = 16;
//  export const Boolean = 17;
//  export const Array = 18;
//  export const Object = 19;
//  export const Key = 20;
//  export const Null = 21;
//  export const EnumMember = 22;
//  export const Struct = 23;
//  export const Event = 24;
//  export const Operator = 25;
//  export const TypeParameter = 26;
// }

use scip::types::{Descriptor, Document};

#[derive(Debug)]
pub enum TagKind {
    Function,
    Class,
}

impl std::str::FromStr for TagKind {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "definition.function" => Ok(Self::Function),
            "definition.type" => Ok(Self::Function),
            _ => anyhow::bail!("unknown tag kind: {}", s),
        }
    }
}

#[derive(Debug)]
pub struct TagEntry {
    pub descriptors: Vec<Descriptor>,
    pub kind: TagKind,
    pub parent: Option<Box<TagEntry>>,

    pub line: usize,
    // pub column: usize,
}

impl TagEntry {
    pub fn from_document(document: Document) -> Vec<TagEntry> {
        todo!("{:?}", document)
    }
}
