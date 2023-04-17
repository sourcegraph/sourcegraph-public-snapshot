import { parseSearchQuery, ParseSuccess, Node, OperatorKind } from './parser'
import { visit, Visitors } from './visitor'

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
    const visitedNodes: Node[] = []
    const visitors: Visitors = {
        visitOperator(left, right, kind: OperatorKind, range, groupRange) {
            visitedNodes.push({ type: 'operator', left, right, kind, range, groupRange })
        },
        visitSequence(nodes: Node[], range) {
            visitedNodes.push({ type: 'sequence', nodes, range })
        },
        visitParameter(field, value, negated, quoted, range) {
            visitedNodes.push({ type: 'parameter', field, value, negated, quoted, range })
        },
        visitPattern(value, kind, range) {
            visitedNodes.push({ type: 'pattern', value, kind, range })
        },
    }
    visit([(parseSearchQuery(input) as ParseSuccess).node], visitors)
    return visitedNodes
}

describe('visit()', () => {
    test('basic visit', () => {
        expect(collect('repo:foo pattern-bar or file:baz')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "left": {
                  "type": "sequence",
                  "nodes": [
                    {
                      "type": "parameter",
                      "field": "repo",
                      "value": "foo",
                      "quoted": false,
                      "negated": false,
                      "range": {
                        "start": 0,
                        "end": 8
                      }
                    },
                    {
                      "type": "pattern",
                      "kind": 1,
                      "value": "pattern-bar",
                      "range": {
                        "start": 9,
                        "end": 20
                      }
                    }
                  ],
                  "range": {
                    "start": 0,
                    "end": 20
                  }
                },
                "right": {
                  "type": "parameter",
                  "field": "file",
                  "value": "baz",
                  "quoted": false,
                  "negated": false,
                  "range": {
                    "start": 24,
                    "end": 32
                  }
                },
                "kind": "OR",
                "range": {
                  "start": 0,
                  "end": 32
                }
              },
              {
                "type": "sequence",
                "nodes": [
                  {
                    "type": "parameter",
                    "field": "repo",
                    "value": "foo",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 0,
                      "end": 8
                    }
                  },
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "pattern-bar",
                    "range": {
                      "start": 9,
                      "end": 20
                    }
                  }
                ],
                "range": {
                  "start": 0,
                  "end": 20
                }
              },
              {
                "type": "parameter",
                "field": "repo",
                "value": "foo",
                "negated": false,
                "quoted": false,
                "range": {
                  "start": 0,
                  "end": 8
                }
              },
              {
                "type": "pattern",
                "value": "pattern-bar",
                "kind": 1,
                "range": {
                  "start": 9,
                  "end": 20
                }
              },
              {
                "type": "parameter",
                "field": "file",
                "value": "baz",
                "negated": false,
                "quoted": false,
                "range": {
                  "start": 24,
                  "end": 32
                }
              }
            ]
        `)
    })
})
