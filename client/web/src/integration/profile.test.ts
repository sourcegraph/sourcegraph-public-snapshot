import assert from 'assert'

import { subDays } from 'date-fns'
import { afterEach, beforeEach, describe, it } from 'mocha'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { UserSettingsAreaUserFields } from '../graphql-operations'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

const now = new Date()

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
    createdAt: subDays(now, 732).toISOString(),
    emails: [{ email: 'test@example.com', verified: true, isPrimary: true }],
    organizations: { nodes: [] },
    scimControlled: false,
    roles: {
        __typename: 'RoleConnection',
        nodes: [],
    },
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
        await driver.page.waitForSelector('[data-testid="user-profile-form-fields"]')
        await percySnapshotWithVariants(driver.page, 'User Profile Settings Page')
        await accessibilityAudit(driver.page)
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

    it('display user account security page', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            UserExternalAccountsWithAccountData: () => ({
                user: {
                    __typename: 'User',
                    externalAccounts: {
                        __typename: 'ExternalAccountConnection',
                        nodes: [],
                    },
                },
            }),
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
                    createdAt: '2020-03-02T11:52:15Z',
                    roles: {
                        __typename: 'RoleConnection',
                        nodes: [],
                    },
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
                    emails: [{ email: 'test@sourcegraph.test', verified: true, isPrimary: true }],
                    organizations: { nodes: [] },
                    permissionsInfo: null,
                    scimControlled: false,
                    roles: {
                        __typename: 'RoleConnection',
                        nodes: [],
                    },
                },
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/user/settings/security')
        await driver.page.waitForSelector('.user-settings-account-security-page')
        await percySnapshotWithVariants(driver.page, 'User Account Security Settings Page')
        await accessibilityAudit(driver.page)
    })
})
