import { createBrowserHistory, History, Location } from 'history'
import { of, Subscription, Observable } from 'rxjs'
import { first, startWith, tap, last } from 'rxjs/operators'

import { resetAllMemoizationCaches } from '@sourcegraph/common'

import { SearchPatternType } from '../graphql-operations'
import { useNavbarQueryState } from '../stores'
import { observeLocation } from '../util/location'

import { parseSearchURL, repoFilterForRepoRevision, updateQueryStateFromLocation } from '.'

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
    const showSearchContext = of(false)

    it('should update patternType if different between URL and context', () => {
        const [location] = createHistoryObservable('q=r:golang/oauth2+test+f:travis&patternType=regexp')
        useNavbarQueryState.setState({ searchPatternType: SearchPatternType.standard })

        return updateQueryStateFromLocation({
            location: location.pipe(first()),
            isSearchContextAvailable,
            showSearchContext,
        })
            .pipe(
                last(),
                tap(() => {
                    expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.regexp)
                })
            )
            .toPromise()
    })

    it('should not update patternType if query is empty', () => {
        const [location] = createHistoryObservable('q=&patternType=regexp')
        useNavbarQueryState.setState({ searchPatternType: SearchPatternType.standard })

        return updateQueryStateFromLocation({
            location: location.pipe(first()),
            isSearchContextAvailable,
            showSearchContext,
        })
            .pipe(
                last(),
                tap(() => {
                    expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.standard)
                })
            )
            .toPromise()
    })

    it('should update caseSensitive if different between URL and context', () => {
        const [location] = createHistoryObservable('q=r:golang/oauth2+test+f:travis case:yes')
        useNavbarQueryState.setState({ searchCaseSensitivity: false })

        return updateQueryStateFromLocation({
            location: location.pipe(first()),
            isSearchContextAvailable,
            showSearchContext,
        })
            .pipe(
                last(),
                tap(() => {
                    expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(true)
                })
            )
            .toPromise()
    })

    it('should not update caseSensitive from filter if query is empty', () => {
        const [location] = createHistoryObservable('q=case:yes')
        useNavbarQueryState.setState({ searchCaseSensitivity: false })

        return updateQueryStateFromLocation({
            location: location.pipe(first()),
            isSearchContextAvailable,
            showSearchContext,
        })
            .pipe(
                last(),
                tap(() => {
                    expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(false)
                })
            )
            .toPromise()
    })

    describe('search context', () => {
        it('should extract the search context from the query', () => {
            const [location] = createHistoryObservable('q=context:me+test')

            return updateQueryStateFromLocation({
                location: location.pipe(first()),
                isSearchContextAvailable,
                showSearchContext,
            })
                .pipe(
                    last(),
                    tap(({ searchContextSpec }) => {
                        expect(searchContextSpec).toEqual('me')
                    })
                )
                .toPromise()
        })

        it('remove the context filter from the URL if search contexts are enabled and available', () => {
            const [location] = createHistoryObservable('q=context:me+test')

            return updateQueryStateFromLocation({
                location: location.pipe(first()),
                isSearchContextAvailable: () => Promise.resolve(true),
                showSearchContext: of(true),
            })
                .pipe(
                    last(),
                    tap(() => {
                        expect(useNavbarQueryState.getState().queryState.query).toBe('test')
                    })
                )
                .toPromise()
        })

        it('should not remove the context filter from the URL if search context is not available', () => {
            const [location] = createHistoryObservable('q=context:me+test')

            return updateQueryStateFromLocation({
                location: location.pipe(first()),
                showSearchContext: of(true),
                isSearchContextAvailable: () => Promise.resolve(false),
            })
                .pipe(
                    last(),
                    tap(() => {
                        expect(useNavbarQueryState.getState().queryState.query).toBe('context:me test')
                    })
                )
                .toPromise()
        })

        it('should not remove the context filter from the URL if search contexts are disabled', () => {
            const [location] = createHistoryObservable('q=context:me+test')

            return updateQueryStateFromLocation({
                location: location.pipe(first()),
                showSearchContext: of(false),
                isSearchContextAvailable: () => Promise.resolve(true),
            })
                .pipe(
                    last(),
                    tap(() => {
                        expect(useNavbarQueryState.getState().queryState.query).toBe('context:me test')
                    })
                )
                .toPromise()
        })
    })
})
