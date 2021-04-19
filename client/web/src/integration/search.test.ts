import expect from 'expect'
import { test } from 'mocha'
import { Key } from 'ts-key-enum'

import { SharedGraphQlOperations, SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import {
    RepoGroupsResult,
    SearchResult,
    SearchSuggestionsResult,
    WebGraphQlOperations,
    AutoDefinedSearchContextsResult,
} from '../graphql-operations'
import { SearchEvent } from '../search/stream'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { createJsContext, siteGQLID, siteID } from './jscontext'
import {
    commitHighlightResult,
    commitSearchStreamEvents,
    diffSearchStreamEvents,
    diffHighlightResult,
    mixedSearchStreamEvents,
    highlightFileResult,
    symbolSearchStreamEvents,
} from './streaming-search-mocks'
import { percySnapshotWithVariants } from './utils'

const searchResults = (): SearchResult => ({
    search: {
        results: {
            __typename: 'SearchResults',
            limitHit: true,
            matchCount: 30,
            approximateResultCount: '30+',
            missing: [],
            cloning: [],
            repositoriesCount: 372,
            timedout: [],
            indexUnavailable: false,
            dynamicFilters: [
                {
                    value: 'archived:yes',
                    label: 'archived:yes',
                    count: 5,
                    limitHit: true,
                    kind: 'repo',
                },
                {
                    value: 'fork:yes',
                    label: 'fork:yes',
                    count: 46,
                    limitHit: true,
                    kind: 'repo',
                },
                {
                    value: 'repo:^github\\.com/Algorilla/manta-ray$',
                    label: 'github.com/Algorilla/manta-ray',
                    count: 1,
                    limitHit: false,
                    kind: 'repo',
                },
            ],
            results: [
                {
                    __typename: 'Repository',
                    id: 'UmVwb3NpdG9yeTozODcxOTM4Nw==',
                    name: 'github.com/Algorilla/manta-ray',
                    label: {
                        html:
                            '\u003Cp\u003E\u003Ca href="/github.com/Algorilla/manta-ray" rel="nofollow"\u003Egithub.com/Algorilla/manta-ray\u003C/a\u003E\u003C/p\u003E\n',
                    },
                    url: '/github.com/Algorilla/manta-ray',
                    detail: { html: '\u003Cp\u003ERepository match\u003C/p\u003E\n' },
                    matches: [],
                },
            ],
            alert: null,
            elapsedMilliseconds: 103,
        },
    },
})

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    Search: searchResults,
    SearchSuggestions: (): SearchSuggestionsResult => ({
        search: {
            suggestions: [],
        },
    }),
    RepoGroups: (): RepoGroupsResult => ({
        repoGroups: [],
    }),
    AutoDefinedSearchContexts: (): AutoDefinedSearchContextsResult => ({
        autoDefinedSearchContexts: [],
    }),
    ConvertVersionContextToSearchContext: ({ name }) => ({
        convertVersionContextToSearchContext: { id: `id${name}`, spec: name },
    }),
}

describe('Search', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const waitAndFocusInput = async () => {
        await driver.page.waitForSelector('.monaco-editor .view-lines')
        await driver.page.click('.monaco-editor .view-lines')
    }

    const getSearchFieldValue = (driver: Driver): Promise<string | undefined> =>
        driver.page.evaluate(() => document.querySelector<HTMLTextAreaElement>('#monaco-query-input textarea')?.value)

    test('Styled correctly on results page', async () => {
        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=foo')
        await driver.page.waitForSelector('#monaco-query-input')
        await percySnapshotWithVariants(driver.page, 'Search results page')
    })

    describe('Search filters', () => {
        test('Search filters are shown on search result pages and clicking them triggers a new search', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            const dynamicFilters = ['archived:yes', 'repo:^github\\.com/Algorilla/manta-ray$']
            const origQuery = 'foo'
            for (const filter of dynamicFilters) {
                await driver.page.goto(
                    `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal`
                )
                await driver.page.waitForSelector(`[data-testid="filter-chip"][value=${JSON.stringify(filter)}]`)
                await driver.page.click(`[data-testid="filter-chip"][value=${JSON.stringify(filter)}]`)
                await driver.page.waitForFunction(
                    expectedQuery => {
                        const url = new URL(document.location.href)
                        const query = url.searchParams.get('q')
                        return query && query.trim() === expectedQuery
                    },
                    { timeout: 5000 },
                    `${origQuery} ${filter}`
                )
            }
        })
    })

    describe('Filter completion', () => {
        test('Completing a negated filter should insert the filter with - prefix', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                SearchSuggestions: () => ({
                    search: {
                        suggestions: [],
                    },
                }),
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('#monaco-query-input')
            await driver.replaceText({
                selector: '#monaco-query-input',
                newText: '-repo',
                enterTextMethod: 'type',
            })
            await driver.page.waitForSelector('.monaco-query-input-container .suggest-widget.visible')
            await driver.findElementWithText('-repo', {
                action: 'click',
                wait: { timeout: 5000 },
                selector: '.monaco-query-input-container .suggest-widget.visible span',
            })
            expect(await getSearchFieldValue(driver)).toStrictEqual('-repo:')
        })
    })

    describe('Suggestions', () => {
        test('Typing in the search field shows relevant suggestions', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                SearchSuggestions: () => ({
                    search: {
                        suggestions: [
                            { __typename: 'Repository', name: 'github.com/auth0/go-jwt-middleware' },
                            {
                                __typename: 'Symbol',
                                name: 'OnError',
                                containerName: 'jwtmiddleware',
                                url: '/github.com/auth0/go-jwt-middleware/-/blob/jwtmiddleware.go#L56:1-56:14',
                                kind: SymbolKind.STRUCT,
                                location: {
                                    resource: {
                                        path: 'jwtmiddleware.go',
                                        repository: { name: 'github.com/auth0/go-jwt-middleware' },
                                    },
                                },
                            },
                            {
                                __typename: 'File',
                                path: 'jwtmiddleware.go',
                                name: 'jwtmiddleware.go',
                                isDirectory: false,
                                url: '/github.com/auth0/go-jwt-middleware/-/blob/jwtmiddleware.go',
                                repository: { name: 'github.com/auth0/go-jwt-middleware' },
                            },
                        ],
                    },
                }),
            })
            // Repo autocomplete from homepage
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            // Using id selector rather than `test-` classes as Monaco doesn't allow customizing classes
            await driver.page.waitForSelector('#monaco-query-input')
            await driver.replaceText({
                selector: '#monaco-query-input',
                newText: 'go-jwt-middlew',
                enterTextMethod: 'type',
            })
            await driver.page.waitForSelector('.monaco-query-input-container .suggest-widget.visible')
            await driver.findElementWithText('github.com/auth0/go-jwt-middleware', {
                action: 'click',
                wait: { timeout: 5000 },
                selector: '.monaco-query-input-container .suggest-widget.visible span',
            })
            expect(await getSearchFieldValue(driver)).toStrictEqual('repo:^github\\.com/auth0/go-jwt-middleware$ ')

            // Submit search
            await driver.page.keyboard.press(Key.Enter)

            // File autocomplete from repo search bar
            await driver.page.waitForSelector('#monaco-query-input')
            await driver.page.focus('#monaco-query-input')
            await driver.page.keyboard.type('jwtmi')
            await driver.page.waitForSelector('.monaco-query-input-container .suggest-widget.visible')
            await driver.findElementWithText('jwtmiddleware.go', {
                selector: '.monaco-query-input-container .suggest-widget.visible span',
                wait: { timeout: 5000 },
            })
            await driver.page.keyboard.press(Key.Tab)
            expect(await getSearchFieldValue(driver)).toStrictEqual(
                'repo:^github\\.com/auth0/go-jwt-middleware$ file:^jwtmiddleware\\.go$ '
            )

            // Symbol autocomplete in top search bar
            await driver.page.keyboard.type('On')
            await driver.page.waitForSelector('.monaco-query-input-container .suggest-widget.visible')
            await driver.findElementWithText('OnError', {
                selector: '.monaco-query-input-container .suggest-widget.visible span',
                wait: { timeout: 5000 },
            })
        })
    })

    describe('Search field value', () => {
        test('Is set from the URL query parameter when loading a search-related page', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=foo')
            await driver.page.waitForSelector('#monaco-query-input')
            expect(await getSearchFieldValue(driver)).toStrictEqual('foo')
            // Field value is cleared when navigating to a non search-related page
            await driver.page.waitForSelector('[data-testid="batch-changes"]')
            await driver.page.click('[data-testid="batch-changes"]')
            expect(await getSearchFieldValue(driver)).toStrictEqual('')
            // Field value is restored when the back button is pressed
            await driver.page.goBack()
            expect(await getSearchFieldValue(driver)).toStrictEqual('foo')
        })
    })

    describe('Case sensitivity toggle', () => {
        test('Clicking toggle turns on case sensitivity', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-case-sensitivity-toggle')
            await waitAndFocusInput()
            await driver.page.type('.test-query-input', 'test')
            await driver.page.click('.test-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal&case=yes')
        })

        test('Clicking toggle turns off case sensitivity and removes case= URL parameter', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=literal&case=yes')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-case-sensitivity-toggle')
            await driver.page.click('.test-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal')
        })
    })

    describe('Structural search toggle', () => {
        test('Clicking toggle turns on structural search', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-structural-search-toggle')
            await waitAndFocusInput()
            await driver.page.type('.test-query-input', 'test')
            await driver.page.click('.test-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=structural')
        })

        test('Clicking toggle turns on structural search and removes existing patternType parameter', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-structural-search-toggle')
            await driver.page.click('.test-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=structural')
        })

        test('Clicking toggle turns off structural saerch and reverts to default pattern type', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=structural')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-structural-search-toggle')
            await driver.page.click('.test-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal')
        })
    })

    describe('Search button', () => {
        test('Clicking search button executes search', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-search-button', { visible: true })
            // Note: Delay added because this test has been intermittently failing without it. Monaco search bar may drop events if it gets too many too fast.
            await driver.page.keyboard.type(' hello', { delay: 50 })
            await driver.page.click('.test-search-button')
            await driver.assertWindowLocation('/search?q=test+hello&patternType=regexp')
        })
    })

    describe('Streaming search', () => {
        const viewerSettingsWithStreamingSearch: Partial<WebGraphQlOperations> = {
            ViewerSettings: () => ({
                viewerSettings: {
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: JSON.stringify({ experimentalFeatures: { searchStreaming: true } }),
                            },
                        },
                        {
                            __typename: 'Site',
                            id: siteGQLID,
                            siteID,
                            latestSettings: {
                                id: 470,
                                contents: JSON.stringify({ experimentalFeatures: { searchStreaming: true } }),
                            },
                            settingsURL: '/site-admin/global-settings',
                            viewerCanAdminister: true,
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
        }

        test('Streaming search', async () => {
            const searchStreamEvents: SearchEvent[] = [
                {
                    type: 'matches',
                    data: [
                        { type: 'repo', repository: 'github.com/sourcegraph/sourcegraph' },
                        {
                            type: 'file',
                            lineMatches: [],
                            name: 'stream.ts',
                            repository: 'github.com/sourcegraph/sourcegraph',
                        },
                        {
                            type: 'file',
                            lineMatches: [],
                            name: 'stream.ts',
                            repository: 'github.com/sourcegraph/sourcegraph',
                            version: 'abcd',
                        },
                        {
                            type: 'file',
                            lineMatches: [],
                            name: 'stream.ts',
                            repository: 'github.com/sourcegraph/sourcegraph',
                            branches: ['test/branch'],
                        },
                    ],
                },
                { type: 'done', data: {} },
            ]

            testContext.overrideGraphQL({ ...commonSearchGraphQLResults, ...viewerSettingsWithStreamingSearch })
            testContext.overrideSearchStreamEvents(searchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-search-result', { visible: true })

            const results = await driver.page.evaluate(() =>
                [...document.querySelectorAll('.test-search-result-label')].map(label =>
                    (label.textContent || '').trim()
                )
            )
            expect(results).toEqual([
                'github.com/sourcegraph/sourcegraph',
                'sourcegraph/sourcegraph › stream.ts',
                'sourcegraph/sourcegraph@abcd › stream.ts',
                'sourcegraph/sourcegraph@test/branch › stream.ts',
            ])
        })

        test('Streaming search with error', async () => {
            const searchStreamEvents: SearchEvent[] = [
                {
                    type: 'error',
                    data: { message: 'Search is invalid' },
                },
            ]

            testContext.overrideGraphQL({ ...commonSearchGraphQLResults, ...viewerSettingsWithStreamingSearch })
            testContext.overrideSearchStreamEvents(searchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('[data-testid="search-results-list-error"]', { visible: true })

            const results = await driver.page.evaluate(
                () => document.querySelector('[data-testid="search-results-list-error"]')?.textContent
            )
            expect(results).toContain('Search is invalid')
        })

        test('Streaming diff search syntax highlighting', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...viewerSettingsWithStreamingSearch,
                ...diffHighlightResult,
            })
            testContext.overrideSearchStreamEvents(diffSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test%20type:diff&patternType=regexp')
            await driver.page.waitForSelector('.search-result-match__code-excerpt .selection-highlight', {
                visible: true,
            })
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })

            await percySnapshotWithVariants(driver.page, 'Streaming diff search syntax highlighting', {
                waitForCodeHighlighting: true,
            })
        })

        test('Streaming commit search syntax highlighting', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...viewerSettingsWithStreamingSearch,
                ...commitHighlightResult,
            })
            testContext.overrideSearchStreamEvents(commitSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=graph%20type:commit&patternType=regexp')
            await driver.page.waitForSelector('.search-result-match__code-excerpt .selection-highlight', {
                visible: true,
            })
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })

            await percySnapshotWithVariants(driver.page, 'Streaming commit search syntax highlighting', {
                waitForCodeHighlighting: true,
            })
        })

        test('Streaming search code, file and repo results with filter suggestions', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...viewerSettingsWithStreamingSearch,
                ...highlightFileResult,
            })
            testContext.overrideSearchStreamEvents(mixedSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.code-excerpt .selection-highlight', {
                visible: true,
            })
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })

            await percySnapshotWithVariants(
                driver.page,
                'Streaming commit code, file and repo results with filter suggestions',
                {
                    waitForCodeHighlighting: true,
                }
            )
        })

        test('Streaming search symbols', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...viewerSettingsWithStreamingSearch,
            })
            testContext.overrideSearchStreamEvents(symbolSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-file-match-children-item', {
                visible: true,
            })
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })

            await percySnapshotWithVariants(driver.page, 'Streaming search symbols', {
                waitForCodeHighlighting: true,
            })
        })
    })

    describe('Search contexts', () => {
        const viewerSettingsWithSearchContexts: Partial<WebGraphQlOperations> = {
            ViewerSettings: () => ({
                viewerSettings: {
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: JSON.stringify({
                                    experimentalFeatures: {
                                        showSearchContext: true,
                                        showSearchContextManagement: true,
                                    },
                                }),
                            },
                        },
                        {
                            __typename: 'Site',
                            id: siteGQLID,
                            siteID,
                            latestSettings: {
                                id: 470,
                                contents: JSON.stringify({
                                    experimentalFeatures: {
                                        showSearchContext: true,
                                        showSearchContextManagement: true,
                                    },
                                }),
                            },
                            settingsURL: '/site-admin/global-settings',
                            viewerCanAdminister: true,
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
        }
        const testContextForSearchContexts: Partial<WebGraphQlOperations> = {
            ...commonSearchGraphQLResults,
            ...viewerSettingsWithSearchContexts,
            AutoDefinedSearchContexts: () => ({
                autoDefinedSearchContexts: [
                    {
                        __typename: 'SearchContext',
                        id: '1',
                        spec: 'global',
                        description: '',
                        autoDefined: true,
                        repositories: [],
                    },
                    {
                        __typename: 'SearchContext',
                        id: '2',
                        spec: '@test',
                        description: '',
                        autoDefined: true,
                        repositories: [],
                    },
                ],
            }),
            UserRepositories: () => ({
                node: {
                    repositories: {
                        totalCount: 1,
                        nodes: [
                            {
                                id: '1',
                                name: 'repo',
                                viewerCanAdminister: false,
                                createdAt: '',
                                url: '',
                                isPrivate: false,
                                mirrorInfo: { cloned: true, cloneInProgress: false, updatedAt: null },
                                externalRepository: { serviceType: '', serviceID: '' },
                            },
                        ],
                        pageInfo: { hasNextPage: false },
                    },
                },
            }),
        }
        const versionContexts = [
            {
                name: 'version-context-1',
                description: 'v1',
                revisions: [],
            },
            {
                name: 'version-context-2',
                description: 'v2',
                revisions: [],
            },
        ]

        beforeEach(() => {
            testContext.overrideGraphQL(testContextForSearchContexts)
            const context = createJsContext({ sourcegraphBaseUrl: driver.sourcegraphBaseUrl })
            testContext.overrideJsContext({
                ...context,
                experimentalFeatures: {
                    versionContexts,
                },
            })
        })

        afterEach(async () => {
            await driver.page.evaluate(() => localStorage.clear())
        })

        const getSelectedSearchContextSpec = () =>
            driver.page.evaluate(() => document.querySelector('.test-selected-search-context-spec')?.textContent)

        const isSearchContextDropdownVisible = () =>
            driver.page.evaluate(
                () => document.querySelector<HTMLButtonElement>('.test-search-context-dropdown') !== null
            )

        const isSearchContextHighlightTourStepVisible = () =>
            driver.page.evaluate(
                () =>
                    document.querySelector<HTMLDivElement>(
                        'div[data-shepherd-step-id="search-contexts-start-tour"]'
                    ) !== null
            )

        const isSearchContextDropdownDisabled = () =>
            driver.page.evaluate(
                () => document.querySelector<HTMLButtonElement>('.test-search-context-dropdown')?.disabled
            )
        test('Search context selected based on URL', async () => {
            testContext.overrideGraphQL({
                ...testContextForSearchContexts,
                IsSearchContextAvailable: () => ({
                    isSearchContextAvailable: true,
                }),
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:%40test+test&patternType=regexp', {
                waitUntil: 'networkidle2',
            })
            await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
            expect(await getSelectedSearchContextSpec()).toStrictEqual('context:@test')
        })

        test('Missing context filter should default to global context', async () => {
            // Initialize localStorage to a valid context, that should not be used
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.evaluate(() => localStorage.setItem('sg-last-search-context', '@test'))
            // Visit the search page with a query without a context filter
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
            expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
        })

        test('Unavailable search context should remain in the query and disable the search context dropdown', async () => {
            await driver.page.goto(
                driver.sourcegraphBaseUrl + '/search?q=context:%40unavailableCtx+test&patternType=regexp',
                { waitUntil: 'networkidle2' }
            )
            await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
            await driver.page.waitForSelector('#monaco-query-input')
            expect(await getSearchFieldValue(driver)).toStrictEqual('context:@unavailableCtx test')
            expect(await isSearchContextDropdownDisabled()).toBeTruthy()
        })

        test('Search context dropdown should not be visible if user has no repositories', async () => {
            testContext.overrideGraphQL({
                ...testContextForSearchContexts,
                UserRepositories: () => ({
                    node: {
                        repositories: {
                            totalCount: 0,
                            nodes: [],
                            pageInfo: { hasNextPage: false },
                        },
                    },
                }),
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('#monaco-query-input')
            expect(await isSearchContextDropdownVisible()).toBeFalsy()
        })

        test('Reset unavailable search context from localStorage if query is not present', async () => {
            // First initialize localStorage on the page
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.evaluate(() => localStorage.setItem('sg-last-search-context', 'doesnotexist'))
            // Visit the page again with localStorage initialized
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search', {
                waitUntil: 'networkidle2',
            })
            await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
            expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
        })

        test('Convert version context', async () => {
            testContext.overrideGraphQL({
                ...testContextForSearchContexts,
                IsSearchContextAvailable: () => ({
                    isSearchContextAvailable: false,
                }),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts?tab=convert-version-contexts')

            await driver.page.waitForSelector('.test-convert-version-context-btn')
            await driver.page.click('.test-convert-version-context-btn')

            await driver.page.waitForSelector('.convert-version-context-node .alert-success')

            const successText = await driver.page.evaluate(
                () => document.querySelector('.convert-version-context-node .alert-success')?.textContent
            )
            expect(successText).toBe('Version context successfully converted.')
        })

        test('Convert all version contexts', async () => {
            testContext.overrideGraphQL({
                ...testContextForSearchContexts,
                IsSearchContextAvailable: () => ({
                    isSearchContextAvailable: false,
                }),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts?tab=convert-version-contexts')

            await driver.page.waitForSelector('.test-convert-all-search-contexts-btn')
            await driver.page.click('.test-convert-all-search-contexts-btn')

            testContext.overrideGraphQL({
                ...testContextForSearchContexts,
                IsSearchContextAvailable: () => ({
                    isSearchContextAvailable: true,
                }),
            })

            // Check that a success message appears with the correct number of converted contexts
            await driver.page.waitForSelector('.test-convert-all-search-contexts-success')
            const successText = await driver.page.evaluate(
                () => document.querySelector('.test-convert-all-search-contexts-success')?.textContent
            )
            expect(successText).toBe(`Sucessfully converted ${versionContexts.length} version contexts.`)

            // Check that individual context nodes have 'Converted' text
            const convertedContexts = await driver.page.evaluate(
                () => document.querySelectorAll('.test-converted-context').length
            )
            expect(convertedContexts).toBe(versionContexts.length)
        })

        test('Highlight tour step should be visible with empty local storage', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:global+test&patternType=regexp')
            await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
            expect(await isSearchContextHighlightTourStepVisible()).toBeTruthy()
        })

        test('Highlight tour step should not be visible if already seen', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:global+test&patternType=regexp')
            await driver.page.evaluate(() =>
                localStorage.setItem('has-seen-search-contexts-dropdown-highlight-tour-step', 'true')
            )
            await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
            expect(await isSearchContextHighlightTourStepVisible()).toBeFalsy()
        })
    })
})
