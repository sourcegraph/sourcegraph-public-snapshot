import { validateFilter } from './filters'
import { Literal, Quoted } from './parser'

describe('validateFilter()', () => {
    interface TestCase {
        description: string
        filterType: string
        expected: ReturnType<typeof validateFilter>
        token: Literal | Quoted
    }
    const TESTCASES: TestCase[] = [
        {
            description: 'Valid repo filter',
            filterType: 'repo',
            expected: { valid: true },
            token: { type: 'literal', value: 'a' },
        },
        {
            description: 'Valid repo filter - quoted value',
            filterType: 'repo',
            expected: { valid: true },
            token: { type: 'quoted', quotedValue: 'a' },
        },
        {
            description: 'Valid repo filter - alias',
            filterType: 'repo',
            expected: { valid: true },
            token: { type: 'quoted', quotedValue: 'a' },
        },
        {
            description: 'Invalid filter type',
            filterType: 'repoo',
            expected: { valid: false, reason: 'Invalid filter type' },
            token: { type: 'literal', value: 'a' },
        },
        {
            description: 'Valid case filter',
            filterType: 'case',
            expected: { valid: true },
            token: { type: 'literal', value: 'yes' },
        },
        {
            description: 'Invalid quoted value for case filter',
            filterType: 'case',
            expected: { valid: false, reason: 'Invalid filter value, expected one of: yes, no' },
            token: { type: 'quoted', quotedValue: 'yes' },
        },
        {
            description: 'Invalid literal value for case filter',
            filterType: 'case',
            expected: { valid: false, reason: 'Invalid filter value, expected one of: yes, no' },
            token: { type: 'literal', value: 'yess' },
        },
    ]

    for (const { description, filterType, expected, token } of TESTCASES) {
        test(description, () => {
            expect(validateFilter(filterType, { token, range: { start: 0, end: 1 } })).toStrictEqual(expected)
        })
    }
})
