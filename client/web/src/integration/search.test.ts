import expect from 'expect'
import { test } from 'mocha'
import { Key } from 'ts-key-enum'

import { SharedGraphQlOperations, SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
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

const mockDefaultStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [{ type: 'repo', repository: 'github.com/Algorilla/manta-ray' }],
    },
    { type: 'progress', data: { matchCount: 30, durationMs: 103, skipped: [] } },
    {
        type: 'filters',
        data: [
            { label: 'archived:yes', value: 'archived:yes', count: 5, kind: 'generic', limitHit: true },
            { label: 'fork:yes', value: 'fork:yes', count: 46, kind: 'generic', limitHit: true },
            // Two repo filters to trigger the repository sidebar section
            {
                label: 'github.com/Algorilla/manta-ray',
                value: 'repo:^github\\.com/Algorilla/manta-ray$',
                count: 1,
                kind: 'repo',
                limitHit: true,
            },
            {
                label: 'github.com/Algorilla/manta-ray2',
                value: 'repo:^github\\.com/Algorilla/manta-ray2$',
                count: 1,
                kind: 'repo',
                limitHit: true,
            },
        ],
    },
    { type: 'done', data: {} },
]

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    IsSearchContextAvailable: () => ({
        isSearchContextAvailable: true,
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
        testContext.overrideGraphQL(commonSearchGraphQLResults)
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const waitAndFocusInput = async () => {
        await driver.page.waitForSelector('.monaco-editor .view-lines')
        await driver.page.click('.monaco-editor .view-lines')
    }

    const getSearchFieldValue = (driver: Driver): Promise<string | undefined> =>
        driver.page.evaluate(() => document.querySelector<HTMLTextAreaElement>('#monaco-query-input textarea')?.value)

    describe('Search filters', () => {
        test('Search filters are shown on search result pages and clicking them triggers a new search', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            const dynamicFilters = ['archived:yes', 'repo:^github\\.com/Algorilla/manta-ray$']
            const origQuery = 'context:global foo'
            for (const filter of dynamicFilters) {
                await driver.page.goto(
                    `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal`
                )
                await driver.page.waitForSelector(`[data-testid="filter-link"][value=${JSON.stringify(filter)}]`)
                await driver.page.click(`[data-testid="filter-link"][value=${JSON.stringify(filter)}]`)
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
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('#monaco-query-input')
            await driver.replaceText({
                selector: '#monaco-query-input',
                newText: '-file',
                enterTextMethod: 'type',
            })
            await driver.page.waitForSelector('#monaco-query-input .suggest-widget.visible')
            await driver.findElementWithText('-file', {
                action: 'click',
                wait: { timeout: 5000 },
                selector: '#monaco-query-input .suggest-widget.visible span',
            })
            expect(await getSearchFieldValue(driver)).toStrictEqual('-file:')
        })
    })

    describe('Suggestions', () => {
        test('Typing in the search field shows relevant suggestions', async () => {
            testContext.overrideSearchStreamEvents([
                {
                    type: 'matches',
                    data: [
                        { type: 'repo', repository: 'github.com/auth0/go-jwt-middleware' },
                        {
                            type: 'symbol',
                            symbols: [
                                {
                                    name: 'OnError',
                                    containerName: 'jwtmiddleware',
                                    url: '/github.com/auth0/go-jwt-middleware/-/blob/jwtmiddleware.go#L56:1-56:14',
                                    kind: SymbolKind.FUNCTION,
                                },
                            ],
                            path: 'jwtmiddleware.go',
                            repository: 'github.com/auth0/go-jwt-middleware',
                        },
                        { type: 'path', path: 'jwtmiddleware.go', repository: 'github.com/auth0/go-jwt-middleware' },
                    ],
                },

                { type: 'done', data: {} },
            ])

            // Repo autocomplete from homepage
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            // Using id selector rather than `test-` classes as Monaco doesn't allow customizing classes
            await driver.page.waitForSelector('#monaco-query-input')
            await driver.replaceText({
                selector: '#monaco-query-input',
                newText: 'go-jwt-middlew',
                enterTextMethod: 'type',
            })
            await driver.page.waitForSelector('#monaco-query-input .suggest-widget.visible')
            await driver.findElementWithText('github.com/auth0/go-jwt-middleware', {
                action: 'click',
                wait: { timeout: 5000 },
                selector: '#monaco-query-input .suggest-widget.visible a.label-name',
            })
            expect(await getSearchFieldValue(driver)).toStrictEqual('repo:^github\\.com/auth0/go-jwt-middleware$ ')

            // Submit search
            await driver.page.keyboard.press(Key.Enter)

            // File autocomplete from repo search bar
            await driver.page.waitForSelector('#monaco-query-input')
            await driver.page.focus('#monaco-query-input')
            await driver.page.keyboard.type('jwtmi')
            await driver.page.waitForSelector('#monaco-query-input .suggest-widget.visible')
            await driver.findElementWithText('jwtmiddleware.go', {
                selector: '#monaco-query-input .suggest-widget.visible span',
                wait: { timeout: 5000 },
            })
            await driver.page.keyboard.press(Key.Tab)
            expect(await getSearchFieldValue(driver)).toStrictEqual(
                'repo:^github\\.com/auth0/go-jwt-middleware$ file:^jwtmiddleware\\.go$ '
            )

            // Symbol autocomplete in top search bar
            await driver.page.keyboard.type('On')
            await driver.page.waitForSelector('#monaco-query-input .suggest-widget.visible')
            await driver.findElementWithText('OnError', {
                selector: '#monaco-query-input .suggest-widget.visible span',
                wait: { timeout: 5000 },
            })
        })
    })

    describe('Search field value', () => {
        test('Is set from the URL query parameter when loading a search-related page', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                RegistryExtensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: { error: null, nodes: [] },
                        featuredExtensions: null,
                    },
                }),
            })
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=foo')
            await driver.page.waitForSelector('#monaco-query-input')
            expect(await getSearchFieldValue(driver)).toStrictEqual('foo')
            // Field value is cleared when navigating to a non search-related page
            await driver.page.waitForSelector('a[href="/extensions"]')
            await driver.page.click('a[href="/extensions"]')
            // Search box is gone when in a non-search page
            expect(await getSearchFieldValue(driver)).toStrictEqual(undefined)
            // Field value is restored when the back button is pressed
            await driver.page.goBack()
            expect(await getSearchFieldValue(driver)).toStrictEqual('foo')
        })
    })

    describe('Case sensitivity toggle', () => {
        test('Clicking toggle turns on case sensitivity', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-case-sensitivity-toggle')
            await waitAndFocusInput()
            await driver.page.type('.test-query-input', 'test')
            await driver.page.click('.test-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=context:global+test&patternType=literal&case=yes')
        })

        test('Clicking toggle turns off case sensitivity and removes case= URL parameter', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=literal&case=yes')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-case-sensitivity-toggle')
            await driver.page.click('.test-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=context:global+test&patternType=literal')
        })
    })

    describe('Structural search toggle', () => {
        test('Clicking toggle turns on structural search', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-structural-search-toggle')
            await waitAndFocusInput()
            await driver.page.type('.test-query-input', 'test')
            await driver.page.click('.test-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=context:global+test&patternType=structural')
        })

        test('Clicking toggle turns on structural search and removes existing patternType parameter', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-structural-search-toggle')
            await driver.page.click('.test-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=context:global+test&patternType=structural')
        })

        test('Clicking toggle turns off structural saerch and reverts to default pattern type', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=structural')
            await driver.page.waitForSelector('.test-query-input', { visible: true })
            await driver.page.waitForSelector('.test-structural-search-toggle')
            await driver.page.click('.test-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=context:global+test&patternType=literal')
        })
    })

    describe('Search button', () => {
        test('Clicking search button executes search', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-search-button', { visible: true })
            // Note: Delay added because this test has been intermittently failing without it. Monaco search bar may drop events if it gets too many too fast.
            await driver.page.keyboard.type(' hello', { delay: 50 })
            await driver.page.click('.test-search-button')
            await driver.assertWindowLocation('/search?q=context:global+test+hello&patternType=regexp')
        })
    })

    describe('Verify search streaming event handling', () => {
        test('Streaming search', async () => {
            const searchStreamEvents: SearchEvent[] = [
                {
                    type: 'matches',
                    data: [
                        { type: 'repo', repository: 'github.com/sourcegraph/sourcegraph' },
                        {
                            type: 'content',
                            lineMatches: [],
                            path: 'stream.ts',
                            repository: 'github.com/sourcegraph/sourcegraph',
                        },
                        {
                            type: 'content',
                            lineMatches: [],
                            path: 'stream.ts',
                            repository: 'github.com/sourcegraph/sourcegraph',
                            commit: 'abcd',
                        },
                        {
                            type: 'content',
                            lineMatches: [],
                            path: 'stream.ts',
                            repository: 'github.com/sourcegraph/sourcegraph',
                            branches: ['test/branch'],
                        },
                    ],
                },
                { type: 'done', data: {} },
            ]

            testContext.overrideSearchStreamEvents(searchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-search-result', { visible: true })

            const results = await driver.page.evaluate(() =>
                [...document.querySelectorAll('.test-search-result-label')].map(label =>
                    (label.textContent || '').trim()
                )
            )
            expect(results).toEqual([
                'sourcegraph/sourcegraph',
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

            testContext.overrideSearchStreamEvents(searchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('[data-testid="search-results-list-error"]', { visible: true })

            const results = await driver.page.evaluate(
                () => document.querySelector('[data-testid="search-results-list-error"]')?.textContent
            )
            expect(results).toContain('Search is invalid')
        })
    })

    describe('Search results snapshots', () => {
        // To avoid covering the Percy snapshots
        const hideCreateCodeMonitorFeatureTour = () =>
            driver.page.evaluate(() => {
                localStorage.setItem('has-seen-create-code-monitor-feature-tour', 'true')
                location.reload()
            })

        test('diff search syntax highlighting', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...diffHighlightResult,
            })
            testContext.overrideSearchStreamEvents(diffSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test%20type:diff&patternType=regexp', {
                waitUntil: 'networkidle0',
            })
            await hideCreateCodeMonitorFeatureTour()
            await driver.page.waitForSelector('[data-testid="search-result-match-code-excerpt"] .selection-highlight', {
                visible: true,
            })
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })

            await percySnapshotWithVariants(driver.page, 'Streaming diff search syntax highlighting', {
                waitForCodeHighlighting: true,
            })
        })

        test('commit search syntax highlighting', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...commitHighlightResult,
            })
            testContext.overrideSearchStreamEvents(commitSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=graph%20type:commit&patternType=regexp', {
                waitUntil: 'networkidle0',
            })
            await hideCreateCodeMonitorFeatureTour()
            await driver.page.waitForSelector('[data-testid="search-result-match-code-excerpt"] .selection-highlight', {
                visible: true,
            })
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })

            await percySnapshotWithVariants(driver.page, 'Streaming commit search syntax highlighting', {
                waitForCodeHighlighting: true,
            })
        })

        test('code, file and repo results with filter suggestions', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
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

        test('symbol results', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
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

    describe('Feature tour', () => {
        const resetCreateCodeMonitorFeatureTour = (dismissSearchContextsFeatureTour = true) =>
            driver.page.evaluate((dismissSearchContextsFeatureTour: boolean) => {
                localStorage.setItem('has-seen-create-code-monitor-feature-tour', 'false')
                localStorage.setItem(
                    'has-seen-search-contexts-dropdown-highlight-tour-step',
                    dismissSearchContextsFeatureTour ? 'true' : 'false'
                )
                location.reload()
            }, dismissSearchContextsFeatureTour)

        const isCreateCodeMonitorFeatureTourVisible = () =>
            driver.page.evaluate(
                () =>
                    document.querySelector<HTMLDivElement>(
                        'div[data-shepherd-step-id="create-code-monitor-feature-tour"]'
                    ) !== null
            )

        test('Do not show create code monitor button feature tour with missing search type', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test', {
                waitUntil: 'networkidle0',
            })
            await resetCreateCodeMonitorFeatureTour()
            await driver.page.waitForSelector('.test-search-result-label', { visible: true })
            expect(await isCreateCodeMonitorFeatureTourVisible()).toBeFalsy()
        })

        test('Show create code monitor button feature tour with valid search type', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test+type:diff', {
                waitUntil: 'networkidle0',
            })
            await resetCreateCodeMonitorFeatureTour()
            await driver.page.waitForSelector('.test-search-result-label', { visible: true })
            expect(await isCreateCodeMonitorFeatureTourVisible()).toBeTruthy()
        })

        test('Do not show create code monitor button feature tour if search contexts feature tour is not dismissed', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test+type:diff', {
                waitUntil: 'networkidle0',
            })
            await resetCreateCodeMonitorFeatureTour(false)
            await driver.page.waitForSelector('.test-search-result-label', { visible: true })
            expect(await isCreateCodeMonitorFeatureTourVisible()).toBeFalsy()
        })
    })
})
