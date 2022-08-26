import delay from 'delay'
import expect from 'expect'
import { test } from 'mocha'

import { SearchAggregationMode, SearchGraphQlOperations } from '@sourcegraph/search'
import { GetSearchAggregationResult } from '@sourcegraph/search-ui'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { createEditorAPI } from './utils'

const aggregationDefaultMock: GetSearchAggregationResult = {
    searchQueryAggregate: {
        __typename: 'SearchQueryAggregate',
        aggregations: {
            __typename: 'ExhaustiveSearchAggregationResult',
            mode: SearchAggregationMode.REPO,
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
}
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
            tags: [],
        },
    }),
}

const QUERY_INPUT_SELECTOR = '[data-testid="searchbox"] .test-query-input'

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
            GetSearchAggregation: () => aggregationDefaultMock,
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
                EvaluateFeatureFlag: variables => {
                    if (variables.flagName === 'search-aggregation-filters') {
                        return { evaluateFeatureFlag: true }
                    }

                    return { evaluateFeatureFlag: false }
                },
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
                { mode: 'REPO', id: 'repo-aggregation-mode' },
                { mode: 'PATH', id: 'file-aggregation-mode' },
                { mode: 'AUTHOR', id: 'author-aggregation-mode' },
                { mode: 'CAPTURE_GROUP', id: 'captureGroup-aggregation-mode' },
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
                    testCase.mode
                )
            }
        })

        test('should open expanded full UI by default if UI mode is set in URL query param', async () => {
            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/search?q=${encodeURIComponent('insights(')}&groupByUI=searchPage`
            )

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
            await driver.page.click('[data-testid="expand-aggregation-ui"]')

            await driver.page.waitForSelector('[aria-label="Aggregation results panel"]')

            await driver.page.waitForFunction(
                (expectedQuery: string) => {
                    const url = new URL(document.location.href)
                    const query = url.searchParams.get('q')
                    const aggregationMode = url.searchParams.get('groupBy')
                    const aggregationUIMode = url.searchParams.get('groupByUI')

                    return (
                        query &&
                        query.trim() === expectedQuery &&
                        aggregationMode === 'PATH' &&
                        aggregationUIMode === 'searchPage'
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
                    const aggregationUIMode = url.searchParams.get('groupByUI')

                    return (
                        query &&
                        query.trim() === expectedQuery &&
                        aggregationMode === 'AUTHOR' &&
                        aggregationUIMode === 'sidebar'
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

            await driver.page.waitForSelector('[aria-label="chart content group"] a')
            await driver.page.click('[aria-label="Sidebar search aggregation chart"] a')

            expect(await editor.getValue()).toStrictEqual('insights repo:sourcegraph/sourcegraph')

            await driver.page.click('[data-testid="expand-aggregation-ui"]')
            await driver.page.waitForSelector('[aria-label="chart content group"] g:nth-child(2) a')
            await driver.page.click(
                '[aria-label="Expanded search aggregation chart"] [aria-label="chart content group"] g:nth-child(2) a'
            )

            expect(await editor.getValue()).toStrictEqual('insights repo:sourecegraph/about')
        })
    })
})
