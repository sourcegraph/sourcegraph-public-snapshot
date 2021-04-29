import { BloomFilterFuzzySearch, allFuzzyParts, fuzzyMatchesQuery } from './BloomFilterFuzzySearch'

const all = [
    '.tsconfig.json',
    'to/the/moon.jpg',
    'lol/business.txt',
    'haha/business.txt',
    'business/crazy.txt',
    'fuzzy/business.txt',
    '.travis/workflows/config.json',
]

const fuzzy = BloomFilterFuzzySearch.fromSearchValues(all.map(f => ({ value: f })))

function checkSearch(query: string, expected: string[]) {
    test(`search-${query}`, () => {
        const actual = fuzzy.search({ value: query, maxResults: 1000 }).values.map(t => t.text)
        expect(actual).toStrictEqual(expected)
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
        obtained.forEach(pos => {
            parts.push(value.substring(pos.startOffset, pos.endOffset))
        })

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

checkSearch('h/bus', ['haha/business.txt'])
checkSearch('moon', ['to/the/moon.jpg'])
checkSearch('t/moon', ['to/the/moon.jpg'])
checkSearch('t/t/moon', ['to/the/moon.jpg'])
checkSearch('t.t.moon', [])
checkSearch('t t moon', [])
checkSearch('jpg', ['to/the/moon.jpg'])
checkSearch('t/m', ['to/the/moon.jpg'])
checkSearch('mo', ['to/the/moon.jpg'])
checkSearch('t', all)

checkFuzzyMatch('consume-delimeter-negative', 'ts/json', '.tsconfig.json', [])
checkFuzzyMatch('consume-delimeter-positive', 'ts/json', '.tsconfig/json', ['ts', '/', 'json'])
checkFuzzyMatch('consume-delimeter-end-of-word', 'ts/', '.tsconfig/json', ['ts', '/'])
checkFuzzyMatch('consume-delimeter-start-of-word', '.ts/', '.tsconfig/json', ['.', 'ts', '/'])

// TODO(olafurpg): treat all-lowercase queries as case insensitive
// checkSearch('readme', ["Documentation/README.md"])
