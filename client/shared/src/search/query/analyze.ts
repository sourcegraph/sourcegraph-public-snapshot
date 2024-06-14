import { isDefined } from '@sourcegraph/common'

import { type Node, OperatorKind, Parameter } from './parser'
import { type CharacterRange, KeywordKind, type Token, PatternKind, Pattern } from './token'

interface SyntheticSequence {
    type: 'sequence'
    nodes: SyntheticNode[]
}
interface SyntheticOperator {
    type: 'operator'
    kind: OperatorKind
    left: SyntheticNode | null
    right: SyntheticNode | null
}

type SyntheticNode = SyntheticSequence | SyntheticOperator | Parameter | Pattern

export interface RelevantTokenResult {
    tokens: Token[]
    sourceMap: Map<Token, CharacterRange>
}

export const EMPTY_RELEVANT_TOKEN_RESULT: RelevantTokenResult = { tokens: [], sourceMap: new Map() }

/**
 * getRelevantTokens processes a given query in a top-down manner and removes any
 * patterns and filters that cannot affect the token at the target character range.
 * This function also accepts a filter callback to control which tokens should be
 * included (it's possible to filter the returned token list instead but that makes
 * handling operators more complicated).
 *
 * The function returns a list of tokens that are relevant to the target range and a
 * source map that maps a token to its range in the original query. Not all tokens
 * will have source range because getRelevantTokens generates new tokens as needed.
 *
 * @param query The query AST to process
 * @param target The position for which to find relevant tokens
 * @param filter A filter function to control which tokens should be included
 * @returns A list of relevant tokens and a source map
 */
export function getRelevantTokens(
    query: Node,
    target: CharacterRange,
    filter: (node: Node) => boolean
): RelevantTokenResult {
    function processNode(node: Node): SyntheticNode | null {
        switch (node.type) {
            case 'parameter':
            case 'pattern': {
                return filter(node) ? node : null
            }
            case 'sequence': {
                const nodes = node.nodes.map(processNode).filter(isDefined)
                return nodes.length > 0 ? { type: 'sequence', nodes } : null
            }
            case 'operator': {
                switch (node.kind) {
                    case OperatorKind.Or: {
                        // If one operand contains the target branche we only
                        // need to keep that operand (the other branch is
                        // irrelevant). But if no operand contains the target
                        // range we need to process all nodes and assume that
                        // this token is ANDed at some level with the target
                        // range.
                        //
                        // Examples:
                        //
                        // filter:a filter:b OR filter:|
                        // ^^^^^^^^^^^^^^^^^
                        //      discard
                        //
                        // (filter:a or filter:b) filter:|
                        // ^^^^^^^^^^^^^^^^^^^^^^
                        // needs to be preserved
                        const operand = [node.left, node.right].find(
                            node => node && node.range.start <= target.start && node.range.end >= target.end
                        )

                        if (operand) {
                            return processNode(operand)
                        }
                        // NOTE: Intentional fallthrough since the logic is the
                        // same.
                    }
                    case OperatorKind.And: {
                        const left = node.left && processNode(node.left)
                        const right = node.right && processNode(node.right)
                        if (left && right) {
                            return {
                                type: 'operator',
                                // needs to be node.kind to properly handle
                                // fallthrough case.
                                kind: node.kind,
                                left,
                                right,
                            }
                        }
                        return left || right
                    }
                    case OperatorKind.Not: {
                        if (!node.right) {
                            return null
                        }
                        const operand = processNode(node.right)
                        if (!operand) {
                            return null
                        }
                        return {
                            type: 'operator',
                            kind: node.kind,
                            left: null,
                            right: operand,
                        }
                    }
                }
            }
        }
    }

    const sourceMap = new Map<Token, CharacterRange>()
    const tokens = alignTokenRanges(tokenize(processNode(query), { sourceMap }))
    return { tokens, sourceMap }
}

const operatorKindToKeywordKind: Record<OperatorKind, KeywordKind> = {
    [OperatorKind.Not]: KeywordKind.Not,
    [OperatorKind.Or]: KeywordKind.Or,
    [OperatorKind.And]: KeywordKind.And,
}

/**
 * Converts a parse node into a sequence of Token's. This function generates
 * new tokens as needed to represent the parse tree in a flat list.
 * The returned tokens have relative ranges that need to be aligned with the
 * {@link alignTokenRanges} function.
 *
 * @param node The parse node to convert
 * @param context A context object to store the source map
 * @returns A list of tokens representing the parse node
 */
function tokenize(node: SyntheticNode | null, context: { sourceMap: Map<Token, CharacterRange> }): Token[] {
    switch (node?.type) {
        case undefined: {
            return []
        }
        case 'parameter': {
            const fieldStart = node.negated ? 1 : 0
            const fieldEnd = fieldStart + node.field.length

            const field: Token = {
                type: 'literal',
                value: node.field,
                quoted: false,
                range: {
                    start: fieldStart,
                    end: fieldEnd,
                },
            }
            // + 1 due to ':' between field and value
            const valueStart = fieldEnd + 1
            const valueEnd = valueStart + node.value.length + (node.quoted ? 2 : 0)
            const value: Token = {
                type: 'literal',
                value: node.value,
                quoted: node.quoted,
                range: { start: valueStart, end: valueEnd },
            }

            const filter: Token = {
                type: 'filter',
                field,
                value,
                negated: node.negated,
                range: { start: 0, end: valueEnd },
            }

            context.sourceMap.set(field, { start: node.range.start + fieldStart, end: node.range.start + fieldEnd })
            context.sourceMap.set(value, { start: node.range.start + valueStart, end: node.range.start + valueEnd })
            context.sourceMap.set(filter, node.range)
            return [filter]
        }
        case 'pattern': {
            const delimited = node.kind === PatternKind.Regexp
            const pattern: Token = {
                ...node,
                delimited,
                range: {
                    start: 0,
                    end: node.value.length + (delimited ? 2 : 0),
                },
            }

            context.sourceMap.set(pattern, node.range)

            return [pattern]
        }
        case 'sequence': {
            const tokens: Token[] = []
            for (const child of node.nodes) {
                if (tokens.length > 0) {
                    tokens.push({ type: 'whitespace', range: { start: 0, end: 1 } })
                }
                tokens.push(...tokenize(child, context))
            }
            return tokens
        }
        case 'operator': {
            switch (node.kind) {
                case OperatorKind.Not: {
                    return [
                        {
                            type: 'keyword',
                            kind: operatorKindToKeywordKind[node.kind],
                            value: node.kind,
                            range: { start: 0, end: 3 },
                        },
                        { type: 'whitespace', range: { start: 0, end: 1 } },
                        ...(node.right ? tokenize(node.right, context) : []),
                    ]
                }
                default: {
                    return [
                        { type: 'openingParen', range: { start: 0, end: 1 } },
                        ...(node.left ? tokenize(node.left, context) : []),
                        { type: 'whitespace', range: { start: 0, end: 1 } },
                        {
                            type: 'keyword',
                            kind: operatorKindToKeywordKind[node.kind],
                            value: node.kind,
                            range: { start: 0, end: node.kind.length },
                        },
                        { type: 'whitespace', range: { start: 0, end: 1 } },
                        ...(node.right ? tokenize(node.right, context) : []),
                        { type: 'closingParen', range: { start: 0, end: 1 } },
                    ]
                }
            }
        }
    }
}

/**
 * Mutates the range of the tokens in place to align them relative to each other.
 * Returns the same tokens for convenience.
 *
 * @param tokens The tokens to shift
 * @returns The passed in tokens
 */
function alignTokenRanges(tokens: Token[]): Token[] {
    let position = 0
    for (const token of tokens) {
        const shift = position - token.range.start

        switch (token.type) {
            case 'filter': {
                token.field.range.start += shift
                token.field.range.end += shift
                if (token.value) {
                    token.value.range.start += shift
                    token.value.range.end += shift
                }
                break
            }
        }

        token.range.start += shift
        token.range.end += shift
        position = token.range.end
    }

    return tokens
}
