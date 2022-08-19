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

    test('scan recognized file.contains syntax', () => {
        expect(scanPredicate('file', 'contains(stuff)')).toMatchInlineSnapshot(
            '{"path":["contains"],"parameters":"(stuff)"}'
        )
    })
})

describe('resolveAccess', () => {
    test('resolves partial access tree', () => {
        expect(resolveAccess(['repo', 'contains'], PREDICATES)).toMatchInlineSnapshot(
            '[{"name":"file"},{"name":"path"},{"name":"content"},{"name":"commit","fields":[{"name":"after"}]}]'
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
