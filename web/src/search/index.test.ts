import { interactiveParseSearchURLQuery } from './index'

describe('interactiveParseSearchURLQuery', () => {
    test('parses the match query', () => {
        const query = '?q=hello+world'
        expect(interactiveParseSearchURLQuery(query)).toBe('hello world')
    })
    test('parses the match and repo query params', () => {
        const query = '?q=hello+world&repo=gorilla/mux'
        expect(interactiveParseSearchURLQuery(query)).toBe('repo:gorilla/mux hello world')
    })
    test('parses the match, repo, and file query params', () => {
        const query = '?q=hello+world&repo=gorilla/mux&file=test.tsx'
        expect(interactiveParseSearchURLQuery(query)).toBe('repo:gorilla/mux file:test.tsx hello world')
    })
})
