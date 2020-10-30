import { parseSearchQuery } from './parser'

describe('parseSearchQuery()', () => {
    test('empty', () =>
        expect(parseSearchQuery('')).toMatchObject({
            token: {
                range: {
                    start: 0,
                    end: 1,
                },
                members: [],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('whitespace', () =>
        expect(parseSearchQuery('  ')).toMatchObject({
            token: {
                range: {
                    start: 0,
                    end: 2,
                },
                members: [
                    {
                        range: {
                            end: 2,
                            start: 0,
                        },
                        type: 'whitespace',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('literal', () =>
        expect(parseSearchQuery('a')).toMatchObject({
            token: {
                range: {
                    start: 0,
                    end: 1,
                },
                members: [
                    {
                        range: {
                            start: 0,
                            end: 1,
                        },
                        type: 'literal',
                        value: 'a',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('triple quotes', () => {
        expect(parseSearchQuery('"""')).toMatchObject({
            token: {
                range: {
                    end: 3,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 3,
                            start: 0,
                        },
                        type: 'literal',
                        value: '"""',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('filter', () =>
        expect(parseSearchQuery('f:b')).toMatchObject({
            token: {
                range: {
                    end: 3,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 3,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 1,
                                start: 0,
                            },
                            type: 'literal',
                            value: 'f',
                        },
                        filterValue: {
                            range: {
                                end: 3,
                                start: 2,
                            },
                            type: 'literal',
                            value: 'b',
                        },
                        type: 'filter',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('negated filter', () =>
        expect(parseSearchQuery('-f:b')).toMatchObject({
            token: {
                range: {
                    end: 4,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 4,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 2,
                                start: 0,
                            },
                            type: 'literal',
                            value: '-f',
                        },
                        filterValue: {
                            range: {
                                end: 4,
                                start: 3,
                            },
                            type: 'literal',
                            value: 'b',
                        },
                        type: 'filter',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('filter with quoted value', () => {
        expect(parseSearchQuery('f:"b"')).toMatchObject({
            token: {
                range: {
                    end: 5,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 5,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 1,
                                start: 0,
                            },
                            type: 'literal',
                            value: 'f',
                        },
                        filterValue: {
                            range: {
                                end: 5,
                                start: 2,
                            },
                            quotedValue: 'b',
                            type: 'quoted',
                        },
                        type: 'filter',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('filter with a value ending with a colon', () => {
        expect(parseSearchQuery('f:a:')).toStrictEqual({
            token: {
                range: {
                    end: 4,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 4,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 1,
                                start: 0,
                            },
                            type: 'literal',
                            value: 'f',
                        },
                        filterValue: {
                            range: {
                                end: 4,
                                start: 2,
                            },
                            type: 'literal',
                            value: 'a:',
                        },
                        type: 'filter',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('filter where the value is a colon', () => {
        expect(parseSearchQuery('f::')).toStrictEqual({
            token: {
                range: {
                    end: 3,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 3,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 1,
                                start: 0,
                            },
                            type: 'literal',
                            value: 'f',
                        },
                        filterValue: {
                            range: {
                                end: 3,
                                start: 2,
                            },
                            type: 'literal',
                            value: ':',
                        },
                        type: 'filter',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('quoted, double quotes', () =>
        expect(parseSearchQuery('"a:b"')).toMatchObject({
            token: {
                range: {
                    end: 5,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 5,
                            start: 0,
                        },
                        quotedValue: 'a:b',
                        type: 'quoted',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('quoted, single quotes', () =>
        expect(parseSearchQuery("'a:b'")).toMatchObject({
            token: {
                range: {
                    end: 5,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 5,
                            start: 0,
                        },
                        quotedValue: 'a:b',
                        type: 'quoted',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('quoted (escaped quotes)', () =>
        expect(parseSearchQuery('"-\\"a\\":b"')).toMatchObject({
            token: {
                range: {
                    end: 10,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 10,
                            start: 0,
                        },
                        quotedValue: '-\\"a\\":b',
                        type: 'quoted',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('complex query', () =>
        expect(parseSearchQuery('repo:^github\\.com/gorilla/mux$ lang:go -file:mux.go Router')).toMatchObject({
            token: {
                range: {
                    end: 58,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 30,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 4,
                                start: 0,
                            },
                            type: 'literal',
                            value: 'repo',
                        },
                        filterValue: {
                            range: {
                                end: 30,
                                start: 5,
                            },
                            type: 'literal',
                            value: '^github\\.com/gorilla/mux$',
                        },
                        type: 'filter',
                    },
                    {
                        range: {
                            end: 31,
                            start: 30,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 38,
                            start: 31,
                        },
                        filterType: {
                            range: {
                                end: 35,
                                start: 31,
                            },
                            type: 'literal',
                            value: 'lang',
                        },
                        filterValue: {
                            range: {
                                end: 38,
                                start: 36,
                            },
                            type: 'literal',
                            value: 'go',
                        },
                        type: 'filter',
                    },
                    {
                        range: {
                            end: 39,
                            start: 38,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 51,
                            start: 39,
                        },
                        filterType: {
                            range: {
                                end: 44,
                                start: 39,
                            },
                            type: 'literal',
                            value: '-file',
                        },
                        filterValue: {
                            range: {
                                end: 51,
                                start: 45,
                            },
                            type: 'literal',
                            value: 'mux.go',
                        },
                        type: 'filter',
                    },
                    {
                        range: {
                            end: 52,
                            start: 51,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 58,
                            start: 52,
                        },
                        type: 'literal',
                        value: 'Router',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('parenthesized parameters', () => {
        expect(parseSearchQuery('repo:a (file:b and c)')).toMatchObject({
            token: {
                range: {
                    end: 21,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 6,
                            start: 0,
                        },
                        filterType: {
                            range: {
                                end: 4,
                                start: 0,
                            },
                            type: 'literal',
                            value: 'repo',
                        },
                        filterValue: {
                            range: {
                                end: 6,
                                start: 5,
                            },
                            type: 'literal',
                            value: 'a',
                        },
                        type: 'filter',
                    },
                    {
                        range: {
                            end: 7,
                            start: 6,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 8,
                            start: 7,
                        },
                        type: 'openingParen',
                    },
                    {
                        range: {
                            end: 14,
                            start: 8,
                        },
                        filterType: {
                            range: {
                                end: 12,
                                start: 8,
                            },
                            type: 'literal',
                            value: 'file',
                        },
                        filterValue: {
                            range: {
                                end: 14,
                                start: 13,
                            },
                            type: 'literal',
                            value: 'b',
                        },
                        type: 'filter',
                    },
                    {
                        range: {
                            end: 15,
                            start: 14,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 18,
                            start: 15,
                        },
                        type: 'operator',
                        value: 'and',
                    },
                    {
                        range: {
                            end: 19,
                            start: 18,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 20,
                            start: 19,
                        },
                        type: 'literal',
                        value: 'c',
                    },
                    {
                        range: {
                            end: 21,
                            start: 20,
                        },
                        type: 'closingParen',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('nested parenthesized parameters', () => {
        expect(parseSearchQuery('(a and (b or c) and d)')).toMatchObject({
            token: {
                range: {
                    end: 22,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 1,
                            start: 0,
                        },
                        type: 'openingParen',
                    },
                    {
                        range: {
                            end: 2,
                            start: 1,
                        },
                        type: 'literal',
                        value: 'a',
                    },
                    {
                        range: {
                            end: 3,
                            start: 2,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 6,
                            start: 3,
                        },
                        type: 'operator',
                    },
                    {
                        range: {
                            end: 7,
                            start: 6,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 8,
                            start: 7,
                        },
                        type: 'openingParen',
                    },
                    {
                        range: {
                            end: 9,
                            start: 8,
                        },
                        type: 'literal',
                        value: 'b',
                    },
                    {
                        range: {
                            end: 10,
                            start: 9,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 12,
                            start: 10,
                        },
                        type: 'operator',
                    },
                    {
                        range: {
                            end: 13,
                            start: 12,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 14,
                            start: 13,
                        },
                        type: 'literal',
                        value: 'c',
                    },
                    {
                        range: {
                            end: 15,
                            start: 14,
                        },
                        type: 'closingParen',
                    },
                    {
                        range: {
                            end: 16,
                            start: 15,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 19,
                            start: 16,
                        },
                        type: 'operator',
                    },
                    {
                        range: {
                            end: 20,
                            start: 19,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 21,
                            start: 20,
                        },
                        type: 'literal',
                        value: 'd',
                    },
                    {
                        range: {
                            end: 22,
                            start: 21,
                        },
                        type: 'closingParen',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('do not treat links as filters', () => {
        expect(parseSearchQuery('http://example.com repo:a')).toMatchObject({
            token: {
                range: {
                    end: 25,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            end: 18,
                            start: 0,
                        },
                        type: 'literal',
                        value: 'http://example.com',
                    },
                    {
                        range: {
                            end: 19,
                            start: 18,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            end: 25,
                            start: 19,
                        },
                        filterType: {
                            range: {
                                end: 23,
                                start: 19,
                            },
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('interpret C-style comments', () => {
        const query = `// saucegraph is best graph
repo:sourcegraph
// search for thing
thing`
        expect(parseSearchQuery(query, true)).toMatchObject({
            token: {
                range: {
                    end: 70,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            start: 0,
                            end: 27,
                        },
                        type: 'comment',
                        value: '// saucegraph is best graph',
                    },
                    {
                        range: {
                            start: 27,
                            end: 28,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            start: 28,
                            end: 44,
                        },
                        filterType: {
                            range: {
                                start: 28,
                                end: 32,
                            },
                            type: 'literal',
                            value: 'repo',
                        },
                        filterValue: {
                            range: {
                                start: 33,
                                end: 44,
                            },
                            type: 'literal',
                            value: 'sourcegraph',
                        },
                        type: 'filter',
                    },
                    {
                        range: {
                            start: 44,
                            end: 45,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            start: 45,
                            end: 64,
                        },
                        type: 'comment',
                        value: '// search for thing',
                    },
                    {
                        range: {
                            start: 64,
                            end: 65,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            start: 65,
                            end: 70,
                        },
                        type: 'literal',
                        value: 'thing',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('do not interpret C-style comments', () => {
        expect(parseSearchQuery('// thing')).toMatchObject({
            token: {
                range: {
                    end: 8,
                    start: 0,
                },
                members: [
                    {
                        range: {
                            start: 0,
                            end: 2,
                        },
                        type: 'literal',
                        value: '//',
                    },
                    {
                        range: {
                            start: 2,
                            end: 3,
                        },
                        type: 'whitespace',
                    },
                    {
                        range: {
                            start: 3,
                            end: 8,
                        },
                        type: 'literal',
                        value: 'thing',
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })
})
