import assert from 'assert'

import { beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
import { GET_DASHBOARD_INSIGHTS_EMPTY, INSIGHTS_DASHBOARDS } from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

describe('Code insights dashboard', () => {
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

    it('can be removed through UI', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                InsightsDashboards: () => INSIGHTS_DASHBOARDS,
                GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS_EMPTY,
                DeleteDashboard: () => ({ deleteInsightsDashboard: { alwaysNil: null } }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/EMPTY_DASHBOARD')
        await driver.page.waitForSelector('[aria-label="dashboard options"]')
        await driver.page.click('[aria-label="dashboard options"]')

        const deleteDashboard = await testContext.waitForGraphQLRequest(async () => {
            const [deleteButton] = await driver.page.$x("//button[contains(., 'Delete')]")
            await deleteButton?.click()

            const [deleteConfirmButton] = await driver.page.$x("//button[contains(., 'Delete forever')]")
            await deleteConfirmButton?.click()
        }, 'DeleteDashboard')

        assert(deleteDashboard.id, 'EMPTY_DASHBOARD')
    })
})
