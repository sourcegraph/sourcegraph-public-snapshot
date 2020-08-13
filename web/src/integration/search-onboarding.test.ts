import { Driver, createDriverForTest } from '../../../shared/src/testing/driver'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { saveScreenshotsUponFailures } from '../../../shared/src/testing/screenshotReporter'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteID, siteGQLID } from './jscontext'
import assert from 'assert'

describe.only('Search onboarding', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            ViewerSettings: () => ({
                viewerSettings: {
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: JSON.stringify({ experimentalFeatures: { showOnboardingTour: true } }),
                            },
                        },
                        {
                            __typename: 'Site',
                            id: siteGQLID,
                            siteID,
                            latestSettings: {
                                id: 470,
                                contents: JSON.stringify({ experimentalFeatures: { showOnboardingTour: true } }),
                            },
                            settingsURL: '/site-admin/global-settings',
                            viewerCanAdminister: true,
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
        })
    })
    saveScreenshotsUponFailures(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('Onboarding', () => {
        it('displays all steps in the language onboarding flow', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.test-tour-language-button')
            await driver.page.click('.test-tour-language-button')
            await driver.page.waitForSelector('#monaco-query-input')
            // eslint-disable-next-line unicorn/prefer-text-content
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'lang:')

            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('typescript')
            await driver.page.waitForSelector('.test-tour-step-3')
            await driver.page.keyboard.type(' test')
            await driver.page.waitForSelector('.test-tour-step-4')
            await driver.page.click('.test-search-help-dropdown-button-icon')
            await driver.page.waitForSelector('.test-tour-step-5')
            await driver.page.click('.test-search-button')
            await driver.assertWindowLocation('/search?q=lang:typescript+test&patternType=literal')
        })
    })
})
