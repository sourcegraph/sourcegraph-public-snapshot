import { treeParse, toString, Node } from './treeParser'

export const treeParseSuccess = (input: string): Node[] => {
    const result = treeParse(input)
    if (result.type === 'success') {
        return result.nodes
    }
    return []
}

export const prettyPrint = (nodes: Node[]): string => {
    const result: string[] = []
    for (const node of nodes) {
        result.push(toString(node))
    }
    return result.join(' ')
}

describe('treeParse', () => {
    test('and basic', () => expect(prettyPrint(treeParseSuccess('a and b'))).toMatch('(and a b)'))
    test('or nesting', () =>
        expect(prettyPrint(treeParseSuccess('a or b or c or d'))).toMatch('(or a (or b (or c d)))'))
    test('or precedence', () =>
        expect(prettyPrint(treeParseSuccess('a and b or c and d'))).toMatch('(or (and a b) (and c d)'))
})
