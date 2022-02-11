import { createURLWithUTM } from './utm'

describe('toUTM', () => {
    it.each([
        [
            'adds all UTM markers',
            createURLWithUTM(new URL('https://example.com'), {
                utm_source: 'a',
                utm_campaign: 'b',
                utm_medium: 'c',
                utm_term: 'd',
                utm_content: 'e',
            }),
            'https://example.com/?utm_source=a&utm_campaign=b&utm_medium=c&utm_term=d&utm_content=e',
        ],
        [
            'encodes UTM values',
            createURLWithUTM(new URL('https://example.com/search'), {
                utm_source: 'something with a space and /',
            }),
            'https://example.com/search?utm_source=something+with+a+space+and+%2F',
        ],
        [
            'correctly handles existing query parameters and fragments',
            createURLWithUTM(new URL('https://example.com/search?foo=bar#baz=qux'), {
                utm_source: 'tadaa',
            }),
            'https://example.com/search?foo=bar&utm_source=tadaa#baz=qux',
        ],
        // eslint-disable-next-line id-length
    ])('%p', (_, expected, actual) => {
        expect(expected.toString()).toEqual(actual)
    })
})
