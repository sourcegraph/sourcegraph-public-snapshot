import { scanPredicate, resolveAccess, predicates } from './predicates'

expect.addSnapshotSerializer({
    serialize: value => (value ? JSON.stringify(value) : 'invalid'),
    test: () => true,
})

describe('scanPredicate', () => {
    test('scan recognized and valid syntax', () => {
        expect(scanPredicate('repo', 'contains.file(content:stuff)', true)).toMatchInlineSnapshot(
            '{"path":["contains","file"],"parameters":"(content:stuff)"}'
        )
    })

    test('scan recognized dot syntax', () => {
        expect(scanPredicate('repo', 'contains.commit.after(stuff)', true)).toMatchInlineSnapshot(
            '{"path":["contains","commit","after"],"parameters":"(stuff)"}'
        )
    })

    test('scan recognized and valid syntax with escapes', () => {
        expect(scanPredicate('repo', 'contains.file(content:\\((stuff))', true)).toMatchInlineSnapshot(
            '{"path":["contains","file"],"parameters":"(content:\\\\((stuff))"}'
        )
    })

    test('scan valid syntax but not recognized', () => {
        expect(scanPredicate('foo', 'contains.path(stuff)', true)).toMatchInlineSnapshot('invalid')
    })

    test('scan unbalanced syntax', () => {
        expect(scanPredicate('repo', 'contains.file(content:', true)).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid nonalphanumeric name', () => {
        expect(scanPredicate('repo', 'contains.yo?inks(stuff)', true)).toMatchInlineSnapshot('invalid')
    })

    test('resolve field aliases for predicates', () => {
        expect(scanPredicate('r', 'contains.file(content:stuff)', true)).toMatchInlineSnapshot(
            '{"path":["contains","file"],"parameters":"(content:stuff)"}'
        )
    })

    test('scan recognized file:contains.content syntax', () => {
        expect(scanPredicate('file', 'contains.content(stuff)', true)).toMatchInlineSnapshot(
            '{"path":["contains","content"],"parameters":"(stuff)"}'
        )
    })

    test('scan invalid repo:contains() syntax', () => {
        expect(scanPredicate('repo', 'contains(content:stuff)', true)).toMatchInlineSnapshot('invalid')
    })

    test('scan invalid file:contains() syntax', () => {
        expect(scanPredicate('file', 'contains(stuff', true)).toMatchInlineSnapshot('invalid')
    })
})

describe('resolveAccess', () => {
    test('resolves partial access tree', () => {
        expect(resolveAccess(['repo'], predicates(true))).toMatchInlineSnapshot(
            '[{"name":"contains","fields":[{"name":"file"},{"name":"path"},{"name":"content"},{"name":"commit","fields":[{"name":"after"}]}]},{"name":"has","fields":[{"name":"file"},{"name":"path"},{"name":"content"},{"name":"commit","fields":[{"name":"after"}]},{"name":"description"},{"name":"tag"},{"name":"key"}]}]'
        )
    })

    test('resolves partial access tree depth 2', () => {
        expect(resolveAccess(['repo', 'contains', 'commit'], predicates(true))).toMatchInlineSnapshot(
            '[{"name":"after"}]'
        )
    })

    test('resolves fully qualified path', () => {
        expect(resolveAccess(['repo', 'contains', 'file'], predicates(true))).toMatchInlineSnapshot('[]')
    })

    test('undefind path', () => {
        expect(resolveAccess(['OCOTILLO', 'contains', 'file'], predicates(true))).toMatchInlineSnapshot('invalid')
    })

    test('invalid predicate syntax', () => {
        expect(resolveAccess(['repo', 'contains'], predicates(true))).toMatchInlineSnapshot('invalid')
    })
})
