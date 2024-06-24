use super::{Descriptor, NonLocalSymbol, Package, Scheme, Symbol};
use nom::branch::alt;
use nom::bytes::complete::{tag, take_while1};
use nom::character::complete::char;
use nom::combinator::{cut, eof, fail};
use nom::error::{context, convert_error, VerboseError};
use nom::multi::many1;
use nom::sequence::{delimited, preceded, tuple};
use nom::{Finish, IResult, Parser};
use std::borrow::Cow;

pub(super) fn parse_symbol(input: &str) -> Result<Symbol<'_>, String> {
    match parse_symbol_inner(input).finish() {
        Ok((_, symbol)) => Ok(symbol),
        Err(err) => Err(format!(
            "Invalid symbol: '{input}'\n{}",
            convert_error(input, err)
        )),
    }
}

type PResult<'a, A> = IResult<&'a str, A, VerboseError<&'a str>>;

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
    tuple((
        parse_space_terminated,
        parse_package,
        many1(parse_descriptor),
    ))
    .map(|(scheme, package, descriptors)| Symbol::NonLocal (NonLocalSymbol {
        scheme: Scheme(scheme),
        package,
        descriptors,
    }))
    .parse(input)
}

fn parse_package(input: &str) -> PResult<'_, Package> {
    tuple((
        parse_space_terminated,
        parse_space_terminated,
        parse_space_terminated,
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
            let (input, disambiguator) = parse_simple_identifier_str(&input[1..])?;
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
    c.is_alphanumeric() || c == '_' || c == '+' || c == '-' || c == '$'
}

fn parse_simple_identifier_str(input: &str) -> PResult<'_, &str> {
    take_while1(is_simple_identifier_char)(input)
}

fn parse_simple_identifier(input: &str) -> PResult<'_, Cow<'_, str>> {
    parse_simple_identifier_str.map(Cow::Borrowed).parse(input)
}

fn parse_escaped_identifier(input: &str) -> PResult<'_, Cow<'_, str>> {
    let (input, _) = char('`')(input)?;
    let (input, name) = parse_terminated(input, '`')?;
    let (input, _) = char('`')(input)?;
    Ok((input, name))
}

fn parse_space_terminated(input: &str) -> PResult<'_, Cow<'_, str>> {
    let (input, terminated) = parse_terminated(input, ' ')?;
    let (input, _) = char(' ')(input)?;
    Ok((input, terminated))
}

fn parse_terminated(input: &str, terminator: char) -> PResult<'_, Cow<'_, str>> {
    let terminator_len = terminator.len_utf8();
    let mut needs_escape = false;
    let mut current = input;
    while let Some(offset) = current.find(terminator) {
        let (_, rest) = current.split_at(offset + terminator_len);
        if rest.starts_with(terminator) {
            needs_escape = true;
            current = &rest[terminator_len..];
        } else {
            let (raw, rest) = input.split_at(input.len() - rest.len() - terminator_len);
            let escaped = if needs_escape {
                Cow::Owned(raw.replace(
                    &format!("{terminator}{terminator}"),
                    &terminator.to_string(),
                ))
            } else {
                Cow::Borrowed(raw)
            };
            return Ok((rest, escaped));
        }
    }
    context("Missing terminator", cut(fail))(current)
}
