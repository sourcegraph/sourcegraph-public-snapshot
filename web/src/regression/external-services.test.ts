/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestFixtures } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { deleteUser, setUserSiteAdmin, getUser, ensureNoTestExternalServices } from './util/api'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
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

        await driver.page.goto(config.sourcegraphBaseUrl + '/site-admin/external-services')
        await retry(() => driver.clickElementWithText('Add external service'), { retries: 3, maxRetryTime: 500 })
        await retry(() => driver.clickElementWithText('Add GitHub.com repositories.'), {
            retries: 3,
            maxRetryTime: 500,
        })
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
        await retry(() => driver.clickElementWithText('Add external service', { tagName: 'button' }), {
            retries: 3,
            maxRetryTime: 500,
        })
        await resourceManager.create({
            type: 'External service',
            name: externalServiceName,
            create: () =>
                // already created above
                Promise.resolve(() =>
                    ensureNoTestExternalServices(gqlClient, {
                        kind: GQL.ExternalServiceKind.GITHUB,
                        uniqueDisplayName: externalServiceName,
                        deleteIfExist: true,
                    })
                ),
        })
    })
})
