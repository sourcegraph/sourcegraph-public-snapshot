import assert from 'assert'

import delay from 'delay'
import { Key } from 'ts-key-enum'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { InsightViewNode } from '../../graphql-operations'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'

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
            customContext: {
                // Enforce using a new gql API for code insights pages
                codeInsightsGqlApiEnabled: true,
            },
        })
    })

    after(() => driver?.close())
    afterEach(() => testContext?.dispose())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('should update user settings if drill-down filters have been persisted', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock back-end insights with standard gql API handler
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' })],
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

                UpdateLineChartSearchInsight: () => ({
                    __typename: 'Mutation',
                    updateLineChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' }),
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await driver.page.waitForSelector('svg circle')

        await driver.page.click('button[aria-label="Filters"]')
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]')
        await driver.page.type('[name="excludeRepoRegexp"]', 'github.com/sourcegraph/sourcegraph')

        // Wait until async validation of the search context field is passed
        await delay(500)

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
                searchContexts: [],
                includeRepoRegex: '',
                excludeRepoRegex: 'github.com/sourcegraph/sourcegraph',
            },
        })
    })

    it('should create a new insight with predefined filters via drill-down flow insight creation', async () => {
        const insightWithFilters: InsightViewNode = {
            ...createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' }),
            appliedFilters: {
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
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [insightWithFilters],
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

                FirstStepCreateSearchBasedInsight: () => ({
                    __typename: 'Mutation',
                    createLineChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' }),
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

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
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
        }, 'UpdateLineChartSearchInsight')

        assert.deepStrictEqual(variables.input, {
            dataSeries: [
                {
                    seriesId: '001',
                    query: 'patternType:regex case:yes \\*\\sas\\sGQL',
                    options: {
                        label: 'Imports of old GQL.* types',
                        lineColor: 'var(--oc-red-7)',
                    },
                    repositoryScope: {
                        repositories: [],
                    },
                    timeScope: {
                        stepInterval: {
                            unit: 'WEEK',
                            value: 6,
                        },
                    },
                },
                {
                    seriesId: '002',
                    query: "patternType:regexp case:yes /graphql-operations'",
                    options: {
                        label: 'Imports of new graphql-operations types',
                        lineColor: 'var(--oc-blue-7)',
                    },
                    repositoryScope: {
                        repositories: [],
                    },
                    timeScope: {
                        stepInterval: {
                            unit: 'WEEK',
                            value: 6,
                        },
                    },
                },
            ],
            presentationOptions: {
                title: 'Migration to new GraphQL TS types',
            },
            viewControls: {
                filters: {
                    searchContexts: [],
                    includeRepoRegex: 'github.com/sourcegraph/sourcegraph',
                    excludeRepoRegex: 'github.com/sourcegraph/sourcegraph',
                },
            },
        })
    })
})
