import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../../context'
import { GET_DASHBOARD_INSIGHTS_EMPTY, INSIGHTS_DASHBOARDS } from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

describe('[Code Insight] Dashboard', () => {
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
    afterEach(() => testContext?.dispose())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('creates dashboard through dashboard creation UI', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                InsightsDashboards: () => INSIGHTS_DASHBOARDS,
                GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS_EMPTY,
                InsightSubjects: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: 'user_001',
                        organizations: { nodes: [] },
                    },
                    site: { __typename: 'Site', id: 'site_id' },
                }),
                CreateDashboard: () => ({
                    createInsightsDashboard: {
                        __typename: 'InsightsDashboardPayload',
                        dashboard: {
                            __typename: 'InsightsDashboard',
                            id: '001',
                            title: '',
                            views: { nodes: [] },
                            grants: {
                                __typename: 'InsightsPermissionGrants',
                                users: [],
                                organizations: [],
                                global: true,
                            },
                        },
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')

        await driver.page.waitForSelector('[aria-label="Add dashboard"]')
        await driver.page.click('[aria-label="Add dashboard"]')

        await driver.page.waitForSelector('form')

        await driver.page.type('[name="name"]', 'New test dashboard')
        await driver.page.click('[name="visibility"][value="site_id"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Add dashboard')]")

            if (button) {
                await button.click()
            }
        }, 'CreateDashboard')

        assert.deepStrictEqual(variables, {
            input: {
                title: 'New test dashboard',
                grants: {
                    users: [],
                    organizations: [],
                    global: true,
                },
            },
        })
    })
})
