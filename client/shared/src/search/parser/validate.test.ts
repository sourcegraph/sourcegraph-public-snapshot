import { findGlobalFilter } from './validate'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

describe('finds a global filter', () => {
    test('valid global filter', () => {
        expect(findGlobalFilter('repo:sg/sg case:yes SauceGraph', 'case')).toMatchInlineSnapshot(`
            {
              "type": "filter",
              "range": {
                "start": 11,
                "end": 19
              },
              "field": {
                "type": "literal",
                "value": "case",
                "range": {
                  "start": 11,
                  "end": 15
                }
              },
              "value": {
                "type": "literal",
                "value": "yes",
                "range": {
                  "start": 16,
                  "end": 19
                }
              },
              "negated": false
            }
        `)
    })

    test('invalid, more than one filter', () => {
        expect(
            findGlobalFilter('patterntype:literal SauceGraph or PatternType:regexp wafflecat', 'patterntype')
        ).toBeUndefined()
    })

    test('invalid, grouped filter', () => {
        expect(findGlobalFilter('repo:sg/sg (case:yes SauceGraph)', 'case')).toBeUndefined()
    })
})
