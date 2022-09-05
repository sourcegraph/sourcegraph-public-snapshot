import delay from 'delay'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { percySnapshotWithVariants } from '../utils'

import { MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE } from './fixtures/calculated-insights'
import {
    createJITMigrationToGQLInsightMetadataFixture,
    SOURCEGRAPH_LANG_STATS_INSIGHT_METADATA_FIXTURE,
    STORYBOOK_GROWTH_INSIGHT_METADATA_FIXTURE,
} from './fixtures/insights-metadata'
import {
    SOURCEGRAPH_LANG_STATS_INSIGHT_DATA_FIXTURE,
    STORYBOOK_GROWTH_INSIGHT_COMMITS_FIXTURE,
    STORYBOOK_GROWTH_INSIGHT_MATCH_DATA_FIXTURE,
} from './fixtures/runtime-insights'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

// Disable these tests since SVG elements rendering is super flaky in the Percy sandbox,
// Enable these visual code insights tests when we finish the investigation around screenshot
// testing tooling.
describe.skip('[VISUAL] Code insights page', () => {
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

    async function takeChartSnapshot(name: string): Promise<void> {
        await driver.page.waitForSelector('svg circle')
        await delay(500)
        await percySnapshotWithVariants(driver.page, name)
        await accessibilityAudit(driver.page)
    }

    it('is styled correctly with back-end insights', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock insight config query
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
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')

        await takeChartSnapshot('Code insights page with back-end insights only')
    })

    it('is styled correctly with just-in-time insights ', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [STORYBOOK_GROWTH_INSIGHT_METADATA_FIXTURE],
                    },
                }),
                BulkSearchCommits: () => STORYBOOK_GROWTH_INSIGHT_COMMITS_FIXTURE,
                BulkSearch: () => STORYBOOK_GROWTH_INSIGHT_MATCH_DATA_FIXTURE,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with search-based insights only')
    })

    it('is styled correctly with errored insight', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [STORYBOOK_GROWTH_INSIGHT_METADATA_FIXTURE],
                    },
                }),
                BulkSearchCommits: () => ({ error: 'Inappropriate data shape will cause an insight error' } as any),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')

        await delay(500)
        await percySnapshotWithVariants(driver.page, 'Code insights page with search-based errored insight')
        await accessibilityAudit(driver.page)
    })

    it('is styled correctly with all types of insight', async () => {
        overrideInsightsGraphQLApi({
            testContext,

            overrides: {
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [
                            createJITMigrationToGQLInsightMetadataFixture({ type: 'calculated' }),
                            STORYBOOK_GROWTH_INSIGHT_METADATA_FIXTURE,
                            SOURCEGRAPH_LANG_STATS_INSIGHT_METADATA_FIXTURE,
                        ],
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

                // Search just-in-time insight mock
                BulkSearchCommits: () => STORYBOOK_GROWTH_INSIGHT_COMMITS_FIXTURE,
                BulkSearch: () => STORYBOOK_GROWTH_INSIGHT_MATCH_DATA_FIXTURE,

                // Code stats just-in-time insight mock
                LangStatsInsightContent: () => SOURCEGRAPH_LANG_STATS_INSIGHT_DATA_FIXTURE,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await takeChartSnapshot('Code insights page with all types of insight')
    })

    describe('Add dashboard page', () => {
        it('is styled correctly', async () => {
            overrideInsightsGraphQLApi({
                testContext,
                overrides: {
                    InsightSubjects: () => ({
                        currentUser: {
                            __typename: 'User',
                            id: '001',
                            displayName: 'Kapica',
                            username: 'kapica@sourcegraph.com',
                            viewerCanAdminister: true,
                            organizations: {
                                nodes: [
                                    {
                                        __typename: 'Org',
                                        name: 'test organization 1',
                                        displayName: 'Test organization 1',
                                        id: 'Org_test_id_001',
                                        viewerCanAdminister: true,
                                    },
                                    {
                                        __typename: 'Org',
                                        name: 'test organization 2',
                                        displayName: 'Test organization 2',
                                        id: 'Org_test_id_002',
                                        viewerCanAdminister: true,
                                    },
                                ],
                            },
                        },
                        site: {
                            __typename: 'Site',
                            id: '003',
                            allowSiteSettingsEdits: true,
                            viewerCanAdminister: true,
                        },
                    }),
                },
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/add-dashboard')
            await driver.page.waitForSelector('input[name="name"]')

            await percySnapshotWithVariants(driver.page, 'Code insights add new dashboard page')
            await accessibilityAudit(driver.page)
        })
    })
})
