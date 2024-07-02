import assert from 'assert'

import delay from 'delay'
import { afterEach, beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../context'

import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from './fixtures/calculated-insights'
import { createJITMigrationToGQLInsightMetadataFixture } from './fixtures/insights-metadata'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

describe('Code insights single insight page', () => {
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

    async function takeChartSnapshot(name: string): Promise<void> {
        await driver.page.waitForSelector('svg circle')
        await delay(500)
    }

    it('is styled correctly with common backend insights', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock insight config query
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ id: '001', type: 'calculated' })],
                    },
                }),
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/insight/001')

        await takeChartSnapshot('Code insights single insight page with common backend insight')
    })

    it('is re-fetched correctly if any of sort and limit filters have been changed', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock insight config query
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ id: '001', type: 'calculated' })],
                    },
                }),
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),

                UpdateLineChartSearchInsight: () => ({
                    __typename: 'Mutation',
                    updateLineChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: createJITMigrationToGQLInsightMetadataFixture({ id: '001', type: 'calculated' }),
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/insight/001')
        await driver.page.waitForSelector('[aria-label="Open filters panel"]')
        await driver.page.click('[aria-label="Open filters panel"]')
        await driver.page.waitForSelector('[aria-label="Sort by name with ascending order"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[aria-label="Sort by name with ascending order"]')
            await driver.page.click('input[aria-label="Number of data series"]', { clickCount: 3 })
            await driver.page.keyboard.type('2')
        }, 'GetInsightView')

        assert.deepStrictEqual(variables.filters, {
            searchContexts: [''],
            includeRepoRegex: '',
            excludeRepoRegex: '',
        })

        assert.deepStrictEqual(variables.seriesDisplayOptions, {
            limit: 2,
            numSamples: null,
            sortOptions: {
                direction: 'ASC',
                mode: 'LEXICOGRAPHICAL',
            },
        })
    })
})
