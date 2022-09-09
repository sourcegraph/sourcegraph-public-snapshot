import expect from 'expect'
import { test } from 'mocha'
import { Key } from 'ts-key-enum'

import {
    SearchGraphQlOperations,
    commitHighlightResult,
    commitSearchStreamEvents,
    diffSearchStreamEvents,
    diffHighlightResult,
    mixedSearchStreamEvents,
    highlightFileResult,
    symbolSearchStreamEvents,
} from '@sourcegraph/search'
import { SharedGraphQlOperations, SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults, createViewerSettingsGraphQLOverride } from './graphQlResults'
import { createEditorAPI, enableEditor, percySnapshotWithVariants, withSearchQueryInput } from './utils'

const mockDefaultStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [{ type: 'repo', repository: 'github.com/Algorilla/manta-ray' }],
    },
    { type: 'progress', data: { matchCount: 30, durationMs: 103, skipped: [] } },
    {
        type: 'filters',
        data: [
            { label: 'archived:yes', value: 'archived:yes', count: 5, kind: 'utility', limitHit: true },
            { label: 'fork:yes', value: 'fork:yes', count: 46, kind: 'utility', limitHit: true },
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

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations> = {
    ...commonWebGraphQlResults,
    IsSearchContextAvailable: () => ({
        isSearchContextAvailable: true,
    }),
}

const commonSearchGraphQLResultsWithUser: Partial<
    WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations
> = {
    ...commonSearchGraphQLResults,
    UserAreaUserProfile: () => ({
        user: {
            __typename: 'User',
            id: 'user123',
            username: 'alice',
            displayName: 'alice',
            url: '/users/test',
            settingsURL: '/users/test/settings',
            avatarURL: '',
            viewerCanAdminister: true,
            builtinAuth: true,
            tags: [],
        },
    }),
}

const queryInputSelector = '[data-testid="searchbox"] .test-query-input'

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
        testContext.overrideGraphQL(commonSearchGraphQLResultsWithUser)
        testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('Search filters', () => {
        test('Search filters are shown on search result pages and clicking them triggers a new search', async () => {
            const dynamicFilters = ['archived:yes', 'repo:^github\\.com/Algorilla/manta-ray$']
            const origQuery = 'context:global foo'
            for (const filter of dynamicFilters) {
                await driver.page.goto(
                    `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal`
                )
                await driver.page.waitForSelector(`[data-testid="filter-link"][value=${JSON.stringify(filter)}]`)
                await driver.page.click(`[data-testid="filter-link"][value=${JSON.stringify(filter)}]`)
                await driver.page.waitForFunction(
                    (expectedQuery: string) => {
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
        withSearchQueryInput(editorName => {
            test(`Completing a negated filter should insert the filter with - prefix (${editorName})`, async () => {
                testContext.overrideGraphQL({
                    ...commonSearchGraphQLResults,
                    ...createViewerSettingsGraphQLOverride({ user: enableEditor(editorName) }),
                })

                await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                const editor = await createEditorAPI(driver, queryInputSelector)
                await editor.replace('-file')
                await editor.selectSuggestion('-file')
                expect(await editor.getValue()).toStrictEqual('-file:')
                await percySnapshotWithVariants(driver.page, `Search home page (${editorName})`)
                await accessibilityAudit(driver.page)
            })
        })
    })

    describe('Suggestions', () => {
        withSearchQueryInput(editorName => {
            test.skip(`Typing in the search field shows relevant suggestions (${editorName})`, async () => {
                testContext.overrideGraphQL({
                    ...commonSearchGraphQLResults,
                    ...createViewerSettingsGraphQLOverride({ user: enableEditor(editorName) }),
                })
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
                                        line: 56,
                                    },
                                ],
                                path: 'jwtmiddleware.go',
                                repository: 'github.com/auth0/go-jwt-middleware',
                            },
                            {
                                type: 'path',
                                path: 'jwtmiddleware.go',
                                repository: 'github.com/auth0/go-jwt-middleware',
                            },
                        ],
                    },

                    { type: 'done', data: {} },
                ])

                // Repo autocomplete from homepage
                await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                const editor = await createEditorAPI(driver, queryInputSelector)
                await editor.focus()
                await editor.replace('repo:go-jwt-middlew')
                await editor.selectSuggestion('github.com/auth0/go-jwt-middleware')
                expect(await editor.getValue()).toStrictEqual('repo:^github\\.com/auth0/go-jwt-middleware$ ')

                // Submit search
                await driver.page.keyboard.press(Key.Enter)

                // File autocomplete from repo search bar
                await editor.focus()
                await driver.page.keyboard.type('file:jwtmi')
                await editor.waitForSuggestion('jwtmiddleware.go')
                // NOTE: This test assumes that the first suggestion is the one
                // to be selected.
                // It doesn't seem to be possible to otherwise "select" a specific
                // entry from the list (other than simulating arrow key presses and
                // somehow comparing the selected entry to the expected one).
                await driver.page.keyboard.press(Key.Tab)
                expect(await editor.getValue()).toStrictEqual(
                    'repo:^github\\.com/auth0/go-jwt-middleware$ file:^jwtmiddleware\\.go$ '
                )

                // Symbol autocomplete in top search bar
                await driver.page.keyboard.type('On')
                await editor.waitForSuggestion('OnError')
            })
        })
    })

    describe('Search field value', () => {
        withSearchQueryInput(editorName => {
            describe(editorName, () => {
                beforeEach(() => {
                    testContext.overrideGraphQL({
                        ...commonSearchGraphQLResults,
                        ...createViewerSettingsGraphQLOverride({ user: enableEditor(editorName) }),
                        RegistryExtensions: () => ({
                            extensionRegistry: {
                                __typename: 'ExtensionRegistry',
                                extensions: { error: null, nodes: [] },
                                featuredExtensions: null,
                            },
                        }),
                    })
                })

                test('Is set from the URL query parameter when loading a search-related page', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=foo')
                    const editor = await createEditorAPI(driver, queryInputSelector)
                    await editor.waitForIt()
                    await driver.page.waitForSelector('[data-testid="results-info-bar"]')
                    expect(await editor.getValue()).toStrictEqual('foo')
                    // Field value is cleared when navigating to a non search-related page
                    await driver.page.waitForSelector('a[href="/extensions"]')
                    await driver.page.click('a[href="/extensions"]')
                    // Search box is gone when in a non-search page
                    expect(await editor.getValue()).toStrictEqual(undefined)
                    // Field value is restored when the back button is pressed
                    await driver.page.goBack()
                    await editor.waitForIt()
                    await driver.page.waitForSelector('[data-testid="results-info-bar"]')
                    expect(await editor.getValue()).toStrictEqual('foo')
                })

                test('Normalizes input with line breaks', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                    const editor = await createEditorAPI(driver, queryInputSelector)
                    await editor.focus()
                    await driver.paste('foo\n\n\n\n\nbar')
                    expect(await editor.getValue()).toBe('foo bar')
                })
            })
        })
    })

    describe('Case sensitivity toggle', () => {
        withSearchQueryInput(editorName => {
            describe(editorName, () => {
                beforeEach(() => {
                    testContext.overrideGraphQL({
                        ...commonSearchGraphQLResults,
                        ...createViewerSettingsGraphQLOverride({ user: enableEditor(editorName) }),
                    })
                })

                test('Clicking toggle turns on case sensitivity', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                    const editor = await createEditorAPI(driver, queryInputSelector)
                    await driver.page.waitForSelector('.test-case-sensitivity-toggle')
                    await editor.focus()
                    await driver.page.keyboard.type('test')
                    await driver.page.click('.test-case-sensitivity-toggle')
                    await driver.assertWindowLocation('/search?q=context:global+test&patternType=standard&case=yes')
                })

                test('Clicking toggle turns off case sensitivity and removes case= URL parameter', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=standard&case=yes')
                    await createEditorAPI(driver, queryInputSelector)
                    await driver.page.waitForSelector('.test-case-sensitivity-toggle')
                    await driver.page.click('.test-case-sensitivity-toggle')
                    await driver.assertWindowLocation('/search?q=context:global+test&patternType=standard')
                })
            })
        })
    })

    describe('Structural search toggle', () => {
        withSearchQueryInput(editorName => {
            describe(editorName, () => {
                beforeEach(() => {
                    testContext.overrideGraphQL({
                        ...commonSearchGraphQLResults,
                        ...createViewerSettingsGraphQLOverride({ user: enableEditor(editorName) }),
                    })
                })

                test('Clicking toggle turns on structural search', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                    const editor = await createEditorAPI(driver, queryInputSelector)
                    await driver.page.waitForSelector('.test-structural-search-toggle')
                    await editor.focus()
                    await driver.page.keyboard.type('test')
                    await driver.page.click('.test-structural-search-toggle')
                    await driver.assertWindowLocation('/search?q=context:global+test&patternType=structural')
                })

                test('Clicking toggle turns on structural search and removes existing patternType parameter', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
                    const editor = await createEditorAPI(driver, queryInputSelector)
                    await editor.focus()
                    await driver.page.waitForSelector('.test-structural-search-toggle')
                    await driver.page.click('.test-structural-search-toggle')
                    await driver.assertWindowLocation('/search?q=context:global+test&patternType=structural')
                })

                test('Clicking toggle turns off structural search and reverts to default pattern type', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=structural')
                    await createEditorAPI(driver, queryInputSelector)
                    await driver.page.waitForSelector('.test-structural-search-toggle')
                    await driver.page.click('.test-structural-search-toggle')
                    await driver.assertWindowLocation('/search?q=context:global+test&patternType=standard')
                })
            })
        })
    })

    describe('Search button', () => {
        test('Clicking search button executes search', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-search-button', { visible: true })
            // Note: Delay added because this test has been intermittently failing without it. Monaco search bar may drop events if it gets too many too fast.
            await driver.page.keyboard.type(' hello', { delay: 500 })
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
        test('diff search syntax highlighting', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...diffHighlightResult,
            })
            testContext.overrideSearchStreamEvents(diffSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test%20type:diff&patternType=regexp', {
                waitUntil: 'networkidle0',
            })
            await driver.page.waitForSelector('[data-testid="search-result-match-code-excerpt"] .match-highlight', {
                visible: true,
            })
            await percySnapshotWithVariants(driver.page, 'Streaming diff search syntax highlighting', {
                waitForCodeHighlighting: true,
            })
            await accessibilityAudit(driver.page)
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
            await driver.page.waitForSelector('[data-testid="search-result-match-code-excerpt"] .match-highlight', {
                visible: true,
            })

            await percySnapshotWithVariants(driver.page, 'Streaming commit search syntax highlighting', {
                waitForCodeHighlighting: true,
            })
            await accessibilityAudit(driver.page)
        })

        test('code, file and repo results with filter suggestions', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...highlightFileResult,
            })
            testContext.overrideSearchStreamEvents(mixedSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('[data-testid="code-excerpt"] .match-highlight', {
                visible: true,
            })

            await percySnapshotWithVariants(
                driver.page,
                'Streaming commit code, file and repo results with filter suggestions',
                {
                    waitForCodeHighlighting: true,
                }
            )
            await accessibilityAudit(driver.page)
        })

        test('symbol results', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ...highlightFileResult,
            })
            testContext.overrideSearchStreamEvents(symbolSearchStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.test-file-match-children-item', {
                visible: true,
            })

            await percySnapshotWithVariants(driver.page, 'Streaming search symbols', {
                waitForCodeHighlighting: true,
            })
            await accessibilityAudit(driver.page)
        })
    })

    describe('Saved searches', () => {
        test('is styled correctly, with saved searches', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                savedSearches: () => ({
                    savedSearches: [
                        {
                            description: 'Demo',
                            id: 'U2F2ZWRTZWFyY2g6NQ==',
                            namespace: { __typename: 'User', id: 'user123', namespaceName: 'test' },
                            notify: false,
                            notifySlack: false,
                            query: 'context:global Batch Change patternType:literal',
                            slackWebhookURL: null,
                        },
                    ],
                }),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('[data-testid="saved-searches-list-page"]')
            await percySnapshotWithVariants(driver.page, 'Saved searches list')
            await accessibilityAudit(driver.page)
        })

        test('is styled correctly, with saved search form', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/searches/add')
            await driver.page.waitForSelector('[data-testid="saved-search-form"]')
            await percySnapshotWithVariants(driver.page, 'Saved search - Form')
            await accessibilityAudit(driver.page)
        })
    })

    describe('Search sidebar', () => {
        withSearchQueryInput(editorName => {
            describe(editorName, () => {
                beforeEach(() => {
                    testContext.overrideGraphQL({
                        ...commonSearchGraphQLResults,
                        ...createViewerSettingsGraphQLOverride({ user: enableEditor(editorName) }),
                    })
                })

                test.skip('updates the query input and triggers suggestions', async () => {
                    await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test')
                    await driver.page.waitForSelector('[data-testid="search-type-suggest"]')
                    await driver.page.click('[data-testid="search-type-suggest"]')
                    const editor = await createEditorAPI(driver, queryInputSelector)
                    await editor.waitForSuggestion()
                    expect(await editor.getValue()).toEqual('test repo:')
                })
            })
        })

        test('updates the query input and submits the query', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test')
            await driver.page.waitForSelector('[data-testid="search-type-submit"]')
            await Promise.all([
                driver.page.waitForNavigation(),
                driver.page.click('[data-testid="search-type-submit"]'),
            ])
            await driver.assertWindowLocation('/search?q=context:global+test+type:commit&patternType=standard')
        })
    })
})
