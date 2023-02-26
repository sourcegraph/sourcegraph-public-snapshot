import { FilterType } from './filters'
import { findFilter, FilterKind } from './query'
import { filterExists } from './validate'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

describe('finds a filter', () => {
    test('valid global filter', () => {
        expect(findFilter('repo:sg/sg case:yes SauceGraph', 'case', FilterKind.Global, true)).toBeTruthy()
    })

    test('invalid global filter for more than one filter', () => {
        expect(
            findFilter(
                'patterntype:literal SauceGraph or PatternType:regexp wafflecat',
                'patterntype',
                FilterKind.Global,
                true
            )
        ).toBeUndefined()
    })

    test('invalid global filter for subexpression filter', () => {
        expect(findFilter('repo:sg/sg (case:yes SauceGraph)', 'case', FilterKind.Global, true)).toBeUndefined()
    })

    test('valid single subexpression filter', () => {
        expect(findFilter('repo:sg/sg (case:yes SauceGraph)', 'case', FilterKind.Subexpression, true)).toBeTruthy()
    })

    test('valid multiple subexpression filters', () => {
        expect(
            findFilter('repo:sg/sg (case:yes SauceGraph) or (case:no derp)', 'case', FilterKind.Subexpression, true)
        ).toBeTruthy()
    })

    test('invalid subexpression filter when global', () => {
        expect(findFilter('repo:sg/sg case:yes', 'case', FilterKind.Subexpression, true)).toBeUndefined()
    })
})

describe('isContextFilterInQuery', () => {
    test('no context filter in query', () => {
        expect(filterExists('foo', FilterType.context, false, true)).toBeFalsy()
    })

    test('context filter in query', () => {
        expect(filterExists('context:@user foo', FilterType.context, false, true)).toBeTruthy()
    })

    test('context filters in both subexpressions', () => {
        expect(filterExists('(context:@user foo) or (context:@test bar)', FilterType.context, false, true)).toBeTruthy()
    })

    test('context filters in one subexpression', () => {
        expect(filterExists('foo or (context:@test bar)', FilterType.context, false, true)).toBeTruthy()
    })
})
