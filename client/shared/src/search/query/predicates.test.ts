import { scanPredicate, resolveAccess, PREDICATES } from './predicates'

expect.addSnapshotSerializer({
    serialize: value => (value ? JSON.stringify(value) : 'invalid'),
    test: () => true,
})

describe('scanPredicate', () => {
    test('scan recognized and valid syntax', () => {
        expect(scanPredicate('repo', 'contains(stuff)')).toMatchInlineSnapshot(
            '{"path":["contains"],"parameters":"(stuff)"}'
        )
    })

    test('scan recognized dot syntax', () => {
        expect(scanPredicate('repo', 'contains.commit.after(stuff)')).toMatchInlineSnapshot(
            '{"path":["contains","commit","after"],"parameters":"(stuff)"}'
        )
    })

    test('scan recognized and valid syntax with escapes', () => {
        expect(scanPredicate('repo', 'contains(\\((stuff))')).toMatchInlineSnapshot(
            '{"path":["contains"],"parameters":"(\\\\((stuff))"}'
        )
    })

    test('scan valid syntax but not recognized', () => {
        expect(scanPredicate('foo', 'contains(stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('scan unbalanced syntax', () => {
        expect(scanPredicate('repo', 'contains(')).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid nonalphanumeric name', () => {
        expect(scanPredicate('repo', 'contains.yo?inks(stuff)')).toMatchInlineSnapshot('invalid')
    })

    test('resolve field aliases for predicates', () => {
        expect(scanPredicate('r', 'contains.file(stuff)')).toMatchInlineSnapshot(
            '{"path":["contains","file"],"parameters":"(stuff)"}'
        )
    })
})

describe('resolveAccess', () => {
    test('resolves partial access tree', () => {
        expect(resolveAccess(['repo', 'contains'], PREDICATES)).toMatchInlineSnapshot(
            '[{"name":"file"},{"name":"content"},{"name":"commit","fields":[{"name":"after"}]}]'
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
})
