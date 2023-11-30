import assert from 'assert'

import { beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
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
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('creates dashboard through dashboard creation UI', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                InsightSubjects: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: 'user_001',
                        organizations: { nodes: [] },
                    },
                    site: { __typename: 'Site', id: 'TestSiteID' },
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

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/all')

        await driver.page.waitForSelector('[aria-label="Add dashboard"]')
        await driver.page.click('[aria-label="Add dashboard"]')

        await driver.page.waitForSelector('form')

        await driver.page.type('[name="name"]', 'New test dashboard')
        await driver.page.click('[name="visibility"][value="TestSiteID"]')

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
