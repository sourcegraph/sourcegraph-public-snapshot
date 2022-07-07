import { scanSearchQuery } from './scanner'
import { PatternKind, Token, KeywordKind, CharacterRange } from './token'

export interface Pattern {
    type: 'pattern'
    kind: PatternKind
    value: string
    quoted: boolean
    negated: boolean
    range: CharacterRange
}

export interface Parameter {
    type: 'parameter'
    field: string
    value: string
    negated: boolean
    range: CharacterRange
}

export enum OperatorKind {
    Or = 'OR',
    And = 'AND',
}

/**
 * A nonterminal node for operators AND and OR.
 */
export interface Operator {
    type: 'operator'
    operands: Node[]
    kind: OperatorKind
    range: CharacterRange
    // Position in the query string including parenthesis if used
    groupRange?: CharacterRange
}

export type Node = Operator | Parameter | Pattern

interface ParseError {
    type: 'error'
    expected: string
}

export interface ParseSuccess {
    type: 'success'
    nodes: Node[]
}

export type ParseResult = ParseError | ParseSuccess

/**
 * State contains the current parse result, and the remaining tokens to parse.
 */
interface State {
    result: ParseResult
    tokens: Token[]
}

const createNodes = (nodes: Node[]): ParseSuccess => ({ type: 'success', nodes })

const createPattern = (
    value: string,
    kind: PatternKind,
    quoted: boolean,
    negated: boolean,
    range: CharacterRange
): ParseSuccess => createNodes([{ type: 'pattern', kind, value, quoted, negated, range }])

const createParameter = (field: string, value: string, negated: boolean, range: CharacterRange): ParseSuccess =>
    createNodes([{ type: 'parameter', field, value, negated, range }])

const createOperator = (nodes: Node[], kind: OperatorKind): ParseSuccess => {
    const range: CharacterRange = nodes.reduce(
        (range, node) => {
            const nodeRange = node.type === 'operator' ? node.groupRange ?? node.range : node.range
            if (nodeRange.start < range.start) {
                range.start = nodeRange.start
            }
            if (nodeRange.end > range.end) {
                range.end = nodeRange.end
            }
            return range
        },
        { start: Infinity, end: -Infinity }
    )

    return createNodes([{ type: 'operator', operands: nodes, kind, range }])
}

const tokenToLeafNode = (token: Token): ParseResult => {
    if (token.type === 'pattern') {
        return createPattern(token.value, token.kind, false, false, token.range)
    }
    if (token.type === 'filter') {
        const filterValue = token.value ? token.value.value : ''
        return createParameter(token.field.value, filterValue, token.negated, token.range)
    }
    return { type: 'error', expected: 'a convertable token to tree node' }
}

export const parseParenthesis = (tokens: Token[]): State => {
    const openingParen = tokens[0]
    tokens = tokens.slice(1) // Consume '('.

    const groupNodes = parseOr(tokens)
    if (groupNodes.result.type === 'error') {
        return { result: groupNodes.result, tokens }
    }
    if (groupNodes.tokens[0]?.type !== 'closingParen') {
        return { result: { type: 'error', expected: 'no unbalanced parentheses' }, tokens }
    }

    const closingParen = groupNodes.tokens[0]
    tokens = groupNodes.tokens.slice(1) // Consume )

    if (groupNodes.result.nodes[0].type === 'operator') {
        groupNodes.result.nodes[0].groupRange = {
            start: openingParen.range.start,
            end: closingParen.range.end,
        }
    }
    return { result: groupNodes.result, tokens }
}

export const parseLeaves = (tokens: Token[]): State => {
    const nodes: Node[] = []
    while (true) {
        const current = tokens[0]
        if (current === undefined) {
            break
        }
        if (current.type === 'openingParen') {
            return parseParenthesis(tokens)
        }
        if (current.type === 'closingParen') {
            break
        }
        if (current.type === 'keyword' && (current.kind === KeywordKind.And || current.kind === KeywordKind.Or)) {
            return { result: createNodes(nodes), tokens } // Caller advances.
        }

        const node = tokenToLeafNode(current)
        if (node.type === 'error') {
            return { result: node, tokens }
        }
        nodes.push(...node.nodes)
        tokens = tokens.slice(1)
    }
    return { result: createNodes(nodes), tokens }
}

/**
 * parseAnd parses and-expressions. And operators bind tighter:
 * (a and b or c) => ((a and b) or c).
 */
export const parseAnd = (tokens: Token[]): State => {
    const left = parseLeaves(tokens)
    if (left.result.type === 'error') {
        return { result: left.result, tokens }
    }
    if (left.tokens[0] === undefined) {
        return { result: left.result, tokens: [] }
    }
    if (!(left.tokens[0].type === 'keyword' && left.tokens[0].kind === KeywordKind.And)) {
        return { result: left.result, tokens: left.tokens }
    }
    tokens = left.tokens.slice(1) // Consume AND token.
    const right = parseAnd(tokens)
    if (right.result.type === 'error') {
        return { result: right.result, tokens }
    }
    return {
        result: createOperator(left.result.nodes.concat(...right.result.nodes), OperatorKind.And),
        tokens: right.tokens,
    }
}

/**
 * parseOr parses or-expressions. Or operators have lower precedence than And
 * operators, therefore this function calls parseAnd.
 */
export const parseOr = (tokens: Token[]): State => {
    const left = parseAnd(tokens)
    if (left.result.type === 'error') {
        return { result: left.result, tokens }
    }
    if (left.tokens[0] === undefined) {
        return { result: left.result, tokens: [] }
    }
    if (!(left.tokens[0].type === 'keyword' && left.tokens[0].kind === KeywordKind.Or)) {
        return { result: left.result, tokens: left.tokens }
    }
    tokens = left.tokens.slice(1) // Consume OR token.
    const right = parseOr(tokens)
    if (right.result.type === 'error') {
        return { result: right.result, tokens }
    }
    return {
        result: createOperator(left.result.nodes.concat(...right.result.nodes), OperatorKind.Or),
        tokens: right.tokens,
    }
}

/**
 * Produces a parse tree from a search query.
 */
export const parseSearchQuery = (input: string): ParseResult => {
    const result = scanSearchQuery(input)
    if (result.type === 'error') {
        return {
            type: 'error',
            expected: result.expected,
        }
    }
    // Scanner can produce empty pattern tokens in some locations which break
    // the parser. Those need to be filtered out.
    // See https://github.com/sourcegraph/sourcegraph/issues/38384
    return parseOr(
        result.term.filter(token => token.type !== 'whitespace' && !(token.type === 'pattern' && token.value === ''))
    ).result
}
