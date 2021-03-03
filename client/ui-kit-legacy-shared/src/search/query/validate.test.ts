import { findFilter, FilterKind, isContextFilterInQuery } from './validate'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

describe('finds a filter', () => {
    test('valid global filter', () => {
        expect(findFilter('repo:sg/sg case:yes SauceGraph', 'case', FilterKind.Global)).toBeTruthy()
    })

    test('invalid global filter for more than one filter', () => {
        expect(
            findFilter(
                'patterntype:literal SauceGraph or PatternType:regexp wafflecat',
                'patterntype',
                FilterKind.Global
            )
        ).toBeUndefined()
    })

    test('invalid global filter for subexpression filter', () => {
        expect(findFilter('repo:sg/sg (case:yes SauceGraph)', 'case', FilterKind.Global)).toBeUndefined()
    })

    test('valid single subexpression filter', () => {
        expect(findFilter('repo:sg/sg (case:yes SauceGraph)', 'case', FilterKind.Subexpression)).toBeTruthy()
    })

    test('valid multiple subexpression filters', () => {
        expect(
            findFilter('repo:sg/sg (case:yes SauceGraph) or (case:no derp)', 'case', FilterKind.Subexpression)
        ).toBeTruthy()
    })

    test('invalid subexpression filter when global', () => {
        expect(findFilter('repo:sg/sg case:yes', 'case', FilterKind.Subexpression)).toBeUndefined()
    })
})

describe('isContextFilterInQuery', () => {
    test('no context filter in query', () => {
        expect(isContextFilterInQuery('foo')).toBeFalsy()
    })

    test('context filter in query', () => {
        expect(isContextFilterInQuery('context:@user foo')).toBeTruthy()
    })

    test('context filters in both subexpressions', () => {
        expect(isContextFilterInQuery('(context:@user foo) or (context:@test bar)')).toBeTruthy()
    })

    test('context filters in one subexpression', () => {
        expect(isContextFilterInQuery('foo or (context:@test bar)')).toBeTruthy()
    })
})
