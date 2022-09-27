import assert from 'assert'

import expect from 'expect'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'

import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from './fixtures/calculated-insights'
import { createJITMigrationToGQLInsightMetadataFixture } from './fixtures/insights-metadata'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

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
            customContext: {
                // Enforce using a new gql API for code insights pages
                codeInsightsGqlApiEnabled: true,
            },
        })
    })

    after(() => driver?.close())
    afterEach(() => testContext?.dispose())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    beforeEach(() => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                GetDashboardInsights: () => ({
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
                }),
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),
                GetAccessibleInsightsList: () => ({
                    insightViews: {
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
                }),
                InsightsDashboards: () => ({
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
                                views: { nodes: [] },
                                grants: {
                                    users: [],
                                    organizations: [],
                                    global: true,
                                },
                            },
                        ],
                    },
                }),
                AddInsightViewToDashboard: () => ({
                    addInsightViewToDashboard: {
                        dashboard: { id: 'codeInsightDashboard001' },
                    },
                }),
            },
        })
    })

    it('"add" and "remove" insight mutation work properly', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await driver.page.waitForSelector('[data-reach-listbox-input]')
        await driver.page.click('[data-reach-listbox-input]')
        await driver.page.click('[data-value="codeInsightDashboard001"]')

        expect(driver.page.url()).toBe(`${driver.sourcegraphBaseUrl}/insights/dashboards/codeInsightDashboard001`)

        await driver.page.click('[data-testid="add-or-remove-insights"]')
        await driver.page.waitForSelector('form')

        await driver.page.click('input[value="insight_003"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Save')]")

            if (button) {
                await button.click()
            }
        }, 'AddInsightViewToDashboard')

        assert.deepStrictEqual(variables, {
            dashboardId: 'codeInsightDashboard001',
            insightViewId: 'insight_003',
        })
    })
})
