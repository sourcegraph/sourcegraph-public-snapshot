interface CharacterRange {
    start: number
    end: number
}

export interface Literal {
    type: 'literal'
    value: string
}

export interface Filter {
    type: 'filter'
    filterType: Pick<ParseSuccess<Literal>, 'range' | 'token'>
    filterValue: Pick<ParseSuccess<Literal | Quoted>, 'range' | 'token'> | undefined
}

export interface Sequence {
    type: 'sequence'
    members: Pick<ParseSuccess<Exclude<Token, Sequence>>, 'range' | 'token'>[]
}

export interface Quoted {
    type: 'quoted'
    quotedValue: string
}

export type Token = { type: 'whitespace' } | Literal | Filter | Sequence | Quoted

interface ParseError {
    type: 'error'
    expected: string
    at: number
}

export interface ParseSuccess<T = Token> {
    type: 'success'
    token: T
    range: CharacterRange
}

type ParserResult<T = Token> = ParseError | ParseSuccess<T>

type Parser<T = Token> = (input: string, start: number) => ParserResult<T>

const flatten = (members: Pick<ParseSuccess, 'range' | 'token'>[]): Sequence['members'] =>
    members.reduce(
        (merged: Sequence['members'], { range, token }) =>
            token.type === 'sequence' ? [...merged, ...flatten(token.members)] : [...merged, { token, range }],
        []
    )

const zeroOrMore = (parse: Parser, parseSeparator: Parser): Parser<Sequence> => (input, start) => {
    const members: Pick<ParseSuccess, 'range' | 'token'>[] = []
    let adjustedStart = start
    let end = start
    // try to start with separator
    const separatorResult = parseSeparator(input, start)
    if (separatorResult.type === 'success') {
        end = separatorResult.range.end
        adjustedStart = separatorResult.range.end + 1
        const { token, range } = separatorResult
        members.push({ token, range })
    }
    let result = parse(input, adjustedStart)
    while (result.type !== 'error') {
        const { token, range } = result
        members.push({ token, range })
        end = result.range.end
        adjustedStart = end + 1
        if (input[adjustedStart] === undefined) {
            // EOF
            break
        }
        // Parse separator
        const separatorResult = parseSeparator(input, adjustedStart)
        if (separatorResult.type === 'error') {
            return separatorResult
        }
        // Try to parse another token.
        end = separatorResult.range.end
        adjustedStart = end + 1
        members.push({ token: separatorResult.token, range: separatorResult.range })
        result = parse(input, adjustedStart)
    }
    return {
        type: 'success',
        range: { start, end },
        token: { type: 'sequence', members: flatten(members) },
    }
}

const oneOf = <T = Token>(...parsers: Parser<T>[]): Parser<T> => (input, start) => {
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
        range: { start, end },
        token: { type: 'quoted', quotedValue: input.substring(start + 1, end) },
    }
}

const character = (c: string): Parser<Literal> => (input, start) => {
    if (input[start] !== c) {
        return { type: 'error', expected: c, at: start }
    }
    return {
        type: 'success',
        range: { start, end: start },
        token: { type: 'literal', value: c },
    }
}

const pattern = <T = Literal>(p: RegExp, output?: T, expected?: string): Parser<T> => {
    if (!p.source.startsWith('^')) {
        p = new RegExp(`^${p.source}`)
    }
    return (input, start) => {
        const matchTarget = input.substring(start)
        if (!matchTarget) {
            return { type: 'error', expected: expected || `/${p.source}/`, at: start }
        }
        const match = input.substring(start).match(p)
        if (!match) {
            return { type: 'error', expected: expected || `/${p.source}/`, at: start }
        }
        return {
            type: 'success',
            range: { start, end: start + match[0].length - 1 },
            token: (output || { type: 'literal', value: match[0] }) as T,
        }
    }
}

const whitespace = pattern(/\s+/, { type: 'whitespace' as const }, 'whitespace')

const literal = pattern(/[^\s]+/)

const filterKeyword = pattern(/-?[a-z]+(?=:)/)

const filterDelimiter = character(':')

const filterValue = oneOf<Quoted | Literal>(quoted, pattern(/[^:\s'"]+/))

const filter: Parser<Filter> = (input, start) => {
    const parsedKeyword = filterKeyword(input, start)
    if (parsedKeyword.type === 'error') {
        return parsedKeyword
    }
    const parsedDelimiter = filterDelimiter(input, parsedKeyword.range.end + 1)
    if (parsedDelimiter.type === 'error') {
        return parsedDelimiter
    }
    const parsedValue =
        input[parsedDelimiter.range.end + 1] === undefined
            ? undefined
            : filterValue(input, parsedDelimiter.range.end + 1)
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

const searchQuery = zeroOrMore(oneOf<Filter | Quoted | Literal>(filter, quoted, literal), whitespace)

export const parseSearchQuery = (query: string): ParserResult<Sequence> => searchQuery(query, 0)
