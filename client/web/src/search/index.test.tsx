import { parseSearchURL, repoFilterForRepoRevision, resolveVersionContext } from '.'
import { SearchPatternType } from '../graphql-operations'

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
            versionContext: undefined,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
            versionContext: undefined,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
            versionContext: undefined,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=literal')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: true,
            versionContext: undefined,
        })

        expect(
            parseSearchURL(
                'q=TEST+repo:sourcegraph/sourcegraph+case:no+patternType:regexp&patternType=literal&case=yes'
            )
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  ',
            patternType: SearchPatternType.regexp,
            caseSensitive: false,
            versionContext: undefined,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=literal')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
            versionContext: undefined,
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
            versionContext: undefined,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=literal&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph ',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
            versionContext: undefined,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
            versionContext: undefined,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=literal', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph  case:yes',
            patternType: SearchPatternType.literal,
            caseSensitive: true,
            versionContext: undefined,
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
            versionContext: undefined,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=literal', { appendCaseFilter: true })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.literal,
            caseSensitive: false,
            versionContext: undefined,
        })
    })

    test('resolveVersionContext', () => {
        expect(
            resolveVersionContext('3.16', [
                { name: '3.16', description: '3.16', revisions: [{ rev: '3.16', repo: 'github.com/example/example' }] },
            ])
        ).toBe('3.16')
        expect(
            resolveVersionContext('3.15', [
                { name: '3.16', description: '3.16', revisions: [{ rev: '3.16', repo: 'github.com/example/example' }] },
            ])
        ).toBe(undefined)
        expect(resolveVersionContext('3.15', undefined)).toBe(undefined)
    })
})

describe('repoFilterForRepoRevision escapes values with spaces', () => {
    test('escapes spaces in value', () => {
        expect(repoFilterForRepoRevision('7 is my final answer', false)).toMatchInlineSnapshot(
            '"^7\\\\ is\\\\ my\\\\ final\\\\ answer$"'
        )
    })
})
