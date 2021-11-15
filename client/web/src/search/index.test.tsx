import { SearchPatternType } from '../graphql-operations'

import { parseSearchURL, repoFilterForRepoRevision } from '.'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

describe('search/index', () => {
    test('parseSearchURL', () => {
        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: true,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=literal')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: true,
        })

        expect(
            parseSearchURL(
                'q=TEST+repo:sourcegraph/sourcegraph+case:no+patternType:regexp&patternType=literal&case=yes'
            )
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  ',
            patternType: SearchPatternType.regexp,
            caseSensitive: false,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=literal')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
        })
    })

    test('parseSearchURL with appendCaseFilter', () => {
        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=literal&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.literal,
            caseSensitive: true,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=literal&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=literal', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.literal,
            caseSensitive: true,
        })

        expect(
            parseSearchURL(
                'q=TEST+repo:sourcegraph/sourcegraph+case:no+patternType:regexp&patternType=literal&case=yes',
                { appendCaseFilter: true }
            )
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  ',
            patternType: SearchPatternType.regexp,
            caseSensitive: false,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=literal', { appendCaseFilter: true })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
        })
    })
})

describe('repoFilterForRepoRevision escapes values with spaces', () => {
    test('escapes spaces in value', () => {
        expect(repoFilterForRepoRevision('7 is my final answer', false)).toMatchInlineSnapshot(
            '"^7\\\\ is\\\\ my\\\\ final\\\\ answer$"'
        )
    })
})
