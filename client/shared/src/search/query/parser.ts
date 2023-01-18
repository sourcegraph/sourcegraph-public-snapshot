import { ScanResult, scanSearchQuery } from './scanner'
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

/**
 * A Sequence represent a sequence of nodes, i.e. 'a b c'. While such as
 * sequence is often thought about as "implicit AND", it's usually _not_
 * equivalent to 'a AND b AND c', which is why this gets its own node type.
 */
export interface Sequence {
    type: 'sequence'
    nodes: Node[]
    range: CharacterRange
}

export enum OperatorKind {
    Or = 'OR',
    And = 'AND',
    Not = 'NOT',
}

/**
 * A nonterminal node for operators AND and OR.
 */
export interface Operator {
    type: 'operator'
    operands: Node[]
    kind: OperatorKind
    range: CharacterRange
    /**
     * Position in the query string including parenthesis if used
     */
    groupRange?: CharacterRange
}

export type Node = Sequence | Operator | Parameter | Pattern

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

const rangeFromNodes = (nodes: Node[]): CharacterRange =>
    nodes.reduce(
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

const createNodes = (nodes: Node[]): ParseSuccess => ({ type: 'success', nodes })

const createPattern = (
    value: string,
    kind: PatternKind,
    quoted: boolean,
    negated: boolean,
    range: CharacterRange
): ParseSuccess => createNodes([{ type: 'pattern', kind, value, quoted, negated, range }])

const createSequence = (nodes: Node[]): ParseSuccess => {
    switch (nodes.length) {
        case 0:
        case 1:
            return createNodes(nodes)
        default:
            return createNodes([{ type: 'sequence', nodes, range: rangeFromNodes(nodes) }])
    }
}

const createParameter = (field: string, value: string, negated: boolean, range: CharacterRange): ParseSuccess =>
    createNodes([{ type: 'parameter', field, value, negated, range }])

const createOperator = (nodes: Node[], kind: OperatorKind, rangeStart?: number): ParseSuccess => {
    const range = rangeFromNodes(nodes)
    if (rangeStart !== undefined) {
        range.start = rangeStart
    }
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

const parseNot = (tokens: Token[]): State => {
    const keyword = tokens[0]

    if (!(keyword.type === 'keyword' && keyword.kind === KeywordKind.Not)) {
        throw new Error('parseNot is called at an invalid token position')
    }

    tokens = tokens.slice(1) // consume NOT

    let nodes: Node[] = []
    const operand = tokens[0]

    if (!operand) {
        return { result: createOperator(nodes, OperatorKind.Not), tokens }
    }

    switch (operand.type) {
        case 'openingParen': {
            const state = parseParenthesis(tokens)
            if (state.result.type === 'error') {
                return { result: state.result, tokens }
            }
            nodes = state.result.nodes
            tokens = state.tokens
            break
        }
        default: {
            const node = tokenToLeafNode(operand)
            if (node.type === 'error') {
                return { result: node, tokens }
            }
            nodes = node.nodes
            tokens = tokens.slice(1)
        }
    }

    return { result: createOperator(nodes, OperatorKind.Not, keyword.range.start), tokens }
}

/**
 * parseSequence parses consecutive tokens. If the sequence has a size smaller
 * than 2 the nodes are returned directly. That means there will never be a
 * sequence of size 1.
 * This also takes care of parsing the NOT keyword
 */
const parseSequence = (tokens: Token[]): State => {
    const nodes: Node[] = []
    while (true) {
        const current = tokens[0]
        if (current === undefined) {
            break
        }
        if (current.type === 'openingParen') {
            const state = parseParenthesis(tokens)
            if (state.result.type === 'error') {
                return { result: state.result, tokens: state.tokens }
            }
            nodes.push(...state.result.nodes)
            tokens = state.tokens
            continue
        }
        if (current.type === 'closingParen') {
            break
        }
        if (current.type === 'keyword') {
            if (current.kind === KeywordKind.And || current.kind === KeywordKind.Or) {
                break // Caller advances.
            }
            if (current.kind === KeywordKind.Not) {
                const state = parseNot(tokens)
                if (state.result.type === 'error') {
                    return { result: state.result, tokens: state.tokens }
                }
                nodes.push(...state.result.nodes)
                tokens = state.tokens
                continue
            }
        }

        const node = tokenToLeafNode(current)
        if (node.type === 'error') {
            return { result: node, tokens }
        }
        nodes.push(...node.nodes)
        tokens = tokens.slice(1)
    }
    return { result: createSequence(nodes), tokens }
}

/**
 * parseAnd parses and-expressions. And operators bind tighter:
 * (a and b or c) => ((a and b) or c).
 */
export const parseAnd = (tokens: Token[]): State => {
    const left = parseSequence(tokens)
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
export const parseSearchQuery = (input: string | ScanResult<Token[]>): ParseResult => {
    const result = typeof input === 'string' ? scanSearchQuery(input) : input
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
