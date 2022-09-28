import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { GetDashboardInsightsResult, InsightsDashboardNode, InsightViewNode } from '../../../graphql-operations'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../../context'
import {
    LANG_STATS_INSIGHT,
    LANG_STAT_INSIGHT_CONTENT,
    CAPTURE_GROUP_INSIGHT,
    GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT,
} from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

interface DashboardMockOptions {
    id?: string
    title?: string
    insightIds?: string[]
}

/**
 * Creates dashboard mock entity in order to mock insight dashboard configurations.
 * It's used for easier mocking dashboards list gql handler see InsightsDashboards entry point.
 */
function createDashboard(options: DashboardMockOptions): InsightsDashboardNode {
    const { id = '001_dashboard', title = 'Default dashboard', insightIds = [] } = options

    return {
        __typename: 'InsightsDashboard',
        id,
        title,
        views: {
            __typename: 'InsightViewConnection',
            nodes: insightIds.map(id => ({ __typename: 'InsightView', id })),
        },
        grants: {
            __typename: 'InsightsPermissionGrants',
            users: [testUserID],
            organizations: [],
            global: false,
        },
    }
}

interface DashboardViewMockOptions {
    id?: string
    insightsMocks?: InsightViewNode[]
}

/**
 * Creates mocks for the dashboard view entity, you may ask what difference between createDashboard
 * and this mock helper. Dashboard view mock contains insight configurations, dashboard mocks contain
 * only insight ids. (we should mock dashboard view only when this dashboard is opened on the page)
 */
function createDashboardViewMock(options: DashboardViewMockOptions): GetDashboardInsightsResult {
    const { id = '001_dashboard', insightsMocks = [] } = options

    return {
        insightsDashboards: {
            __typename: 'InsightsDashboardConnection',
            nodes: [
                {
                    __typename: 'InsightsDashboard',
                    id,
                    views: {
                        __typename: 'InsightViewConnection',
                        nodes: insightsMocks.map(mock => ({ ...mock, __typename: 'InsightView' })),
                    },
                },
            ],
        },
    }
}

describe('Code insights [Dashboard card]', () => {
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

    it('renders lang stats insight card with proper options context', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock list of possible code insights dashboards on the dashboard page
                InsightsDashboards: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: testUserID,
                        organizations: { nodes: [] },
                    },
                    insightsDashboards: {
                        __typename: 'InsightsDashboardConnection',
                        nodes: [
                            createDashboard({ id: 'DASHBOARD_WITH_LANG_INSIGHT', insightIds: [LANG_STATS_INSIGHT.id] }),
                        ],
                    },
                }),

                // Mock dashboard configuration (dashboard content)
                GetDashboardInsights: () =>
                    createDashboardViewMock({ id: 'DASHBOARD_WITH_LANG_INSIGHT', insightsMocks: [LANG_STATS_INSIGHT] }),

                // Mock lang stats insight content
                LangStatsInsightContent: () => LANG_STAT_INSIGHT_CONTENT,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_LANG_INSIGHT')
        await driver.page.waitForSelector('[aria-label="Pie chart"]')

        const numberOfArcs = await driver.page.$$eval('[aria-label="Pie chart"] path', elements => elements.length)
        const numberHeadings = await driver.page.$$eval('[aria-label="Pie chart"] h3', elements => elements.length)

        // Why 12?, because LANG_STAT_INSIGHT_CONTENT mock has 14 entries and only five of them
        // are rendered because all other (7 groups) are too small, and they are grouped and presented
        // by special "Other" category (5 original rendered groups + one special Other group = 6)
        // 12 paths because we render 2 paths per pie part (one for the pie arc itself and one for the annotation line)
        assert.strictEqual(numberOfArcs, 12)

        // Why 6?, because LANG_STAT_INSIGHT_CONTENT mock has 14 entries and only five of them
        // are rendered because all other (7 groups) are too small, and they are grouped and presented
        // by special "Other" category (5 original rendered groups + one special Other group = 6)
        assert.strictEqual(numberHeadings, 6)

        const filtersButton = await driver.page.$('[aria-label="Active filters"], [aria-label="Filters"]')

        // Lang's stats insight doesn't support filters functionality, so we shouldn't have any filter UI
        assert.strictEqual(filtersButton, null)

        await driver.page.click('[aria-label="Insight options"]')

        const menuOptions = await driver.page.$$eval('[role="dialog"][aria-modal="true"] [role="menuitem"]', elements =>
            elements.map(element => element.textContent)
        )

        // Check that Pie chart doesn't have anything non pie chart specific menu options
        assert.deepStrictEqual(menuOptions, ['Edit', 'Get shareable link', 'Remove from this dashboard', 'Delete'])
    })

    it('renders capture group insight card with proper options context', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock list of possible code insights dashboards on the dashboard page
                InsightsDashboards: () => ({
                    currentUser: {
                        __typename: 'User',
                        id: testUserID,
                        organizations: { nodes: [] },
                    },
                    insightsDashboards: {
                        __typename: 'InsightsDashboardConnection',
                        nodes: [
                            createDashboard({
                                id: 'DASHBOARD_WITH_CAPTURE_GROUP',
                                insightIds: [CAPTURE_GROUP_INSIGHT.id],
                            }),
                        ],
                    },
                }),
                // Mock dashboard configuration (dashboard content) with one capture group insight configuration
                GetDashboardInsights: () =>
                    createDashboardViewMock({
                        id: 'DASHBOARD_WITH_CAPTURE_GROUP',
                        insightsMocks: [CAPTURE_GROUP_INSIGHT],
                    }),

                // Mock capture group insight content
                GetInsightView: () => GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_CAPTURE_GROUP')
        await driver.page.waitForSelector('[aria-label="Line chart"]')

        const numberOfLines = await driver.page.$$eval('[aria-label="Line chart"] path', elements => elements.length)
        const numberOfPointLinks = await driver.page.$$eval('[aria-label="Line chart"] a', elements => elements.length)

        // Why 20 and 189? See GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT dataset mock, it has 20 lines and 189 points
        assert.strictEqual(numberOfLines, 20)
        assert.strictEqual(numberOfPointLinks, 189)

        await driver.page.click('[aria-label="Filters"]')
        const filterPanel = await driver.page.$('[aria-label="Drill-down filters panel"]')

        // Should open filter panel on filter panel icon click
        assert.strictEqual(filterPanel !== null, true)

        // // Toggle insight filters (close filters panel)
        await driver.page.click('[aria-label="Filters"]')

        await driver.page.click('[aria-label="Insight options"]')

        const menuOptions = await driver.page.$$eval('[role="dialog"][aria-modal="true"] [role="menuitem"]', elements =>
            elements.map(element => element.textContent)
        )

        // Check that Line chart doesn't have anything non-related to capture group menu options
        assert.deepStrictEqual(menuOptions, ['Edit', 'Get shareable link', 'Remove from this dashboard', 'Delete'])
    })
})
