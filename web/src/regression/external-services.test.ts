/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { setUserSiteAdmin, getUser, ensureNoTestExternalServices } from './util/api'
import * as GQL from '../../../shared/src/graphql/schema'

describe('External services regression test suite', () => {
    const testUsername = 'test-extsvc'
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

    test('GitHub.com external service', async () => {
        const externalServiceName = '[TEST] Regression test: GitHub.com'
        await ensureNoTestExternalServices(gqlClient, {
            kind: GQL.ExternalServiceKind.GITHUB,
            uniqueDisplayName: externalServiceName,
            deleteIfExist: true,
        })

        resourceManager.add(
            'External service',
            externalServiceName,
            await (async () => {
                await driver.page.goto(config.sourcegraphBaseUrl + '/site-admin/external-services')
                await (await driver.findElementWithText('Add external service', { wait: { timeout: 500 } })).click()
                await (await driver.findElementWithText('Add GitHub.com repositories.', {
                    wait: { timeout: 500 },
                })).click()
                const repoSlugs = ['gorilla/mux']
                const githubConfig = `{
                    "url": "https://github.com",
                    "token": ${JSON.stringify(config.gitHubToken)},
                    "repos": ${JSON.stringify(repoSlugs)},
                    "repositoryQuery": ["none"],
                }`
                await driver.replaceText({
                    selector: '#e2e-external-service-form-display-name',
                    newText: externalServiceName,
                    selectMethod: 'selectall',
                    enterTextMethod: 'paste',
                })
                await driver.replaceText({
                    selector: '.monaco-editor',
                    newText: githubConfig,
                    selectMethod: 'keyboard',
                    enterTextMethod: 'paste',
                })
                await (await driver.findElementWithText('Add external service', {
                    tagName: 'button',
                    wait: { timeout: 500 },
                })).click()
                return () =>
                    ensureNoTestExternalServices(gqlClient, {
                        kind: GQL.ExternalServiceKind.GITHUB,
                        uniqueDisplayName: externalServiceName,
                        deleteIfExist: true,
                    })
            })()
        )
    })
})
