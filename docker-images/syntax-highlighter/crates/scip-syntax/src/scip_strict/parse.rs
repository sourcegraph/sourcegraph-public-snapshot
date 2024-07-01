use std::{borrow::Cow, str};

use nom::{
    branch::alt,
    bytes::complete::{tag, take_while1},
    character::complete::char,
    combinator::{cut, eof, fail, opt, verify},
    error::context,
    multi::many1,
    sequence::{delimited, preceded, tuple},
    Finish, IResult, Parser,
};

use super::{context_error::CtxError, Descriptor, NonLocalSymbol, Package, Scheme, Symbol};

pub(super) fn parse_symbol(input: &str) -> Result<Symbol<'_>, String> {
    match parse_symbol_inner(input).finish() {
        Ok((_, symbol)) => Ok(symbol),
        Err(err) => Err(format!("Invalid symbol: '{input}'\n{err}",)),
    }
}

type PResult<'a, A> = IResult<&'a str, A, CtxError<&'a str>>;

fn parse_symbol_inner(input: &str) -> PResult<'_, Symbol<'_>> {
    let (input, symbol) = alt((parse_local_symbol, parse_nonlocal_symbol))(input)?;
    eof(input)?;
    Ok((input, symbol))
}

fn parse_local_symbol(input: &str) -> PResult<'_, Symbol<'_>> {
    preceded(tag("local "), parse_simple_identifier_str)
        .map(|local_id| Symbol::Local { local_id })
        .parse(input)
}

fn parse_nonlocal_symbol(input: &str) -> PResult<'_, Symbol<'_>> {
    tuple((parse_scheme, parse_package, many1(parse_descriptor)))
        .map(|(scheme, package, descriptors)| {
            Symbol::NonLocal(NonLocalSymbol {
                scheme,
                package,
                descriptors,
            })
        })
        .parse(input)
}

fn parse_scheme(input: &str) -> PResult<'_, Scheme> {
    context(
        "Invalid scheme",
        verify(parse_space_terminated, |s: &Cow<'_, str>| {
            !s.starts_with("local")
        }),
    )
    .map(Scheme)
    .parse(input)
}

fn parse_package(input: &str) -> PResult<'_, Package> {
    tuple((
        context("Invalid package manager", parse_space_terminated),
        context("Invalid package name", parse_space_terminated),
        context("Invalid package version", parse_space_terminated),
    ))
    .map(|(manager, package_name, version)| Package {
        manager,
        package_name,
        version,
    })
    .parse(input)
}

fn parse_descriptor(input: &str) -> PResult<'_, Descriptor> {
    alt((
        parse_parameter_descriptor,
        parse_type_parameter_descriptor,
        parse_named_descriptor,
    ))(input)
}

fn parse_type_parameter_descriptor(input: &str) -> PResult<'_, Descriptor> {
    delimited(char('['), parse_name, char(']'))
        .map(Descriptor::TypeParameter)
        .parse(input)
}

fn parse_parameter_descriptor(input: &str) -> PResult<'_, Descriptor> {
    delimited(char('('), parse_name, char(')'))
        .map(Descriptor::Parameter)
        .parse(input)
}

fn parse_named_descriptor(input: &str) -> PResult<'_, Descriptor> {
    let (input, name) = parse_name(input)?;
    match input.chars().next() {
        Some('/') => Ok((&input[1..], Descriptor::Namespace(name))),
        Some('#') => Ok((&input[1..], Descriptor::Type(name))),
        Some('.') => Ok((&input[1..], Descriptor::Term(name))),
        Some(':') => Ok((&input[1..], Descriptor::Meta(name))),
        Some('!') => Ok((&input[1..], Descriptor::Macro(name))),
        Some('(') => {
            let (input, disambiguator) = opt(parse_simple_identifier_str)(&input[1..])?;
            let (input, _) = tag(").")(input)?;
            Ok((
                input,
                Descriptor::Method {
                    name,
                    disambiguator,
                },
            ))
        }
        _ => context("Missing descriptor suffix", cut(fail))(input),
    }
}

fn parse_name(input: &str) -> PResult<'_, Cow<'_, str>> {
    alt((parse_escaped_identifier, parse_simple_identifier))(input)
}

pub fn is_simple_identifier_char(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_' || c == '+' || c == '-' || c == '$'
}

fn parse_simple_identifier_str(input: &str) -> PResult<'_, &str> {
    take_while1(is_simple_identifier_char)(input)
}

fn parse_simple_identifier(input: &str) -> PResult<'_, Cow<'_, str>> {
    parse_simple_identifier_str.map(Cow::Borrowed).parse(input)
}

fn parse_escaped_identifier(input: &str) -> PResult<'_, Cow<'_, str>> {
    let (input, _) = char('`')(input)?;
    let (input, name) = parse_terminated(input, b'`')?;
    let (input, _) = char('`')(input)?;
    Ok((input, name))
}

fn parse_space_terminated(input: &str) -> PResult<'_, Cow<'_, str>> {
    let (input, terminated) = parse_terminated(input, b' ')?;
    let (input, _) = char(' ')(input)?;
    Ok((input, terminated))
}

fn parse_terminated(input: &str, terminator: u8) -> PResult<'_, Cow<'_, str>> {
    let mut needs_escape = false;
    let mut current = input;
    while let Some(offset) = current.find(terminator as char) {
        let (_, rest) = current.split_at(offset + 1);
        if rest.starts_with(terminator as char) {
            needs_escape = true;
            current = &rest[1..];
        } else {
            let (raw, rest) = input.split_at(input.len() - rest.len() - 1);
            let escaped = if needs_escape {
                Cow::Owned(raw.replace(
                    str::from_utf8(&[terminator, terminator]).unwrap(),
                    str::from_utf8(&[terminator]).unwrap(),
                ))
            } else {
                Cow::Borrowed(raw)
            };
            return Ok((rest, escaped));
        }
    }
    context("Missing terminator", cut(fail))(current)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parsing_symbols() {
        assert_eq!(
            Symbol::parse("scip-java . . . Dude#lol!waow.")
                .unwrap()
                .to_string(),
            "scip-java . . . Dude#lol!waow."
        );
        assert_eq!(
            Symbol::parse("scip  java . . . Dude#lol!waow.")
                .unwrap()
                .to_string(),
            "scip  java . . . Dude#lol!waow."
        );
        assert_eq!(
            Symbol::parse("scip  java . . . `Dude```#`lol`!waow.")
                .unwrap()
                .to_string(),
            "scip  java . . . `Dude```#lol!waow."
        );
        assert_eq!(Symbol::parse("local 1").unwrap().to_string(), "local 1");
        assert_eq!(
            Symbol::parse("rust-analyzer cargo test_rust_dependency 0.1.0 println!")
                .unwrap()
                .to_string(),
            "rust-analyzer cargo test_rust_dependency 0.1.0 println!"
        );
        assert_eq!(
            Symbol::NonLocal(NonLocalSymbol {
                scheme: Default::default(),
                package: Default::default(),
                descriptors: vec![Descriptor::Type("hi".into())]
            })
            .to_string(),
            " . . . hi#"
        );
    }
}
