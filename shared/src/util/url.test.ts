import {
    buildSearchURLQuery,
    lprToSelectionsZeroIndexed,
    makeRepoURI,
    parseHash,
    parseRepoURI,
    toPrettyBlobURL,
    withWorkspaceRootInputRevision,
    interactiveBuildSearchURLQuery,
} from './url'
import { SearchPatternType } from '../graphql/schema'
import { SuggestionTypes } from '../search/suggestions/util'

/**
 * Asserts deep object equality using node's assert.deepEqual, except it (1) ignores differences in the
 * prototype (because that causes 2 object literals to fail the test) and (2) treats undefined properties as
 * missing.
 */
function assertDeepStrictEqual(actual: any, expected: any, message?: string): void {
    actual = JSON.parse(JSON.stringify(actual))
    expected = JSON.parse(JSON.stringify(expected))
    expect(actual).toEqual(expected)
}

describe('parseRepoURI', () => {
    test('should parse repo', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
        })
    })

    test('should parse repo with rev', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
        })
    })

    test('should parse repo with commitID', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        })
    })

    test('should parse repo with rev and file', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
        })
    })

    test('should parse repo with rev and file and line', () => {
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

    test('should parse repo with rev and file and position', () => {
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

    test('should parse repo with rev and file and range', () => {
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

    test('should parse a file with special characters', () => {
        const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#space%20here.go')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'space here.go',
        })
    })
})

describe('makeRepoURI', () => {
    test('should make repo', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux')
    })

    test('should make repo with rev', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch')
    })

    test('should make repo with commitID', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
    })

    test('should make repo with rev and file', () => {
        const uri = makeRepoURI({
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
        })
        assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go')
    })

    test('should make repo with rev and file and line', () => {
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

    test('should make repo with rev and file and position', () => {
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
    const refMode = { ...lineCharPosition, viewState: 'references' }
    const ctx = {
        repoName: 'github.com/gorilla/mux',
        rev: '',
        commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        filePath: 'mux.go',
    }

    describe('parseHash', () => {
        test('parses empty hash', () => {
            expect(parseHash('')).toEqual({})
        })

        test('parses unexpectedly formatted hash', () => {
            expect(parseHash('L-53')).toEqual({})
            expect(parseHash('L53:')).toEqual({})
            expect(parseHash('L1:2-')).toEqual({})
            expect(parseHash('L1:2-3')).toEqual({})
            expect(parseHash('L1:2-3:')).toEqual({})
            expect(parseHash('L1:-3:')).toEqual({})
            expect(parseHash('L1:-3:4')).toEqual({})
            expect(parseHash('L1-2:3')).toEqual({})
            expect(parseHash('L1-2:')).toEqual({})
            expect(parseHash('L1:-2')).toEqual({})
            expect(parseHash('L1:2--3:4')).toEqual({})
            expect(parseHash('L53:a')).toEqual({})
        })

        test('parses hash with leading octothorpe', () => {
            expect(parseHash('#L1')).toEqual(linePosition)
        })

        test('parses hash with line', () => {
            expect(parseHash('L1')).toEqual(linePosition)
        })

        test('parses hash with line and character', () => {
            expect(parseHash('L1:1')).toEqual(lineCharPosition)
        })

        test('parses hash with range', () => {
            expect(parseHash('L1-2')).toEqual({ line: 1, endLine: 2 })
            expect(parseHash('L1:2-3:4')).toEqual({ line: 1, character: 2, endLine: 3, endCharacter: 4 })
            expect(parseHash('L47-L55')).toEqual({ line: 47, endLine: 55 })
            expect(parseHash('L34:2-L38:3')).toEqual({ line: 34, character: 2, endLine: 38, endCharacter: 3 })
        })

        test('parses hash with references', () => {
            expect(parseHash('$references')).toEqual({ viewState: 'references' })
            expect(parseHash('L1:1$references')).toEqual(refMode)
            expect(parseHash('L1:1$references')).toEqual(refMode)
        })
        test('parses modern hash with references', () => {
            expect(parseHash('tab=references')).toEqual({ viewState: 'references' })
            expect(parseHash('L1:1&tab=references')).toEqual(refMode)
            expect(parseHash('L1:1&tab=references')).toEqual(refMode)
        })
    })

    describe('toPrettyBlobURL', () => {
        test('formats url for empty rev', () => {
            expect(toPrettyBlobURL(ctx)).toBe('/github.com/gorilla/mux/-/blob/mux.go')
        })

        test('formats url for specified rev', () => {
            expect(toPrettyBlobURL({ ...ctx, rev: 'branch' })).toBe('/github.com/gorilla/mux@branch/-/blob/mux.go')
        })

        test('formats url with position', () => {
            expect(toPrettyBlobURL({ ...ctx, position: lineCharPosition })).toBe(
                '/github.com/gorilla/mux/-/blob/mux.go#L1:1'
            )
        })

        test('formats url with view state', () => {
            expect(toPrettyBlobURL({ ...ctx, position: lineCharPosition, viewState: 'references' })).toBe(
                '/github.com/gorilla/mux/-/blob/mux.go#L1:1&tab=references'
            )
        })
    })
})

describe('withWorkspaceRootInputRevision', () => {
    test('uses input revision for URI inside root with input revision', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: 'v' }], parseRepoURI('git://r?c#f'))
        ).toEqual(parseRepoURI('git://r?v#f')))

    test('does not change URI outside root (different repoName)', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: 'v' }], parseRepoURI('git://r2?c#f'))
        ).toEqual(parseRepoURI('git://r2?c#f')))

    test('does not change URI outside root (different rev)', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: 'v' }], parseRepoURI('git://r?c2#f'))
        ).toEqual(parseRepoURI('git://r?c2#f')))

    test('uses empty string input revision (treats differently from undefined)', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: '' }], parseRepoURI('git://r?c#f'))
        ).toEqual({ ...parseRepoURI('git://r?c#f'), rev: '' }))

    test('does not change URI if root has undefined input revision', () =>
        expect(
            withWorkspaceRootInputRevision(
                [{ uri: 'git://r?c', inputRevision: undefined }],
                parseRepoURI('git://r?c#f')
            )
        ).toEqual(parseRepoURI('git://r?c#f')))
})

describe('buildSearchURLQuery', () => {
    it('builds the URL query for a regular expression search', () =>
        expect(buildSearchURLQuery('foo', SearchPatternType.regexp)).toBe('q=foo&patternType=regexp'))
    it('builds the URL query for a literal search', () =>
        expect(buildSearchURLQuery('foo', SearchPatternType.literal)).toBe('q=foo&patternType=literal'))
    it('handles an empty query', () =>
        expect(buildSearchURLQuery('', SearchPatternType.regexp)).toBe('q=&patternType=regexp'))
    it('handles characters that need encoding', () =>
        expect(buildSearchURLQuery('foo bar%baz', SearchPatternType.regexp)).toBe('q=foo+bar%25baz&patternType=regexp'))
    it('preserves / and : for readability', () =>
        expect(buildSearchURLQuery('repo:foo/bar', SearchPatternType.regexp)).toBe('q=repo:foo/bar&patternType=regexp'))
    it('overrides the patternType parameter if a patternType field exists in the query', () =>
        expect(buildSearchURLQuery('foo patternType:literal', SearchPatternType.regexp)).toBe(
            'q=foo+&patternType=literal'
        ))
})

describe('lprToSelectionsZeroIndexed', () => {
    test('converts an LPR with only a start line', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 5,
            }),
            [
                {
                    start: {
                        line: 4,
                        character: 0,
                    },
                    end: {
                        line: 4,
                        character: 0,
                    },
                    anchor: {
                        line: 4,
                        character: 0,
                    },
                    active: {
                        line: 4,
                        character: 0,
                    },
                    isReversed: false,
                },
            ]
        )
    })

    test('converts an LPR with a line and a character', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 5,
                character: 45,
            }),
            [
                {
                    start: {
                        line: 4,
                        character: 44,
                    },
                    end: {
                        line: 4,
                        character: 44,
                    },
                    anchor: {
                        line: 4,
                        character: 44,
                    },
                    active: {
                        line: 4,
                        character: 44,
                    },
                    isReversed: false,
                },
            ]
        )
    })

    test('converts an LPR with a start and end line', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 12,
                endLine: 15,
            }),
            [
                {
                    start: {
                        line: 11,
                        character: 0,
                    },
                    end: {
                        line: 14,
                        character: 0,
                    },
                    anchor: {
                        line: 11,
                        character: 0,
                    },
                    active: {
                        line: 14,
                        character: 0,
                    },
                    isReversed: false,
                },
            ]
        )
    })

    test('converts an LPR with a start and end line and characters', () => {
        assertDeepStrictEqual(
            lprToSelectionsZeroIndexed({
                line: 12,
                character: 30,
                endLine: 15,
                endCharacter: 60,
            }),
            [
                {
                    start: {
                        line: 11,
                        character: 29,
                    },
                    end: {
                        line: 14,
                        character: 59,
                    },
                    anchor: {
                        line: 11,
                        character: 29,
                    },
                    active: {
                        line: 14,
                        character: 59,
                    },
                    isReversed: false,
                },
            ]
        )
    })
})

describe('interactiveBuildSearchURLQuery', () => {
    test('builds search URL query correctly with no filter inputs', () => {
        const result = new URLSearchParams()
        expect(interactiveBuildSearchURLQuery({}).toString()).toBe(result.toString())
    })

    test('builds search URL query correctly with a repo filter input', () => {
        const result = new URLSearchParams('repo=gorilla/mux')
        expect(
            interactiveBuildSearchURLQuery({
                'repo 1': { type: SuggestionTypes.repo, value: 'gorilla/mux', editable: false },
            }).toString()
        ).toBe(result.toString())
    })

    test('builds search URL query correctly with multiple repo filter inputs', () => {
        const result = new URLSearchParams('repo=gorilla/mux&repo=gorilla/muxy')
        expect(
            interactiveBuildSearchURLQuery({
                'repo 1': { type: SuggestionTypes.repo, value: 'gorilla/mux', editable: false },
                'repo 2': { type: SuggestionTypes.repo, value: 'gorilla/muxy', editable: false },
            }).toString()
        ).toBe(result.toString())
    })

    test('builds search URL query correctly with repo and file filter inputs', () => {
        const result = new URLSearchParams('repo=gorilla/mux&file=test')
        expect(
            interactiveBuildSearchURLQuery({
                'repo 1': { type: SuggestionTypes.repo, value: 'gorilla/mux', editable: false },
                'file 2': { type: SuggestionTypes.file, value: 'test', editable: false },
            }).toString()
        ).toBe(result.toString())
    })

    test('builds search URL query correctly with repo and file filter inputs, and a repo filter in the navbar query', () => {
        const result = new URLSearchParams('repo=gorilla/mux&file=test')
        expect(
            interactiveBuildSearchURLQuery({
                'repo 1': { type: SuggestionTypes.repo, value: 'gorilla/mux', editable: false },
                'file 2': { type: SuggestionTypes.file, value: 'test', editable: false },
            }).toString()
        ).toBe(result.toString())
    })
})
