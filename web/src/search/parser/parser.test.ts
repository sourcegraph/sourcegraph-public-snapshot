import { parseSearchQuery } from './parser'

describe('parseSearchQuery()', () => {
    test('empty', () =>
        expect(parseSearchQuery('')).toMatchObject({
            range: {
                end: 0,
                start: 0,
            },
            token: {
                members: [],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('whitespace', () =>
        expect(parseSearchQuery('  ')).toMatchObject({
            range: {
                end: 1,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 1,
                            start: 0,
                        },
                        token: {
                            type: 'whitespace',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('literal', () =>
        expect(parseSearchQuery('a')).toMatchObject({
            range: {
                end: 0,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 0,
                            start: 0,
                        },
                        token: {
                            type: 'literal',
                            value: 'a',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('filter', () =>
        expect(parseSearchQuery('a:b')).toMatchObject({
            range: {
                end: 2,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 2,
                            start: 0,
                        },
                        token: {
                            filterType: {
                                range: {
                                    end: 0,
                                    start: 0,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'a',
                                },
                                type: 'success',
                            },
                            filterValue: {
                                range: {
                                    end: 2,
                                    start: 2,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'b',
                                },
                                type: 'success',
                            },
                            type: 'filter',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('negated filter', () =>
        expect(parseSearchQuery('-a:b')).toMatchObject({
            range: {
                end: 3,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 3,
                            start: 0,
                        },
                        token: {
                            filterType: {
                                range: {
                                    end: 1,
                                    start: 0,
                                },
                                token: {
                                    type: 'literal',
                                    value: '-a',
                                },
                                type: 'success',
                            },
                            filterValue: {
                                range: {
                                    end: 3,
                                    start: 3,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'b',
                                },
                                type: 'success',
                            },
                            type: 'filter',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('filter with quoted value', () => {
        expect(parseSearchQuery('a:"b"')).toMatchObject({
            range: {
                end: 4,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 4,
                            start: 0,
                        },
                        token: {
                            filterType: {
                                range: {
                                    end: 0,
                                    start: 0,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'a',
                                },
                                type: 'success',
                            },
                            filterValue: {
                                range: {
                                    end: 4,
                                    start: 2,
                                },
                                token: {
                                    quotedValue: 'b',
                                    type: 'quoted',
                                },
                                type: 'success',
                            },
                            type: 'filter',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        })
    })

    test('quoted', () =>
        expect(parseSearchQuery('"a:b"')).toMatchObject({
            range: {
                end: 4,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 4,
                            start: 0,
                        },
                        token: {
                            quotedValue: 'a:b',
                            type: 'quoted',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('quoted (escaped quotes)', () =>
        expect(parseSearchQuery('"-\\"a\\":b"')).toMatchObject({
            range: {
                end: 9,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 9,
                            start: 0,
                        },
                        token: {
                            quotedValue: '-\\"a\\":b',
                            type: 'quoted',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))

    test('complex query', () =>
        expect(parseSearchQuery('repo:^github\\.com/gorilla/mux$ lang:go -file:mux.go Router')).toMatchObject({
            range: {
                end: 57,
                start: 0,
            },
            token: {
                members: [
                    {
                        range: {
                            end: 29,
                            start: 0,
                        },
                        token: {
                            filterType: {
                                range: {
                                    end: 3,
                                    start: 0,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'repo',
                                },
                                type: 'success',
                            },
                            filterValue: {
                                range: {
                                    end: 29,
                                    start: 5,
                                },
                                token: {
                                    type: 'literal',
                                    value: '^github\\.com/gorilla/mux$',
                                },
                                type: 'success',
                            },
                            type: 'filter',
                        },
                    },
                    {
                        range: {
                            end: 30,
                            start: 30,
                        },
                        token: {
                            type: 'whitespace',
                        },
                    },
                    {
                        range: {
                            end: 37,
                            start: 31,
                        },
                        token: {
                            filterType: {
                                range: {
                                    end: 34,
                                    start: 31,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'lang',
                                },
                                type: 'success',
                            },
                            filterValue: {
                                range: {
                                    end: 37,
                                    start: 36,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'go',
                                },
                                type: 'success',
                            },
                            type: 'filter',
                        },
                    },
                    {
                        range: {
                            end: 38,
                            start: 38,
                        },
                        token: {
                            type: 'whitespace',
                        },
                    },
                    {
                        range: {
                            end: 50,
                            start: 39,
                        },
                        token: {
                            filterType: {
                                range: {
                                    end: 43,
                                    start: 39,
                                },
                                token: {
                                    type: 'literal',
                                    value: '-file',
                                },
                                type: 'success',
                            },
                            filterValue: {
                                range: {
                                    end: 50,
                                    start: 45,
                                },
                                token: {
                                    type: 'literal',
                                    value: 'mux.go',
                                },
                                type: 'success',
                            },
                            type: 'filter',
                        },
                    },
                    {
                        range: {
                            end: 51,
                            start: 51,
                        },
                        token: {
                            type: 'whitespace',
                        },
                    },
                    {
                        range: {
                            end: 57,
                            start: 52,
                        },
                        token: {
                            type: 'literal',
                            value: 'Router',
                        },
                    },
                ],
                type: 'sequence',
            },
            type: 'success',
        }))
})
