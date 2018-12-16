import assert from 'assert'
import { buildSearchURLQuery, makeRepoURI, parseHash, parseRepoURI, toPrettyBlobURL } from './url'

/**
 * Asserts deep object equality using node's assert.deepStrictEqual, except it (1) ignores differences in the
 * prototype (because that causes 2 object literals to fail the test) and (2) treats undefined properties as
 * missing.
 */
function assertDeepStrictEqual(actual: any, expected: any, message?: string): void {
    actual = JSON.parse(JSON.stringify(actual))
    expected = JSON.parse(JSON.stringify(expected))
    assert.deepStrictEqual(actual, expected, message)
}

describe('parseRepoURI', () => {
    it('should parse repo', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
        })
    })

    it('should parse repo with rev', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
        })
    })

    it('should parse repo with commitID', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        })
    })

    it('should parse repo with rev and file', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
        })
    })

    it('should parse repo with rev and file and line', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            position: {
                line: 3,
                character: 0,
            },
        })
    })

    it('should parse repo with rev and file and position', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3,5')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            position: {
                line: 3,
                character: 5,
            },
        })
    })

    it('should parse repo with rev and file and range', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3,5-6,9')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            range: {
                start: {
                    line: 3,
                    character: 5,
                },
                end: {
                    line: 6,
                    character: 9,
                },
            },
        })
    })
})

describe('makeRepoURI', () => {
    it('should make repo', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux')
    })

    it('should make repo with rev', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch')
    })

    it('should make repo with commitID', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
    })

    it('should make repo with rev and file', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go')
    })

    it('should make repo with rev and file and line', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            position: {
                line: 3,
                character: 0,
            },
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go:3')
    })

    it('should make repo with rev and file and position', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            position: {
                line: 3,
                character: 5,
            },
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go:3,5')
    })
})

describe('util/url', () => {
    const linePosition = { line: 1 }
    const lineCharPosition = { line: 1, character: 1 }
    const localRefMode = { ...lineCharPosition, viewState: 'references' }
    const externalRefMode = { ...lineCharPosition, viewState: 'references' }
    const ctx = {
        repoName: 'github.com/gorilla/mux',
        rev: '',
        commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        filePath: 'mux.go',
    }

    describe('parseHash', () => {
        it('parses empty hash', () => {
            assert.deepStrictEqual(parseHash(''), {})
        })

        it('parses unexpectedly formatted hash', () => {
            assert.deepStrictEqual(parseHash('L-53'), {})
            assert.deepStrictEqual(parseHash('L53:'), {})
            assert.deepStrictEqual(parseHash('L1:2-'), {})
            assert.deepStrictEqual(parseHash('L1:2-3'), {})
            assert.deepStrictEqual(parseHash('L1:2-3:'), {})
            assert.deepStrictEqual(parseHash('L1:-3:'), {})
            assert.deepStrictEqual(parseHash('L1:-3:4'), {})
            assert.deepStrictEqual(parseHash('L1-2:3'), {})
            assert.deepStrictEqual(parseHash('L1-2:'), {})
            assert.deepStrictEqual(parseHash('L1:-2'), {})
            assert.deepStrictEqual(parseHash('L1:2--3:4'), {})
            assert.deepStrictEqual(parseHash('L53:a'), {})
        })

        it('parses hash with leading octothorpe', () => {
            assert.deepStrictEqual(parseHash('#L1'), linePosition)
        })

        it('parses hash with line', () => {
            assert.deepStrictEqual(parseHash('L1'), linePosition)
        })

        it('parses hash with line and character', () => {
            assert.deepStrictEqual(parseHash('L1:1'), lineCharPosition)
        })

        it('parses hash with range', () => {
            assert.deepStrictEqual(parseHash('L1-2'), { line: 1, endLine: 2 })
            assert.deepStrictEqual(parseHash('L1:2-3:4'), { line: 1, character: 2, endLine: 3, endCharacter: 4 })
        })

        it('parses hash with local references', () => {
            assert.deepStrictEqual(parseHash('$references'), { viewState: 'references' })
            assert.deepStrictEqual(parseHash('L1:1$references'), localRefMode)
            assert.deepStrictEqual(parseHash('L1:1$references'), localRefMode)
        })
        it('parses modern hash with local references', () => {
            assert.deepStrictEqual(parseHash('tab=references'), { viewState: 'references' })
            assert.deepStrictEqual(parseHash('L1:1&tab=references'), localRefMode)
            assert.deepStrictEqual(parseHash('L1:1&tab=references'), localRefMode)
        })

        it('parses hash with external references', () => {
            assert.deepStrictEqual(parseHash('L1:1$references'), externalRefMode)
        })
        it('parses modern hash with external references', () => {
            assert.deepStrictEqual(parseHash('L1:1&tab=references'), externalRefMode)
        })
    })

    describe('toPrettyBlobURL', () => {
        it('formats url for empty rev', () => {
            assert.strictEqual(toPrettyBlobURL(ctx), '/github.com/gorilla/mux/-/blob/mux.go')
        })

        it('formats url for specified rev', () => {
            assert.strictEqual(
                toPrettyBlobURL({ ...ctx, rev: 'branch' }),
                '/github.com/gorilla/mux@branch/-/blob/mux.go'
            )
        })

        it('formats url with position', () => {
            assert.strictEqual(
                toPrettyBlobURL({ ...ctx, position: lineCharPosition }),
                '/github.com/gorilla/mux/-/blob/mux.go#L1:1'
            )
        })

        it('formats url with view state', () => {
            assert.strictEqual(
                toPrettyBlobURL({ ...ctx, position: lineCharPosition, viewState: 'references' }),
                '/github.com/gorilla/mux/-/blob/mux.go#L1:1&tab=references'
            )
        })
    })
})

describe('buildSearchURLQuery', () => {
    it('builds the URL query for a search', () => assert.strictEqual(buildSearchURLQuery('foo'), 'q=foo'))
    it('handles an empty query', () => assert.strictEqual(buildSearchURLQuery(''), 'q='))
    it('handles characters that need encoding', () =>
        assert.strictEqual(buildSearchURLQuery('foo bar%baz'), 'q=foo+bar%25baz'))
    it('preserves / and : for readability', () =>
        assert.strictEqual(buildSearchURLQuery('repo:foo/bar'), 'q=repo:foo/bar'))
})
