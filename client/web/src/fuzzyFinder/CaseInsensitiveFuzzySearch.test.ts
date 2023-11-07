import { describe, expect, test } from '@jest/globals'

import { CaseInsensitiveFuzzySearch } from './CaseInsensitiveFuzzySearch'

function fuzzyMatches(query: string, values: string[]): string[] {
    const fuzzy = new CaseInsensitiveFuzzySearch(
        values.map(value => ({ text: value })),
        undefined
    )

    // Reproduce a real-world scenario where the user types one character at a
    // time.  CaseInsensitiveFuzzySearch has an internal cache that may contain
    // bugs that don't reproduce when it only gets one `search()` request.
    for (let index = 1; index < query.length; index++) {
        fuzzy.search({ query: query.slice(0, index), maxResults: 100 })
    }

    const results = fuzzy.search({ query, maxResults: 100 })
    return results.links.map(link => link.text)
}

function checkFuzzyMatches(name: string, query: string, inputs: string[], expected: string[]): void {
    test(`fuzzyMatches-${name}`, () => {
        expect(fuzzyMatches(query, inputs)).toStrictEqual(expected)
    })
}

describe('case-insensitive fuzzy search', () => {
    checkFuzzyMatches('no-separator', 'hahabusiness', ['haha/business.txt', 'crab/cakes.png'], ['haha/business.txt'])
    checkFuzzyMatches(
        'exact-match-first',
        'executor.go',
        ['executor/batch.go', 'batches/executor.go', 'ignore.me'],
        ['batches/executor.go', 'executor/batch.go']
    )
    checkFuzzyMatches(
        'space-union',
        'executor batches',
        ['executor/batch.go', 'batches/executor.go', 'ignore.me'],
        ['batches/executor.go']
    )
    checkFuzzyMatches('exact-match', 'src/hello.ts', ['src/hello.ts', 'ignore.me'], ['src/hello.ts'])
    checkFuzzyMatches('no-smart-case', 'getSlocEntry', ['getSLocEntry'], ['getSLocEntry'])

    // Buggy cache test. Previously, the cached list of candidates only included
    // results that had a fuzzy score above 0.2 causing the 'a' query to filter
    // out a lot of strings containing the character 'a' because their score was
    // below 0.2.
    checkFuzzyMatches('buggy-cache', 'attr', ['.gitattributes', 'aaa24'], ['.gitattributes'])
})
