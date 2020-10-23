export interface Leaf {
    type: 'leaf'
    value: string
}

enum OperatorKind {
    Or = 'or',
    And = 'and',
}

/**
 * A nonterminal node for operators 'and' and 'or'.
 */
export interface Operator {
    type: 'operator'
    operands: Node[]
    kind: OperatorKind
}

export type Node = Leaf | Operator

interface ParseError {
    type: 'error'
    expected: string
}

export interface ParseSuccess {
    type: 'success'
    nodes: Node[]
}

export type ParseResult = ParseError | ParseSuccess

export const toString = (node: Node): string => {
    const result: string[] = []
    switch (node.type) {
        case 'operator':
            for (const operand of node.operands) {
                result.push(toString(operand))
            }
            switch (node.kind) {
                case OperatorKind.Or:
                    return `(or ${result.join(' ')})`
                case OperatorKind.And:
                    return `(and ${result.join(' ')})`
            }
        case 'leaf':
            return node.value
    }
}

interface State {
    result: ParseResult
    advance: number
}

const match = (input: string, value: string): boolean => input.startsWith(value)

const newNodes = (nodes: Node[]): ParseResult => ({
    type: 'success',
    nodes,
})

const newOperator = (nodes: Node[], kind: OperatorKind): ParseResult => ({
    type: 'success',
    nodes: [
        {
            type: 'operator',
            operands: nodes,
            kind,
        },
    ],
})

const newLeaf = (value: string): ParseResult => ({
    type: 'success',
    nodes: [
        {
            type: 'leaf',
            value,
        },
    ],
})

/**
 * parses tokens that are leaves in the tree. {@link State} returns the result of the parse, and how much the caller should advance input.
 */
export const scanLeaf = (input: string): State => {
    let position = 0
    let current = ''
    const result: string[] = []
    while (input.length > 0) {
        current = input[0]
        position++
        input = input.slice(1)
        if (current === ' ') {
            position-- // Backtrack.
            break
        }
        result.push(current)
    }
    return { result: newLeaf(result.join('')), advance: position }
}

/**
 * parses tokens that are leaves in the tree. {@link State} returns the result of the parse, and how much the caller should advance input.
 */
export const parseLeaves = (input: string): State => {
    const nodes: Node[] = []
    let position = 0
    while (true) {
        const current = input[0]
        if (current === undefined) {
            break
        }
        if (current === ' ') {
            input = input.slice(1)
            position++
        }
        if (match(input, OperatorKind.And) || match(input, OperatorKind.Or)) {
            return { result: newNodes(nodes), advance: position } // Caller advances.
        }
        const leaf = scanLeaf(input)
        input = input.slice(leaf.advance)
        position += leaf.advance
        if (leaf.result.type === 'error') {
            return { result: leaf.result, advance: position }
        }
        nodes.push(leaf.result.nodes[0])
    }
    return { result: newNodes(nodes), advance: position }
}

/**
 * parses and-expressions. {@link State} returns the result of the parse, and how much the caller should advance input.
 */
export const parseAnd = (input: string): State => {
    const left = parseLeaves(input)
    let position = 0
    input = input.slice(left.advance)
    position += left.advance
    if (left.result.type === 'error') {
        return { result: left.result, advance: position }
    }
    if (!match(input, OperatorKind.And)) {
        return { result: left.result, advance: position }
    }
    // Consume 'and'.
    input = input.slice(OperatorKind.And.length)
    position += OperatorKind.And.length

    const right = parseAnd(input)
    position += right.advance
    if (right.result.type === 'error') {
        return { result: right.result, advance: position }
    }
    return { result: newOperator([left.result.nodes[0], right.result.nodes[0]], OperatorKind.And), advance: position }
}

/**
 * parses or-expressions. Or-operators have lower precedence than and-operators, therefore this function calls parseAnd.
 * {@link State} returns the result of the parse, and how much the caller should advance input.
 */
export const parseOr = (input: string): State => {
    const left = parseAnd(input)
    let position = 0
    input = input.slice(left.advance)
    position += left.advance
    if (left.result.type === 'error') {
        return { result: left.result, advance: position }
    }
    if (!match(input, OperatorKind.Or)) {
        return { result: left.result, advance: position }
    }
    // Consume 'or'.
    input = input.slice(OperatorKind.Or.length)
    position += OperatorKind.Or.length

    const right = parseOr(input)
    position += right.advance
    if (right.result.type === 'error') {
        return { result: right.result, advance: position }
    }
    return { result: newOperator([left.result.nodes[0], right.result.nodes[0]], OperatorKind.Or), advance: position }
}

export const treeParse = (input: string): ParseResult => parseOr(input).result
