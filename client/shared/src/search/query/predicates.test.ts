import { scanPredicate } from './predicates'

expect.addSnapshotSerializer({
    serialize: value => (value ? JSON.stringify(value) : 'invalid'),
    test: () => true,
})

describe('scanPredicate()', () => {
    test('scan recognized and valid syntax', () => {
        expect(scanPredicate('repo', 'contains(stuff)')).toMatchInlineSnapshot(
            '{"name":"contains","parameters":"(stuff)"}'
        )
    })

    test('scan recognized and valid syntax with escapes', () => {
        expect(scanPredicate('repo', 'contains(\\((stuff))')).toMatchInlineSnapshot(
            '{"name":"contains","parameters":"(\\\\((stuff))"}'
        )
    })

    test('scan valid syntax but not recognized', () => {
        expect(scanPredicate('foo', 'contains(stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('scan unbalanced syntax', () => {
        expect(scanPredicate('repo', 'contains(')).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid nonalphanumeric name', () => {
        expect(scanPredicate('repo', 'contains.yoinks(stuff)')).toMatchInlineSnapshot('invalid')
    })
})
