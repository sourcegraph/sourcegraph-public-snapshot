import { Key } from 'ts-key-enum'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { Driver } from '../../../shared/src/e2e/driver'
import { GraphQLClient, createGraphQLClient } from './util/GraphQLClient'
import { TestResourceManager } from './util/TestResourceManager'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { getUser, setUserSiteAdmin, fetchAllOrganizations, deleteOrganization, getViewerSettings } from './util/api'
import { PlatformContext } from '../../../shared/src/platform/context'
import * as GQL from '../../../shared/src/graphql/schema'
import { parse } from '@sqs/jsonc-parser'
import { parseJSONCOrError } from '../../../shared/src/util/jsonc'
import { Settings, QuickLink } from '../schema/settings.schema'
import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { QuickLinks } from '../search/QuickLinks'

/**
 * @jest-environment node
 */

async function deleteOrganizationByName(
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    name: string
): Promise<void> {
    const orgs = await fetchAllOrganizations({ requestGraphQL }, { first: 1000, query: '' }).toPromise()
    const matches = orgs.nodes.filter((org: GQL.IOrg) => org.name === name)
    await Promise.all(matches.map(org => deleteOrganization({ requestGraphQL }, org.id).toPromise()))
}

/**
 * Test plan:
 * Test 1: org UI
 * Test 2: org settings (user settings cascade) (API-driven)
 * Test 3: auth.userOrgMap (API-driven)
 */

describe('Organizations regression test suite', () => {
    describe('Organizations GUI', () => {
        const testUsername = 'test-org'
        const testOrg = {
            name: 'test-org-1',
            displayName: 'Test Org 1',
        }
        const config = getConfig(
            'sudoToken',
            'sudoUsername',
            'headless',
            'slowMo',
            'keepBrowser',
            'noCleanup',
            'sourcegraphBaseUrl',
            'testUserPassword',
            'logBrowserConsole',
            'managementConsoleUrl'
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
            await deleteOrganizationByName(gqlClient, testOrg.name)
        })
        afterAll(async () => {
            if (!config.noCleanup) {
                await resourceManager.destroyAll()
            }
            if (driver) {
                await driver.close()
            }
        })

        test('Create an organization and test settings cascade', async () => {
            const quicklink = { name: 'Test Org 1 Quicklink', url: 'http://test-org-1-link.local' }
            const userGQLClient = createGraphQLClient({
                baseUrl: config.sourcegraphBaseUrl,
                token: config.sudoToken,
                sudoUsername: testUsername,
            })
            const getQuickLinks = async (): Promise<QuickLink[] | undefined> => {
                const rawSettings = await getViewerSettings(userGQLClient)
                const settingsOrError = parseJSONCOrError<Settings>(rawSettings.final)
                if (isErrorLike(settingsOrError)) {
                    throw settingsOrError
                }
                return settingsOrError.quicklinks
            }

            await driver.page.goto(config.sourcegraphBaseUrl + '/site-admin/organizations')
            await (await driver.findElementWithText('Create organization', { wait: { timeout: 2000 } })).click()
            await driver.replaceText({
                selector: '.e2e-new-org-name-input',
                newText: testOrg.name,
            })
            await driver.replaceText({
                selector: '.e2e-new-org-display-name-input',
                newText: testOrg.displayName,
            })
            await (await driver.findElementWithText('Create organization')).click()
            resourceManager.add('Organization', testOrg.name, () => deleteOrganizationByName(gqlClient, testOrg.name))
            await driver.page.waitForSelector('.monaco-editor')
            await driver.replaceText({
                selector: '.monaco-editor',
                newText: `{"quicklinks": [${JSON.stringify(quicklink)}]}`,
                selectMethod: 'keyboard',
                enterTextMethod: 'paste',
            })

            await driver.page.keyboard.down(Key.Control)
            await driver.page.keyboard.down(Key.Shift)
            await driver.page.keyboard.press('i')
            await driver.page.keyboard.up(Key.Shift)
            await driver.page.keyboard.up(Key.Control)
            await (await driver.findElementWithText('Save changes')).click()
            await new Promise(resolve => setTimeout(resolve, 500)) // Wait for save
            await driver.findElementWithText('Save changes')

            {
                const quicklinks = await getQuickLinks()
                if (!quicklinks) {
                    throw new Error('No quicklinks found')
                }
                if (!quicklinks.some(l => l.name === quicklink.name && l.url === quicklink.url)) {
                    throw new Error(
                        `Did not find quicklink found ${JSON.stringify(quicklink)} in quicklinks: ${JSON.stringify(
                            quicklinks
                        )}`
                    )
                }
            }

            // Remove user from org
            await (await driver.findElementWithText('Members')).click()
            // eslint-disable-next-line @typescript-eslint/no-misused-promises
            driver.page.once('dialog', async dialog => {
                await dialog.accept()
            })
            await (await driver.findElementWithText('Leave organization', { wait: { timeout: 1000 } })).click()

            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            await driver.page.waitForFunction(() => !(window as any).find('Leave organization'))

            {
                const quicklinks = await getQuickLinks()
                if (quicklinks && quicklinks.some(l => l.name === quicklink.name && l.url === quicklink.url)) {
                    throw new Error(
                        `Found quicklink ${JSON.stringify(quicklink)} in quicklinks: ${JSON.stringify(quicklinks)}`
                    )
                }
            }
        })
    })
})
