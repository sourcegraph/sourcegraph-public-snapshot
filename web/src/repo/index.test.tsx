import * as assert from 'assert'
import * as _repo from '.'
import { createTestBundle } from '../util/unit-test-utils'

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

describe('repo/index', () => {
    let repo: typeof _repo

    before(async function(): Promise<void> {
        this.timeout(20000)
        repo = (await createTestBundle(__dirname + '/index.tsx')).load().module
    })

    describe('parseRepoURI', () => {
        it('should parse repo', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
            })
        })

        it('should parse repo with rev', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux?branch')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
            })
        })

        it('should parse repo with commitID', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
        })

        it('should parse repo with rev and file', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux?branch#mux.go')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
        })

        it('should parse repo with rev and file and line', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 0,
                },
            })
        })

        it('should parse repo with rev and file and position', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3,5')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 5,
                },
            })
        })

        it('should parse repo with rev and file and range', () => {
            const parsed = repo.parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3,5-6,9')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
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

    describe('parseBrowserRepoURL', () => {
        it('should parse github repo', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
            })
        })
        it('should parse repo', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux')
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
            })
        })

        it('should parse github repo with rev', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@branch')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
            })
        })
        it('should parse repo with rev', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch')
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
                rev: 'branch',
            })
        })

        it('should parse github repo with multi-path-part rev', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@foo/baz/bar')
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'foo/baz/bar',
            })
            const parsed2 = repo.parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@foo/baz/bar/-/blob/mux.go'
            )
            assertDeepStrictEqual(parsed2, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'foo/baz/bar',
                filePath: 'mux.go',
            })
        })
        it('should parse repo with multi-path-part rev', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@foo/baz/bar')
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
                rev: 'foo/baz/bar',
            })
            const parsed2 = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@foo/baz/bar/-/blob/mux.go')
            assertDeepStrictEqual(parsed2, {
                repoPath: 'gorilla/mux',
                rev: 'foo/baz/bar',
                filePath: 'mux.go',
            })
        })

        it('should parse github repo with commitID', () => {
            const parsed = repo.parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7'
            )
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
        })
        it('should parse repo with commitID', () => {
            const parsed = repo.parseBrowserRepoURL(
                'https://sourcegraph.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7'
            )
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
                rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
        })

        it('should parse github repo with rev and file', () => {
            const parsed = repo.parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go'
            )
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
        })
        it('should parse repo with rev and file', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch/-/blob/mux.go')
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
        })

        it('should parse github repo with rev and file and line', () => {
            const parsed = repo.parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go#L3'
            )
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 0,
                },
            })
        })
        it('should parse repo with rev and file and line', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch/-/blob/mux.go#L3')
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 0,
                },
            })
        })

        it('should parse github repo with rev and file and position', () => {
            const parsed = repo.parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go#L3:5'
            )
            assertDeepStrictEqual(parsed, {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 5,
                },
            })
        })
        it('should parse repo with rev and file and position', () => {
            const parsed = repo.parseBrowserRepoURL('https://sourcegraph.com/gorilla/mux@branch/-/blob/mux.go#L3:5')
            assertDeepStrictEqual(parsed, {
                repoPath: 'gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 5,
                },
            })
        })
    })

    describe('makeRepoURI', () => {
        it('should make repo', () => {
            const uri = repo.makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
            })
            assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux')
        })

        it('should make repo with rev', () => {
            const uri = repo.makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
            })
            assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch')
        })

        it('should make repo with commitID', () => {
            const uri = repo.makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
            assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
        })

        it('should make repo with rev and file', () => {
            const uri = repo.makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
            assertDeepStrictEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go')
        })

        it('should make repo with rev and file and line', () => {
            const uri = repo.makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
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
            const uri = repo.makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
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
})
