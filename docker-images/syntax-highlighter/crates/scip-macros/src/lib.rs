use proc_macro::TokenStream;
use quote::quote;
use syn::{
    parse::{Parse, ParseStream},
    parse_macro_input, Ident, LitStr, Result, Token,
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
