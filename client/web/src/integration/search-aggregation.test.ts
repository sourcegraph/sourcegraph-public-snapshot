import delay from 'delay'
import expect from 'expect'
import { afterEach, beforeEach, describe, test } from 'mocha'

import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import type { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import {
    type GetSearchAggregationResult,
    type WebGraphQlOperations,
    SearchAggregationMode,
} from '../graphql-operations'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { createEditorAPI, removeContextFromQuery } from './utils'

const aggregationDefaultMock = (mode: SearchAggregationMode): GetSearchAggregationResult => ({
    searchQueryAggregate: {
        __typename: 'SearchQueryAggregate',
        aggregations: {
            __typename: 'ExhaustiveSearchAggregationResult',
            mode,
            otherGroupCount: 100,
            groups: [
                {
                    __typename: 'AggregationGroup',
                    label: 'sourcegraph/sourcegraph',
                    count: 100,
                    query: 'context:global insights repo:sourcegraph/sourcegraph',
                },
                {
                    __typename: 'AggregationGroup',
                    label: 'sourcegraph/about',
                    count: 80,
                    query: 'context:global insights repo:sourecegraph/about',
                },
                {
                    __typename: 'AggregationGroup',
                    label: 'sourcegraph/search-insight',
                    count: 60,
                    query: 'context:global insights repo:sourecegraph/search-insight',
                },
                {
                    __typename: 'AggregationGroup',
                    label: 'sourcegraph/lang-stats',
                    count: 40,
                    query: 'context:global insights repo:sourecegraph/lang-stats',
                },
            ],
        },
        modeAvailability: [
            {
                __typename: 'AggregationModeAvailability',
                mode: SearchAggregationMode.REPO,
                available: true,
                reasonUnavailable: null,
            },
            {
                __typename: 'AggregationModeAvailability',
                mode: SearchAggregationMode.PATH,
                available: true,
                reasonUnavailable: null,
            },
            {
                __typename: 'AggregationModeAvailability',
                mode: SearchAggregationMode.AUTHOR,
                available: true,
                reasonUnavailable: null,
            },
            {
                __typename: 'AggregationModeAvailability',
                mode: SearchAggregationMode.CAPTURE_GROUP,
                available: true,
                reasonUnavailable: null,
            },
        ],
    },
})

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

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    IsSearchContextAvailable: () => ({ isSearchContextAvailable: true }),
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
            createdAt: '2020-03-02T11:52:15Z',
            roles: {
                __typename: 'RoleConnection',
                nodes: [],
            },
        },
    }),
}

const QUERY_INPUT_SELECTOR = '.test-query-input'

describe('Search aggregation', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    before(async () => {
        driver = await createDriverForTest()
    })
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
            GetSearchAggregation: ({ mode }) => aggregationDefaultMock(mode ?? SearchAggregationMode.REPO),
        })
        testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)
    })

    after(() => driver?.close())
    afterEach(() => testContext?.dispose())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    test('should be hidden if feature flag is off', async () => {
        await driver.page.goto(
            `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent('context:global foo')}&patternType=literal`
        )

        await driver.page.waitForSelector('[data-testid="filter-link"]')
        const aggregationSidebar = await driver.page.$x("//button[contains(., 'Grouped by')]")

        expect(aggregationSidebar).toStrictEqual([])
    })

    describe('with aggregation feature flag on', () => {
        beforeEach(() =>
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                ViewerSettings: () => ({
                    viewerSettings: {
                        __typename: 'SettingsCascade',
                        subjects: [
                            {
                                __typename: 'DefaultSettings',
                                id: 'TestDefaultSettingsID',
                                settingsURL: null,
                                viewerCanAdminister: false,
                                latestSettings: {
                                    id: 0,
                                    contents: JSON.stringify({
                                        experimentalFeatures: {
                                            searchResultsAggregations: true,
                                            searchQueryInput: 'v1',
                                        },
                                    }),
                                },
                            },
                        ],
                        final: JSON.stringify({}),
                    },
                }),
            })
        )

        test('should sync aggregation settings across different UI via URL', async () => {
            const origQuery = 'context:global insights('

            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal`
            )

            await driver.page.waitForSelector('[aria-label="Aggregation mode picker"]')

            // Wait for FE sets correct aggregation mode based on BE response
            await delay(100)

            const aggregationCases = [
                { mode: 'REPO', urlKey: 'repo', id: 'repo-aggregation-mode' },
                { mode: 'PATH', urlKey: 'path', id: 'file-aggregation-mode' },
                { mode: 'AUTHOR', urlKey: 'author', id: 'author-aggregation-mode' },
                { mode: 'CAPTURE_GROUP', urlKey: 'group', id: 'captureGroup-aggregation-mode' },
            ]

            for (const testCase of aggregationCases) {
                await driver.page.click(`[data-testid="${testCase.id}"]`)

                await driver.page.waitForFunction(
                    (expectedQuery: string, mode: string) => {
                        const url = new URL(document.location.href)
                        const query = url.searchParams.get('q')
                        const aggregationMode = url.searchParams.get('groupBy')

                        return query && query.trim() === expectedQuery && aggregationMode === mode
                    },
                    { timeout: 5000 },
                    `${origQuery}`,
                    testCase.urlKey
                )
            }
        })

        test('should open expanded full UI by default if UI mode is set in URL query param', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent('insights(')}&expanded`)

            await driver.page.waitForSelector('[aria-label="Aggregation results panel"]')
        })

        test('should expand the full UI mode with the current aggregation mode', async () => {
            const origQuery = 'context:global insights('

            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal`
            )

            await driver.page.waitForSelector('[aria-label="Aggregation mode picker"]')

            // Wait for FE sets correct aggregation mode based on BE response
            await delay(100)

            await driver.page.click('[data-testid="file-aggregation-mode"]')

            await driver.page.waitForSelector('[data-testid="expand-aggregation-ui"]')
            await driver.page.click('[data-testid="expand-aggregation-ui"]')

            await driver.page.waitForSelector('[aria-label="Aggregation results panel"]')

            await driver.page.waitForFunction(
                (expectedQuery: string) => {
                    const url = new URL(document.location.href)
                    const query = url.searchParams.get('q')
                    const aggregationMode = url.searchParams.get('groupBy')
                    const aggregationUIMode = url.searchParams.get('expanded')

                    return (
                        query &&
                        query.trim() === expectedQuery &&
                        aggregationMode === 'path' &&
                        aggregationUIMode === ''
                    )
                },
                { timeout: 5000 },
                `${origQuery}`
            )

            await driver.page.click('[data-testid="author-aggregation-mode"]')
            await driver.page.click('[aria-label="Close aggregation full UI mode"]')

            await driver.page.waitForSelector('[aria-label="Aggregation mode picker"]')

            await driver.page.waitForFunction(
                (expectedQuery: string) => {
                    const url = new URL(document.location.href)
                    const query = url.searchParams.get('q')
                    const aggregationMode = url.searchParams.get('groupBy')
                    const aggregationUIMode = url.searchParams.get('expanded')

                    return (
                        query &&
                        query.trim() === expectedQuery &&
                        aggregationMode === 'author' &&
                        aggregationUIMode === null
                    )
                },
                { timeout: 5000 },
                `${origQuery}`
            )
        })

        test('should update the search box query when user clicks on one of aggregation bars', async () => {
            const origQuery = 'context:global insights('

            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal`
            )

            const editor = await createEditorAPI(driver, QUERY_INPUT_SELECTOR)
            await editor.waitForIt()

            await driver.page.waitForSelector('[aria-label="Bar chart content"] a')

            // waitForSelector checks for dom element, but it doesn't track visual representation of the element
            // Wait until chart is visually rendered and only then click the element, otherwise it may be possible
            // to encounter a puppeter bug https://github.com/puppeteer/puppeteer/issues/8627
            await delay(200)
            await driver.page.click('[aria-label="Sidebar search aggregation chart"] a')

            expect(removeContextFromQuery((await editor.getValue()) ?? '')).toStrictEqual(
                'insights repo:sourcegraph/sourcegraph'
            )

            await driver.page.waitForSelector('[data-testid="expand-aggregation-ui"]')
            await driver.page.click('[data-testid="expand-aggregation-ui"]')
            await driver.page.waitForSelector(
                '[aria-label="Expanded search aggregation chart"] [aria-label="Bar chart content"] g:nth-child(2) a'
            )

            // waitForSelector checks for dom element, but it doesn't track visual representation of the element
            // Wait until chart is visually rendered and only then click the element, otherwise it may be possible
            // to encounter a puppeter bug https://github.com/puppeteer/puppeteer/issues/8627
            await delay(200)

            await driver.page.click(
                '[aria-label="Expanded search aggregation chart"] [aria-label="Bar chart content"] g:nth-child(2) a'
            )

            expect(removeContextFromQuery((await editor.getValue()) ?? '')).toStrictEqual(
                'insights repo:sourecegraph/about'
            )
        })

        test('should preserve case sensitive filter in a query', async () => {
            const origQuery = 'context:global insights('

            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent(origQuery)}&patternType=literal&case=yes`
            )

            const variables = await testContext.waitForGraphQLRequest(() => {}, 'GetSearchAggregation')

            expect(variables.query).toEqual(`${origQuery} case:yes`)

            const variablesForFileMode = await testContext.waitForGraphQLRequest(async () => {
                await driver.page.waitForSelector('[aria-label="Aggregation mode picker"]')
                await driver.page.click('[data-testid="file-aggregation-mode"]')
            }, 'GetSearchAggregation')

            expect(variablesForFileMode.query).toEqual(`${origQuery} case:yes`)

            const variablesWithoutCaseSensitivity = await testContext.waitForGraphQLRequest(
                async () => driver.page.click('.test-case-sensitivity-toggle'),
                'GetSearchAggregation'
            )

            expect(variablesWithoutCaseSensitivity.query).toEqual(origQuery)
        })
    })
})
