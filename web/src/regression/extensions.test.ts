/**
 * @jest-environment node
 */

import { TestResourceManager, ResourceDestructor } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestFixtures } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { editUserSettings } from './util/settings'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { ensureTestExternalService } from './util/api'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'

/**
 * Activates the extension with the given ID in user settings.
 * Optionally sets additional configuration values for the extension.
 */
async function activateAndConfigureExtension(
    {
        username,
        extensionID,
        extensionConfig = {},
    }: { username: string; extensionID: string; extensionConfig?: { [key: string]: any } },
    graphQLClient: GraphQLClient
): Promise<ResourceDestructor> {
    await editUserSettings(
        username,
        { keyPath: [{ property: 'extensions' }, { property: extensionID }], value: true },
        graphQLClient
    )
    for (const [property, value] of Object.entries(extensionConfig)) {
        await editUserSettings(username, { keyPath: [{ property }], value }, graphQLClient)
    }
    // No need to clean up: the test user will be destroyed.
    return () => Promise.resolve()
}

describe('Sourcegraph extensions regression test suite', () => {
    const testUsername = 'test-sg-extensions'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'keepBrowser',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'gitHubToken'
    )
    const externalService = {
        kind: ExternalServiceKind.GITHUB,
        uniqueDisplayName: '[TEST] Github (extensions.test.ts)',
    }
    const repos = ['theupdateframework/notary']

    let driver: Driver
    let graphQLClient: GraphQLClient
    let resourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, gqlClient: graphQLClient, resourceManager } = await getTestFixtures(config))
        resourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, graphQLClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        resourceManager.add(
            'External service',
            externalService.uniqueDisplayName,
            await ensureTestExternalService(graphQLClient, {
                ...externalService,
                config: {
                    url: 'https://github.com',
                    token: config.gitHubToken,
                    repos,
                    repositoryQuery: ['none'],
                },
                waitForRepos: ['github.com/theupdateframework/notary'],
            })
        )
    })

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    })

    describe('Codecov extension', () => {
        test(
            'it works',
            async () => {
                resourceManager.add(
                    'Extension',
                    'sourcegraph/codecov',
                    await activateAndConfigureExtension(
                        {
                            username: testUsername,
                            extensionID: 'sourcegraph/codecov',
                        },
                        graphQLClient
                    )
                )
                await driver.page.goto(
                    new URL(
                        'github.com/theupdateframework/notary@62258bc0beb3bdc41de1e927a57acaee06bebe4b/-/blob/cmd/notary/delegations.go#L60',
                        config.sourcegraphBaseUrl
                    ).href
                )
                // No lines should be decorated upon page load
                expect(await driver.page.$$('tr[style]')).toHaveLength(0)
                expect(await driver.page.$$('.line-decoration-attachment')).toHaveLength(0)

                // Wait for the "Coverage: X%" button to appear and click it
                await retry(() => driver.findElementWithText('Coverage: 80%'))
                await driver.clickElementWithText('Coverage: 80%')

                // Lines should get decorated, but without line/hit branch counts
                await retry(async () => expect(await driver.page.$$('tr[style]')).toHaveLength(264))
                expect(await driver.page.$$('.line-decoration-attachment')).toHaveLength(0)

                // Open the command palette and click "Show line/hit branch counts"
                await driver.page.click('.command-list-popover-button')
                await driver.clickElementWithText('Codecov: Show line hit/branch counts')

                // Line/hit branch counts should now show up
                await retry(async () => expect(await driver.page.$$('.line-decoration-attachment')).toHaveLength(264))

                // Check that the the "View commit report" button links to the correct location
                await driver.page.click('.command-list-popover-button')
                await driver.clickElementWithText('Codecov: View commit report')
                const codecovCommitURL =
                    'https://codecov.io/gh/theupdateframework/notary/commit/62258bc0beb3bdc41de1e927a57acaee06bebe4b'
                if (driver.page.url() === codecovCommitURL) {
                    return
                }
                await driver.page.waitForNavigation()
                expect(driver.page.url()).toEqual(codecovCommitURL)
            },
            30 * 1000
        )
    })
})
