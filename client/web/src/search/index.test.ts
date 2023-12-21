import { type Location, createPath } from 'react-router-dom'
import { Subscription, Subject } from 'rxjs'
import { tap, last } from 'rxjs/operators'
import { afterEach, beforeEach, describe, expect, it, test } from 'vitest'

import { logger, resetAllMemoizationCaches } from '@sourcegraph/common'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { createBarrier } from '@sourcegraph/testing'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SearchPatternType } from '../graphql-operations'

import {
    parseSearchURL,
    repoFilterForRepoRevision,
    getQueryStateFromLocation,
    searchQueryForRepoRevision,
    fileFilterForFilePath,
} from '.'

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
        expect(repoFilterForRepoRevision('7 is my final answer')).toMatchInlineSnapshot(
            '"^7\\\\ is\\\\ my\\\\ final\\\\ answer$"'
        )
    })
})

describe('searchQueryForRepoRevision', () => {
    test('respects pattern type', () => {
        expect(searchQueryForRepoRevision('foo bar', undefined, SearchPatternType.standard)).toStrictEqual(
            'repo:^foo\\ bar$ '
        )
        expect(searchQueryForRepoRevision('foo bar', undefined, SearchPatternType.newStandardRC1)).toStrictEqual(
            'repo:"foo bar" '
        )
        expect(searchQueryForRepoRevision('foo bar', '1ef3b', SearchPatternType.newStandardRC1)).toStrictEqual(
            'repo:"foo bar"@1ef3b '
        )
        expect(searchQueryForRepoRevision('foobar', '1ef3b', SearchPatternType.newStandardRC1)).toStrictEqual(
            'repo:foobar@1ef3b '
        )
    })
})

describe('filterForFilePath', () => {
    test('respects pattern type', () => {
        expect(fileFilterForFilePath('foo/bar/never panic', SearchPatternType.standard)).toStrictEqual(
            'file:^foo/bar/never\\ panic'
        )
        expect(fileFilterForFilePath('foo/bar/never panic', SearchPatternType.newStandardRC1)).toStrictEqual(
            'file:"foo/bar/never panic"'
        )
        expect(fileFilterForFilePath('foo/bar/never_panic', SearchPatternType.newStandardRC1)).toStrictEqual(
            'file:foo/bar/never_panic'
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

    function createHistoryObservable(search: string): [Subject<Location>, Location] {
        const { locationRef } = renderWithBrandedContext(null, { route: createPath({ search }) })
        const locationSubject = new Subject<Location>()

        return [locationSubject, locationRef.current!]
    }

    const isSearchContextAvailable = () => Promise.resolve(true)

    describe('search context', () => {
        it('should extract the search context from the query', async () => {
            const { wait, done } = createBarrier()
            const [locationSubject, location] = createHistoryObservable('q=context:me+test')

            getQueryStateFromLocation({
                location: locationSubject,
                isSearchContextAvailable,
            })
                .pipe(
                    last(),
                    tap(({ searchContextSpec, query }) => {
                        expect(searchContextSpec?.spec).toEqual('me')
                        expect(query).toEqual('context:me test')
                        done()
                    })
                )
                .toPromise()
                .catch(logger.error)

            locationSubject.next(location)
            locationSubject.complete()
            await wait
        })
    })
})
