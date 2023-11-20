use std::path::PathBuf;

use proc_macro::TokenStream;
use quote::quote;
use syn::{
    parse::{Parse, ParseStream},
    parse_macro_input,
    token::Comma,
    Ident, LitStr, Result, Token,
};

struct ScipQuery {
    pub lang: String,
    pub query: String,
}
impl Parse for ScipQuery {
    fn parse(input: ParseStream) -> Result<Self> {
        let lang: String = match input.parse::<Ident>() {
            Ok(lang) => lang.to_string().to_lowercase(),
            Err(_) => {
                let lang: LitStr = input.parse()?;
                lang.value()
            }
        };

        input.parse::<Token![,]>()?;
        let query: LitStr = input.parse()?;

        Ok(Self {
            lang,
            query: query.value(),
        })
    }
}

/// Use to get a particular query from the scip-semantic repo.
///     Will do this at compile time and directly include
///
/// Example:
/// > include_scip_query!("rust", "scip-tags");
#[proc_macro]
pub fn include_scip_query(input: TokenStream) -> TokenStream {
    let ScipQuery { lang, query } = parse_macro_input!(input as ScipQuery);
    // let base = std::env::var("CARGO_MANIFEST_DIR").unwrap() + "/";

    quote! { include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/queries/", #lang, "/", #query, ".scm")) }.into()
}

struct IncludeOptional {
    literals: Vec<String>,
}

impl Parse for IncludeOptional {
    fn parse(input: ParseStream) -> syn::Result<Self> {
        let mut literals = Vec::new();

        // parse while we still have inputs.
        loop {
            if input.peek(LitStr) {
                let lit_str: LitStr = input.parse()?;
                literals.push(lit_str.value());
            } else if input.peek(Ident) {
                let ident: Ident = input.parse()?;
                literals.push(ident.to_string());
            } else {
                panic!("Should make error here but it's OK for now");
            }

            if input.is_empty() {
                break;
            }

            let _: Comma = input.parse()?;
            if input.is_empty() {
                break;
            }
        }

        Ok(Self { literals })
    }
}

#[allow(unused)]
#[proc_macro]
pub fn include_project_file_optional(input: TokenStream) -> TokenStream {
    if true {
        todo!("don't use this yet, i have to come back to this");
    }

    let literals = parse_macro_input!(input as IncludeOptional).literals;

    // project files are always relative to the Cargo.toml of the compiling project.
    // let base = std::env::var("CARGO_MANIFEST_DIR").unwrap() + "/";
    panic!("{:?}", std::env::var("CARGO_MANIFEST_DIR"));
    let base = "/".to_string();
    let filepath: PathBuf = literals.iter().fold(base, |acc, lit| acc + lit).into();

    if filepath.exists() {
        let filepath = filepath
            .to_str()
            .expect("Filepath must be expandable at this point");

        quote! { include_str!(concat!(env!("CARGO_MANIFEST_DIR"), #filepath)) }.into()
    } else {
        quote! { "" }.into()
    }
}
