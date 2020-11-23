import * as Monaco from 'monaco-editor'
import { Token, Pattern, CharacterRange, PatternKind } from './scanner'
import { RegExpParser, visitRegExpAST } from 'regexpp'
import {
    Alternative,
    Assertion,
    CapturingGroup,
    Character,
    CharacterClass,
    CharacterClassRange,
    CharacterSet,
    Group,
    Quantifier,
} from 'regexpp/ast'

export enum RegexpMetaKind {
    Assertion = 'Assertion', // like ^ or \b
    Alternative = 'Alternative', // like |
    Delimited = 'Delimited', // like ( or )
    EscapedCharacter = 'EscapedCharacter', // like \(
    CharacterSet = 'CharacterSet', // like \s
    CharacterClass = 'CharacterClass', // like [a-z]
    LazyQuantifier = 'LazyQuantifier', // the ? after a range quantifier
    RangeQuantifier = 'RangeQuantifier', // like +
}

export interface RegexpMeta {
    type: 'regexpMeta'
    range: CharacterRange
    kind: RegexpMetaKind
    value: string
}

export enum StructuralMetaKind {
    Hole = 'Hole',
}

export interface StructuralMeta {
    type: 'structuralMeta'
    range: CharacterRange
    kind: StructuralMetaKind
    value: string
}

export interface Field {
    type: 'field'
    range: CharacterRange
    value: string
}

export type MetaToken = RegexpMeta | StructuralMeta

type DecoratedToken = Token | Field | MetaToken

const mapRegexpMeta = (pattern: Pattern): DecoratedToken[] => {
    const tokens: DecoratedToken[] = []
    try {
        const ast = new RegExpParser().parsePattern(pattern.value)
        const offset = pattern.range.start
        visitRegExpAST(ast, {
            onAlternativeEnter(node: Alternative) {
                // regexpp doesn't tell us where a '|' operator is. We infer it by visiting any
                // pattern of an Alternative node, and for a '|' directly after it. Based on
                // regexpp's implementation, we know this is a true '|' operator, and _not_ an
                // escaped \| or part of a character class like [abcd|].
                if (pattern.value[node.end] && pattern.value[node.end] === '|') {
                    tokens.push({
                        type: 'regexpMeta',
                        range: { start: offset + node.end, end: offset + node.end + 1 },
                        value: '|',
                        kind: RegexpMetaKind.Alternative,
                    })
                }
            },
            onAssertionEnter(node: Assertion) {
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.start, end: offset + node.end },
                    value: node.raw,
                    kind: RegexpMetaKind.Assertion,
                })
            },
            onGroupEnter(node: Group) {
                // Push the leading '('
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.start, end: offset + node.start + 1 },
                    value: '(',
                    kind: RegexpMetaKind.Delimited,
                })
                // Push the trailing ')'
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.end - 1, end: offset + node.end },
                    value: ')',
                    kind: RegexpMetaKind.Delimited,
                })
            },
            onCapturingGroupEnter(node: CapturingGroup) {
                // Push the leading '('
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.start, end: offset + node.start + 1 },
                    value: '(',
                    kind: RegexpMetaKind.Delimited,
                })
                // Push the trailing ')'
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.end - 1, end: offset + node.end },
                    value: ')',
                    kind: RegexpMetaKind.Delimited,
                })
            },
            onCharacterSetEnter(node: CharacterSet) {
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.start, end: offset + node.end },
                    value: node.raw,
                    kind: RegexpMetaKind.CharacterSet,
                })
            },
            onCharacterClassEnter(node: CharacterClass) {
                // Push the leading '['
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.start, end: offset + node.start + 1 },
                    value: '[',
                    kind: RegexpMetaKind.CharacterClass,
                })
                // Push the trailing ']'
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.end - 1, end: offset + node.end },
                    value: ']',
                    kind: RegexpMetaKind.CharacterClass,
                })
            },
            onCharacterClassRangeEnter(node: CharacterClassRange) {
                // highlight the '-' in [a-z]. Take care to use node.min.end, because we
                // don't want to highlight the first '-' in [--z], nor an escaped '-' with a
                // two-character offset as in [\--z].
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.min.end, end: offset + node.min.end + 1 },
                    value: '-',
                    kind: RegexpMetaKind.CharacterClass,
                })
            },
            onQuantifierEnter(node: Quantifier) {
                // the lazy quantifier ? adds one
                const lazyQuantifierOffset = node.greedy ? 0 : 1
                if (!node.greedy) {
                    tokens.push({
                        type: 'regexpMeta',
                        range: { start: offset + node.end - 1, end: offset + node.end },
                        value: '?',
                        kind: RegexpMetaKind.LazyQuantifier,
                    })
                }

                const quantifier = node.raw[node.raw.length - lazyQuantifierOffset - 1]
                if (quantifier === '+' || quantifier === '*' || quantifier === '?') {
                    tokens.push({
                        type: 'regexpMeta',
                        range: {
                            start: offset + node.end - 1 - lazyQuantifierOffset,
                            end: offset + node.end - lazyQuantifierOffset,
                        },
                        value: node.raw,
                        kind: RegexpMetaKind.RangeQuantifier,
                    })
                } else {
                    // regexpp provides no easy way to tell whether the quantifier is a range '{number, number}',
                    // nor the offsets of this range.
                    // At this point we know it is none of +, *, or ?, so it is a ranged quantifier.
                    // We need to then find the opening brace of {number, number}, and go backwards from the end
                    // of this quantifier to avoid dealing with other leading braces that are not part of it.
                    let openBrace = node.end - 1 - lazyQuantifierOffset
                    while (pattern.value[openBrace] && pattern.value[openBrace] !== '{') {
                        openBrace = openBrace - 1
                    }
                    tokens.push({
                        type: 'regexpMeta',
                        range: { start: offset + openBrace, end: offset + node.end },
                        value: pattern.value.slice(offset + openBrace, offset + node.end - lazyQuantifierOffset),
                        kind: RegexpMetaKind.RangeQuantifier,
                    })
                }
            },
            onCharacterEnter(node: Character) {
                if (node.end - node.start > 1 && node.raw.startsWith('\\')) {
                    // This is an escaped value like `\.`, `\u0065`, `\x65`.
                    tokens.push({
                        type: 'regexpMeta',
                        range: { start: offset + node.start, end: offset + node.end },
                        value: node.raw,
                        kind: RegexpMetaKind.EscapedCharacter,
                    })
                    return
                }
                tokens.push({
                    type: 'pattern',
                    range: { start: offset + node.start, end: offset + node.end },
                    value: node.raw,
                    kind: PatternKind.Regexp,
                })
            },
        })
    } catch {
        tokens.push(pattern)
    }
    // The AST is not necessarily traversed in increasing range. We need
    // to sort by increasing range because the ordering is significant to Monaco.
    tokens.sort((left, right) => {
        if (left.range.start < right.range.start) {
            return -1
        }
        return 0
    })
    return tokens
}

const mapStructuralMeta = (pattern: Pattern): DecoratedToken[] => {
    const offset = pattern.range.start

    const decorated: DecoratedToken[] = []
    let current = ''
    let start = 0
    let token: string[] = []

    // Track context of whether we are inside an opening hole, e.g., after
    // ':['. Value is greater than 1 when inside.
    let open = 0
    // Track whether we are balanced inside a regular expression character
    // set like '[a]' inside an open hole, e.g., :[foo~[a]]. Value is greater
    // than 1 when inside.
    let inside = 0

    const nextChar = (): string => {
        current = pattern.value[start]
        start = start + 1
        return current
    }

    // Appends a decorated token to the list of tokens, and resets the current token to be empty.
    const appendDecoratedToken = (endIndex: number, kind: PatternKind.Literal | StructuralMetaKind): void => {
        const value = token.join('')
        const range = { start: offset + endIndex - value.length, end: offset + endIndex }
        if (kind === PatternKind.Literal) {
            decorated.push({ type: 'pattern', kind, value, range })
        } else {
            decorated.push({ type: 'structuralMeta', kind, value, range })
        }
        token = []
    }

    while (pattern.value[start] !== undefined) {
        current = nextChar()
        switch (current) {
            case '.':
                // Look ahead and see if this is a ... hole alias.
                if (pattern.value.slice(start, start + 2) === '..') {
                    // It is a ... hole.
                    if (token.length > 0) {
                        // Append the value before this '...'.
                        appendDecoratedToken(start - 1, PatternKind.Literal)
                    }
                    start = start + 2
                    // Append the value of '...' after advancing.
                    appendDecoratedToken(start - 3, StructuralMetaKind.Hole)
                    continue
                }
                token.push('.')
                break
            case ':':
                if (open > 0) {
                    // ':' inside a hole, likely part of a regexp pattern.
                    token.push(':')
                    continue
                }
                if (pattern.value[start] !== undefined) {
                    if (pattern.value[start] === ':') {
                        // '::' case, so push the first ':' and continue.
                        token.push(':')
                        continue
                    }
                    // Look ahead and see if this is the start of a hole.
                    current = nextChar()
                    if (current === '[') {
                        // It is the start of a hole.
                        open = open + 1
                        // Persist the literal token scanned up to this point.
                        appendDecoratedToken(start - 2, PatternKind.Literal)
                        token.push(':[')
                        continue
                    }
                    // Something else, push the ':' we saw, backtrack, and continue.
                    token.push(':')
                    start = start - 1
                    continue
                }
                // Trailing ':'.
                token.push(current)
                break
            case '\\':
                if (pattern.value[start] !== undefined && open > 0) {
                    // Assume this is an escape sequence inside a regexp hole.
                    current = nextChar()
                    token.push('\\', current)
                    continue
                }
                token.push('\\')
                break
            case '[':
                if (open > 0) {
                    // Assume this is a character set inside a regexp hole.
                    inside = inside + 1
                    token.push('[')
                    continue
                }
                token.push('[')
                break
            case ']':
                if (open > 0 && inside > 0) {
                    // This ']' closes a regular expression inside a hole.
                    inside = inside - 1
                    token.push(current)
                    continue
                }
                if (open > 0) {
                    // This ']' closes a hole.
                    open = open - 1
                    token.push(']')
                    appendDecoratedToken(start, StructuralMetaKind.Hole)
                    continue
                }
                token.push(current)
                break
            default:
                token.push(current)
        }
    }
    if (token.length > 0) {
        // Append any left over literal at the end.
        appendDecoratedToken(start, PatternKind.Literal)
    }
    return decorated
}

/**
 * Returns true for filter values that have regexp values, e.g., repo, file.
 * Excludes FilterType.content because that depends on the pattern kind.
 */
export const hasRegexpValue = (field: string): boolean => {
    const fieldName = field.startsWith('-') ? field.slice(1) : field
    switch (fieldName.toLocaleLowerCase()) {
        case 'repo':
        case 'r':
        case 'file':
        case 'f':
        case 'repohasfile':
        case 'message':
        case 'msg':
        case 'm':
        case 'commiter':
        case 'author':
            return true
        default:
            return false
    }
}

const decorateTokens = (tokens: Token[]): DecoratedToken[] => {
    const decorated: DecoratedToken[] = []
    for (const token of tokens) {
        switch (token.type) {
            case 'pattern':
                switch (token.kind) {
                    case PatternKind.Regexp:
                        decorated.push(...mapRegexpMeta(token))
                        break
                    case PatternKind.Structural:
                        decorated.push(...mapStructuralMeta(token))
                        break
                    case PatternKind.Literal:
                        decorated.push(token)
                        break
                }
                break
            case 'filter': {
                decorated.push({
                    type: 'field',
                    range: token.range,
                    value: token.field.value,
                })
                if (token.value && token.value.type === 'literal' && hasRegexpValue(token.field.value)) {
                    // Highlight fields with regexp values.
                    decorated.push(
                        ...decorateTokens([
                            {
                                type: 'pattern',
                                kind: PatternKind.Regexp,
                                value: token.value.value,
                                range: token.value.range,
                            },
                        ])
                    )
                } else if (token.value) {
                    decorated.push(token.value)
                }
                break
            }
            default:
                decorated.push(token)
        }
    }
    return decorated
}

const fromDecoratedTokens = (tokens: DecoratedToken[]): Monaco.languages.IToken[] => {
    const monacoTokens: Monaco.languages.IToken[] = []
    for (const token of tokens) {
        switch (token.type) {
            case 'field':
            case 'whitespace':
            case 'keyword':
            case 'comment':
            case 'openingParen':
            case 'closingParen':
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: token.type,
                })
                break
            case 'regexpMeta':
            case 'structuralMeta':
                /** The scopes value is derived from the token type and its kind.
                 * E.g., regexpMetaDelimited derives from {@link RegexpMeta} and {@link RegexpMetaKind}.
                 */
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: `${token.type}${token.kind}`,
                })
                break
            default:
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: 'identifier',
                })
                break
        }
    }
    return monacoTokens
}

const fromTokens = (tokens: Token[]): Monaco.languages.IToken[] => {
    const monacoTokens: Monaco.languages.IToken[] = []
    for (const token of tokens) {
        switch (token.type) {
            case 'filter':
                {
                    monacoTokens.push({
                        startIndex: token.field.range.start,
                        scopes: 'field',
                    })
                    if (token.value) {
                        monacoTokens.push({
                            startIndex: token.value.range.start,
                            scopes: 'identifier',
                        })
                    }
                }
                break
            case 'whitespace':
            case 'keyword':
            case 'comment':
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: token.type,
                })
                break
            default:
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: 'identifier',
                })
                break
        }
    }
    return monacoTokens
}

/**
 * Returns the tokens in a scanned search query displayed in the Monaco query input. If the experimental
 * decorate flag is true, a list of {@link DecoratedToken} provides more contextual highlighting for patterns.
 */
export const getMonacoTokens = (tokens: Token[], decorate = false): Monaco.languages.IToken[] =>
    decorate ? fromDecoratedTokens(decorateTokens(tokens)) : fromTokens(tokens)
