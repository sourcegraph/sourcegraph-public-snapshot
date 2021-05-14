import { BloomFilterFuzzySearch, allFuzzyParts, fuzzyMatchesQuery } from './BloomFilterFuzzySearch'

const all = [
    't1/README.md',
    't2/Readme.md',
    't1/READMES.md',
    '.tsconfig.json',
    'to/the/moon.jpg',
    'lol/business.txt',
    'haha/business.txt',
    't3/siteConfig.json',
    'business/crazy.txt',
    'fuzzy/business.txt',
    '.travis/workflows/config.json',
    'test/WorkspaceSymbolProvider.scala',
]

const fuzzy = BloomFilterFuzzySearch.fromSearchValues(all.map(text => ({ text })))

function checkSearch(query: string, expected: string[]) {
    test(`search-${query}`, () => {
        const queryProps = { value: query, maxResults: 1000 }
        const actual = fuzzy.search(queryProps).values.map(highlightedText => highlightedText.text)
        expect(actual).toStrictEqual(expected)
        for (const result of expected) {
            const individualFuzzy = BloomFilterFuzzySearch.fromSearchValues([{ text: result }])
            const individualActual = individualFuzzy
                .search(queryProps)
                .values.map(highlightedText => highlightedText.text)
            expect(individualActual).toStrictEqual([result])
        }
    })
}

function checkParts(name: string, original: string, expected: string[]) {
    test(`allFuzzyParts-${name}`, () => {
        expect(allFuzzyParts(original, false)).toStrictEqual(expected)
    })
}

function checkFuzzyMatch(name: string, query: string, value: string, expected: string[]) {
    test(`fuzzyMatchesQuery-${name}`, () => {
        const obtained = fuzzyMatchesQuery(query, value)
        const parts: string[] = []
        for (const position of obtained) {
            parts.push(value.slice(position.startOffset, position.endOffset))
        }
        expect(parts).toStrictEqual(expected)
    })
}
checkParts('basic', 'haha/business.txt', ['haha', 'business', 'txt'])
checkParts('snake_case', 'haha_business.txt', ['haha', 'business', 'txt'])
checkParts('camelCase', 'hahaBusiness.txt', ['haha', 'Business', 'txt'])
checkParts('CamelCase', 'HahaBusiness.txt', ['Haha', 'Business', 'txt'])
checkParts('kebab-case', 'haha-business.txt', ['haha', 'business', 'txt'])
checkParts('kebab-case', 'haha-business.txt', ['haha', 'business', 'txt'])
checkParts('dotfile', '.tsconfig.json', ['tsconfig', 'json'])
checkFuzzyMatch('dotfile', 'ts', '.tsconfig.json', ['ts'])

checkFuzzyMatch('basic', 'ha/busi', 'haha/business.txt', ['ha', '/', 'busi'])
checkFuzzyMatch('all-lowercase', 'readme', 't1/README.md', ['README'])
checkFuzzyMatch('all-lowercase2', 'readme', 't2/Readme.md', ['Readme'])

checkSearch('h/bus', ['haha/business.txt'])
checkSearch('moon', ['to/the/moon.jpg'])
checkSearch('t/moon', ['to/the/moon.jpg'])
checkSearch('t/t/moon', ['to/the/moon.jpg'])
checkSearch('t.t.moon', [])
checkSearch('t t moon', [])
checkSearch('jpg', ['to/the/moon.jpg'])
checkSearch('t/mo', ['to/the/moon.jpg'])
checkSearch('mo', ['to/the/moon.jpg'])
checkSearch('t', all)
checkSearch('readme', ['t1/README.md', 't2/Readme.md', 't1/READMES.md'])
checkSearch('README', ['t1/README.md', 't1/READMES.md'])
checkSearch('Readme', ['t2/Readme.md'])
checkSearch('WSProvider', ['test/WorkspaceSymbolProvider.scala'])
checkFuzzyMatch('digits', 't2', 't2/Readme.md', ['t2'])
checkSearch('t2', ['t2/Readme.md'])

// // Known caveat: the bloom filter fuzzy finder is quite strict with capitalization
checkSearch('sitecon', [])
checkFuzzyMatch('sitecon', 'sitecon', 'website/siteConfig.js', [])

checkFuzzyMatch('consume-delimeter-negative', 'ts/json', '.tsconfig.json', [])
checkFuzzyMatch('consume-delimeter-positive', 'ts/json', '.tsconfig/json', ['ts', '/', 'json'])
checkFuzzyMatch('consume-delimeter-end-of-word', 'ts/', '.tsconfig/json', ['ts', '/'])
checkFuzzyMatch('consume-delimeter-start-of-word', '.ts/', '.tsconfig/json', ['.', 'ts', '/'])
