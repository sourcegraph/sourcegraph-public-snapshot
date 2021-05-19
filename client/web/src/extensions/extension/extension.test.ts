import { extensionsQuery, urlToExtensionsQuery, validCategories } from './extension'

describe('validCategories', () => {
    test('selects only known categories and dedupes', () =>
        expect(validCategories(['Programming languages', 'Other', 'x', 'Other'])).toEqual([
            'Programming languages',
            'Other',
        ]))

    test('returns undefined for undefined', () => expect(validCategories(undefined)).toEqual(undefined))

    test('returns undefined for empty', () => expect(validCategories([])).toEqual(undefined))

    test('returns undefined when no categories are known', () => expect(validCategories(['x'])).toEqual(undefined))
})

describe('extensionsQuery', () => {
    test('tag (unquoted)', () => expect(extensionsQuery({ tag: 't' })).toBe('tag:t'))
    test('tag (quoted)', () => expect(extensionsQuery({ tag: 't t' })).toBe('tag:"t t"'))
    test('none', () => expect(extensionsQuery({})).toBe(''))
})

describe('urlToExtensionsQuery', () => {
    test('only query', () => expect(urlToExtensionsQuery({ query: 'foo bar' })).toBe('/extensions?query=foo+bar'))
    test('only category', () =>
        expect(urlToExtensionsQuery({ category: 'Linters' })).toBe('/extensions?category=Linters'))
    test('both query and category', () =>
        expect(urlToExtensionsQuery({ query: 'foo bar', category: 'Linters' })).toBe(
            '/extensions?query=foo+bar&category=Linters'
        ))
    test('neither query nor category', () => expect(urlToExtensionsQuery({ query: undefined })).toBe('/extensions'))
})
