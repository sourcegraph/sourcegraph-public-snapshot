import { DEFAULT_SOURCEGRAPH_URL } from './context'
import { toAbsoluteBlobURL } from './url'

describe('toAbsoluteBlobURL', () => {
    const ctx = {
        repoName: 'github.com/gorilla/mux',
        rev: '',
        commitID: '24fca303ac6da784b9e8269f724ddeb0b2eea5e7',
        filePath: 'mux.go',
    }

    test('default sourcegraph URL, default context', () => {
        expect(toAbsoluteBlobURL(DEFAULT_SOURCEGRAPH_URL, ctx)).toBe(
            'https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go'
        )
    })

    test('default sourcegraph URL, specified rev', () => {
        expect(toAbsoluteBlobURL(DEFAULT_SOURCEGRAPH_URL, { ...ctx, rev: 'branch' })).toBe(
            'https://sourcegraph.com/github.com/gorilla/mux@branch/-/blob/mux.go'
        )
    })

    test('default sourcegraph URL, with position', () => {
        expect(toAbsoluteBlobURL(DEFAULT_SOURCEGRAPH_URL, { ...ctx, position: { line: 1, character: 1 } })).toBe(
            'https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go#L1:1'
        )
    })
})
