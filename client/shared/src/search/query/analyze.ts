import { isDefined } from '@sourcegraph/common'

import { type Node, OperatorKind } from './parser'
import { type CharacterRange, KeywordKind, type Token } from './token'

// Empty range to create valid nodes
const placeholderRange: CharacterRange = { start: 0, end: 0 }

/**
 * getRelevantTokens processes a given query in a top-down manner and removes any
 * patterns and filters that cannot affect the token at the target character range.
 * This function also accepts a filter callback to control which tokens should be
 * included (it's possible to filter the returned token list instead but that makes
 * handling operators more complicated).
 */
export function getRelevantTokens(query: Node, target: CharacterRange, filter: (node: Node) => boolean): Token[] {
    function processNode(node: Node): Node | null {
        switch (node.type) {
            case 'parameter':
            case 'pattern': {
                return filter(node) ? node : null
            }
            case 'sequence': {
                const nodes = node.nodes.map(processNode).filter(isDefined)
                return nodes.length > 0 ? { type: 'sequence', nodes, range: placeholderRange } : null
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
                                range: placeholderRange,
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
                            range: placeholderRange,
                        }
                    }
                }
            }
        }
    }

    return tokenize(processNode(query))
}

const operatorKindToKeywordKind: Record<OperatorKind, KeywordKind> = {
    [OperatorKind.Not]: KeywordKind.Not,
    [OperatorKind.Or]: KeywordKind.Or,
    [OperatorKind.And]: KeywordKind.And,
}

/**
 * Converts a parse node into a sequence of Token's
 */
function tokenize(node: Node | null): Token[] {
    switch (node?.type) {
        case undefined: {
            return []
        }
        case 'parameter': {
            return [
                {
                    type: 'filter',
                    field: { type: 'literal', value: node.field, quoted: false, range: placeholderRange },
                    value: { type: 'literal', value: node.value, quoted: node.quoted, range: placeholderRange },
                    negated: node.negated,
                    range: placeholderRange,
                },
            ]
        }
        case 'pattern': {
            return [node]
        }
        case 'sequence': {
            const tokens: Token[] = []
            for (const child of node.nodes) {
                if (tokens.length > 0) {
                    tokens.push({ type: 'whitespace', range: placeholderRange })
                }
                tokens.push(...tokenize(child))
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
                            value: 'NOT',
                            range: placeholderRange,
                        },
                        { type: 'whitespace', range: placeholderRange },
                        ...(node.right ? tokenize(node.right) : []),
                    ]
                }
                default: {
                    return [
                        { type: 'openingParen', range: placeholderRange },
                        ...(node.left ? tokenize(node.left) : []),
                        { type: 'whitespace', range: placeholderRange },
                        {
                            type: 'keyword',
                            kind: operatorKindToKeywordKind[node.kind],
                            value: node.kind,
                            range: placeholderRange,
                        },
                        { type: 'whitespace', range: placeholderRange },
                        ...(node.right ? tokenize(node.right) : []),
                        { type: 'closingParen', range: placeholderRange },
                    ]
                }
            }
        }
    }
}
