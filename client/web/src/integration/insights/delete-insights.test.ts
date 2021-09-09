import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { emptyResponse } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'

import { SEARCH_INSIGHT_COMMITS_MOCK, SEARCH_INSIGHT_RESULT_MOCK } from './utils/insight-mock-data'
import { overrideGraphQLExtensions } from './utils/override-insights-graphql'

describe('Code insights page', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    before(async () => {
        driver = await createDriverForTest({
            devtools: true,
        })
    })

    after(() => driver?.close())

    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('should update user/org settings if insight delete happened', async () => {
        const settings = {
            'searchInsights.insight.graphQLTypesMigration': {
                title: 'The First search-based insight',
                repositories: ['github.com/sourcegraph/sourcegraph'],
                series: [
                    {
                        name: 'The first series of the first chart',
                        stroke: 'var(--oc-grape-7)',
                        query: 'Kapica',
                    },
                ],
                step: {
                    months: 8,
                },
            },
            'searchInsights.insight.teamSize': {
                title: 'The Second search-based insight',
                repositories: ['github.com/sourcegraph/sourcegraph'],
                series: [
                    {
                        name: 'The second series of the second chart',
                        stroke: 'var(--oc-blue-7)',
                        query: 'Korolev',
                    },
                ],
                step: {
                    months: 8,
                },
            },
        }

        overrideGraphQLExtensions({
            testContext,

            // Since search insight and code stats insights work via user/org
            // settings. We have to mock them by mocking user settings cascade.
            userSettings: settings,
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: [] } }),

                // Mock built-in search-based insight
                BulkSearchCommits: () => SEARCH_INSIGHT_COMMITS_MOCK,
                BulkSearch: () => SEARCH_INSIGHT_RESULT_MOCK,

                OverwriteSettings: () => ({
                    settingsMutation: {
                        overwriteSettings: {
                            empty: emptyResponse,
                        },
                    },
                }),

                SubjectSettings: () => ({
                    settingsSubject: {
                        latestSettings: {
                            id: 310,
                            contents: JSON.stringify(settings),
                        },
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click(
                '[data-testid="insight-card.searchInsights.insight.graphQLTypesMigration"] [data-testid="InsightContextMenuButton"]'
            )

            await Promise.all([
                driver.acceptNextDialog(),
                driver.page.click('[data-testid="insight-context-menu-delete-button"]'),
            ])
        }, 'OverwriteSettings')

        assert.deepStrictEqual(JSON.parse(variables.contents), {
            'searchInsights.insight.teamSize': {
                title: 'The Second search-based insight',
                repositories: ['github.com/sourcegraph/sourcegraph'],
                series: [
                    {
                        name: 'The second series of the second chart',
                        stroke: 'var(--oc-blue-7)',
                        query: 'Korolev',
                    },
                ],
                step: {
                    months: 8,
                },
            },
        })
    })
})
