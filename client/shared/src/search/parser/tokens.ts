import * as Monaco from 'monaco-editor'
import { Token, Pattern, CharacterRange, PatternKind } from './scanner'
import { RegExpParser, visitRegExpAST } from 'regexpp'
import { Character, CharacterSet, CapturingGroup, Assertion, Quantifier } from 'regexpp/ast'

export enum RegexpMetaKind {
    Delimited = 'Delimited',
    CharacterSet = 'CharacterSet',
    Quantifier = 'Quantifier',
    Assertion = 'Assertion',
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

export type MetaToken = RegexpMeta | StructuralMeta

type DecoratedToken = Token | MetaToken

const mapRegexpMeta = (pattern: Pattern): DecoratedToken[] => {
    const tokens: DecoratedToken[] = []
    try {
        const ast = new RegExpParser().parsePattern(pattern.value)
        const offset = pattern.range.start
        visitRegExpAST(ast, {
            onAssertionEnter(node: Assertion) {
                tokens.push({
                    type: 'regexpMeta',
                    range: { start: offset + node.start, end: offset + node.end },
                    value: node.raw,
                    kind: RegexpMetaKind.Assertion,
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
            onQuantifierEnter(node: Quantifier) {
                // the lazy quantifier ? adds one
                const lazyQuantifierOffset = node.greedy ? 0 : 1
                const quantifier = node.raw[node.raw.length - lazyQuantifierOffset - 1]
                if (quantifier === '+' || quantifier === '*' || quantifier === '?') {
                    tokens.push({
                        type: 'regexpMeta',
                        range: { start: offset + node.end - 1 - lazyQuantifierOffset, end: offset + node.end },
                        value: node.raw,
                        kind: RegexpMetaKind.Quantifier,
                    })
                } else {
                    // regexpp provides no easy way to tell whether the quantifier is a range '{number, number}'.
                    // At this point we know it is none of +, *, or ?, so it is a ranged quantifer.
                    // We skip highlighting for now; it's trickier.
                    tokens.push({
                        type: 'pattern',
                        range: { start: offset + node.start, end: offset + node.end },
                        value: node.raw,
                        kind: PatternKind.Regexp,
                    })
                }
            },
            onCharacterEnter(node: Character) {
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

const mapStructuralMeta = (pattern: Pattern): DecoratedToken[] => [pattern]

const decorateTokens = (tokens: Token[]): DecoratedToken[] => {
    const decorated: DecoratedToken[] = []
    for (const token of tokens) {
        if (token.type === 'pattern') {
            switch (token.kind) {
                case PatternKind.Regexp:
                    decorated.push(...mapRegexpMeta(token))
                    break
                case PatternKind.Structural:
                    decorated.push(...mapStructuralMeta(token))
                    break
                default:
                    decorated.push(token)
            }
            continue
        }
        decorated.push(token)
    }
    return decorated
}

const fromDecoratedTokens = (tokens: DecoratedToken[]): Monaco.languages.IToken[] => {
    const monacoTokens: Monaco.languages.IToken[] = []
    for (const token of tokens) {
        switch (token.type) {
            case 'filter':
                {
                    monacoTokens.push({
                        startIndex: token.filterType.range.start,
                        scopes: 'filterKeyword',
                    })
                    if (token.filterValue) {
                        monacoTokens.push({
                            startIndex: token.filterValue.range.start,
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
            case 'regexpMeta':
            case 'structuralMeta':
                /** The scopes value is derived from the token type and its kind.
                 * E.g., regexpMetaDelimited dervies from {@link RegexpMeta} and {@link RegexpMetaKind}.
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
                        startIndex: token.filterType.range.start,
                        scopes: 'filterKeyword',
                    })
                    if (token.filterValue) {
                        monacoTokens.push({
                            startIndex: token.filterValue.range.start,
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
