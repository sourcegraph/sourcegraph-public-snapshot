import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../../context'
import {
    GET_DASHBOARD_INSIGHTS_POPULATED,
    GET_INSIGHT_VIEW_1,
    GET_INSIGHT_VIEW_2,
    INSIGHTS_DASHBOARDS,
    LANG_STAT_INSIGHT_CONTENT,
} from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

describe('Code insights populated dashboard', () => {
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

    it('renders a dashboard with each of the available insight types', async () => {
        overrideInsightsGraphQLApi({ testContext })
        testContext.overrideGraphQL({
            GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS_POPULATED,
            GetInsightView: () => GET_INSIGHT_VIEW_1,
            InsightsDashboards: () => INSIGHTS_DASHBOARDS,
            LangStatsInsightContent: () => LANG_STAT_INSIGHT_CONTENT,
        })
        testContext.overrideGraphQL({ GetInsightView: () => GET_INSIGHT_VIEW_2 })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/EACH_TYPE_OF_INSIGHT')

        const dashboardSelectButton = await driver.page.waitForSelector('[data-testid="dashboard-select-button"')

        assert(dashboardSelectButton)

        const dashboardSelectButtonText: string = (await driver.page.evaluate(
            button => button.textContent,
            dashboardSelectButton
        )) as string

        assert(/each type of insight/i.test(dashboardSelectButtonText))
    })
})
