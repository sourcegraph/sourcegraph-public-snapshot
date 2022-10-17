import { CaseInsensitiveFuzzySearch } from './CaseInsensitiveFuzzySearch'

function fuzzyMatches(query: string, values: string[]): string[] {
    const fuzzy = new CaseInsensitiveFuzzySearch(
        values.map(value => ({ text: value })),
        undefined
    )
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
})
