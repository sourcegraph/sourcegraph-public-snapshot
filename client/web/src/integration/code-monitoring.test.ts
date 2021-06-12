import assert from 'assert'

import expect from 'expect'

import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteID, siteGQLID } from './jscontext'
import { percySnapshotWithVariants } from './utils'

describe('Code monitoring', () => {
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
            SearchSuggestions: () => ({
                search: {
                    suggestions: [{ __typename: 'Repository', name: '^github\\.com/sourcegraph/sourcegraph$' }],
                },
            }),
            RepoGroups: () => ({
                repoGroups: [],
            }),
            AutoDefinedSearchContexts: () => ({
                autoDefinedSearchContexts: [],
            }),
            ViewerSettings: () => ({
                viewerSettings: {
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: JSON.stringify({
                                    experimentalFeatures: { showOnboardingTour: true, codeMonitoring: true },
                                }),
                            },
                        },
                        {
                            __typename: 'Site',
                            id: siteGQLID,
                            siteID,
                            latestSettings: {
                                id: 470,
                                contents: JSON.stringify({
                                    experimentalFeatures: { showOnboardingTour: true, codeMonitoring: true },
                                }),
                            },
                            settingsURL: '/site-admin/global-settings',
                            viewerCanAdminister: true,
                            allowSiteSettingsEdits: true,
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('Code monitoring form advances sequentially', () => {
        it('validates trigger query input', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/code-monitoring/new')
            await driver.page.waitForSelector('.test-name-input')

            await percySnapshotWithVariants(driver.page, 'Code monitoring - Form')

            await driver.page.type('.test-name-input', 'test monitor')

            await driver.page.waitForSelector('.test-action-button')
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-action-button')!.disabled
                ),
                true,
                'Expected action button to be disabled'
            )

            await driver.page.waitForSelector('.test-trigger-button')
            await driver.page.click('.test-trigger-button')

            await driver.page.waitForSelector('.test-trigger-input')
            await driver.page.type('.test-trigger-input', 'foobar')
            await driver.page.waitForSelector('.test-is-invalid')

            await driver.page.type('.test-trigger-input', ' type:diff')
            await driver.page.waitForSelector('.test-is-invalid')

            await driver.page.type('.test-trigger-input', ' repo:test')
            await driver.page.waitForSelector('.test-is-valid')
            await driver.page.waitForSelector('.test-preview-link')
            expect(
                await driver.page.evaluate(() => document.querySelectorAll('.test-preview-link').length)
            ).toBeGreaterThan(0)
        })

        it('disables the actions area until trigger is complete', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/code-monitoring/new')
            await driver.page.waitForSelector('.test-name-input')
            await driver.page.type('.test-name-input', 'test monitor')

            await driver.page.waitForSelector('.test-action-button')
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-action-button')!.disabled
                ),
                true,
                'Expected action button to be disabled'
            )

            await driver.page.waitForSelector('.test-trigger-button')
            await driver.page.click('.test-trigger-button')

            await driver.page.waitForSelector('.test-trigger-input')
            await driver.page.type('.test-trigger-input', 'foobar type:diff repo:test')
            await driver.page.waitForSelector('.test-is-valid')
            await driver.page.waitForSelector('.test-preview-link')
            await driver.page.waitForSelector('.test-submit-trigger')
            await driver.page.click('.test-submit-trigger')

            await driver.page.waitForSelector('.test-action-button')
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-action-button')!.disabled
                ),
                false,
                'Expected action button to be enabled'
            )

            await driver.page.click('.test-action-button')
            await driver.page.waitForSelector('.test-action-form')
        })

        it('disables submitting the code monitor area until trigger and action are complete', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/code-monitoring/new')
            await driver.page.waitForSelector('.test-name-input')
            await driver.page.type('.test-name-input', 'test monitor')

            await driver.page.waitForSelector('.test-submit-monitor')
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-submit-monitor')!.disabled
                ),
                true,
                'Expected submit monitor button to be disabled'
            )

            await driver.page.waitForSelector('.test-trigger-button')
            await driver.page.click('.test-trigger-button')

            await driver.page.waitForSelector('.test-trigger-input')
            await driver.page.type('.test-trigger-input', 'foobar type:diff repo:test')
            await driver.page.waitForSelector('.test-is-valid')
            await driver.page.waitForSelector('.test-preview-link')
            await driver.page.waitForSelector('.test-submit-trigger')
            await driver.page.click('.test-submit-trigger')

            await driver.page.waitForSelector('.test-action-button')
            await driver.page.click('.test-action-button')
            await driver.page.waitForSelector('.test-action-form')
            await driver.page.waitForSelector('.test-submit-action')
            await driver.page.click('.test-submit-action')

            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-submit-monitor')!.disabled
                ),
                false,
                'Expected submit monitor button to be enabled'
            )
        })
    })
})
