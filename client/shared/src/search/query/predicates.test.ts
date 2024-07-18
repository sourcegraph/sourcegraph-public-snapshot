import { describe, expect, test } from 'vitest'

import { scanPredicate } from './predicates'

expect.addSnapshotSerializer({
    serialize: value => (value ? JSON.stringify(value) : 'invalid'),
    test: () => true,
})

describe('scanPredicate', () => {
    test('scan recognized and valid syntax', () => {
        expect(scanPredicate('repo', 'contains.file(content:stuff)')).toMatchInlineSnapshot(
            '{"field":"repo","name":"contains.file","parameters":"(content:stuff)"}'
        )
    })

    test('scan recognized dot syntax', () => {
        expect(scanPredicate('repo', 'contains.commit.after(stuff)')).toMatchInlineSnapshot(
            '{"field":"repo","name":"contains.commit.after","parameters":"(stuff)"}'
        )
    })

    test('scan recognized and valid syntax with escapes', () => {
        expect(scanPredicate('repo', 'contains.file(content:\\((stuff))')).toMatchInlineSnapshot(
            '{"field":"repo","name":"contains.file","parameters":"(content:\\\\((stuff))"}'
        )
    })

    test('scan valid syntax but not recognized', () => {
        expect(scanPredicate('foo', 'contains.path(stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('scan unbalanced syntax', () => {
        expect(scanPredicate('repo', 'contains.file(content:')).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid nonalphanumeric name', () => {
        expect(scanPredicate('repo', 'contains.yo?inks(stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('resolve field aliases for predicates', () => {
        expect(scanPredicate('r', 'contains.file(content:stuff)')).toMatchInlineSnapshot(
            '{"field":"repo","name":"contains.file","parameters":"(content:stuff)"}'
        )
    })

    test('scan recognized file:contains.content syntax', () => {
        expect(scanPredicate('file', 'contains.content(stuff)')).toMatchInlineSnapshot(
            '{"field":"file","name":"contains.content","parameters":"(stuff)"}'
        )
    })

    test('scan invalid repo:contains() syntax', () => {
        expect(scanPredicate('repo', 'contains(content:stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid file:contains() syntax', () => {
        expect(scanPredicate('file', 'contains(stuff')).toMatchInlineSnapshot('invalid')
    })

    test('scan repo:has.meta with regex', () => {
        expect(scanPredicate('repo', 'has.meta(/abc.*/:/def.*/)')).toMatchInlineSnapshot(
            '{"field":"repo","name":"has.meta","parameters":"(/abc.*/:/def.*/)"}'
        )
    })
})
