import { SearchPatternType } from '../graphql-operations'

import { parseSearchURL, repoFilterForRepoRevision } from '.'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

describe('search/index', () => {
    test('parseSearchURL', () => {
        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.standard,
            caseSensitive: true,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=standard&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.standard,
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

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=standard')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
        })
    })

    test('parseSearchURL with appendCaseFilter', () => {
        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.standard,
            caseSensitive: true,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=standard&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.standard,
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
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.standard,
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
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=standard', { appendCaseFilter: true })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
        })
    })

    test('parseSearchURL preserves literal search compatibility', () => {
        expect(parseSearchURL('q=/a literal pattern/&patternType=literal')).toStrictEqual({
            query: 'content:"/a literal pattern/"',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
        })

        expect(parseSearchURL('q=not /a literal pattern/&patternType=literal')).toStrictEqual({
            query: 'not content:"/a literal pattern/"',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
        })

        expect(parseSearchURL('q=un.*touched&patternType=literal')).toStrictEqual({
            query: 'un.*touched',
            patternType: SearchPatternType.standard,
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
