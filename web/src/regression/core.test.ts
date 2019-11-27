/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient, createGraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser, getGlobalSettings } from './util/helpers'
import { setUserEmailVerified } from './util/api'
import { ScreenshotVerifier } from './util/ScreenshotVerifier'
import { gql, dataOrThrowErrors } from '../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { applyEdits, parse } from '@sqs/jsonc-parser'
import { overwriteSettings } from '../../../shared/src/settings/edit'

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
    let screenshots: ScreenshotVerifier
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
        resourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        screenshots = new ScreenshotVerifier(driver)
    })
    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
        if (screenshots.screenshots.length > 0) {
            console.log(screenshots.verificationInstructions())
        }
    })

    let alwaysCleanupManager: TestResourceManager
    beforeEach(() => {
        alwaysCleanupManager = new TestResourceManager()
    })
    afterEach(async () => {
        await alwaysCleanupManager.destroyAll()
    })

    test('2.2.1 User settings are saved and applied', async () => {
        const getSettings = async () => {
            await driver.page.waitForSelector('.view-line')
            return driver.page.evaluate(() => {
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
        await (await driver.findElementWithText('Save changes')).click()
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
        await (await driver.findElementWithText('Save changes')).click()
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

    test('2.2.2 User profile page', async () => {
        const aviURL =
            'https://media2.giphy.com/media/26tPplGWjN0xLybiU/giphy.gif?cid=790b761127d52fa005ed23fdcb09d11a074671ac90146787&rid=giphy.gif'
        const displayName = 'Test Display Name'

        await driver.page.goto(driver.sourcegraphBaseUrl + `/users/${testUsername}/settings/profile`)
        await driver.replaceText({
            selector: '.e2e-user-settings-profile-page__display-name',
            newText: displayName,
        })
        await driver.replaceText({
            selector: '.e2e-user-settings-profile-page__avatar_url',
            newText: aviURL,
            enterTextMethod: 'paste',
        })
        await (await driver.findElementWithText('Update profile')).click()
        await driver.page.reload()
        await driver.page.waitForFunction(
            displayName => {
                const el = document.querySelector('.e2e-user-area-header__display-name')
                return el?.textContent && el.textContent.trim() === displayName
            },
            undefined,
            displayName
        )

        await screenshots.verifySelector(
            'navbar-toggle-is-bart-simpson.png',
            'Navbar toggle avatar is Bart Simpson',
            '.e2e-user-nav-item-toggle'
        )
    })

    test('2.2.3. User emails page', async () => {
        const testEmail = 'sg-test-account@protonmail.com'
        await driver.page.goto(driver.sourcegraphBaseUrl + `/users/${testUsername}/settings/emails`)
        await driver.replaceText({ selector: '.e2e-user-email-add-input', newText: 'sg-test-account@protonmail.com' })
        await (await driver.findElementWithText('Add')).click()
        await driver.findElementWithText(testEmail, { wait: true })
        try {
            await driver.findElementWithText('Verification pending')
        } catch (err) {
            await driver.findElementWithText('Not verified')
        }
        await setUserEmailVerified(gqlClient, testUsername, testEmail, true)
        await driver.page.reload()
        await driver.findElementWithText('Verified', { wait: true })
    })

    test('2.2.4 Access tokens work and invalid access tokens return "401 Unauthorized"', async () => {
        await driver.page.goto(config.sourcegraphBaseUrl + `/users/${testUsername}/settings/tokens`)
        await (await driver.findElementWithText('Generate new token', { wait: { timeout: 5000 } })).click()
        await driver.findElementWithText('New access token', { wait: { timeout: 1000 } })
        await driver.replaceText({
            selector: '.e2e-create-access-token-description',
            newText: 'test-regression',
        })
        await (await driver.findElementWithText('Generate token', { wait: { timeout: 1000 } })).click()
        await driver.findElementWithText("Copy the new access token now. You won't be able to see it again.", {
            wait: { timeout: 1000 },
        })
        await (await driver.findElementWithText('Copy')).click()
        const token = await driver.page.evaluate(() => {
            const tokenEl = document.querySelector('.e2e-access-token')
            if (!tokenEl) {
                return null
            }
            const inputEl = tokenEl.querySelector('input')
            if (!inputEl) {
                return null
            }
            return inputEl.value
        })
        if (!token) {
            throw new Error('Could not obtain access token')
        }
        const gqlClientWithToken = createGraphQLClient({
            baseUrl: config.sourcegraphBaseUrl,
            token,
        })
        await new Promise(resolve => setTimeout(resolve, 2000))
        const currentUsernameQuery = gql`
            query {
                currentUser {
                    username
                }
            }
        `
        const response = await gqlClientWithToken
            .queryGraphQL(currentUsernameQuery)
            .pipe(map(dataOrThrowErrors))
            .toPromise()
        expect(response).toEqual({ currentUser: { username: testUsername } })

        const gqlClientWithInvalidToken = createGraphQLClient({
            baseUrl: config.sourcegraphBaseUrl,
            token: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
        })

        await expect(
            gqlClientWithInvalidToken
                .queryGraphQL(currentUsernameQuery)
                .pipe(map(dataOrThrowErrors))
                .toPromise()
        ).rejects.toThrowError('401 Unauthorized')
    })

    test('2.5 Quicklinks: add a quicklink, test that it appears on the front page and works.', async () => {
        const quicklinkInfo = {
            name: 'Quicklink',
            url: config.sourcegraphBaseUrl + '/api/console',
            description: 'This is a quicklink',
        }

        const { subjectID, settingsID, contents: oldContents } = await getGlobalSettings(gqlClient)
        if (parse(oldContents).quicklinks) {
            throw new Error('Global setting quicklinks already exists, aborting test')
        }
        const newContents = applyEdits(
            oldContents,
            setProperty(oldContents, ['quicklinks'], [quicklinkInfo], {
                eol: '\n',
                insertSpaces: true,
                tabSize: 2,
            })
        )
        await overwriteSettings(gqlClient, subjectID, settingsID, newContents)
        alwaysCleanupManager.add('Global setting', 'quicklinks', async () => {
            const { subjectID: currentSubjectID, settingsID: currentSettingsID } = await getGlobalSettings(gqlClient)
            await overwriteSettings(gqlClient, currentSubjectID, currentSettingsID, oldContents)
        })

        await driver.page.goto(config.sourcegraphBaseUrl + '/search')
        await (
            await driver.findElementWithText(quicklinkInfo.name, {
                selector: 'a',
                wait: { timeout: 1000 },
            })
        ).hover()
        await driver.findElementWithText(quicklinkInfo.description, {
            wait: { timeout: 1000 },
        })
        await (
            await driver.findElementWithText(quicklinkInfo.name, {
                selector: 'a',
                wait: { timeout: 1000 },
            })
        ).click()
        await driver.page.waitForNavigation()
        expect(driver.page.url()).toEqual(quicklinkInfo.url)
    })

    test('2.4 Explore page', async () => {
        // TODO(@sourcegraph/web)
    })
})
