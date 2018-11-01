// Polyfill URLSearchParams because we still use node 8 on CI where
// URLSearchParams is not a global.
import { URLSearchParams } from 'whatwg-url'
Object.assign(global, { URLSearchParams })

import * as assert from 'assert'
import { parseHash, toPrettyBlobURL, toTreeURL } from './url'

describe('util/url', () => {
    const linePosition = { line: 1 }
    const lineCharPosition = { line: 1, character: 1 }
    const localRefMode = { ...lineCharPosition, viewState: 'references' }
    const externalRefMode = { ...lineCharPosition, viewState: 'references:external' }
    const ctx = {
        repoPath: 'github.com/gorilla/mux',
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
            assert.deepStrictEqual(parseHash('L1:1$references:external'), externalRefMode)
        })
        it('parses modern hash with external references', () => {
            assert.deepStrictEqual(parseHash('L1:1&tab=references:external'), externalRefMode)
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
                toPrettyBlobURL({ ...ctx, position: lineCharPosition, viewState: 'references:external' }),
                '/github.com/gorilla/mux/-/blob/mux.go#L1:1&tab=references:external'
            )
        })
    })

    describe('toTreeURL', () => {
        it('formats url', () => {
            assert.strictEqual(
                toTreeURL(ctx),
                '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/tree/mux.go'
            )
        })

        // other cases are gratuitous given tests for other URL functions
    })
})
