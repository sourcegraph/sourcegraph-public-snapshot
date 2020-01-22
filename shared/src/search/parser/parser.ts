import { IRange } from 'monaco-editor'

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

/**
 * Represents a literal in a search query.
 *
 * Example: `Conn`.
 */
export interface Literal {
    type: 'literal'
    value: string
}

/**
 * Represents a filter in a search query.
 *
 * Example: `repo:^github\.com\/sourcegraph\/sourcegraph$`.
 */
export interface Filter {
    type: 'filter'
    filterType: Pick<ParseSuccess<Literal>, 'range' | 'token'>
    filterValue: Pick<ParseSuccess<Literal | Quoted>, 'range' | 'token'> | undefined
}

/**
 * Represents a sequence of tokens in a search query.
 */
export interface Sequence {
    type: 'sequence'
    members: Pick<ParseSuccess<Exclude<Token, Sequence>>, 'range' | 'token'>[]
}

/**
 * Represents a quoted string in a search query.
 *
 * Example: "Conn".
 */
export interface Quoted {
    type: 'quoted'
    quotedValue: string
}

export type Token = { type: 'whitespace' } | Literal | Filter | Sequence | Quoted

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
export interface ParseSuccess<T = Token> {
    type: 'success'

    /**
     * The parsed token.
     */
    token: T

    /**
     * The character range that was successfully parsed.
     */
    range: CharacterRange
}

/**
 * Represents the result of running a {@link Parser} on a search query.
 */
type ParserResult<T = Token> = ParseError | ParseSuccess<T>

type Parser<T = Token> = (input: string, start: number) => ParserResult<T>

/**
 * Returns a {@link Parser} that succeeds if zero or more of the given `parseToken` parsers,
 * separated by `parsedSeparator`, are found in a search query.
 */
const zeroOrMore = (
    parseToken: Parser<Exclude<Token, Sequence>>,
    parseSeparator: Parser<Exclude<Token, Sequence>>
): Parser<Sequence> => (input, start) => {
    const members: Pick<ParseSuccess<Exclude<Token, Sequence>>, 'range' | 'token'>[] = []
    let end = start + 1
    let adjustedStart = start
    // try to start with separator
    const separatorResult = parseSeparator(input, start)
    if (separatorResult.type === 'success') {
        end = separatorResult.range.end
        adjustedStart = separatorResult.range.end
        const { token, range } = separatorResult
        members.push({ token, range })
    }
    let result = parseToken(input, adjustedStart)
    while (result.type !== 'error') {
        const { token, range } = result
        members.push({ token, range })
        end = result.range.end
        if (input[end] === undefined) {
            // EOF
            break
        }
        // Parse separator
        const separatorResult = parseSeparator(input, end)
        if (separatorResult.type === 'error') {
            return separatorResult
        }
        // Try to parse another token.
        end = separatorResult.range.end
        members.push({ token: separatorResult.token, range: separatorResult.range })
        result = parseToken(input, end)
    }
    return {
        type: 'success',
        range: { start, end },
        token: { type: 'sequence', members },
    }
}

/**
 * Returns a {@link Parser} that succeeds if any of the given parsers succeeds.
 */
const oneOf = <T extends Exclude<Token, Sequence>>(...parsers: Parser<T>[]): Parser<T> => (input, start) => {
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
 * A {@link Parser} that will attempt to parse quoted strings in a search query.
 */
const quoted: Parser<Quoted> = (input, start) => {
    if (input[start] !== '"') {
        return { type: 'error', expected: '"', at: start }
    }
    let end = start + 1
    while (input[end] && (input[end] !== '"' || input[end - 1] === '\\')) {
        end = end + 1
    }
    if (!input[end]) {
        return { type: 'error', expected: '"', at: end }
    }
    return {
        type: 'success',
        // end + 1 as `end` is currently the index of the quote in the string.
        range: { start, end: end + 1 },
        token: { type: 'quoted', quotedValue: input.substring(start + 1, end) },
    }
}

/**
 * Returns a {@link Parser} that will attempt to parse tokens matching
 * the given character in a search query.
 */
const character = (c: string): Parser<Literal> => (input, start) => {
    if (input[start] !== c) {
        return { type: 'error', expected: c, at: start }
    }
    return {
        type: 'success',
        range: { start, end: start + 1 },
        token: { type: 'literal', value: c },
    }
}

/**
 * Returns a {@link Parser} that will attempt to parse
 * tokens matching the given RegExp pattern in a search query.
 */
const pattern = <T = Literal>(p: RegExp, output?: T, expected?: string): Parser<T> => {
    if (!p.source.startsWith('^')) {
        p = new RegExp(`^${p.source}`)
    }
    return (input, start) => {
        const matchTarget = input.substring(start)
        if (!matchTarget) {
            return { type: 'error', expected: expected || `/${p.source}/`, at: start }
        }
        const match = matchTarget.match(p)
        if (!match) {
            return { type: 'error', expected: expected || `/${p.source}/`, at: start }
        }
        return {
            type: 'success',
            range: { start, end: start + match[0].length },
            token: (output || { type: 'literal', value: match[0] }) as T,
        }
    }
}

const whitespace = pattern(/\s+/, { type: 'whitespace' as const }, 'whitespace')

const literal = pattern(/[^\s]+/)

const filterKeyword = pattern(/-?[a-z]+(?=:)/i)

const filterDelimiter = character(':')

const filterValue = oneOf<Quoted | Literal>(quoted, pattern(/[^:\s'"]+/))

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
    const parsedDelimiter = filterDelimiter(input, parsedKeyword.range.end)
    if (parsedDelimiter.type === 'error') {
        return parsedDelimiter
    }
    const parsedValue =
        input[parsedDelimiter.range.end] === undefined ? undefined : filterValue(input, parsedDelimiter.range.end)
    if (parsedValue && parsedValue.type === 'error') {
        return parsedValue
    }
    return {
        type: 'success',
        range: { start, end: parsedValue ? parsedValue.range.end : parsedDelimiter.range.end },
        token: {
            type: 'filter',
            filterType: parsedKeyword,
            filterValue: parsedValue,
        },
    }
}

/**
 * A {@link Parser} for a Sourcegraph search query.
 */
const searchQuery = zeroOrMore(oneOf<Filter | Quoted | Literal>(filter, quoted, literal), whitespace)

/**
 * Parses a search query string.
 */
export const parseSearchQuery = (query: string): ParserResult<Sequence> => searchQuery(query, 0)
