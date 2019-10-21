/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import * as jsonc from '@sqs/jsonc-parser'
import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { setUserSiteAdmin, getUser, getManagementConsoleState } from './util/api'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'

describe('Auth regression test suite', () => {
    const testUsername = 'test-auth'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'headless',
        'slowMo',
        'keepBrowser',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logBrowserConsole',
        'managementConsoleUrl',
        'gitHubClientID',
        'gitHubClientSecret',
        'gitHubUserAmyPassword'
    )

    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
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
        const user = await getUser(gqlClient, testUsername)
        if (!user) {
            throw new Error(`test user ${testUsername} does not exist`)
        }
        await setUserSiteAdmin(gqlClient, user.id, true)
    })

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    }, 10 * 1000)

    test(
        'Sign in via GitHub.com',
        async () => {
            const testAuthProvider = {
                type: 'github',
                displayName: '[TEST] GitHub.com',
                clientID: config.gitHubClientID,
                clientSecret: config.gitHubClientSecret,
                allowSignup: true,
            }

            const { plaintextPassword: managementConsolePassword } = await getManagementConsoleState(gqlClient)
            if (!managementConsolePassword) {
                throw new Error('empty management console password')
            }
            const authHeaders = {
                Authorization: `Basic ${new Buffer(`:${managementConsolePassword}`).toString('base64')}`,
            }
            const gotoManagementConsole = async () => {
                try {
                    await driver.page.goto(config.managementConsoleUrl)
                } catch (err) {
                    if (!err.message.includes('net::ERR_CERT_AUTHORITY_INVALID')) {
                        throw err
                    }
                    await driver.page.waitForSelector('#details-button')
                    await driver.page.click('#details-button')
                    await driver.clickElementWithText('Proceed to')
                }
                await driver.page.waitForSelector('.monaco-editor')
            }

            resourceManager.add(
                'Authentication provider',
                '[TEST] GitHub',
                await (async () => {
                    await driver.page.setExtraHTTPHeaders(authHeaders)
                    await gotoManagementConsole()

                    const oldCriticalConfig = await driver.page.evaluate(async managementConsoleUrl => {
                        const res = await fetch(managementConsoleUrl + '/api/get', { method: 'GET' })
                        return (await res.json()).Contents
                    }, config.managementConsoleUrl)
                    const parsedOldConfig = jsonc.parse(oldCriticalConfig)
                    const authProviders = parsedOldConfig['auth.providers'] as any[]
                    if (
                        authProviders.filter(
                            p => p.type === testAuthProvider.type && p.displayName === testAuthProvider.displayName
                        ).length > 0
                    ) {
                        return () => Promise.resolve()
                    }

                    const newCriticalConfig = jsonc.applyEdits(
                        oldCriticalConfig,
                        jsoncEdit.setProperty(oldCriticalConfig, ['auth.providers', -1], testAuthProvider, {
                            eol: '\n',
                            insertSpaces: true,
                            tabSize: 2,
                        })
                    )
                    await driver.replaceText({
                        selector: '.monaco-editor',
                        newText: newCriticalConfig,
                        selectMethod: 'keyboard',
                        enterTextMethod: 'paste',
                    })
                    await driver.clickElementWithText('Save changes')
                    await retry(() => driver.findElementWithText('Saved!'), { retries: 3, maxRetryTime: 500 })
                    await driver.page.setExtraHTTPHeaders({})

                    return async () => {
                        await driver.page.setExtraHTTPHeaders(authHeaders)
                        await gotoManagementConsole()

                        await driver.replaceText({
                            selector: '.monaco-editor',
                            newText: oldCriticalConfig,
                            selectMethod: 'keyboard',
                            enterTextMethod: 'paste',
                        })

                        await driver.clickElementWithText('Save changes')
                        await retry(() => driver.findElementWithText('Saved!'), { retries: 3, maxRetryTime: 500 })

                        await driver.page.setExtraHTTPHeaders({})
                    }
                })()
            )

            await driver.page.goto(config.sourcegraphBaseUrl + '/-/sign-out')

            await driver.page.goto(config.sourcegraphBaseUrl)
            await driver.page.reload()
            await driver.page.waitForNavigation()
            await driver.clickElementWithText('Sign in with ' + testAuthProvider.displayName, { tagName: 'a' })
            await driver.page.waitForSelector('#login_field')
            await driver.replaceText({
                selector: '#login_field',
                newText: 'sg-e2e-regression-test-amy',
                selectMethod: 'keyboard',
                enterTextMethod: 'paste',
            })
            await driver.replaceText({
                selector: '#password',
                newText: config.gitHubUserAmyPassword,
                selectMethod: 'keyboard',
                enterTextMethod: 'paste',
            })
            await driver.page.keyboard.press('Enter')
            await driver.page.waitForNavigation()
            if (driver.page.url() !== config.sourcegraphBaseUrl + '/search') {
                throw new Error('unsuccessful login')
            }
        },
        20 * 1000
    )
})
