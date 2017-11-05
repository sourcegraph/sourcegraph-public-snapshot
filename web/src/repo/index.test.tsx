import * as assert from 'assert'
import { JSDOM } from 'jsdom'
const pickBy = require('lodash/pickBy')
import '../util/polyfill'
import { makeRepoURI, parseBrowserRepoURL, parseRepoURI } from './index'

describe('repo module', () => {
    // remove undefined values from an object
    const compact = (obj: any) => pickBy(obj, (val: any) => val !== undefined)

    describe('parseRepoURI', () => {
        it('should parse repo', () => {
            const parsed = parseRepoURI('git://github.com/gorilla/mux')
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
            })
        })

        it('should parse repo with rev', () => {
            const parsed = parseRepoURI('git://github.com/gorilla/mux?branch')
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
            })
        })

        it('should parse repo with commitID', () => {
            const parsed = parseRepoURI('git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
        })

        it('should parse repo with rev and file', () => {
            const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go')
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
        })

        it('should parse repo with rev and file and line', () => {
            const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3')
            assert.deepEqual(compact(parsed), {
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
            const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3,5')
            assert.deepEqual(compact(parsed), {
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
            const parsed = parseRepoURI('git://github.com/gorilla/mux?branch#mux.go:3,5-6,9')
            assert.deepEqual(compact(parsed), {
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
        let jsdom: JSDOM
        before(() => {
            jsdom = new JSDOM('<!DOCTYPE html><body />')
        })

        it('should parse repo', () => {
            const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux', jsdom.window)
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
            })
        })

        it('should parse repo with rev', () => {
            const parsed = parseBrowserRepoURL('https://sourcegraph.com/github.com/gorilla/mux@branch', jsdom.window)
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
            })
        })

        it('should parse repo with multi-path-part rev', () => {
            const parsed = parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@foo/baz/bar',
                jsdom.window
            )
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: 'foo/baz/bar',
            })
            const parsed2 = parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@foo/baz/bar/-/blob/mux.go',
                jsdom.window
            )
            assert.deepEqual(compact(parsed2), {
                repoPath: 'github.com/gorilla/mux',
                rev: 'foo/baz/bar',
                filePath: 'mux.go',
            })
        })

        it('should parse repo with commitID', () => {
            const parsed = parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
                jsdom.window
            )
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
        })

        it('should parse repo with rev and file', () => {
            const parsed = parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go',
                jsdom.window
            )
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
        })

        it('should parse repo with rev and file and line', () => {
            const parsed = parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go#L3',
                jsdom.window
            )
            assert.deepEqual(compact(parsed), {
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
            const parsed = parseBrowserRepoURL(
                'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go#L3:5',
                jsdom.window
            )
            assert.deepEqual(compact(parsed), {
                repoPath: 'github.com/gorilla/mux',
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
            const uri = makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
            })
            assert.deepEqual(uri, 'git://github.com/gorilla/mux')
        })

        it('should make repo with rev', () => {
            const uri = makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
            })
            assert.deepEqual(uri, 'git://github.com/gorilla/mux?branch')
        })

        it('should make repo with commitID', () => {
            const uri = makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
            })
            assert.deepEqual(uri, 'git://github.com/gorilla/mux?24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
        })

        it('should make repo with rev and file', () => {
            const uri = makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
            })
            assert.deepEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go')
        })

        it('should make repo with rev and file and line', () => {
            const uri = makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 0,
                },
            })
            assert.deepEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go:3')
        })

        it('should make repo with rev and file and position', () => {
            const uri = makeRepoURI({
                repoPath: 'github.com/gorilla/mux',
                rev: 'branch',
                filePath: 'mux.go',
                position: {
                    line: 3,
                    character: 5,
                },
            })
            assert.deepEqual(uri, 'git://github.com/gorilla/mux?branch#mux.go:3,5')
        })
    })
})
