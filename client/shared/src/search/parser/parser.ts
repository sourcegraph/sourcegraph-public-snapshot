import { IRange } from 'monaco-editor'
import { filterTypeKeysWithAliases } from '../interactive/util'

/**
 * Represents a zero-indexed character range in a single-line search query.
 */
export interface CharacterRange {
    /** Zero-based character on the line */
    start: number
    /** Zero-based character on the line */
    end: number
}

/**
 * Converts a zero-indexed, single-line {@link CharacterRange} to a Monaco {@link IRange}.
 */
export const toMonacoRange = ({ start, end }: CharacterRange): IRange => ({
    startLineNumber: 1,
    endLineNumber: 1,
    startColumn: start + 1,
    endColumn: end + 1,
})

enum PatternKind {
    Literal = 1,
    Regexp,
    Structural,
}

export interface Pattern {
    type: 'pattern'
    range: CharacterRange
    kind: PatternKind
    value: string
}

/**
 * Represents a literal in a search query.
 *
 * Example: `Conn`.
 */
export interface Literal {
    type: 'literal'
    range: CharacterRange
    value: string
}

/**
 * Represents a filter in a search query.
 *
 * Example: `repo:^github\.com\/sourcegraph\/sourcegraph$`.
 */
export interface Filter {
    type: 'filter'
    range: CharacterRange
    filterType: Literal
    filterValue: Quoted | Literal | undefined
}

/**
 * Represents an operator in a search query.
 *
 * Example: AND, OR, NOT.
 */
export interface Operator {
    type: 'operator'
    range: CharacterRange
    value: string
}

/**
 * Represents a sequence of tokens in a search query.
 */
export interface Sequence {
    type: 'sequence'
    range: CharacterRange
    members: Token[]
}

/**
 * Represents a quoted string in a search query.
 *
 * Example: "Conn".
 */
export interface Quoted {
    type: 'quoted'
    range: CharacterRange
    quotedValue: string
}

/**
 * Represents a C-style comment, terminated by a newline.
 *
 * Example: `// Oh hai`
 */
export interface Comment {
    type: 'comment'
    range: CharacterRange
    value: string
}

export interface Whitespace {
    type: 'whitespace'
    range: CharacterRange
}

export interface OpeningParen {
    type: 'openingParen'
    range: CharacterRange
}

export interface ClosingParen {
    type: 'closingParen'
    range: CharacterRange
}

export type Token = Whitespace | OpeningParen | ClosingParen | Operator | Comment | Literal | Pattern | Filter | Quoted

export type Term = Token | Sequence

/**
 * Represents the failed result of running a {@link Parser} on a search query.
 */
interface ParseError {
    type: 'error'

    /**
     * A string representing the token that would have been expected
     * for successful parsing at {@link ParseError#at}.
     */
    expected: string

    /**
     * The index in the search query string where parsing failed.
     */
    at: number
}

/**
 * Represents the successful result of running a {@link Parser} on a search query.
 */
export interface ParseSuccess<T = Term> {
    type: 'success'

    /**
     * The resulting term.
     */
    token: T
}

/**
 * Represents the result of running a {@link Parser} on a search query.
 */
export type ParserResult<T = Term> = ParseError | ParseSuccess<T>

type Parser<T = Term> = (input: string, start: number) => ParserResult<T>

/**
 * Returns a {@link Parser} that succeeds if zero or more tokens parsed
 * by the given `parseToken` parsers are found in a search query.
 */
const zeroOrMore = (parseToken: Parser<Term>): Parser<Sequence> => (input, start) => {
    const members: Token[] = []
    let adjustedStart = start
    let end = start + 1
    while (input[adjustedStart] !== undefined) {
        const result = parseToken(input, adjustedStart)
        if (result.type === 'error') {
            return result
        }
        if (result.token.type === 'sequence') {
            for (const member of result.token.members) {
                members.push(member)
            }
        } else {
            members.push(result.token)
        }
        end = result.token.range.end
        adjustedStart = end
    }
    return {
        type: 'success',
        token: { type: 'sequence', members, range: { start, end } },
    }
}

/**
 * Returns a {@link Parser} that succeeds if any of the given parsers succeeds.
 */
const oneOf = <T>(...parsers: Parser<T>[]): Parser<T> => (input, start) => {
    const expected: string[] = []
    for (const parser of parsers) {
        const result = parser(input, start)
        if (result.type === 'success') {
            return result
        }
        expected.push(result.expected)
    }
    return {
        type: 'error',
        expected: `One of: ${expected.join(', ')}`,
        at: start,
    }
}

/**
 * A {@link Parser} that will attempt to parse delimited strings for an arbitrary
 * delimiter. `\` is treated as an escape character for the delimited string.
 */
const quoted = (delimiter: string): Parser<Quoted> => (input, start) => {
    if (input[start] !== delimiter) {
        return { type: 'error', expected: delimiter, at: start }
    }
    let end = start + 1
    while (input[end] && (input[end] !== delimiter || input[end - 1] === '\\')) {
        end = end + 1
    }
    if (!input[end]) {
        return { type: 'error', expected: delimiter, at: end }
    }
    return {
        type: 'success',
        // end + 1 as `end` is currently the index of the quote in the string.
        token: { type: 'quoted', quotedValue: input.slice(start + 1, end), range: { start, end: end + 1 } },
    }
}

/**
 * Returns a {@link Parser} that will attempt to parse tokens matching
 * the given character in a search query.
 */
const character = (character: string): Parser<Literal> => (input, start) => {
    if (input[start] !== character) {
        return { type: 'error', expected: character, at: start }
    }
    return {
        type: 'success',
        token: { type: 'literal', value: character, range: { start, end: start + 1 } },
    }
}

/**
 * Returns a {@link Parser} that will attempt to parse
 * tokens matching the given RegExp pattern in a search query.
 */
const scanToken = <T extends Term = Literal>(
    regexp: RegExp,
    output?: T | ((input: string, range: CharacterRange) => T),
    expected?: string
): Parser<T> => {
    if (!regexp.source.startsWith('^')) {
        regexp = new RegExp(`^${regexp.source}`, regexp.flags)
    }
    return (input, start) => {
        const matchTarget = input.slice(Math.max(0, start))
        if (!matchTarget) {
            return { type: 'error', expected: expected || `/${regexp.source}/`, at: start }
        }
        const match = matchTarget.match(regexp)
        if (!match) {
            return { type: 'error', expected: expected || `/${regexp.source}/`, at: start }
        }
        const range = { start, end: start + match[0].length }
        return {
            type: 'success',
            token: output
                ? typeof output === 'function'
                    ? output(input, range)
                    : output
                : ({ type: 'literal', value: match[0], range } as T),
        }
    }
}

const whitespace = scanToken(
    /\s+/,
    (_input, range): Whitespace => ({
        type: 'whitespace',
        range,
    })
)

const literal = scanToken(/[^\s)]+/)

const operator = scanToken(
    /(and|AND|or|OR|not|NOT)/,
    (input, { start, end }): Operator => ({ type: 'operator', value: input.slice(start, end), range: { start, end } })
)

const comment = scanToken(
    /\/\/.*/,
    (input, { start, end }): Comment => ({ type: 'comment', value: input.slice(start, end), range: { start, end } })
)

const filterKeyword = scanToken(new RegExp(`-?(${filterTypeKeysWithAliases.join('|')})+(?=:)`, 'i'))

const filterDelimiter = character(':')

const filterValue = oneOf<Quoted | Literal>(quoted('"'), quoted("'"), literal)

const openingParen = scanToken(/\(/, (_input, range): OpeningParen => ({ type: 'openingParen', range }))

const closingParen = scanToken(/\)/, (_input, range): ClosingParen => ({ type: 'closingParen', range }))

/**
 * Returns a {@link Parser} that succeeds if a token parsed by `parseToken`,
 * followed by whitespace or EOF, is found in the search query.
 */
const followedBy = (parseToken: Parser<Token>, parseNext: Parser<Token>): Parser<Sequence> => (input, start) => {
    const members: Token[] = []
    const tokenResult = parseToken(input, start)
    if (tokenResult.type === 'error') {
        return tokenResult
    }
    members.push(tokenResult.token)
    let { end } = tokenResult.token.range
    if (input[end] !== undefined) {
        const separatorResult = parseNext(input, end)
        if (separatorResult.type === 'error') {
            return separatorResult
        }
        members.push(separatorResult.token)
        end = separatorResult.token.range.end
    }
    return {
        type: 'success',
        token: { type: 'sequence', members, range: { start, end } },
    }
}

/**
 * A {@link Parser} that will attempt to parse {@link Filter} tokens
 * (consisting a of a filter type and a filter value, separated by a colon)
 * in a search query.
 */
const filter: Parser<Filter> = (input, start) => {
    const parsedKeyword = filterKeyword(input, start)
    if (parsedKeyword.type === 'error') {
        return parsedKeyword
    }
    const parsedDelimiter = filterDelimiter(input, parsedKeyword.token.range.end)
    if (parsedDelimiter.type === 'error') {
        return parsedDelimiter
    }
    const parsedValue =
        input[parsedDelimiter.token.range.end] === undefined
            ? undefined
            : filterValue(input, parsedDelimiter.token.range.end)
    if (parsedValue && parsedValue.type === 'error') {
        return parsedValue
    }
    return {
        type: 'success',
        token: {
            type: 'filter',
            range: { start, end: parsedValue ? parsedValue.token.range.end : parsedDelimiter.token.range.end },
            filterType: parsedKeyword.token,
            filterValue: parsedValue?.token,
        },
    }
}

const baseTerms: Parser<Token>[] = [operator, filter, quoted('"'), quoted("'"), literal]

const createParser = (terms: Parser<Token>[]): Parser<Sequence> =>
    zeroOrMore(
        oneOf<Term>(
            whitespace,
            openingParen,
            closingParen,
            ...terms.map(token => followedBy(token, oneOf<Whitespace | ClosingParen>(whitespace, closingParen)))
        )
    )

/**
 * A {@link Parser} for a Sourcegraph search query.
 */
const searchQuery = createParser(baseTerms)

/**
 * A {@link Parser} for a Sourcegraph search query containing comments.
 */
const searchQueryWithComments = createParser([comment, ...baseTerms])

/**
 * Parses a search query string.
 */
export const parseSearchQuery = (query: string, interpretComments?: boolean): ParserResult<Sequence> =>
    interpretComments ? searchQueryWithComments(query, 0) : searchQuery(query, 0)
