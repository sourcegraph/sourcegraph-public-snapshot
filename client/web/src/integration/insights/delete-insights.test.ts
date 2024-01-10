import assert from 'assert'

import { afterEach, beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../context'

import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from './fixtures/calculated-insights'
import { createJITMigrationToGQLInsightMetadataFixture } from './fixtures/insights-metadata'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

describe('Code insights page', () => {
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

    it('should update user/org settings if insight delete happened', async () => {
        overrideInsightsGraphQLApi({
            testContext,

            // Since search insight and code stats insights work via user/org
            // settings. We have to mock them by mocking user settings cascade.
            // userSettings: settings,
            overrides: {
                GetAllInsightConfigurations: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' })],
                        pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
                        totalCount: 1,
                    },
                }),
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),

                DeleteInsightView: () => ({
                    __typename: 'Mutation',
                    deleteInsightView: {
                        __typename: 'EmptyResponse',
                        alwaysNil: '',
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/all')
        await driver.page.waitForSelector('svg circle')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-card.001"] [data-testid="InsightContextMenuButton"]')

            await driver.page.click('[data-testid="insight-context-menu-delete-button"]')
            const [button] = await driver.page.$x("//button[contains(., 'Delete forever')]")
            if (button) {
                await button.click()
            }
        }, 'DeleteInsightView')

        assert.strictEqual(variables.id, '001')
    })
})
