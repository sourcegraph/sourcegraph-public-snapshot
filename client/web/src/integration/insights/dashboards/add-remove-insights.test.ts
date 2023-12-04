import assert from 'assert'

import expect from 'expect'
import { beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { FindInsightsBySearchTermResult } from '../../../graphql-operations'
import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
import { GET_DASHBOARD_INSIGHTS_EMPTY, INSIGHTS_DASHBOARDS } from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

const ALL_AVAILABLE_INSIGHTS_LIST: FindInsightsBySearchTermResult = {
    insightViews: {
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 3,
        nodes: [
            {
                __typename: 'InsightView',
                id: 'insight_001',
                presentation: {
                    __typename: 'LineChartInsightViewPresentation',
                    title: 'First Insight',
                },
                dataSeriesDefinitions: [
                    {
                        __typename: 'SearchInsightDataSeriesDefinition',
                        query: 'Test query 1',
                        groupBy: null,
                        generatedFromCaptureGroups: false,
                    },
                ],
            },
            {
                __typename: 'InsightView',
                id: 'insight_002',
                presentation: {
                    __typename: 'LineChartInsightViewPresentation',
                    title: 'Second Insight',
                },
                dataSeriesDefinitions: [
                    {
                        __typename: 'SearchInsightDataSeriesDefinition',
                        query: 'Test query 2',
                        groupBy: null,
                        generatedFromCaptureGroups: false,
                    },
                ],
            },
            {
                __typename: 'InsightView',
                id: 'insight_003',
                presentation: {
                    __typename: 'LineChartInsightViewPresentation',
                    title: 'Third Insight',
                },
                dataSeriesDefinitions: [
                    {
                        __typename: 'SearchInsightDataSeriesDefinition',
                        query: 'Test query 3',
                        groupBy: null,
                        generatedFromCaptureGroups: false,
                    },
                ],
            },
        ],
    },
}

describe('Code insights empty dashboard', () => {
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
    })

    after(() => driver?.close())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('"add" and "remove" insight mutation work properly', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                GetAllInsightConfigurations: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [],
                        pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
                        totalCount: 0,
                    },
                }),
                InsightsDashboards: () => INSIGHTS_DASHBOARDS,
                GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS_EMPTY,
                FindInsightsBySearchTerm: () => ALL_AVAILABLE_INSIGHTS_LIST,
                AddInsightViewToDashboard: () => ({
                    addInsightViewToDashboard: {
                        dashboard: { id: 'EMPTY_DASHBOARD' },
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/all')
        await driver.page.waitForSelector('button[role="tab"]')
        await (await driver.page.$x("//button[contains(., 'Dashboards')]"))[0].click()

        // Check that first personal dashboard redirection works as expected (pick first personal dashboard
        // if URL doesn't have dashboard id
        await driver.page.waitForSelector('[aria-label="Choose a dashboard, Empty Dashboard"]')
        expect(driver.page.url()).toBe(`${driver.sourcegraphBaseUrl}/insights/dashboards/EMPTY_DASHBOARD`)

        await (await driver.page.$x("//button[contains(., 'Add or remove insights')]"))[0].click()
        // Wait for the suggestion list item are rendered
        await driver.page.waitForSelector('form ul li')

        await (await driver.page.$x("//li[contains(., 'Second Insight')]"))[0].click()

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Save')]")

            if (button) {
                await button.click()
            }
        }, 'AddInsightViewToDashboard')

        assert.deepStrictEqual(variables, {
            dashboardId: 'EMPTY_DASHBOARD',
            insightViewId: 'insight_002',
        })
    })
})
