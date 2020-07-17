import assert from 'assert'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import { commonWebGraphQlResults } from './graphQlResults'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { saveScreenshotsUponFailures } from '../../../shared/src/testing/screenshotReporter'

describe('User profile page', () => {
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

    it('updates display name', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            User: () => ({
                user: {
                    __typename: 'User',
                    id: 'VXNlcjoxODkyNw==',
                    username: 'test',
                    displayName: 'Test',
                    url: '/users/test',
                    settingsURL: '/users/test/settings',
                    avatarURL: '',
                    viewerCanAdminister: true,
                    siteAdmin: true,
                    builtinAuth: true,
                    createdAt: '2020-04-10T21:11:42Z',
                    emails: [{ email: 'test@example.com', verified: true }],
                    organizations: { nodes: [] },
                    permissionsInfo: null,
                },
            }),
            UserForProfilePage: () => ({
                node: {
                    id: 'VXNlcjoxODkyNw==',
                    username: 'test',
                    displayName: 'Test',
                    avatarURL: '',
                    viewerCanChangeUsername: true,
                },
            }),
            updateUser: () => ({ updateUser: { alwaysNil: null } }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/settings/profile')
        await driver.page.waitForSelector('.user-settings-profile-page')
        await driver.replaceText({
            selector: '.test-user-settings-profile-page__display-name',
            newText: 'Test2',
            selectMethod: 'selectall',
        })

        const requestVariables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('.test-user-settings-profile-page-update-profile')
        }, 'updateUser')

        assert.strictEqual(requestVariables.displayName, 'Test2')
    })
})
