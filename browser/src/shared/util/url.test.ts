import { toAbsoluteBlobURL } from './url'

describe('toAbsoluteBlobURL', () => {
    const ctx = {
        repoName: 'github.com/gorilla/mux',
        rev: '',
        commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        filePath: 'mux.go',
    }
    // const cache = {
    //     [ctx.repoName]: 'https://sourcegraph.private.org',
    // }
    test('default sourcegraph URL, default context', () => {
        expect(toAbsoluteBlobURL(ctx)).toBe('https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go')
    })

    test('default sourcegraph URL, specified rev', () => {
        expect(toAbsoluteBlobURL({ ...ctx, rev: 'branch' })).toBe(
            'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go'
        )
    })

    test('default sourcegraph URL, with position', () => {
        expect(toAbsoluteBlobURL({ ...ctx, position: { line: 1, character: 1 } })).toBe(
            'https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go#L1:1'
        )
    })

    // test('sourcegraph URL from cache, default context', () => {
    //     expect(toAbsoluteBlobURL(ctx, cache)).toBe(
    //         'https://sourcegraph.private.org/github.com/gorilla/mux/-/blob/mux.go'
    //     )
    // })

    // test('sourcegraph URL from cache, specified rev', () => {
    //     expect(toAbsoluteBlobURL({ ...ctx, rev: 'branch' }, cache)).toBe(
    //         'https://sourcegraph.private.org/github.com/gorilla/mux@branch/-/blob/mux.go'
    //     )
    // })

    // test('sourcegraph URL from cache, with position', () => {
    //     expect(toAbsoluteBlobURL({ ...ctx, position: { line: 1, character: 1 } }, cache)).toBe(
    //         'https://sourcegraph.private.org/github.com/gorilla/mux/-/blob/mux.go#L1:1'
    //     )
    // })
})
