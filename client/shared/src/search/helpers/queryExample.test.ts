import { describe, expect, it } from '@jest/globals'

import { FilterType } from '../query/filters'

import { createQueryExampleFromString, updateQueryWithFilterAndExample } from './queryExample'

describe('example helpers', () => {
    describe('createExampleFromString', () => {
        it('parses examples without placeholder', () => {
            expect(createQueryExampleFromString('foo bar')).toMatchInlineSnapshot(`
                Object {
                  "tokens": Array [
                    Object {
                      "end": 7,
                      "start": 0,
                      "type": "text",
                      "value": "foo bar",
                    },
                  ],
                  "value": "foo bar",
                }
            `)
        })

        it('parses examples with placeholder', () => {
            expect(createQueryExampleFromString('{foo}')).toMatchInlineSnapshot(`
                Object {
                  "tokens": Array [
                    Object {
                      "end": 3,
                      "start": 0,
                      "type": "placeholder",
                      "value": "foo",
                    },
                  ],
                  "value": "foo",
                }
            `)
            expect(createQueryExampleFromString('({foo})')).toMatchInlineSnapshot(`
                Object {
                  "tokens": Array [
                    Object {
                      "end": 1,
                      "start": 0,
                      "type": "text",
                      "value": "(",
                    },
                    Object {
                      "end": 4,
                      "start": 1,
                      "type": "placeholder",
                      "value": "foo",
                    },
                    Object {
                      "end": 5,
                      "start": 4,
                      "type": "text",
                      "value": ")",
                    },
                  ],
                  "value": "(foo)",
                }
            `)
        })
    })

    describe('updateQueryWithFilterExample', () => {
        describe('repeatable filters', () => {
            it('appends placeholder filter and selects placeholder', () => {
                expect(
                    updateQueryWithFilterAndExample('foo', FilterType.after, createQueryExampleFromString('({test})'))
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 16,
                        "start": 4,
                      },
                      "placeholderRange": Object {
                        "end": 15,
                        "start": 11,
                      },
                      "query": "foo after:(test)",
                    }
                `)
            })

            it('appends filter with empty value', () => {
                expect(
                    updateQueryWithFilterAndExample('foo', FilterType.after, createQueryExampleFromString('({test})'), {
                        emptyValue: true,
                    })
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 10,
                        "start": 4,
                      },
                      "placeholderRange": Object {
                        "end": 10,
                        "start": 10,
                      },
                      "query": "foo after:",
                    }
                `)
            })

            it('appends negated filter', () => {
                expect(
                    updateQueryWithFilterAndExample('foo', FilterType.after, createQueryExampleFromString('({test})'), {
                        negate: true,
                    })
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 17,
                        "start": 4,
                      },
                      "placeholderRange": Object {
                        "end": 16,
                        "start": 12,
                      },
                      "query": "foo -after:(test)",
                    }
                `)
            })
        })

        describe('unique filters', () => {
            it('appends placeholder filter and selects placeholder', () => {
                expect(
                    updateQueryWithFilterAndExample('foo', FilterType.after, createQueryExampleFromString('({test})'), {
                        singular: true,
                    })
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 16,
                        "start": 4,
                      },
                      "placeholderRange": Object {
                        "end": 15,
                        "start": 11,
                      },
                      "query": "foo after:(test)",
                    }
                `)
            })

            it('selects value of existing placeholder', () => {
                expect(
                    updateQueryWithFilterAndExample(
                        'after:value foo',
                        FilterType.after,
                        createQueryExampleFromString('({test})'),
                        { singular: true }
                    )
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 11,
                        "start": 0,
                      },
                      "placeholderRange": Object {
                        "end": 11,
                        "start": 6,
                      },
                      "query": "after:value foo",
                    }
                `)
            })

            it('updates existing filter with empty value', () => {
                expect(
                    updateQueryWithFilterAndExample(
                        'after:value foo',
                        FilterType.after,
                        createQueryExampleFromString('({test})'),
                        { singular: true, emptyValue: true }
                    )
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 6,
                        "start": 0,
                      },
                      "placeholderRange": Object {
                        "end": 6,
                        "start": 6,
                      },
                      "query": "after: foo",
                    }
                `)
            })

            it('updates existing empty filter with empty value', () => {
                expect(
                    updateQueryWithFilterAndExample(
                        'after: foo',
                        FilterType.after,
                        createQueryExampleFromString('({test})'),
                        { singular: true, emptyValue: true }
                    )
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 6,
                        "start": 0,
                      },
                      "placeholderRange": Object {
                        "end": 6,
                        "start": 6,
                      },
                      "query": "after: foo",
                    }
                `)
            })

            it('selects value of existing negated filter', () => {
                expect(
                    updateQueryWithFilterAndExample(
                        '-after:value foo',
                        FilterType.after,
                        createQueryExampleFromString('({test})'),
                        { singular: true, negate: true }
                    )
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 12,
                        "start": 0,
                      },
                      "placeholderRange": Object {
                        "end": 12,
                        "start": 7,
                      },
                      "query": "-after:value foo",
                    }
                `)
            })

            it('updates existing negated filter with empty value', () => {
                expect(
                    updateQueryWithFilterAndExample(
                        '-after:value foo',
                        FilterType.after,
                        createQueryExampleFromString('({test})'),
                        { singular: true, negate: true, emptyValue: true }
                    )
                ).toMatchInlineSnapshot(`
                    Object {
                      "filterRange": Object {
                        "end": 7,
                        "start": 0,
                      },
                      "placeholderRange": Object {
                        "end": 7,
                        "start": 7,
                      },
                      "query": "-after: foo",
                    }
                `)
            })
        })
    })
})
