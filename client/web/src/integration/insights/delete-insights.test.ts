import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailedWithJest } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'

import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from './fixtures/calculated-insights'
import { createJITMigrationToGQLInsightMetadataFixture } from './fixtures/insights-metadata'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

describe('Code insights page', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    beforeAll(async () => {
        driver = await createDriverForTest()
    })

    beforeEach(async () => {
        testContext = await createWebIntegrationTestContext({
            driver,
            directory: __dirname,
            customContext: {
                // Enforce using a new gql API for code insights pages
                codeInsightsGqlApiEnabled: true,
            },
        })
    })

    afterAll(() => driver?.close())
    afterEach(() => testContext?.dispose())
    afterEachSaveScreenshotIfFailedWithJest(() => driver.page)

    it('should update user/org settings if insight delete happened', async () => {
        overrideInsightsGraphQLApi({
            testContext,

            // Since search insight and code stats insights work via user/org
            // settings. We have to mock them by mocking user settings cascade.
            // userSettings: settings,
            overrides: {
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' })],
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

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-card.001"] [data-testid="InsightContextMenuButton"]')

            await Promise.all([
                driver.acceptNextDialog(),
                driver.page.click('[data-testid="insight-context-menu-delete-button"]'),
            ])
        }, 'DeleteInsightView')

        assert.strictEqual(variables.id, '001')
    })
})
