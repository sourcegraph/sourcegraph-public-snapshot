import assert from 'assert'

import { beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { GetDashboardInsightsResult, InsightsDashboardsResult } from '../../../graphql-operations'
import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from '../fixtures/calculated-insights'
import { createJITMigrationToGQLInsightMetadataFixture } from '../fixtures/insights-metadata'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

const DASHBOARD_PAGE_METADATA: InsightsDashboardsResult = {
    currentUser: {
        __typename: 'User',
        id: 'user001',
        organizations: { nodes: [] },
    },
    insightsDashboards: {
        nodes: [
            {
                __typename: 'InsightsDashboard',
                id: 'codeInsightDashboard001',
                title: 'Dashboard 001',
                grants: {
                    __typename: 'InsightsPermissionGrants',
                    users: [],
                    organizations: [],
                    global: true,
                },
            },
        ],
    },
}

const DASHBOARD_WITH_TWO_INSIGHTS_METADATA: GetDashboardInsightsResult = {
    insightsDashboards: {
        nodes: [
            {
                __typename: 'InsightsDashboard',
                id: 'codeInsightDashboard001',
                views: {
                    nodes: [
                        createJITMigrationToGQLInsightMetadataFixture({
                            id: 'insight_001',
                            type: 'calculated',
                        }),
                        createJITMigrationToGQLInsightMetadataFixture({
                            id: 'insight_002',
                            type: 'calculated',
                        }),
                    ],
                },
            },
        ],
    },
}

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

    it('can be updated through UI', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock dashboard page (dashboard list)
                InsightsDashboards: () => DASHBOARD_PAGE_METADATA,

                // Mock particular (picked) dashboard insights configuration list
                GetDashboardInsights: () => DASHBOARD_WITH_TWO_INSIGHTS_METADATA,

                // Mock dashboard insights datasets (for both insights we will use one
                // dataset)
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),

                // Mock dashboard list of possible users/orgs where this dashboard can be
                // included
                InsightSubjects: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: 'user_001',
                        organizations: { nodes: [] },
                    },
                    site: { __typename: 'Site', id: 'TestSiteID' },
                }),

                UpdateDashboard: () => ({
                    updateInsightsDashboard: {
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

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/codeInsightDashboard001')
        await driver.page.waitForSelector('[aria-label="dashboard options"]')
        await driver.page.click('[aria-label="dashboard options"]')

        const [configureButton] = await driver.page.$x("//button[contains(., 'Configure dashboard')]")
        await configureButton?.click()

        await driver.page.waitForSelector('form')
        await driver.page.type('[name="name"]', ' Edited test dashboard title')
        await driver.page.click('[name="visibility"][value="user_001"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Save changes')]")

            await button?.click()
        }, 'UpdateDashboard')

        assert.deepStrictEqual(variables, {
            id: 'codeInsightDashboard001',
            input: {
                title: 'Dashboard 001 Edited test dashboard title',
                grants: {
                    users: ['user_001'],
                    organizations: [],
                    global: false,
                },
            },
        })
    })
})
