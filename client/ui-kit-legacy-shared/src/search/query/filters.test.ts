import { escapeSpaces, validateFilter } from './filters'
import { Literal, Quoted } from './token'

expect.addSnapshotSerializer({
    serialize: value => value as string,
    test: () => true,
})

describe('validateFilter()', () => {
    interface TestCase {
        description: string
        filterType: string
        expected: ReturnType<typeof validateFilter>
        token: Literal | Quoted
    }
    const range = { start: 0, end: 1 }
    const TESTCASES: TestCase[] = [
        {
            description: 'Valid repo filter',
            filterType: 'repo',
            expected: { valid: true },
            token: { type: 'literal', value: 'a', range },
        },
        {
            description: 'Valid repo filter - quoted value',
            filterType: 'repo',
            expected: { valid: true },
            token: { type: 'quoted', quotedValue: 'a', range },
        },
        {
            description: 'Valid repo filter - alias',
            filterType: 'repo',
            expected: { valid: true },
            token: { type: 'quoted', quotedValue: 'a', range },
        },
        {
            description: 'Invalid filter type',
            filterType: 'repoo',
            expected: { valid: false, reason: 'Invalid filter type.' },
            token: { type: 'literal', value: 'a', range },
        },
        {
            description: 'Valid case filter',
            filterType: 'case',
            expected: { valid: true },
            token: { type: 'literal', value: 'yes', range },
        },
        {
            description: 'Valid quoted value for case filter',
            filterType: 'case',
            expected: { valid: true },
            token: { type: 'quoted', quotedValue: 'yes', range },
        },
        {
            description: 'Invalid literal value for case filter',
            filterType: 'case',
            expected: { valid: false, reason: 'Invalid filter value, expected one of: yes, no.' },
            token: { type: 'literal', value: 'yess', range },
        },
        {
            description: 'Valid case-insensitive repo filter',
            filterType: 'RePo',
            expected: { valid: true },
            token: { type: 'literal', value: 'a', range },
        },
    ]

    for (const { description, filterType, expected, token } of TESTCASES) {
        test(description, () => {
            expect(validateFilter(filterType, token)).toStrictEqual(expected)
        })
    }
})

describe('escapeSpaces', () => {
    test('escapes spaces in value', () => {
        expect(escapeSpaces('contains a space')).toMatchInlineSnapshot('contains\\ a\\ space')
    })

    test('skips escaped values', () => {
        expect(escapeSpaces('\\\\already\\ escaped')).toMatchInlineSnapshot('\\\\already\\ escaped')
    })

    test('terminates with \\', () => {
        expect(escapeSpaces('last\\')).toMatchInlineSnapshot('last\\')
    })
})
