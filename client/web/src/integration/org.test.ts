import assert from 'assert'

import { subtypeOf } from '@sourcegraph/common'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { emptyResponse } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'
import { retry } from '@sourcegraph/shared/src/testing/utils'

import { WebGraphQlOperations, OrganizationResult } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

describe('Organizations', () => {
    const testOrg = subtypeOf<OrganizationResult['organization']>()({
        name: 'test-org-1',
        displayName: 'Test Org 1',
        __typename: 'Org',
        createdAt: '2020-08-07T00:00',
        settingsURL: '/organizations/test-org-1/settings',
        id: 'TestOrg',
        url: '/organizations/test-org-1',
        viewerCanAdminister: true,
        viewerIsMember: false,
        viewerPendingInvitation: null,
        viewerNeedsCodeHostUpdate: false,
    })

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

    const settingsID = 12345

    describe('Site admin organizations page', () => {
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
                SettingsCascade: () => ({
                    settingsSubject: {
                        settingsCascade: {
                            subjects: [
                                {
                                    latestSettings: {
                                        id: settingsID,
                                        contents: JSON.stringify({}),
                                    },
                                },
                            ],
                        },
                    },
                }),
                CreateOrganization: ({ name }) => ({
                    createOrganization: {
                        id: 'TestOrg',
                        name,
                        settingsURL: '/organizations/test-org-1/settings',
                    },
                }),
            }
            testContext.overrideGraphQL(graphQLResults)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/site-admin/organizations')

            await driver.page.waitForSelector('.test-create-org-button')

            await percySnapshotWithVariants(driver.page, 'Site admin org page')
            await accessibilityAudit(driver.page)

            await driver.page.click('.test-create-org-button')

            await driver.replaceText({
                selector: '[data-testid="test-new-org-name-input"]',
                newText: testOrg.name,
            })
            await driver.replaceText({
                selector: '[data-testid="test-new-org-display-name-input"]',
                newText: testOrg.displayName,
            })

            const variables = await testContext.waitForGraphQLRequest(async () => {
                await driver.page.click('.test-create-org-submit-button')
            }, 'CreateOrganization')
            assert.deepStrictEqual(variables, {
                displayName: testOrg.displayName,
                name: testOrg.name,
            })

            testContext.overrideGraphQL({
                ...graphQLResults,
                Organization: () => ({ organization: testOrg }),
            })

            await driver.waitUntilURL(`${driver.sourcegraphBaseUrl}/organizations/${testOrg.name}/settings`)
        })
    })

    describe('Organization area', () => {
        describe('Settings tab', () => {
            it('allows to change organization-wide settings', async () => {
                testContext.overrideGraphQL({
                    ...commonWebGraphQlResults,
                    Organization: () => ({
                        organization: testOrg,
                    }),
                    SettingsCascade: () => ({
                        settingsSubject: {
                            settingsCascade: {
                                subjects: [
                                    {
                                        latestSettings: {
                                            id: settingsID,
                                            contents: JSON.stringify({}),
                                        },
                                    },
                                ],
                            },
                        },
                    }),
                    OverwriteSettings: () => ({
                        settingsMutation: {
                            overwriteSettings: {
                                empty: emptyResponse,
                            },
                        },
                    }),
                    GetStartedInfo: () => ({
                        membersSummary: { membersCount: 1, invitesCount: 1, __typename: 'OrgMembersSummary' },
                        repoCount: { total: { totalCount: 1, __typename: 'RepositoryConnection' }, __typename: 'Org' },
                        extServices: { totalCount: 1, __typename: 'ExternalServiceConnection' },
                    }),
                })
                await driver.page.goto(driver.sourcegraphBaseUrl + '/organizations/sourcegraph/settings')
                const updatedSettings = '// updated'
                await driver.page.waitForSelector('.test-settings-file .monaco-editor')
                await driver.replaceText({
                    selector: '.test-settings-file .monaco-editor',
                    newText: updatedSettings,
                    selectMethod: 'keyboard',
                    enterTextMethod: 'paste',
                })

                const variables = await testContext.waitForGraphQLRequest(async () => {
                    await driver.page.click('.test-save-toolbar-save')
                }, 'OverwriteSettings')

                assert.deepStrictEqual(variables, {
                    subject: 'TestOrg',
                    lastID: settingsID,
                    contents: updatedSettings,
                })

                await percySnapshotWithVariants(driver.page, 'Organization settings page')
                await accessibilityAudit(driver.page)
            })
        })
        describe('Members tab', () => {
            it('allows to remove a member', async () => {
                const testMember = {
                    id: 'TestMember',
                    displayName: 'Test member',
                    username: 'testmember',
                    avatarURL: null,
                }
                const testMember2 = {
                    id: 'TestMember2',
                    displayName: 'Test member 2',
                    username: 'testmember2',
                    avatarURL: null,
                }
                const graphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
                    ...commonWebGraphQlResults,
                    Organization: () => ({
                        organization: testOrg,
                    }),
                    OrganizationMembers: () => ({
                        node: {
                            viewerCanAdminister: true,
                            members: {
                                totalCount: 2,
                                nodes: [testMember, testMember2],
                            },
                        },
                    }),
                    RemoveUserFromOrganization: () => ({
                        removeUserFromOrganization: emptyResponse,
                    }),
                    GetStartedInfo: () => ({
                        membersSummary: { membersCount: 1, invitesCount: 1, __typename: 'OrgMembersSummary' },
                        repoCount: { total: { totalCount: 1, __typename: 'RepositoryConnection' }, __typename: 'Org' },
                        extServices: { totalCount: 1, __typename: 'ExternalServiceConnection' },
                    }),
                }
                testContext.overrideGraphQL(graphQlResults)

                await driver.page.goto(driver.sourcegraphBaseUrl + '/organizations/sourcegraph/settings/members')

                await driver.page.waitForSelector('.test-remove-org-member')

                assert.strictEqual(
                    await driver.page.evaluate(
                        () => document.querySelectorAll('.test-org-members [data-test-username]').length
                    ),
                    2,
                    'Expected members list to show 2 members.'
                )

                await percySnapshotWithVariants(driver.page, 'Organization members list')
                await accessibilityAudit(driver.page)

                // Override for the fetch post-removal
                testContext.overrideGraphQL({
                    ...graphQlResults,
                    OrganizationMembers: () => ({
                        node: {
                            viewerCanAdminister: true,
                            members: {
                                totalCount: 1,
                                nodes: [testMember2],
                            },
                        },
                    }),
                })

                const variables = await testContext.waitForGraphQLRequest(async () => {
                    await Promise.all([
                        driver.acceptNextDialog(),
                        driver.page.click('[data-test-username="testmember"] .test-remove-org-member'),
                    ])
                }, 'RemoveUserFromOrganization')

                assert.deepStrictEqual(variables, {
                    user: testMember.id,
                    organization: testOrg.id,
                })

                await retry(async () => {
                    assert.strictEqual(
                        await driver.page.evaluate(
                            () => document.querySelectorAll('.test-org-members [data-test-username]').length
                        ),
                        1,
                        'Expected members list to show 1 member.'
                    )
                })

                assert(
                    await driver.page.evaluate(
                        () => !document.querySelector('.test-org-members [data-test-username="testmember"]')
                    ),
                    'Expected user "testmember" to have disappeared from the member list'
                )
            })
        })
    })
})
