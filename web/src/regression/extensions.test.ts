/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser, clickAndWaitForNavigation } from './util/helpers'
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
) {
    await editUserSettings(
        username,
        { keyPath: [{ property: 'extensions' }, { property: extensionID }], value: true },
        graphQLClient
    )
    for (const [property, value] of Object.entries(extensionConfig)) {
        await editUserSettings(username, { keyPath: [{ property }], value }, graphQLClient)
    }
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
    const repos = ['theupdateframework/notary', 'GetStream/Winds', 'codecov/sourcegraph-codecov']

    let driver: Driver
    let graphQLClient: GraphQLClient
    let resourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, gqlClient: graphQLClient, resourceManager } = await getTestTools(config))
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
                waitForRepos: repos.map(s => `github.com/${s}`),
            })
        )
    }, 60 * 1000)

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    })

    test(
        'Codecov extension',
        async () => {
            await activateAndConfigureExtension(
                {
                    username: testUsername,
                    extensionID: 'sourcegraph/codecov',
                },
                graphQLClient
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
            await (await driver.findElementWithText('Coverage: 80%', { wait: true })).click()

            // Lines should get decorated, but without line/hit branch counts
            await retry(async () => expect(await driver.page.$$('tr[style]')).toHaveLength(264))
            expect(await driver.page.$$('.line-decoration-attachment')).toHaveLength(0)

            // Open the command palette and click "Show line/hit branch counts"
            await driver.page.click('.command-list-popover-button')
            await (await driver.findElementWithText('Codecov: Show line hit/branch counts')).click()

            // Line/hit branch counts should now show up
            await retry(async () => expect(await driver.page.$$('.line-decoration-attachment')).toHaveLength(264))

            // Check that the the "View commit report" button links to the correct location
            await driver.page.click('.command-list-popover-button')
            await clickAndWaitForNavigation(
                await driver.findElementWithText('Codecov: View commit report'),
                driver.page
            )
            expect(driver.page.url()).toEqual(
                'https://codecov.io/gh/theupdateframework/notary/commit/62258bc0beb3bdc41de1e927a57acaee06bebe4b'
            )
        },
        30 * 1000
    )

    test(
        'Datadog extension',
        async () => {
            await activateAndConfigureExtension(
                {
                    username: testUsername,
                    extensionID: 'sourcegraph/datadog-metrics',
                },
                graphQLClient
            )

            // Visit a file that contains statsd calls
            await driver.page.goto(
                new URL(
                    'github.com/GetStream/Winds@acd1f5661aae461d33a28c55b54015c20ff49ed7/-/blob/api/src/workers/podcast.js#L97',
                    config.sourcegraphBaseUrl
                ).href
            )
            // Verify datadog decorations appear
            await retry(async () =>
                expect(await driver.page.$$('[data-contents=" View metric (Datadog) » "]')).toHaveLength(9)
            )
        },
        10 * 1000
    )

    test(
        'Sentry extension',
        async () => {
            await activateAndConfigureExtension(
                {
                    username: testUsername,
                    extensionID: 'sourcegraph/sentry',
                    extensionConfig: {
                        'sentry.decorations.inline': true,
                        'sentry.organization': 'sourcegraph',
                        'sentry.projects': [
                            {
                                projectId: '1334031',
                            },
                        ],
                    },
                },
                graphQLClient
            )

            // Visit a file containing throw statements that should be matched by the Sentry extension
            await driver.page.goto(
                new URL(
                    'github.com/codecov/sourcegraph-codecov@f398e73d3e8f49834c09050d1898369762b2c51e/-/blob/src/uri.ts#L19-33',
                    config.sourcegraphBaseUrl
                ).href
            )

            // Verify Sentry decorations appear
            await retry(async () =>
                expect(await driver.page.$$('[data-contents=" View logs in Sentry » "]')).toHaveLength(1)
            )
        },
        10 * 1000
    )
})
