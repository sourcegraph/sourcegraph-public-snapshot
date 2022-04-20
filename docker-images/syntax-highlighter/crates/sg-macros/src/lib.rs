use std::path::PathBuf;

use proc_macro::TokenStream;
use quote::quote;
use syn::{
    parse::{Parse, ParseStream},
    parse_macro_input,
    token::Comma,
    Ident, LitStr,
};

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

#[proc_macro]
pub fn include_project_file_optional(input: TokenStream) -> TokenStream {
    let literals = parse_macro_input!(input as IncludeOptional).literals;

    // project files are always relative to the Cargo.toml of the compiling project.
    let base = std::env::var("CARGO_MANIFEST_DIR").unwrap() + "/";
    let filepath: PathBuf = literals.iter().fold(base, |acc, lit| acc + lit).into();

    if filepath.exists() {
        let filepath = filepath
            .to_str()
            .expect("Filepath must be expandable at this point");

        quote! { include_str!(#filepath) }.into()
    } else {
        quote! { "" }.into()
    }
}
