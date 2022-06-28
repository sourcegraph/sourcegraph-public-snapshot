import assert from 'assert'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { UserSettingsAreaUserFields } from '../graphql-operations'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

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
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('updates display name', async () => {
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
        await percySnapshotWithVariants(driver.page, 'User Profile Settings Page')
        await accessibilityAudit(driver.page)
        await driver.page.waitForSelector('[data-testid="user-profile-form-fields"]')
        await driver.replaceText({
            selector: '[data-testid="test-UserProfileFormFields__displayName"]',
            newText: 'Test2',
            selectMethod: 'selectall',
        })

        const requestVariables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('#test-EditUserProfileForm__save')
        }, 'UpdateUser')

        assert.strictEqual(requestVariables.displayName, 'Test2')
    })
})

describe('User Different Settings Page', () => {
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
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('display user email setting page', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            UserAreaUserProfile: () => ({
                user: USER,
            }),
            UserSettingsAreaUserProfile: () => ({
                node: USER,
            }),
            UserEmails: () => ({
                node: {
                    __typename: 'User',
                    emails: [
                        {
                            email: 'test@example.com',
                            isPrimary: true,
                            verified: true,
                            verificationPending: false,
                            viewerCanManuallyVerify: false,
                        },
                    ],
                },
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/settings/emails')
        await driver.page.waitForSelector('[data-testid="user-settings-emails-page"]')
        await percySnapshotWithVariants(driver.page, 'User Email Settings Page')
        await accessibilityAudit(driver.page)
    })

    it('display user password setting page', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            UserAreaUserProfile: () => ({
                user: {
                    __typename: 'User',
                    id: testUserID,
                    username: 'test',
                    displayName: null,
                    url: '/users/test',
                    settingsURL: '/users/test/settings',
                    avatarURL: null,
                    viewerCanAdminister: true,
                    builtinAuth: true,
                    tags: [],
                },
            }),
            UserSettingsAreaUserProfile: () => ({
                node: {
                    __typename: 'User',
                    id: testUserID,
                    username: 'test',
                    displayName: null,
                    url: '/users/test',
                    settingsURL: '/users/test/settings',
                    avatarURL: null,
                    viewerCanAdminister: true,
                    viewerCanChangeUsername: true,
                    siteAdmin: true,
                    builtinAuth: true,
                    createdAt: '2020-03-02T11:52:15Z',
                    emails: [{ email: 'test@sourcegraph.test', verified: true }],
                    organizations: { nodes: [] },
                    permissionsInfo: null,
                    tags: [],
                },
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/user/settings/password')
        await driver.page.waitForSelector('.user-settings-password-page')
        await percySnapshotWithVariants(driver.page, 'User Password Settings Page')
        await accessibilityAudit(driver.page)
    })
})
