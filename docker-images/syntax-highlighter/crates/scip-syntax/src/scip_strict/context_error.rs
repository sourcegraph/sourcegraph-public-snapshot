use std::fmt;

use nom::error::{ContextError, ErrorKind, FromExternalError, ParseError};

/// default error type, only contains the error's location and code
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CtxError<I> {
    /// position of the error in the input data
    pub input: I,
    /// contextual error message
    pub context: &'static str,
}

impl<I> ParseError<I> for CtxError<I> {
    fn from_error_kind(input: I, _kind: ErrorKind) -> Self {
        CtxError {
            input,
            context: "no context set yet",
        }
    }

    fn append(_: I, _: ErrorKind, other: Self) -> Self {
        other
    }
}

impl<I> ContextError<I> for CtxError<I> {
    fn add_context(_input: I, context: &'static str, mut other: Self) -> Self {
        other.context = context;
        other
    }
}

impl<I, E> FromExternalError<I, E> for CtxError<I> {
    /// Create a new error from an input position and an external error
    fn from_external_error(input: I, _kind: ErrorKind, _e: E) -> Self {
        CtxError {
            input,
            context: "no context set yet",
        }
    }
}

/// The Display implementation allows the std::error::Error implementation
impl<I: fmt::Display> fmt::Display for CtxError<I> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "error '{}' at: {}", self.context, self.input)
    }
}
