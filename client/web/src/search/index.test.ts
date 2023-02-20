import { createBrowserHistory, History, Location } from 'history'
import { Subscription, Observable } from 'rxjs'
import { first, startWith, tap, last } from 'rxjs/operators'

import { resetAllMemoizationCaches } from '@sourcegraph/common'
import { SearchMode } from '@sourcegraph/shared/src/search'

import { SearchPatternType } from '../graphql-operations'
import { observeLocation } from '../util/location'

import { parseSearchURL, repoFilterForRepoRevision, getQueryStateFromLocation } from '.'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

describe('search/index', () => {
    test('parseSearchURL', () => {
        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: true,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=standard&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes')
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
            searchMode: SearchMode.Precise,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: true,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL(
                'q=TEST+repo:sourcegraph/sourcegraph+case:no+patternType:regexp&patternType=literal&case=yes'
            )
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.regexp,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })

        expect(parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=standard')).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })
    })

    test('parseSearchURL with appendCaseFilter', () => {
        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph case:yes',
            patternType: SearchPatternType.standard,
            caseSensitive: true,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:no&patternType=standard&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+patternType:regexp&patternType=literal&case=yes', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph case:yes',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph+case:yes&patternType=standard', {
                appendCaseFilter: true,
            })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph case:yes',
            patternType: SearchPatternType.standard,
            caseSensitive: true,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL(
                'q=TEST+repo:sourcegraph/sourcegraph+case:no+patternType:regexp&patternType=literal&case=yes',
                { appendCaseFilter: true }
            )
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.regexp,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })

        expect(
            parseSearchURL('q=TEST+repo:sourcegraph/sourcegraph&patternType=standard', { appendCaseFilter: true })
        ).toStrictEqual({
            query: 'TEST repo:sourcegraph/sourcegraph',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })
    })

    test('parseSearchURL preserves literal search compatibility', () => {
        expect(parseSearchURL('q=/a literal pattern/&patternType=literal')).toStrictEqual({
            query: 'content:"/a literal pattern/"',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })

        expect(parseSearchURL('q=not /a literal pattern/&patternType=literal')).toStrictEqual({
            query: 'not content:"/a literal pattern/"',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
        })

        expect(parseSearchURL('q=un.*touched&patternType=literal')).toStrictEqual({
            query: 'un.*touched',
            patternType: SearchPatternType.standard,
            caseSensitive: false,
            searchMode: SearchMode.Precise,
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

describe('updateQueryStateFromURL', () => {
    let subscription: Subscription

    beforeEach(() => {
        subscription = new Subscription()
    })

    afterEach(() => {
        subscription.unsubscribe()
        // Ugly implementation detail
        resetAllMemoizationCaches()
    })

    function createHistoryObservable(search: string): [Observable<Location>, History] {
        const history = createBrowserHistory()
        history.replace({ search })

        return [observeLocation(history).pipe(startWith(history.location)), history]
    }

    const isSearchContextAvailable = () => Promise.resolve(true)

    describe('search context', () => {
        it('should extract the search context from the query', () => {
            const [location] = createHistoryObservable('q=context:me+test')

            return getQueryStateFromLocation({
                location: location.pipe(first()),
                isSearchContextAvailable,
            })
                .pipe(
                    last(),
                    tap(({ searchContextSpec, query }) => {
                        expect(searchContextSpec?.spec).toEqual('me')
                        expect(query).toEqual('context:me test')
                    })
                )
                .toPromise()
        })
    })
})
