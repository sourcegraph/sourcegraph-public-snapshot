// TODO: Eventually we'll import the actual lsif-typed protobuf file in this project,
// but it doesn't make sense to do so right now.

export interface JsonDocument {
    occurrences: JsonOccurrence[]
}

export interface JsonOccurrence {
    range: number[]
    syntaxKind: number
}

export class Position {
    constructor(public readonly line: number, public readonly character: number) {}
}

export class Range {
    constructor(public readonly start: Position, public readonly end: Position) {}
}

export class Occurrence {
    public range: Range
    public kind: SyntaxKind

    constructor(occ: JsonOccurrence) {
        this.range = new Range(
            new Position(occ.range[0], occ.range[1]),
            // Handle 3 vs 4 length meaning different things
            occ.range.length === 3
                ? // 3 means same row
                  new Position(occ.range[0], occ.range[2])
                : // 4 means could be different rows
                  new Position(occ.range[2], occ.range[3])
        )

        this.kind = occ.syntaxKind
    }
}

// This is copy & pasted from the enum defined in lsif.proto.
// Make sure to update it whenever you update lsif.proto
export enum SyntaxKind {
    UnspecifiedSyntaxKind = 0,

    // Comment, including comment markers and text
    Comment = 1,

    // `,` `.` `,`
    PunctuationDelimiter = 2,
    // (), {}, [] when used syntactically
    PunctuationBracket = 3,

    // `if`, `else`, `return`, `class`, etc.
    IdentifierKeyword = 4,

    // `+`, `*`, etc.
    IdentifierOperator = 5,

    // non-specific catch-all for any identifier not better described elsewhere
    Identifier = 6,
    // Identifiers builtin to the language: `min`, `print` in Python.
    IdentifierBuiltin = 7,
    // Identifiers representing `null`-like values: `None` in Python, `nil` in Go.
    IdentifierNull = 8,
    // `xyz` in `const xyz = "hello"`
    IdentifierConstant = 9,
    // `var X = "hello"` in Go
    IdentifierMutableGlobal = 10,
    // both parameter definition and references
    IdentifierParameter = 11,
    // identifiers for variable definitions and references within a local scope
    IdentifierLocal = 12,
    // Used when identifier shadowes some other identifier within the scope
    IdentifierShadowed = 13,
    // `package main`
    IdentifierModule = 14,

    // Function call/reference
    IdentifierFunction = 15,
    // Function definition only
    IdentifierFunctionDefinition = 16,

    // Macro call/reference
    IdentifierMacro = 17,
    // Macro definition only
    IdentifierMacroDefinition = 18,

    // non-builtin types, including namespaces
    IdentifierType = 19,
    // builtin types only, such as `str` for Python or `int` in Go
    IdentifierBuiltinType = 20,

    // Python decorators, c-like __attribute__
    IdentifierAttribute = 21,

    // `\b`
    RegexEscape = 22,
    // `*`, `+`
    RegexRepeated = 23,
    // `.`
    RegexWildcard = 24,
    // `(`, `)`, `[`, `]`
    RegexDelimiter = 25,
    // `|`, `-`
    RegexJoin = 26,

    // Literal strings: "Hello, world!"
    StringLiteral = 27,
    // non-regex escapes: "\t", "\n"
    StringLiteralEscape = 28,
    // datetimes within strings, special words within a string, `{}` in format strings
    StringLiteralSpecial = 29,
    // "key" in { "key": "value" }, useful for example in JSON
    StringLiteralKey = 30,
    // 'c' or similar, in languages that differentiate strings and characters
    CharacterLiteral = 31,
    // Literal numbers, both floats and integers
    NumericLiteral = 32,
    // `true`, `false`
    BooleanLiteral = 33,

    // Used for XML-like tags
    Tag = 34,
    // Attribute name in XML-like tags
    TagAttribute = 35,
    // Delimiters for XML-like tags
    TagDelimiter = 36,
}
