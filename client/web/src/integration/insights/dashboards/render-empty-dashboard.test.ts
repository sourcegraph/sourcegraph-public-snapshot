import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../../context'
import { GET_DASHBOARD_INSIGHTS, INSIGHTS_DASHBOARDS } from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

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
            customContext: {
                // Enforce using a new gql API for code insights pages
                codeInsightsGqlApiEnabled: true,
            },
        })
    })

    after(() => driver?.close())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('renders empty dashboard', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                InsightsDashboards: () => INSIGHTS_DASHBOARDS,
                GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/EMPTY_DASHBOARD')

        const dashboardSelectButton = await driver.page.waitForSelector('[data-testid="dashboard-select-button"')
        const addInsightsButtonCard = await driver.page.waitForSelector('[data-testid="add-insights-button-card"')

        assert(dashboardSelectButton)

        const dashboardSelectButtonText: string = (await driver.page.evaluate(
            button => button.textContent,
            dashboardSelectButton
        )) as string
        const addInsightsButtonCardText: string = (await driver.page.evaluate(
            button => button.textContent,
            addInsightsButtonCard
        )) as string

        assert(/empty dashboard/i.test(dashboardSelectButtonText))
        assert(/add insights/i.test(addInsightsButtonCardText))
    })
})
