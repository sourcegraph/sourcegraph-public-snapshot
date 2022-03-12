import assert from 'assert'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailedWithJest } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { UserSettingsAreaUserFields } from '../graphql-operations'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'

describe('User profile page', () => {
    let driver: Driver
    beforeAll(async () => {
        driver = await createDriverForTest()
    })
    afterAll(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async () => {
        testContext = await createWebIntegrationTestContext({
            driver,
            directory: __dirname,
        })
    })
    afterEachSaveScreenshotIfFailedWithJest(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('updates display name', async () => {
        const USER: UserSettingsAreaUserFields = {
            __typename: 'User',
            id: 'VXNlcjoxODkyNw==',
            username: 'test',
            displayName: 'Test',
            url: '/users/test',
            settingsURL: '/users/test/settings',
            avatarURL: '',
            viewerCanAdminister: true,
            viewerCanChangeUsername: true,
            siteAdmin: true,
            builtinAuth: true,
            createdAt: '2020-04-10T21:11:42Z',
            emails: [{ email: 'test@example.com', verified: true }],
            organizations: { nodes: [] },
            tags: [],
        }
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            UserAreaUserProfile: () => ({
                user: USER,
            }),
            UserSettingsAreaUserProfile: () => ({
                node: USER,
            }),
            UpdateUser: () => ({ updateUser: { ...USER, displayName: 'Test2' } }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/settings/profile')
        await driver.page.waitForSelector('[data-testid="user-profile-form-fields"]')
        await driver.replaceText({
            selector: '.test-UserProfileFormFields__displayName',
            newText: 'Test2',
            selectMethod: 'selectall',
        })

        const requestVariables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('#test-EditUserProfileForm__save')
        }, 'UpdateUser')

        assert.strictEqual(requestVariables.displayName, 'Test2')
    })
})
