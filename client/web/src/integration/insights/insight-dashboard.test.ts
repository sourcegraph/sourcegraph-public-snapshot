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

    it('creates dashboard through dashboard creation UI', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                ...testContext.overrideGraphQL,
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

        await driver.page.waitForSelector('[data-testid="add-dashboard-button"]')
        await driver.page.click('[data-testid="add-dashboard-button"]')
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

    it('updates existing dashboard properly', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                ...testContext.overrideGraphQL,
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
                                    __typename: 'InsightsPermissionGrants',
                                    users: [],
                                    organizations: [],
                                    global: true,
                                },
                            },
                        ],
                    },
                }),
                InsightSubjects: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: 'user_001',
                        organizations: { nodes: [] },
                    },
                    site: { __typename: 'Site', id: 'site_id' },
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
        await driver.page.waitForSelector('[data-testid="dashboard-context-menu"]')
        await driver.page.click('[data-testid="dashboard-context-menu"]')
        await driver.page.click('[data-testid="configure-dashboard"]')

        await driver.page.waitForSelector('form')

        await driver.page.type('[name="name"]', ' Edited test dashboard title')
        await driver.page.click('[name="visibility"][value="user_001"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            const [button] = await driver.page.$x("//button[contains(., 'Save changes')]")

            if (button) {
                await button.click()
            }
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
