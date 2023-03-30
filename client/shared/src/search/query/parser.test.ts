import { parseSearchQuery, Node, ParseSuccess } from './parser'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, '  '),
    test: () => true,
})

export const parse = (input: string): Node => (parseSearchQuery(input) as ParseSuccess).node

describe('parseSearchQuery', () => {
    test('query with leaves', () =>
        expect(parse('repo:foo a b c')).toMatchInlineSnapshot(`
            {
              "type": "sequence",
              "nodes": [
                {
                  "type": "filter",
                  "range": {
                    "start": 0,
                    "end": 8
                  },
                  "field": {
                    "type": "literal",
                    "value": "repo",
                    "range": {
                      "start": 0,
                      "end": 4
                    }
                  },
                  "value": {
                    "type": "literal",
                    "value": "foo",
                    "range": {
                      "start": 5,
                      "end": 8
                    },
                    "quoted": false
                  },
                  "negated": false
                },
                {
                  "type": "pattern",
                  "range": {
                    "start": 9,
                    "end": 10
                  },
                  "kind": 1,
                  "value": "a",
                  "delimited": false
                },
                {
                  "type": "pattern",
                  "range": {
                    "start": 11,
                    "end": 12
                  },
                  "kind": 1,
                  "value": "b",
                  "delimited": false
                },
                {
                  "type": "pattern",
                  "range": {
                    "start": 13,
                    "end": 14
                  },
                  "kind": 1,
                  "value": "c",
                  "delimited": false
                }
              ],
              "range": {
                "start": 0,
                "end": 14
              }
            }
        `))

    test('query with and', () =>
        expect(parse('a b and c')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "sequence",
                "nodes": [
                  {
                    "type": "pattern",
                    "range": {
                      "start": 0,
                      "end": 1
                    },
                    "kind": 1,
                    "value": "a",
                    "delimited": false
                  },
                  {
                    "type": "pattern",
                    "range": {
                      "start": 2,
                      "end": 3
                    },
                    "kind": 1,
                    "value": "b",
                    "delimited": false
                  }
                ],
                "range": {
                  "start": 0,
                  "end": 3
                }
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 8,
                  "end": 9
                },
                "kind": 1,
                "value": "c",
                "delimited": false
              },
              "kind": "AND",
              "range": {
                "start": 0,
                "end": 9
              }
            }
        `))

    test('query with or', () =>
        expect(parse('a or b')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "pattern",
                "range": {
                  "start": 0,
                  "end": 1
                },
                "kind": 1,
                "value": "a",
                "delimited": false
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 5,
                  "end": 6
                },
                "kind": 1,
                "value": "b",
                "delimited": false
              },
              "kind": "OR",
              "range": {
                "start": 0,
                "end": 6
              }
            }
        `))

    test('query with not', () =>
        expect(parse('not a')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": null,
              "right": {
                "type": "pattern",
                "range": {
                  "start": 4,
                  "end": 5
                },
                "kind": 1,
                "value": "a",
                "delimited": false
              },
              "kind": "NOT",
              "range": {
                "start": 0,
                "end": 5
              }
            }
        `))

    test('query with and/or operator precedence', () =>
        expect(parse('a or b and not c')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "pattern",
                "range": {
                  "start": 0,
                  "end": 1
                },
                "kind": 1,
                "value": "a",
                "delimited": false
              },
              "right": {
                "type": "operator",
                "left": {
                  "type": "pattern",
                  "range": {
                    "start": 5,
                    "end": 6
                  },
                  "kind": 1,
                  "value": "b",
                  "delimited": false
                },
                "right": {
                  "type": "operator",
                  "left": null,
                  "right": {
                    "type": "pattern",
                    "range": {
                      "start": 15,
                      "end": 16
                    },
                    "kind": 1,
                    "value": "c",
                    "delimited": false
                  },
                  "kind": "NOT",
                  "range": {
                    "start": 11,
                    "end": 16
                  }
                },
                "kind": "AND",
                "range": {
                  "start": 5,
                  "end": 16
                }
              },
              "kind": "OR",
              "range": {
                "start": 0,
                "end": 16
              }
            }
        `))

    test('query with parentheses that override precedence', () => {
        expect(parse('a and (b or c)')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "pattern",
                "range": {
                  "start": 0,
                  "end": 1
                },
                "kind": 1,
                "value": "a",
                "delimited": false
              },
              "right": {
                "type": "operator",
                "left": {
                  "type": "pattern",
                  "range": {
                    "start": 7,
                    "end": 8
                  },
                  "kind": 1,
                  "value": "b",
                  "delimited": false
                },
                "right": {
                  "type": "pattern",
                  "range": {
                    "start": 12,
                    "end": 13
                  },
                  "kind": 1,
                  "value": "c",
                  "delimited": false
                },
                "kind": "OR",
                "range": {
                  "start": 7,
                  "end": 13
                },
                "groupRange": {
                  "start": 6,
                  "end": 14
                }
              },
              "kind": "AND",
              "range": {
                "start": 0,
                "end": 14
              }
            }
        `)

        expect(parse('(a or b) and c')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "operator",
                "left": {
                  "type": "pattern",
                  "range": {
                    "start": 1,
                    "end": 2
                  },
                  "kind": 1,
                  "value": "a",
                  "delimited": false
                },
                "right": {
                  "type": "pattern",
                  "range": {
                    "start": 6,
                    "end": 7
                  },
                  "kind": 1,
                  "value": "b",
                  "delimited": false
                },
                "kind": "OR",
                "range": {
                  "start": 1,
                  "end": 7
                },
                "groupRange": {
                  "start": 0,
                  "end": 8
                }
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 13,
                  "end": 14
                },
                "kind": 1,
                "value": "c",
                "delimited": false
              },
              "kind": "AND",
              "range": {
                "start": 0,
                "end": 14
              }
            }
        `)

        expect(parse('not (a or b) and c')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "operator",
                "left": null,
                "right": {
                  "type": "operator",
                  "left": {
                    "type": "pattern",
                    "range": {
                      "start": 5,
                      "end": 6
                    },
                    "kind": 1,
                    "value": "a",
                    "delimited": false
                  },
                  "right": {
                    "type": "pattern",
                    "range": {
                      "start": 10,
                      "end": 11
                    },
                    "kind": 1,
                    "value": "b",
                    "delimited": false
                  },
                  "kind": "OR",
                  "range": {
                    "start": 5,
                    "end": 11
                  },
                  "groupRange": {
                    "start": 4,
                    "end": 12
                  }
                },
                "kind": "NOT",
                "range": {
                  "start": 0,
                  "end": 12
                }
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 17,
                  "end": 18
                },
                "kind": 1,
                "value": "c",
                "delimited": false
              },
              "kind": "AND",
              "range": {
                "start": 0,
                "end": 18
              }
            }
        `)
        expect(parse('not (a and b) or c')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "operator",
                "left": null,
                "right": {
                  "type": "operator",
                  "left": {
                    "type": "pattern",
                    "range": {
                      "start": 5,
                      "end": 6
                    },
                    "kind": 1,
                    "value": "a",
                    "delimited": false
                  },
                  "right": {
                    "type": "pattern",
                    "range": {
                      "start": 11,
                      "end": 12
                    },
                    "kind": 1,
                    "value": "b",
                    "delimited": false
                  },
                  "kind": "AND",
                  "range": {
                    "start": 5,
                    "end": 12
                  },
                  "groupRange": {
                    "start": 4,
                    "end": 13
                  }
                },
                "kind": "NOT",
                "range": {
                  "start": 0,
                  "end": 13
                }
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 17,
                  "end": 18
                },
                "kind": 1,
                "value": "c",
                "delimited": false
              },
              "kind": "OR",
              "range": {
                "start": 0,
                "end": 18
              }
            }
        `)
    })

    test('query with nested parentheses', () =>
        expect(parse('(a and (b or c))')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "pattern",
                "range": {
                  "start": 1,
                  "end": 2
                },
                "kind": 1,
                "value": "a",
                "delimited": false
              },
              "right": {
                "type": "operator",
                "left": {
                  "type": "pattern",
                  "range": {
                    "start": 8,
                    "end": 9
                  },
                  "kind": 1,
                  "value": "b",
                  "delimited": false
                },
                "right": {
                  "type": "pattern",
                  "range": {
                    "start": 13,
                    "end": 14
                  },
                  "kind": 1,
                  "value": "c",
                  "delimited": false
                },
                "kind": "OR",
                "range": {
                  "start": 8,
                  "end": 14
                },
                "groupRange": {
                  "start": 7,
                  "end": 15
                }
              },
              "kind": "AND",
              "range": {
                "start": 1,
                "end": 15
              },
              "groupRange": {
                "start": 0,
                "end": 16
              }
            }
        `))

    test('query with mixed parenthesis and implicit AND', () =>
        expect(parse('(a AND b) c d')).toMatchInlineSnapshot(`
            {
              "type": "sequence",
              "nodes": [
                {
                  "type": "operator",
                  "left": {
                    "type": "pattern",
                    "range": {
                      "start": 1,
                      "end": 2
                    },
                    "kind": 1,
                    "value": "a",
                    "delimited": false
                  },
                  "right": {
                    "type": "pattern",
                    "range": {
                      "start": 7,
                      "end": 8
                    },
                    "kind": 1,
                    "value": "b",
                    "delimited": false
                  },
                  "kind": "AND",
                  "range": {
                    "start": 1,
                    "end": 8
                  },
                  "groupRange": {
                    "start": 0,
                    "end": 9
                  }
                },
                {
                  "type": "pattern",
                  "range": {
                    "start": 10,
                    "end": 11
                  },
                  "kind": 1,
                  "value": "c",
                  "delimited": false
                },
                {
                  "type": "pattern",
                  "range": {
                    "start": 12,
                    "end": 13
                  },
                  "kind": 1,
                  "value": "d",
                  "delimited": false
                }
              ],
              "range": {
                "start": 0,
                "end": 13
              }
            }
        `))

    test('query with mixed explicit OR and implicit AND operators', () =>
        expect(parse('aaa bbb or ccc')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "sequence",
                "nodes": [
                  {
                    "type": "pattern",
                    "range": {
                      "start": 0,
                      "end": 3
                    },
                    "kind": 1,
                    "value": "aaa",
                    "delimited": false
                  },
                  {
                    "type": "pattern",
                    "range": {
                      "start": 4,
                      "end": 7
                    },
                    "kind": 1,
                    "value": "bbb",
                    "delimited": false
                  }
                ],
                "range": {
                  "start": 0,
                  "end": 7
                }
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 11,
                  "end": 14
                },
                "kind": 1,
                "value": "ccc",
                "delimited": false
              },
              "kind": "OR",
              "range": {
                "start": 0,
                "end": 14
              }
            }
        `))

    test('query with mixed explicit and implicit operators inside parens', () =>
        expect(parse('(aaa bbb and ccc)')).toMatchInlineSnapshot(`
            {
              "type": "operator",
              "left": {
                "type": "sequence",
                "nodes": [
                  {
                    "type": "pattern",
                    "range": {
                      "start": 1,
                      "end": 4
                    },
                    "kind": 1,
                    "value": "aaa",
                    "delimited": false
                  },
                  {
                    "type": "pattern",
                    "range": {
                      "start": 5,
                      "end": 8
                    },
                    "kind": 1,
                    "value": "bbb",
                    "delimited": false
                  }
                ],
                "range": {
                  "start": 1,
                  "end": 8
                }
              },
              "right": {
                "type": "pattern",
                "range": {
                  "start": 13,
                  "end": 16
                },
                "kind": 1,
                "value": "ccc",
                "delimited": false
              },
              "kind": "AND",
              "range": {
                "start": 1,
                "end": 16
              },
              "groupRange": {
                "start": 0,
                "end": 17
              }
            }
        `))
})
