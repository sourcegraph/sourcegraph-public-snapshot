import { convertPlainTextToInteractiveQuery, escapeDoubleQuotes } from './helpers'

describe('Search input helpers', () => {
    describe('convertPlainTextToInteractiveQuery', () => {
        test('converts query with no filters', () => {
            const newQuery = convertPlainTextToInteractiveQuery('foo')
            expect(newQuery.navbarQuery === 'foo' && newQuery.filtersInQuery === {})
        })
        test('converts query with one filter', () => {
            const newQuery = convertPlainTextToInteractiveQuery('foo case:yes')
            expect(
                newQuery.navbarQuery === 'foo' &&
                    newQuery.filtersInQuery ===
                        {
                            case: {
                                type: 'case' as const,
                                value: 'yes',
                                editable: false,
                                negated: false,
                            },
                        }
            )
        })
        test('converts query with multiple filters', () => {
            const newQuery = convertPlainTextToInteractiveQuery('foo case:yes archived:no')
            expect(
                newQuery.navbarQuery === 'foo' &&
                    newQuery.filtersInQuery ===
                        {
                            case: {
                                type: 'case' as const,
                                value: 'yes',
                                editable: false,
                                negated: false,
                            },
                            archived: {
                                type: 'archived' as const,
                                value: 'no',
                                editable: false,
                                negated: false,
                            },
                        }
            )
        })

        test('converts query with invalid filters, without adding invalid filters to filtersInQuery', () => {
            const newQuery = convertPlainTextToInteractiveQuery('foo case:yes archived:no asdf:no')
            expect(
                newQuery.navbarQuery === 'foo asdf:no' &&
                    newQuery.filtersInQuery ===
                        {
                            case: {
                                type: 'case' as const,
                                value: 'yes',
                                editable: false,
                                negated: false,
                            },
                            archived: {
                                type: 'archived' as const,
                                value: 'no',
                                editable: false,
                                negated: false,
                            },
                        }
            )
        })
    })

    describe('escapeDoubleQuotes', () => {
        test('Escapes one double quote input', () => {
            expect(escapeDoubleQuotes('"') === '"\\""')
        })
        test('Escapes quoted input', () => {
            expect(escapeDoubleQuotes('"test query"') === '"\\"test query\\""')
        })
        test('Escapes multiple consecutive double quotes', () => {
            expect(escapeDoubleQuotes('""') === '"\\"\\""')
            expect(escapeDoubleQuotes('"""') === '"\\"\\"\\""')
            expect(escapeDoubleQuotes('""""') === '"\\"\\"\\"\\""')
        })
        test('Escapes double quotes within input', () => {
            expect(escapeDoubleQuotes('test " query') === 'test \\" query')
        })

        test('Escapes multiple double quotes within input', () => {
            expect(escapeDoubleQuotes('"test" "query"') === '\\"test\\" \\"query\\"')
        })
    })
})
