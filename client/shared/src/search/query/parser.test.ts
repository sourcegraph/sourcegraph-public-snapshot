import { parseSearchQuery, Node, ParseSuccess } from './parser'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, '  '),
    test: () => true,
})

export const parse = (input: string): Node[] => (parseSearchQuery(input) as ParseSuccess).nodes

describe('parseSearchQuery', () => {
    test('query with leaves', () =>
        expect(parse('repo:foo a b c')).toMatchInlineSnapshot(`
            [
              {
                "type": "parameter",
                "field": "repo",
                "value": "foo",
                "negated": false,
                "range": {
                  "start": 0,
                  "end": 8
                }
              },
              {
                "type": "pattern",
                "kind": 1,
                "value": "a",
                "quoted": false,
                "negated": false,
                "range": {
                  "start": 9,
                  "end": 10
                }
              },
              {
                "type": "pattern",
                "kind": 1,
                "value": "b",
                "quoted": false,
                "negated": false,
                "range": {
                  "start": 11,
                  "end": 12
                }
              },
              {
                "type": "pattern",
                "kind": 1,
                "value": "c",
                "quoted": false,
                "negated": false,
                "range": {
                  "start": 13,
                  "end": 14
                }
              }
            ]
        `))

    test('query with and', () =>
        expect(parse('a b and c')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "a",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 0,
                      "end": 1
                    }
                  },
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "b",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 2,
                      "end": 3
                    }
                  },
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "c",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 8,
                      "end": 9
                    }
                  }
                ],
                "kind": "AND",
                "range": {
                  "start": 0,
                  "end": 9
                }
              }
            ]
        `))

    test('query with or', () =>
        expect(parse('a or b')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "a",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 0,
                      "end": 1
                    }
                  },
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "b",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 5,
                      "end": 6
                    }
                  }
                ],
                "kind": "OR",
                "range": {
                  "start": 0,
                  "end": 6
                }
              }
            ]
        `))

    test('query with and/or operator precedence', () =>
        expect(parse('a or b and c')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "a",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 0,
                      "end": 1
                    }
                  },
                  {
                    "type": "operator",
                    "operands": [
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "b",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 5,
                          "end": 6
                        }
                      },
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "c",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 11,
                          "end": 12
                        }
                      }
                    ],
                    "kind": "AND",
                    "range": {
                      "start": 5,
                      "end": 12
                    }
                  }
                ],
                "kind": "OR",
                "range": {
                  "start": 0,
                  "end": 12
                }
              }
            ]
        `))

    test('query with parentheses that override precedence', () => {
        expect(parse('a and (b or c)')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "a",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 0,
                      "end": 1
                    }
                  },
                  {
                    "type": "operator",
                    "operands": [
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "b",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 7,
                          "end": 8
                        }
                      },
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "c",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 12,
                          "end": 13
                        }
                      }
                    ],
                    "kind": "OR",
                    "range": {
                      "start": 7,
                      "end": 13
                    },
                    "groupRange": {
                      "start": 6,
                      "end": 14
                    }
                  }
                ],
                "kind": "AND",
                "range": {
                  "start": 0,
                  "end": 14
                }
              }
            ]
        `)

        expect(parse('(a or b) and c')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "operator",
                    "operands": [
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "a",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 1,
                          "end": 2
                        }
                      },
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "b",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 6,
                          "end": 7
                        }
                      }
                    ],
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
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "c",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 13,
                      "end": 14
                    }
                  }
                ],
                "kind": "AND",
                "range": {
                  "start": 0,
                  "end": 14
                }
              }
            ]
        `)
    })

    test('query with nested parentheses', () =>
        expect(parse('(a and (b or c))')).toMatchInlineSnapshot(`
            [
              {
                "type": "operator",
                "operands": [
                  {
                    "type": "pattern",
                    "kind": 1,
                    "value": "a",
                    "quoted": false,
                    "negated": false,
                    "range": {
                      "start": 1,
                      "end": 2
                    }
                  },
                  {
                    "type": "operator",
                    "operands": [
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "b",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 8,
                          "end": 9
                        }
                      },
                      {
                        "type": "pattern",
                        "kind": 1,
                        "value": "c",
                        "quoted": false,
                        "negated": false,
                        "range": {
                          "start": 13,
                          "end": 14
                        }
                      }
                    ],
                    "kind": "OR",
                    "range": {
                      "start": 8,
                      "end": 14
                    },
                    "groupRange": {
                      "start": 7,
                      "end": 15
                    }
                  }
                ],
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
            ]
        `))

    test('query with mixed explicit and implicit operators inside parens', () =>
        expect(parse('(aaa bbb and ccc)')).toMatchInlineSnapshot(`
                            [
                              {
                                "type": "operator",
                                "operands": [
                                  {
                                    "type": "pattern",
                                    "kind": 1,
                                    "value": "aaa",
                                    "quoted": false,
                                    "negated": false,
                                    "range": {
                                      "start": 1,
                                      "end": 4
                                    }
                                  },
                                  {
                                    "type": "pattern",
                                    "kind": 1,
                                    "value": "bbb",
                                    "quoted": false,
                                    "negated": false,
                                    "range": {
                                      "start": 5,
                                      "end": 8
                                    }
                                  },
                                  {
                                    "type": "pattern",
                                    "kind": 1,
                                    "value": "ccc",
                                    "quoted": false,
                                    "negated": false,
                                    "range": {
                                      "start": 13,
                                      "end": 16
                                    }
                                  }
                                ],
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
                            ]
                    `))
})
