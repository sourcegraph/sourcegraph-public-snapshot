import { describe, expect, test } from 'vitest'

import { scanPredicate, resolveAccess, PREDICATES } from './predicates'

expect.addSnapshotSerializer({
    serialize: value => (value ? JSON.stringify(value) : 'invalid'),
    test: () => true,
})

describe('scanPredicate', () => {
    test('scan recognized and valid syntax', () => {
        expect(scanPredicate('repo', 'contains.file(content:stuff)')).toMatchInlineSnapshot(
            '{"path":["contains","file"],"parameters":"(content:stuff)"}'
        )
    })

    test('scan recognized dot syntax', () => {
        expect(scanPredicate('repo', 'contains.commit.after(stuff)')).toMatchInlineSnapshot(
            '{"path":["contains","commit","after"],"parameters":"(stuff)"}'
        )
    })

    test('scan recognized and valid syntax with escapes', () => {
        expect(scanPredicate('repo', 'contains.file(content:\\((stuff))')).toMatchInlineSnapshot(
            '{"path":["contains","file"],"parameters":"(content:\\\\((stuff))"}'
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
            '{"path":["contains","file"],"parameters":"(content:stuff)"}'
        )
    })

    test('scan recognized file:contains.content syntax', () => {
        expect(scanPredicate('file', 'contains.content(stuff)')).toMatchInlineSnapshot(
            '{"path":["contains","content"],"parameters":"(stuff)"}'
        )
    })

    test('scan invalid repo:contains() syntax', () => {
        expect(scanPredicate('repo', 'contains(content:stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid file:contains() syntax', () => {
        expect(scanPredicate('file', 'contains(stuff')).toMatchInlineSnapshot('invalid')
    })
})

describe('resolveAccess', () => {
    test('resolves partial access tree', () => {
        expect(resolveAccess(['repo'], PREDICATES)).toMatchInlineSnapshot(
            '[{"name":"contains","fields":[{"name":"file"},{"name":"path"},{"name":"content"},{"name":"commit","fields":[{"name":"after"}]}]},{"name":"has","fields":[{"name":"file"},{"name":"path"},{"name":"content"},{"name":"commit","fields":[{"name":"after"}]},{"name":"description"},{"name":"tag"},{"name":"key"},{"name":"meta"},{"name":"topic"}]}]'
        )
    })

    test('resolves partial access tree depth 2', () => {
        expect(resolveAccess(['repo', 'contains', 'commit'], PREDICATES)).toMatchInlineSnapshot('[{"name":"after"}]')
    })

    test('resolves fully qualified path', () => {
        expect(resolveAccess(['repo', 'contains', 'file'], PREDICATES)).toMatchInlineSnapshot('[]')
    })

    test('undefind path', () => {
        expect(resolveAccess(['OCOTILLO', 'contains', 'file'], PREDICATES)).toMatchInlineSnapshot('invalid')
    })

    test('invalid predicate syntax', () => {
        expect(resolveAccess(['repo', 'contains'], PREDICATES)).toMatchInlineSnapshot('invalid')
    })
})
