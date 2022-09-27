import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../../context'
import { EMPTY_DASHBOARD, GET_DASHBOARD_INSIGHTS, INSIGHTS_DASHBOARDS } from '../fixtures/dashboards'
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

    it('updates empty dashboard', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                InsightsDashboards: () => INSIGHTS_DASHBOARDS,
                GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS,
                InsightSubjects: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: testUserID,
                        organizations: { nodes: [] },
                    },
                    site: { __typename: 'Site', id: 'site_id' },
                }),
                UpdateDashboard: () => ({
                    updateInsightsDashboard: {
                        __typename: 'InsightsDashboardPayload',
                        dashboard: {
                            __typename: 'InsightsDashboard',
                            id: EMPTY_DASHBOARD.id,
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

        await driver.page.goto(driver.sourcegraphBaseUrl + `/insights/dashboards/${EMPTY_DASHBOARD.id}`)

        await driver.page.waitForSelector('[data-testid="dashboard-context-menu"]')
        await driver.page.click('[data-testid="dashboard-context-menu"]')
        await driver.page.click('[data-testid="configure-dashboard"]')

        await driver.page.waitForSelector('form')

        await driver.page.type('[name="name"]', ' Edited test dashboard title')
        await driver.page.click(`[name="visibility"][value="${testUserID}"]`)

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Save changes')]")

            if (button) {
                await button.click()
            }
        }, 'UpdateDashboard')

        assert.deepStrictEqual(variables, {
            id: EMPTY_DASHBOARD.id,
            input: {
                title: 'Empty Dashboard Edited test dashboard title',
                grants: {
                    users: [testUserID],
                    organizations: [],
                    global: false,
                },
            },
        })
    })
})
