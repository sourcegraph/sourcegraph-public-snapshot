import assert from 'assert'
import { commonWebGraphQlResults } from './graphQlResults'
import { Driver, createDriverForTest } from '../../../shared/src/testing/driver'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { saveScreenshotsUponFailures } from '../../../shared/src/testing/screenshotReporter'
import { SharedGraphQlOperations } from '../../../shared/src/graphql-operations'
import { WebGraphQlOperations } from '../graphql-operations'

describe('Site admin org UI', () => {
    const testOrg = {
        name: 'test-org-1',
        displayName: 'Test Org 1',
    }

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
    })
    saveScreenshotsUponFailures(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('allows to create new organizations', async () => {
        const graphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
            ...commonWebGraphQlResults,
            Organization: () => ({
                organization: null,
            }),
            Organizations: () => ({
                organizations: {
                    nodes: [],
                    totalCount: 0,
                },
            }),
            createOrganization: ({ name }) => ({
                createOrganization: {
                    id: 'TestOrg',
                    name,
                },
            }),
        }
        testContext.overrideGraphQL(graphQLResults)
        await driver.page.goto(driver.sourcegraphBaseUrl + '/site-admin/organizations')
        await driver.findElementWithText('Create organization', { action: 'click', wait: { timeout: 2000 } })
        await driver.replaceText({
            selector: '.test-new-org-name-input',
            newText: testOrg.name,
        })
        await driver.replaceText({
            selector: '.test-new-org-display-name-input',
            newText: testOrg.displayName,
        })

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.findElementWithText('Create organization', { action: 'click' })
        }, 'createOrganization')
        assert.deepStrictEqual(variables, {
            displayName: testOrg.displayName,
            name: testOrg.name,
        })

        testContext.overrideGraphQL({
            ...graphQLResults,
            Organization: () => ({
                organization: {
                    __typename: 'Org',
                    createdAt: '2020-08-07T00:00',
                    displayName: testOrg.displayName,
                    settingsURL: `/organizations/${testOrg.name}/settings`,
                    id: 'TestOrg',
                    name: testOrg.name,
                    url: `/organizations/${testOrg.name}`,
                    viewerCanAdminister: true,
                    viewerIsMember: false,
                    viewerPendingInvitation: null,
                },
            }),
        })

        await driver.waitUntilURL(`${driver.sourcegraphBaseUrl}/organizations/${testOrg.name}/settings`)
    })
})
