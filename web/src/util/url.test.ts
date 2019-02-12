import { lprToSelectionsZeroIndexed, parseBrowserRepoURL, toTreeURL } from './url'

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

const ctx = {
    repoName: 'github.com/gorilla/mux',
    rev: '',
    commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
    filePath: 'mux.go',
}

describe('toTreeURL', () => {
    test('formats url', () => {
        expect(toTreeURL(ctx)).toBe('/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/tree/mux.go')
    })

    // other cases are gratuitous given tests for other URL functions
})

describe('parseBrowserRepoURL', () => {
    test('should parse github repo', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
        })
    })
    test('should parse repo', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux')
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
        })
    })

    test('should parse github repo with rev', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@branch')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
        })
    })
    test('should parse repo with rev', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch')
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
            rev: 'branch',
        })
    })

    test('should parse github repo with multi-path-part rev', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@foo/baz/bar')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'foo/baz/bar',
        })
        const parsed2 = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@foo/baz/bar/-/blob/mux.go')
        assertDeepStrictEqual(parsed2, {
            repoName: 'github.com/gorilla/mux',
            rev: 'foo/baz/bar',
            filePath: 'mux.go',
        })
    })
    test('should parse repo with multi-path-part rev', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@foo/baz/bar')
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
            rev: 'foo/baz/bar',
        })
        const parsed2 = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@foo/baz/bar/-/blob/mux.go')
        assertDeepStrictEqual(parsed2, {
            repoName: 'gorilla/mux',
            rev: 'foo/baz/bar',
            filePath: 'mux.go',
        })
    })

    test('should parse github repo with commitID', () => {
        const parsed = parseBrowserRepoURL(
            'https://sourcegraph.com/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7'
        )
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        })
    })
    test('should parse repo with commitID', () => {
        const parsed = parseBrowserRepoURL(
            'https://sourcegraph.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7'
        )
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
            rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        })
    })

    test('should parse github repo with rev and file', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go')
        assertDeepStrictEqual(parsed, {
            repoName: 'github.com/gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
        })
    })
    test('should parse repo with rev and file', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch/-/blob/mux.go')
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
        })
    })

    test('should parse github repo with rev and file and line', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go#L3')
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
    test('should parse repo with rev and file and line', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch/-/blob/mux.go#L3')
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            position: {
                line: 3,
                character: 0,
            },
        })
    })

    test('should parse github repo with rev and file and position', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go#L3:5')
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
    test('should parse repo with rev and file and position', () => {
        const parsed = parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch/-/blob/mux.go#L3:5')
        assertDeepStrictEqual(parsed, {
            repoName: 'gorilla/mux',
            rev: 'branch',
            filePath: 'mux.go',
            position: {
                line: 3,
                character: 5,
            },
        })
    })
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
