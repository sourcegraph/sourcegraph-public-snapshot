import delay from 'delay'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { percySnapshotWithVariants } from '../utils'

import {
    BACKEND_INSIGHTS,
    CODE_STATS_RESULT_MOCK,
    SEARCH_INSIGHT_COMMITS_MOCK,
    SEARCH_INSIGHT_RESULT_MOCK,
} from './utils/insight-mock-data'
import { overrideGraphQLExtensions } from './utils/override-insights-graphql'

describe('[VISUAL] Code insights page', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    before(async () => {
        driver = await createDriverForTest({
            defaultViewport: { width: 1920 },
        })
    })

    after(() => driver?.close())

    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    async function takeChartSnapshot(name: string): Promise<void> {
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')
        await delay(500)
        await percySnapshotWithVariants(driver.page, name)
    }

    it('is styled correctly with back-end insights', async () => {
        overrideGraphQLExtensions({
            testContext,
            userSettings: {
                'insights.allrepos': {
                    'searchInsights.insight.backend_ID_001': {},
                },
            },
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')

        await takeChartSnapshot('Code insights page with back-end insights only')
    })

    it('is styled correctly with search-based insights ', async () => {
        overrideGraphQLExtensions({
            testContext,

            // Since search insight and code stats insight are working via user/org
            // settings. We have to mock them by mocking user settings and provide
            // mock data - mocking extension work.
            userSettings: {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'The First search-based insight',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [
                        {
                            name: 'The first series of the first chart',
                            stroke: 'var(--oc-grape-7)',
                            query: 'Kapica',
                        },
                    ],
                    step: {
                        months: 8,
                    },
                },
                'searchInsights.insight.teamSize': {
                    title: 'The Second search-based insight',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [
                        {
                            name: 'The second series of the second chart',
                            stroke: 'var(--oc-blue-7)',
                            query: 'Korolev',
                        },
                    ],
                    step: {
                        months: 8,
                    },
                },
                'insights.allrepos': {},
            },
            overrides: {
                BulkSearchCommits: () => SEARCH_INSIGHT_COMMITS_MOCK,
                BulkSearch: () => SEARCH_INSIGHT_RESULT_MOCK,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with search-based insights only')
    })

    it('is styled correctly with errored insight', async () => {
        overrideGraphQLExtensions({
            testContext,

            // Since search insight and code stats insights work via user/org
            // settings. We have to mock them by mocking user settings and provide
            // mock settings cascade data.
            userSettings: {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'The First search-based insight',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [
                        {
                            name: 'The first series of the first chart',
                            stroke: 'var(--oc-grape-7)',
                            query: 'Kapica',
                        },
                    ],
                    step: {
                        months: 8,
                    },
                },
                'insights.allrepos': {
                    'searchInsights.insight.backend_ID_001': {},
                },
            },
            overrides: {
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),
                BulkSearchCommits: () => ({ error: 'Inappropriate data shape will cause an insight error' } as any),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with search-based errored insight')
    })

    it('is styled correctly with all types of insight', async () => {
        overrideGraphQLExtensions({
            testContext,

            // Since search insight and code stats insight are working via user/org
            // settings. We have to mock them by mocking user settings and provide
            // mock data - mocking extension work.
            userSettings: {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'The First search-based insight',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [
                        {
                            name: 'The first series of the first chart',
                            stroke: 'var(--oc-grape-7)',
                            query: 'Kapica',
                        },
                    ],
                    step: {
                        months: 8,
                    },
                },
                'codeStatsInsights.insight.langUsage': {
                    title: 'Adobe lang stats usage',
                    repository: 'ghe.sgdev.org/sourcegraph/adobe-adobe.github.com',
                    otherThreshold: 0.03,
                },
                'insights.allrepos': {
                    'searchInsights.insight.backend_ID_001': {},
                },
            },
            overrides: {
                // Backend insight mock
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),

                // Search built-in insight mock
                BulkSearchCommits: () => SEARCH_INSIGHT_COMMITS_MOCK,
                BulkSearch: () => SEARCH_INSIGHT_RESULT_MOCK,

                // Code stats built-in insight mock
                LangStatsInsightContent: () => CODE_STATS_RESULT_MOCK,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with all types of insight')
    })
})
