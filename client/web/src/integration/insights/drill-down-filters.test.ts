import assert from 'assert'

import delay from 'delay'
import { afterEach, beforeEach, describe, it } from 'mocha'
import { Key } from 'ts-key-enum'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { InsightViewNode } from '../../graphql-operations'
import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../context'

import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from './fixtures/calculated-insights'
import { createJITMigrationToGQLInsightMetadataFixture } from './fixtures/insights-metadata'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

describe('Backend insight drill down filters', () => {
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

    it('should update the insight configuration if drill-down filters have been persisted', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock back-end insights with standard gql API handler
                GetAllInsightConfigurations: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' })],
                        pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
                        totalCount: 1,
                    },
                }),

                // Calculated insight mock
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),

                GetSearchContexts: () => ({
                    __typename: 'Query',
                    searchContexts: {
                        __typename: 'SearchContextConnection',
                        nodes: [],
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                }),

                GetSearchContextByName: () => ({
                    searchContexts: {
                        __typename: 'SearchContextConnection',
                        nodes: [
                            {
                                __typename: 'SearchContext',
                                spec: '@sourcegraph/sourcegraph',
                                query: 'repo:github.com/sourcegraph/sourcegraph',
                            },
                        ],
                    },
                }),

                UpdateLineChartSearchInsight: () => ({
                    __typename: 'Mutation',
                    updateLineChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' }),
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/all')
        await driver.page.waitForSelector('svg circle')

        await driver.page.click('button[aria-label="Filters"]')

        // fill in the excludeRepoRegexp filter
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]')
        await driver.page.type('[name="excludeRepoRegexp"]', 'github.com/sourcegraph/sourcegraph')

        // fill in the search context filter regexp
        await driver.page.click('button[aria-label="search context filter section"]')
        await driver.page.type('[name="context"]', '@sourcegraph/sourcegraph')

        // Wait until async validation of the search context field is passed
        await delay(1000)

        // Close the drill-down filter panel
        await driver.page.keyboard.press(Key.Escape)
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]', {
            hidden: true,
        })

        // In this time we should see active button state (filter dot should appear if we've got some filters)
        await driver.page.click('button[aria-label="Active filters"]')

        // Wait until async validation of the search context field is passed
        await delay(500)

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[role="dialog"][aria-label="Drill-down filters panel"] button[type="submit"]')
        }, 'UpdateLineChartSearchInsight')

        assert.deepStrictEqual(variables.input.viewControls, {
            filters: {
                searchContexts: ['@sourcegraph/sourcegraph'],
                includeRepoRegex: '',
                excludeRepoRegex: 'github.com/sourcegraph/sourcegraph',
            },
            seriesDisplayOptions: {
                limit: null,
                numSamples: null,
                sortOptions: {
                    direction: 'DESC',
                    mode: 'RESULT_COUNT',
                },
            },
        })
    })

    it('should create a new insight with predefined filters via drill-down flow insight creation', async () => {
        const insightWithFilters: InsightViewNode = {
            ...createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated', id: 'view_1' }),
            defaultFilters: {
                __typename: 'InsightViewFilters',
                searchContexts: [],
                includeRepoRegex: '',
                excludeRepoRegex: 'github.com/sourcegraph/sourcegraph',
            },
        }

        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock back-end insights with standard gql API handler
                GetAllInsightConfigurations: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [insightWithFilters],
                        pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
                        totalCount: 1,
                    },
                }),

                // Calculated insight mock
                GetInsightView: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE],
                    },
                }),

                GetSearchContexts: () => ({
                    __typename: 'Query',
                    searchContexts: {
                        __typename: 'SearchContextConnection',
                        nodes: [],
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                }),

                SaveInsightAsNewView: () => ({
                    __typename: 'Mutation',
                    saveInsightAsNewView: {
                        __typename: 'InsightViewPayload',
                        view: createJITMigrationToGQLInsightMetadataFixture({ id: 'view_2', type: 'calculated' }),
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/all')
        await driver.page.waitForSelector('svg circle')

        await driver.page.click('button[aria-label="Active filters"]')
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]')

        await driver.page.type('[name="includeRepoRegexp"]', 'github.com/sourcegraph/sourcegraph')

        // Wait until async validation of the search context field is passed
        await delay(500)

        await driver.page.click(
            '[role="dialog"][aria-label="Drill-down filters panel"] button[data-testid="save-as-new-view-button"]'
        )

        await driver.page.type('[name="insightName"]', 'Insight with filters')

        // Wait until async validation of the insight name field will pass
        await delay(500)

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[role="dialog"][aria-label="Drill-down filters panel"] button[type="submit"]')
        }, 'SaveInsightAsNewView')

        assert.deepStrictEqual(variables.input, {
            insightViewId: 'view_1',
            dashboard: null,
            options: {
                title: 'Insight with filters',
            },
            viewControls: {
                filters: {
                    includeRepoRegex: 'github.com/sourcegraph/sourcegraph',
                    excludeRepoRegex: 'github.com/sourcegraph/sourcegraph',
                    searchContexts: [''],
                },
                seriesDisplayOptions: {
                    limit: null,
                    numSamples: null,
                    sortOptions: {
                        direction: 'DESC',
                        mode: 'RESULT_COUNT',
                    },
                },
            },
        })
    })
})
