import * as Monaco from 'monaco-editor'
import { Token, Pattern, Literal, PatternKind, CharacterRange } from './token'
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

/**
 * A DecoratedToken is a type of token used for syntax highlighting, hovers, and diagnostics. All
 * standard Token types are compatible where DecoratedTokens are used. A DecoratedToken extends
 * the definition of standard tokens to language-specific metasyntax tokens, like .* in regexp,
 * or :[holes] in structural search.
 */
export type DecoratedToken = Token | MetaToken

/**
 * A MetaToken defines a token that is associated with some language-specific metasyntax.
 */
export type MetaToken = MetaRegexp | MetaStructural | MetaField | MetaRepoRevisionSeparator | MetaRevision

/**
 * Defines common properties for meta tokens.
 */
export interface BaseMetaToken {
    type: MetaToken['type']
    range: CharacterRange
    value: string
}

/**
 * A token that is labeled and interpreted as regular expression syntax.
 */
export interface MetaRegexp extends BaseMetaToken {
    type: 'metaRegexp'
    groupRange?: CharacterRange
    kind: MetaRegexpKind
}

/**
 * Classifications of the kinds of regexp metasyntax.
 */
export enum MetaRegexpKind {
    Assertion = 'Assertion', // like ^ or \b
    Alternative = 'Alternative', // like |
    Delimited = 'Delimited', // like ( or )
    EscapedCharacter = 'EscapedCharacter', // like \(
    CharacterSet = 'CharacterSet', // like \s
    CharacterClass = 'CharacterClass', // like [a-z]
    LazyQuantifier = 'LazyQuantifier', // the ? after a range quantifier
    RangeQuantifier = 'RangeQuantifier', // like +
}

/**
 * A token that is labeled and interpreted as structural search syntax.
 */
export interface MetaStructural extends BaseMetaToken {
    type: 'metaStructural'
    kind: MetaStructuralKind
}

/**
 * Classifications of the kinds of structural metasyntax.
 */
export enum MetaStructuralKind {
    Hole = 'Hole',
}

/**
 * A token that is labeled and interpreted as a field, like "repo:" in the Sourcegraph language syntax.
 */
export interface MetaField extends BaseMetaToken {
    type: 'field'
}

/**
 * A token that is labeled and interpreted as repository revision syntax.
 * See https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions.
 */
export interface MetaRevision extends BaseMetaToken {
    type: 'metaRevision'
    kind: MetaRevisionKind
}

export enum MetaRevisionKind {
    Separator = 'Separator', // is a ':'
    Label = 'Label', // a branch or tag
    CommitHash = 'CommitHash', // a commit hash
    PathLike = 'PathLike', // a path-like pattern, e.g., the refs/heads/ part in *refs/heads/*
    Wildcard = 'Wildcard', // a '*' in glob syntax
    Negate = 'Negate', // a '!' in glob syntax
}

/**
 * A token that denotes a revision separator in the Sourcegraph query language. I.e., the '@' in
 * "repo:^foo$@revision" syntax.
 */
export interface MetaRepoRevisionSeparator extends BaseMetaToken {
    type: 'metaRepoRevisionSeparator'
}

/**
 * Coalesces consecutive pattern tokens. Used, for example, when parsing
 * literal characters like 'f', 'o', 'o' in regular expressions, which are
 * coalesced to 'foo' for hovers.
 *
 * @param tokens
 */
const coalescePatterns = (tokens: DecoratedToken[]): DecoratedToken[] => {
    let previous: Pattern | undefined
    const newTokens: DecoratedToken[] = []
    for (const token of tokens) {
        if (token.type === 'pattern') {
            if (previous === undefined) {
                previous = token
                continue
            }
            // Merge with previous
            previous.value = previous.value + token.value
            previous.range = { start: previous.range.start, end: token.range.end }
            continue
        }
        if (previous) {
            newTokens.push(previous)
            previous = undefined
        }
        newTokens.push(token)
    }
    if (previous) {
        newTokens.push(previous)
    }
    return newTokens
}

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
                        type: 'metaRegexp',
                        range: { start: offset + node.end, end: offset + node.end + 1 },
                        value: '|',
                        kind: MetaRegexpKind.Alternative,
                    })
                }
            },
            onAssertionEnter(node: Assertion) {
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.start, end: offset + node.end },
                    value: node.raw,
                    kind: MetaRegexpKind.Assertion,
                })
            },
            onGroupEnter(node: Group) {
                // Push the leading '('
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.start, end: offset + node.start + 1 },
                    value: '(',
                    kind: MetaRegexpKind.Delimited,
                })
                // Push the trailing ')'
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.end - 1, end: offset + node.end },
                    value: ')',
                    kind: MetaRegexpKind.Delimited,
                })
            },
            onCapturingGroupEnter(node: CapturingGroup) {
                // Push the leading '('
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.start, end: offset + node.start + 1 },
                    groupRange: { start: offset + node.start, end: offset + node.end },
                    value: '(',
                    kind: MetaRegexpKind.Delimited,
                })
                // Push the trailing ')'
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.end - 1, end: offset + node.end },
                    groupRange: { start: offset + node.start, end: offset + node.end },
                    value: ')',
                    kind: MetaRegexpKind.Delimited,
                })
            },
            onCharacterSetEnter(node: CharacterSet) {
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.start, end: offset + node.end },
                    value: node.raw,
                    kind: MetaRegexpKind.CharacterSet,
                })
            },
            onCharacterClassEnter(node: CharacterClass) {
                const negatedOffset = node.negate ? 1 : 0
                // Push the leading '['
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.start, end: offset + node.start + 1 + negatedOffset },
                    groupRange: { start: offset + node.start, end: offset + node.end },
                    value: node.negate ? '[^' : '[',
                    kind: MetaRegexpKind.CharacterClass,
                })
                // Push the trailing ']'
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.end - 1, end: offset + node.end },
                    groupRange: { start: offset + node.start, end: offset + node.end },
                    value: ']',
                    kind: MetaRegexpKind.CharacterClass,
                })
            },
            onCharacterClassRangeEnter(node: CharacterClassRange) {
                // highlight the '-' in [a-z]. Take care to use node.min.end, because we
                // don't want to highlight the first '-' in [--z], nor an escaped '-' with a
                // two-character offset as in [\--z].
                tokens.push({
                    type: 'metaRegexp',
                    range: { start: offset + node.min.end, end: offset + node.min.end + 1 },
                    value: '-',
                    kind: MetaRegexpKind.CharacterClass,
                })
            },
            onQuantifierEnter(node: Quantifier) {
                // the lazy quantifier ? adds one
                const lazyQuantifierOffset = node.greedy ? 0 : 1
                if (!node.greedy) {
                    tokens.push({
                        type: 'metaRegexp',
                        range: { start: offset + node.end - 1, end: offset + node.end },
                        value: '?',
                        kind: MetaRegexpKind.LazyQuantifier,
                    })
                }

                const quantifier = node.raw[node.raw.length - lazyQuantifierOffset - 1]
                if (quantifier === '+' || quantifier === '*' || quantifier === '?') {
                    tokens.push({
                        type: 'metaRegexp',
                        range: {
                            start: offset + node.end - 1 - lazyQuantifierOffset,
                            end: offset + node.end - lazyQuantifierOffset,
                        },
                        value: quantifier,
                        kind: MetaRegexpKind.RangeQuantifier,
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
                        type: 'metaRegexp',
                        range: { start: offset + openBrace, end: offset + node.end - lazyQuantifierOffset },
                        value: pattern.value.slice(openBrace, node.end - lazyQuantifierOffset),
                        kind: MetaRegexpKind.RangeQuantifier,
                    })
                }
            },
            onCharacterEnter(node: Character) {
                if (node.end - node.start > 1 && node.raw.startsWith('\\')) {
                    // This is an escaped value like `\.`, `\u0065`, `\x65`.
                    tokens.push({
                        type: 'metaRegexp',
                        range: { start: offset + node.start, end: offset + node.end },
                        value: node.raw,
                        kind: MetaRegexpKind.EscapedCharacter,
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
    return coalescePatterns(tokens)
}

const mapRevisionMeta = (token: Literal): DecoratedToken[] => {
    const offset = token.range.start

    const decorated: DecoratedToken[] = []
    let current = ''
    let start = 0
    let accumulator: string[] = []

    const nextChar = (): string => {
        current = token.value[start]
        start = start + 1
        return current
    }

    // Appends a decorated token to the list of tokens, and resets the current accumulator to be empty.
    const appendDecoratedToken = (endIndex: number): void => {
        const value = accumulator.join('')
        let kind
        switch (value) {
            case ':':
                kind = MetaRevisionKind.Separator
                break
            case '*':
                kind = MetaRevisionKind.Wildcard
                break
            case '!':
                kind = MetaRevisionKind.Negate
                break
            default:
                if (value.includes('/')) {
                    kind = MetaRevisionKind.PathLike
                } else if (value.match(/^[\dA-Fa-f]+$/) && value.length > 6) {
                    kind = MetaRevisionKind.CommitHash
                } else {
                    kind = MetaRevisionKind.Label
                }
        }
        const range = { start: offset + endIndex - value.length, end: offset + endIndex }
        decorated.push({ type: 'metaRevision', kind, value, range })
        accumulator = []
    }

    while (token.value[start] !== undefined) {
        current = nextChar()
        switch (current) {
            case ':':
            case '*':
            case '!':
                appendDecoratedToken(start - 1) // append up to this special character
                accumulator.push(current)
                appendDecoratedToken(start)
                break
            default:
                accumulator.push(current)
        }
    }
    appendDecoratedToken(start)
    return decorated
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
    const appendDecoratedToken = (endIndex: number, kind: PatternKind.Literal | MetaStructuralKind): void => {
        const value = token.join('')
        const range = { start: offset + endIndex - value.length, end: offset + endIndex }
        if (kind === PatternKind.Literal) {
            decorated.push({ type: 'pattern', kind, value, range })
        } else {
            decorated.push({ type: 'metaStructural', kind, value, range })
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
                    appendDecoratedToken(start - 3, MetaStructuralKind.Hole)
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
                    // Look ahead and see if this is the start of a hole.
                    if (pattern.value[start] === '[') {
                        // It is the start of a hole, consume the '['.
                        current = nextChar()
                        open = open + 1
                        // Persist the literal token scanned up to this point.
                        appendDecoratedToken(start - 2, PatternKind.Literal)
                        token.push(':[')
                        continue
                    }
                    // Something else, push the ':' we saw and continue.
                    token.push(':')
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
                    appendDecoratedToken(start, MetaStructuralKind.Hole)
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

const specifiesRevision = (value: string): boolean => value.match(/@/) !== null

const decorateRepoRevision = (token: Literal): DecoratedToken[] => {
    const [repo, revision] = token.value.split('@', 2)
    const offset = token.range.start

    return [
        ...decorate({
            type: 'pattern',
            kind: PatternKind.Regexp,
            value: repo,
            range: { start: offset, end: offset + repo.length },
        }),

        {
            type: 'metaRepoRevisionSeparator',
            value: '@',
            range: {
                start: offset + repo.length,
                end: offset + repo.length + 1,
            },
        },
        ...mapRevisionMeta({
            type: 'literal',
            value: revision,
            range: {
                start: token.range.start + repo.length + 1,
                end: token.range.start + repo.length + 1 + revision.length,
            },
        }),
    ]
}

export const decorate = (token: Token): DecoratedToken[] => {
    const decorated: DecoratedToken[] = []
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
                range: token.field.range,
                value: token.field.value,
            })
            if (
                token.value &&
                token.field.value.toLowerCase().match(/^-?(repo|r)$/i) &&
                token.value.type === 'literal' &&
                specifiesRevision(token.value.value)
            ) {
                decorated.push(...decorateRepoRevision(token.value))
            } else if (
                token.value &&
                token.field.value.toLowerCase().match(/rev|revision/i) &&
                token.value.type === 'literal'
            ) {
                decorated.push(
                    ...mapRevisionMeta({
                        type: 'literal',
                        value: token.value.value,
                        range: token.value.range,
                    })
                )
            } else if (token.value && token.value.type === 'literal' && hasRegexpValue(token.field.value)) {
                // Highlight fields with regexp values.
                decorated.push(
                    ...decorate({
                        type: 'pattern',
                        kind: PatternKind.Regexp,
                        value: token.value.value,
                        range: token.value.range,
                    })
                )
            } else if (token.value) {
                decorated.push(token.value)
            }
            break
        }
        default:
            decorated.push(token)
    }
    return decorated
}

const decoratedToMonaco = (token: DecoratedToken): Monaco.languages.IToken => {
    switch (token.type) {
        case 'field':
        case 'whitespace':
        case 'keyword':
        case 'comment':
        case 'openingParen':
        case 'closingParen':
        case 'metaRepoRevisionSeparator':
            return {
                startIndex: token.range.start,
                scopes: token.type,
            }
        case 'metaRevision':
        case 'metaRegexp':
        case 'metaStructural':
            // The scopes value is derived from the token type and its kind.
            // E.g., regexpMetaDelimited derives from {@link RegexpMeta} and {@link RegexpMetaKind}.
            return {
                startIndex: token.range.start,
                scopes: `${token.type}${token.kind}`,
            }

        default:
            return {
                startIndex: token.range.start,
                scopes: 'identifier',
            }
    }
}

const toMonaco = (token: Token): Monaco.languages.IToken[] => {
    switch (token.type) {
        case 'filter': {
            const monacoTokens: Monaco.languages.IToken[] = []
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
            return monacoTokens
        }
        case 'whitespace':
        case 'keyword':
        case 'comment':
            return [
                {
                    startIndex: token.range.start,
                    scopes: token.type,
                },
            ]
        default:
            return [
                {
                    startIndex: token.range.start,
                    scopes: 'identifier',
                },
            ]
    }
}

/**
 * Returns the tokens in a scanned search query displayed in the Monaco query input. If the experimental
 * decorate flag is true, a list of {@link DecoratedToken} provides more contextual highlighting for patterns.
 */
export const getMonacoTokens = (tokens: Token[], toDecorate = false): Monaco.languages.IToken[] =>
    toDecorate ? tokens.flatMap(token => decorate(token).map(decoratedToMonaco)) : tokens.flatMap(toMonaco)

/**
 * Converts a zero-indexed, single-line {@link CharacterRange} to a Monaco {@link IRange}.
 */
export const toMonacoRange = ({ start, end }: CharacterRange): Monaco.IRange => ({
    startLineNumber: 1,
    endLineNumber: 1,
    startColumn: start + 1,
    endColumn: end + 1,
})
