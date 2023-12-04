import assert from 'assert'

import { beforeEach, describe, it } from 'mocha'
import type { Page } from 'puppeteer'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { GetInsightViewResult } from '../../../graphql-operations'
import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
import {
    CAPTURE_GROUP_INSIGHT,
    COMPUTE_INSIGHT,
    GET_DASHBOARD_INSIGHTS_POPULATED,
    GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT,
    GET_INSIGHT_VIEW_COMPUTE_INSIGHT,
    GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT,
    INSIGHTS_DASHBOARDS,
    LANG_STATS_INSIGHT,
    LANG_STAT_INSIGHT_CONTENT,
    SEARCH_BASED_INSIGHT,
} from '../fixtures/dashboards'
import { overrideInsightsGraphQLApi } from '../utils/override-insights-graphql-api'

describe('Code insights populated dashboard', () => {
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

    it('renders a dashboard with each of the available insight types', async () => {
        overrideInsightsGraphQLApi({ testContext })
        testContext.overrideGraphQL({
            ViewerSettings: () => ({
                viewerSettings: {
                    __typename: 'SettingsCascade',
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            id: 'TestDefaultSettingsID',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: JSON.stringify({
                                    experimentalFeatures: { codeInsightsCompute: true },
                                }),
                            },
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
            GetDashboardInsights: () => GET_DASHBOARD_INSIGHTS_POPULATED,
            GetInsightView: ({ id }) => getInsightViewById(id),
            InsightsDashboards: () => INSIGHTS_DASHBOARDS,
            LangStatsInsightContent: () => LANG_STAT_INSIGHT_CONTENT,
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/EACH_TYPE_OF_INSIGHT')

        const dashboardSelectButton = await driver.page.waitForSelector(
            '[aria-label="Choose a dashboard, Each Type of Insight"]'
        )

        assert(dashboardSelectButton)

        const dashboardSelectButtonText: string = (await driver.page.evaluate(
            button => button.textContent,
            dashboardSelectButton
        )) as string

        assert(/each type of insight/i.test(dashboardSelectButtonText))

        const expectedLinks = [
            CAPTURE_GROUP_INSIGHT.presentation.title,
            LANG_STATS_INSIGHT.presentation.title,
            SEARCH_BASED_INSIGHT.presentation.title,
            COMPUTE_INSIGHT.presentation.title,
        ]

        await driver.page.waitForSelector(`[aria-label="${CAPTURE_GROUP_INSIGHT.presentation.title} insight`)
        await driver.page.waitForSelector(`[aria-label="${LANG_STATS_INSIGHT.presentation.title} insight`)
        await driver.page.waitForSelector(`[aria-label="${SEARCH_BASED_INSIGHT.presentation.title} insight`)
        await driver.page.waitForSelector(`[aria-label="${COMPUTE_INSIGHT.presentation.title} insight`)

        const foundLinks = await getLinks(driver.page, expectedLinks)

        assert.deepStrictEqual(expectedLinks, foundLinks)
    })
})

/**
 * Helper function to get all links on the page that match a list of strings.
 *
 * @param page - Reference to the puppeteer page
 * @param labels - Array of labels that should be found on the page
 * @returns - Array of labels that were found on the page
 */
async function getLinks(page: Page, labels: string[]): Promise<string[]> {
    const foundLinks: string[] = []
    const links = await page.$$('a')

    for (const link of links) {
        const valueHandle = await link.getProperty('innerText')
        const linkText = await valueHandle.jsonValue()
        if (typeof linkText === 'string' && labels.includes(linkText)) {
            foundLinks.push(linkText)
        }
    }

    return foundLinks
}

/**
 * Helper function to get insight view by id.
 *
 * @param id - Insight id
 * @returns a GetInsightViewResult
 */
function getInsightViewById(id: string | null): GetInsightViewResult {
    const insightView = [
        GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT,
        GET_INSIGHT_VIEW_COMPUTE_INSIGHT,
        GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT,
    ].find(insight => insight.insightViews.nodes[0]?.id === id)

    if (!insightView) {
        throw new Error(`Insight view with id ${id} not found`)
    }

    return insightView
}
