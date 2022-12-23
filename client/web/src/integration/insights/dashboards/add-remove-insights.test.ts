import assert from 'assert'

import expect from 'expect'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { GetDashboardAccessibleInsightsResult } from '../../../graphql-operations'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../../context'
import { GET_DASHBOARD_INSIGHTS_EMPTY, INSIGHTS_DASHBOARDS } from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

const ALL_AVAILABLE_INSIGHTS_LIST: GetDashboardAccessibleInsightsResult = {
    dashboardInsightsIds: { nodes: [{ views: { nodes: [] } }] },
    accessibleInsights: {
        nodes: [
            {
                __typename: 'InsightView',
                id: 'insight_001',
                presentation: {
                    __typename: 'LineChartInsightViewPresentation',
                    title: 'First Insight',
                },
            },
            {
                __typename: 'InsightView',
                id: 'insight_002',
                presentation: {
                    __typename: 'LineChartInsightViewPresentation',
                    title: 'Second Insight',
                },
            },
            {
                __typename: 'InsightView',
                id: 'insight_003',
                presentation: {
                    __typename: 'LineChartInsightViewPresentation',
                    title: 'Third Insight',
                },
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
                InsightsDashboards: () => INSIGHTS_DASHBOARDS,
                GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS_EMPTY,
                GetDashboardAccessibleInsights: () => ALL_AVAILABLE_INSIGHTS_LIST,
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
        await driver.page.waitForSelector('form')

        await driver.page.click('input[value="insight_003"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Save')]")

            if (button) {
                await button.click()
            }
        }, 'AddInsightViewToDashboard')

        assert.deepStrictEqual(variables, {
            dashboardId: 'EMPTY_DASHBOARD',
            insightViewId: 'insight_003',
        })
    })
})
