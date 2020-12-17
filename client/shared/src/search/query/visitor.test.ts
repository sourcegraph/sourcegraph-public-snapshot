import { visit, Visitors } from './visitor'
import { parseSearchQuery, ParseSuccess, Node, OperatorKind } from './parser'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

/**
 * A function that defines a simple visitor to collect visited nodes
 * in a query string, and returns the string result.
 *
 * @param input The input query string.
 */
const collect = (input: string): Node[] => {
    const nodes: Node[] = []
    const visitors: Visitors = {
        visitOperator(operands: Node[], kind: OperatorKind) {
            nodes.push({ type: 'operator', operands, kind })
        },
        visitParameter(field, value, negated) {
            nodes.push({ type: 'parameter', field, value, negated })
        },
        visitPattern(value, kind, negated, quoted) {
            nodes.push({ type: 'pattern', value, kind, negated, quoted })
        },
    }
    visit((parseSearchQuery(input) as ParseSuccess).nodes, visitors)
    return nodes
}

describe('visit()', () => {
    test('basic visit', () => {
        expect(collect('repo:foo pattern-bar or file:baz')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "parameter",
                    "field": "repo",
                    "value": "foo",
                    "negated": false
                  },
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "pattern-bar",
                    "quoted": false,
                    "negated": false
                  },
                  {
                    "type": "parameter",
                    "field": "file",
                    "value": "baz",
                    "negated": false
                  }
                ],
                "kind": "OR"
              },
              {
                "type": "parameter",
                "field": "repo",
                "value": "foo",
                "negated": false
              },
              {
                "type": "pattern",
                "value": "pattern-bar",
                "kind": 1,
                "negated": false,
                "quoted": false
              },
              {
                "type": "parameter",
                "field": "file",
                "value": "baz",
                "negated": false
              }
            ]
        `)
    })
})
