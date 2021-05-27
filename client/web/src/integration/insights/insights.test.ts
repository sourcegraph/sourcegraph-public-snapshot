
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { percySnapshotWithVariants } from '../utils'

import {
    BACKEND_INSIGHTS,
    INSIGHT_VIEW_TEAM_SIZE,
    INSIGHT_VIEW_TYPES_MIGRATION
} from './utils/insight-mock-data';
import { overrideGraphQLExtensions } from './utils/override-graphql-with-extensions';

describe('Code insights page', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    before(async () => {
        driver = await createDriverForTest({ sourcegraphBaseUrl: 'https://sourcegraph.test:3443', devtools: true })
    })

    after(() => driver?.close())

    beforeEach(async function () {

        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })

        overrideGraphQLExtensions({
            testContext,

            /**
             * Since search insight and code stats insight are working via user/org
             * settings. We have to mock them by mocking user settings and provide
             * mock data - mocking extension work.
             * */
            userSettings: {
                'searchInsights.insight.graphQLTypesMigration': {},
                'searchInsights.insight.teamSize': {},
            },
            insightExtensionsMocks: {
                'searchInsights.insight.teamSize': INSIGHT_VIEW_TEAM_SIZE,
                'searchInsights.insight.graphQLTypesMigration': {
                    ...INSIGHT_VIEW_TYPES_MIGRATION,
                    title: 'Migration to new GraphQL TS types',
                }
            },
            overrides: {
                /**
                 * Mock back-end insights with standard gql API handler.
                 * */
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),
            }
        })
    })

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('is styled correctly', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        await percySnapshotWithVariants(driver.page, 'Code insights page')
    })
})
