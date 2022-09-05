// TODO: Eventually we'll import the actual lsif-typed protobuf file in this project,
// but it doesn't make sense to do so right now.

export interface JsonDocument {
    occurrences?: JsonOccurrence[]
}

export interface DocumentInfo {
    content: string
    lsif?: string
}

export interface JsonOccurrence {
    range: [number, number, number] | [number, number, number, number]
    syntaxKind?: SyntaxKind
}

export class Position {
    constructor(public readonly line: number, public readonly character: number) {}

    public isSmaller(other: Position): boolean {
        return this.compare(other) < 0
    }
    public isSmallerOrEqual(other: Position): boolean {
        return this.compare(other) <= 0
    }
    public isGreater(other: Position): boolean {
        return this.compare(other) > 0
    }
    public isGreaterOrEqual(other: Position): boolean {
        return this.compare(other) >= 0
    }
    public compare(other: Position): number {
        if (this.line !== other.line) {
            return this.line - other.line
        }
        return this.character - other.character
    }
}

export class Range {
    constructor(public readonly start: Position, public readonly end: Position) {}
    public withStart(newStart: Position): Range {
        return new Range(newStart, this.end)
    }
    public withEnd(newEnd: Position): Range {
        return new Range(this.start, newEnd)
    }
    public isZeroWidth(): boolean {
        return this.start.compare(this.end) === 0
    }
    public isOverlapping(other: Range): boolean {
        return this.start.isSmallerOrEqual(other.start) && this.end.isGreater(other.start)
    }
    public isSingleLine(): boolean {
        return this.start.line === this.end.line
    }
    public compare(other: Range): number {
        const byStart = this.start.compare(other.start)
        if (byStart !== 0) {
            return byStart
        }
        // Both ranges have the same start position, sort by inverse end
        // position so that the longer range appears first. The motivation for
        // using inverse order is so that we handle the larger range first when
        // dealing with overlapping ranges.
        return -this.end.compare(other.end)
    }
}

export class Occurrence {
    constructor(public readonly range: Range, public readonly kind?: SyntaxKind) {}

    public withStartPosition(newStartPosition: Position): Occurrence {
        return this.withRange(this.range.withStart(newStartPosition))
    }
    public withEndPosition(newEndPosition: Position): Occurrence {
        return this.withRange(this.range.withEnd(newEndPosition))
    }
    public withRange(newRange: Range): Occurrence {
        return new Occurrence(newRange, this.kind)
    }

    public static fromJson(occ: JsonOccurrence): Occurrence {
        const range = new Range(
            new Position(occ.range[0], occ.range[1]),
            // Handle 3 vs 4 length meaning different things
            occ.range.length === 3
                ? // 3 means same row
                  new Position(occ.range[0], occ.range[2])
                : // 4 means could be different rows
                  new Position(occ.range[2], occ.range[3])
        )

        return new Occurrence(range, occ.syntaxKind)
    }
    public static fromInfo(info: DocumentInfo): Occurrence[] {
        const sortedSingleLineOccurrences = parseJsonOccurrencesIntoSingleLineOccurrences(info)
        return nonOverlappingOccurrences(sortedSingleLineOccurrences)
    }
}

// Converts an array of potentially overlapping occurrences into an array of
// non-overlapping occurrences.  The most narrow occurrence "wins", meaning that
// when two ranges overlap, we pick the syntax kind of the occurrence with the
// shortest distance between start/end.
function nonOverlappingOccurrences(occurrences: Occurrence[]): Occurrence[] {
    // NOTE: we can't guarantee that the occurrences are sorted from the server
    // or after splitting multiline occurrences into single-line occurrences.
    const stack: Occurrence[] = occurrences.sort((a, b) => a.range.compare(b.range)).reverse()
    const result: Occurrence[] = []
    const pushResult = (occ: Occurrence): void => {
        if (!occ.range.isZeroWidth()) {
            result.push(occ)
        }
    }
    while (true) {
        const current = stack.pop()
        if (!current) {
            break
        }
        const next = stack.pop()
        if (next) {
            if (current.range.isOverlapping(next.range)) {
                pushResult(current.withEndPosition(next.range.start))
                stack.push(current.withStartPosition(next.range.end))
            } else {
                pushResult(current)
            }
            stack.push(next)
        } else {
            pushResult(current)
        }
    }
    return result.sort((a, b) => a.range.compare(b.range))
}

// Returns a list of occurrences that are guaranteed to only consiste of
// single-line ranges.  A multiline occurrence gets split into multiple
// occurrences where each range encloses a single line.
function parseJsonOccurrencesIntoSingleLineOccurrences(info: DocumentInfo): Occurrence[] {
    if (!info.lsif) {
        return []
    }

    const jsonOccurrences = (JSON.parse(info.lsif) as JsonDocument).occurrences ?? []
    const lines = info.content.split(/\r?\n/g)
    const result: Occurrence[] = []
    for (const jsonOccurrence of jsonOccurrences) {
        const occurrence = Occurrence.fromJson(jsonOccurrence)
        if (occurrence.range.isSingleLine()) {
            result.push(occurrence)
            continue
        }
        for (let line = occurrence.range.start.line; line <= occurrence.range.end.line; line++) {
            const startCharacter = line === occurrence.range.start.line ? occurrence.range.start.character : 0
            const endCharacter =
                line === occurrence.range.end.line ? occurrence.range.end.character : lines[line].length
            const singleLineOccurrence = new Occurrence(
                new Range(new Position(line, startCharacter), new Position(line, endCharacter)),
                occurrence.kind
            )
            result.push(singleLineOccurrence)
        }
    }
    return result
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
