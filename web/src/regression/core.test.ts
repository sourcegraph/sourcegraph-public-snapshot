/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestFixtures } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { deleteUser } from './util/api'

describe('Core functionality regression test suite', () => {
    const testUsername = 'test-core'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser'
    )

    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestFixtures(config))
        await resourceManager.create({
            type: 'User',
            name: testUsername,
            create: async () => {
                await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                    username: testUsername,
                    deleteIfExists: true,
                    ...config,
                })
                return () => deleteUser(gqlClient, testUsername, false)
            },
        })
    })

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    })

    test('User settings are saved and applied', async () => {
        const getSettings = async () => {
            await driver.page.waitForSelector('.view-line')
            return await driver.page.evaluate(() => {
                const editor = document.querySelector('.monaco-editor') as HTMLElement
                return editor ? editor.innerText : null
            })
        }

        await driver.page.goto(config.sourcegraphBaseUrl + `/users/${testUsername}/settings`)
        const previousSettings = await getSettings()
        if (!previousSettings) {
            throw new Error('Previous settings were null')
        }
        const newSettings = '{\xa0/*\xa0These\xa0are\xa0new\xa0settings\xa0*/}'
        await driver.replaceText({
            selector: '.monaco-editor',
            newText: newSettings,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.page.reload()

        const currentSettings = await getSettings()
        if (currentSettings !== previousSettings) {
            throw new Error(
                `Settings ${JSON.stringify(currentSettings)} did not match (old) saved settings ${JSON.stringify(
                    previousSettings
                )}`
            )
        }

        await driver.replaceText({
            selector: '.monaco-editor',
            newText: newSettings,
            selectMethod: 'keyboard',
            enterTextMethod: 'type',
        })
        await driver.clickElementWithText('Save changes')
        await driver.page.waitForFunction(
            () => document.evaluate("//*[text() = ' Saving...']", document).iterateNext() === null
        )
        await driver.page.reload()

        const currentSettings2 = await getSettings()
        if (JSON.stringify(currentSettings2) !== JSON.stringify(newSettings)) {
            throw new Error(
                `Settings ${JSON.stringify(currentSettings2)} did not match (new) saved settings ${JSON.stringify(
                    newSettings
                )}`
            )
        }

        // Restore old settings
        await driver.replaceText({
            selector: '.monaco-editor',
            newText: previousSettings,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.clickElementWithText('Save changes')
        await driver.page.waitForFunction(
            () => document.evaluate("//*[text() = ' Saving...']", document).iterateNext() === null
        )
        const previousSettings2 = await getSettings()
        await driver.page.reload()

        const currentSettings3 = await getSettings()
        if (currentSettings3 !== previousSettings2) {
            throw new Error(
                `Settings ${JSON.stringify(currentSettings3)} did not match (old) saved settings ${JSON.stringify(
                    previousSettings2
                )}`
            )
        }
    })

    test('User profile page', async () => {
        // TODO(@sourcegraph/web)
    })
    test('User emails page', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Access tokens work', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Organizations (admin user)', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Organizations (non-admin user)', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Explore page', async () => {
        // TODO(@sourcegraph/web)
    })
    test('Quicklinks', async () => {
        // TODO(@sourcegraph/web)
    })
})
