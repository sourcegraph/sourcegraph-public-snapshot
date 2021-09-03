import delay from 'delay'
import { View } from 'sourcegraph'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { percySnapshotWithVariants } from '../utils'

import {
    BACKEND_INSIGHTS,
    CODE_STATS_INSIGHT_LANG_USAGE,
    INSIGHT_VIEW_TEAM_SIZE,
    INSIGHT_VIEW_TYPES_MIGRATION,
} from './utils/insight-mock-data'
import { overrideGraphQLExtensions } from './utils/override-graphql-with-extensions'

describe.skip('[VISUAL] Code insights page', () => {
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
                    repositories: [],
                    series: [],
                },
                'searchInsights.insight.teamSize': {
                    title: 'The Second search-based insight',
                    repositories: [],
                    series: [],
                },
                'insights.allrepos': {},
            },
            insightExtensionsMocks: {
                'searchInsights.insight.teamSize': INSIGHT_VIEW_TEAM_SIZE,
                'searchInsights.insight.graphQLTypesMigration': INSIGHT_VIEW_TYPES_MIGRATION,
            },
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: [] } }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with search-based insights only')
    })

    it('is styled correctly with errored insight', async () => {
        overrideGraphQLExtensions({
            testContext,

            // Since search insight and code stats insight are working via user/org
            // settings. We have to mock them by mocking user settings and provide
            // mock data - mocking extension work.
            userSettings: {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'The First search-based insight',
                    repositories: [],
                    series: [],
                },
                'searchInsights.insight.teamSize': {
                    title: 'The Second search-based insight',
                    repositories: [],
                    series: [],
                },
                'insights.allrepos': {},
            },
            insightExtensionsMocks: {
                'searchInsights.insight.teamSize': ({ message: 'Error message', name: 'hello' } as unknown) as View,
                'searchInsights.insight.graphQLTypesMigration': INSIGHT_VIEW_TYPES_MIGRATION,
            },
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: [] } }),
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
                    repositories: [],
                    series: [],
                },
                'searchInsights.insight.teamSize': {
                    title: 'The Second search-based insight',
                    repositories: [],
                    series: [],
                },
                'codeStatsInsights.insight.langUsage': {},
                'insights.allrepos': {},
            },
            insightExtensionsMocks: {
                'codeStatsInsights.insight.langUsage': CODE_STATS_INSIGHT_LANG_USAGE,
                'searchInsights.insight.teamSize': INSIGHT_VIEW_TEAM_SIZE,
                'searchInsights.insight.graphQLTypesMigration': INSIGHT_VIEW_TYPES_MIGRATION,
            },
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with all types of insight')
    })
})
