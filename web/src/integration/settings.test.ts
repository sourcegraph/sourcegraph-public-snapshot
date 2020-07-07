import { createDriverForTest } from '../../../shared/src/testing/driver'
import MockDate from 'mockdate'
import { getConfig } from '../../../shared/src/testing/config'
import assert from 'assert'
import {
    ExternalServiceKind,
    IQuery,
    IOrgConnection,
    IUserEmail,
    IOrg,
    IMutation,
    StatusMessage,
    IAlert,
} from '../../../shared/src/graphql/schema'
import { describeIntegration } from './helpers'
import { retry } from '../../../shared/src/testing/utils'

describeIntegration('Settings', ({ initGeneration, describe }) => {
    initGeneration(async () => {
        // Reset date mocking
        MockDate.reset()
        const { gitHubToken, sourcegraphBaseUrl, headless, slowMo, testUserPassword } = getConfig(
            'gitHubToken',
            'sourcegraphBaseUrl',
            'headless',
            'slowMo',
            'testUserPassword'
        )

        // Start browser
        const driver = await createDriverForTest({
            sourcegraphBaseUrl,
            logBrowserConsole: true,
            headless,
            slowMo,
        })
        const repoSlugs = ['gorilla/mux', 'sourcegraph/jsonrpc2']
        await driver.ensureLoggedIn({ username: 'test', password: testUserPassword, email: 'test@test.com' })
        await driver.resetUserSettings()
        await driver.ensureHasExternalService({
            kind: ExternalServiceKind.GITHUB,
            displayName: 'e2e-test-github',
            config: JSON.stringify({
                url: 'https://github.com',
                token: gitHubToken,
                repos: repoSlugs,
            }),
            ensureRepos: repoSlugs.map(slug => `github.com/${slug}`),
        })
        return { driver, sourcegraphBaseUrl }
    })

    describe('User settings page', ({ it }) => {
        it('updates user settings', async ({ driver, sourcegraphBaseUrl, overrideGraphQL, waitForGraphQLRequest }) => {
            const testUserID = 'TestUserID'
            const settingsID = 123
            overrideGraphQL({
                SettingsCascade: {
                    data: {
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
                    } as IQuery,
                    errors: undefined,
                },
                OverwriteSettings: {
                    data: {
                        settingsMutation: {
                            overwriteSettings: {
                                empty: {
                                    alwaysNil: null,
                                },
                            },
                        },
                    } as IMutation,
                    errors: undefined,
                },
                User: {
                    data: {
                        user: {
                            __typename: 'User',
                            id: testUserID,
                            username: 'test',
                            displayName: null,
                            url: '/users/test',
                            settingsURL: '/users/test/settings',
                            avatarURL: null,
                            viewerCanAdminister: true,
                            siteAdmin: true,
                            builtinAuth: true,
                            createdAt: '2020-03-02T11:52:15Z',
                            emails: [{ email: 'test@sourcegraph.test', verified: true } as IUserEmail],
                            organizations: { nodes: [] as IOrg[] } as IOrgConnection,
                            permissionsInfo: null,
                        },
                    } as IQuery,
                    errors: undefined,
                },
                ViewerSettings: {
                    data: {
                        viewerSettings: {
                            subjects: [
                                {
                                    __typename: 'DefaultSettings',
                                    latestSettings: {
                                        id: 0,
                                        contents: JSON.stringify({}),
                                    },
                                    settingsURL: null,
                                    viewerCanAdminister: false,
                                },
                                {
                                    __typename: 'Site',
                                    id: 'U2l0ZToic2l0ZSI=',
                                    siteID: 'e00d94ff-adc1-432c-ab53-5181c664b1ed',
                                    latestSettings: {
                                        id: 470,
                                        contents: JSON.stringify({}),
                                    },
                                    settingsURL: '/site-admin/global-settings',
                                    viewerCanAdminister: true,
                                },
                                {
                                    __typename: 'User',
                                    id: testUserID,
                                    username: 'test',
                                    displayName: null,
                                    latestSettings: {
                                        id: settingsID,
                                        contents: JSON.stringify({}),
                                    },
                                    settingsURL: '/users/test/settings',
                                    viewerCanAdminister: true,
                                },
                            ],
                            final: JSON.stringify({}),
                        },
                    } as IQuery,
                    errors: undefined,
                },
                CurrentAuthState: {
                    data: {
                        currentUser: {
                            __typename: 'User',
                            id: testUserID,
                            databaseID: 1,
                            username: 'test',
                            avatarURL: null,
                            email: 'felix@sourcegraph.com',
                            displayName: null,
                            siteAdmin: true,
                            tags: [] as string[],
                            url: '/users/test',
                            settingsURL: '/users/test/settings',
                            organizations: { nodes: [] as IOrg[] },
                            session: { canSignOut: true },
                            viewerCanAdminister: true,
                        },
                    } as IQuery,
                    errors: undefined,
                },
                StatusMessages: {
                    data: {
                        statusMessages: [] as StatusMessage[],
                    } as IQuery,
                    errors: undefined,
                },
                ActivationStatus: {
                    data: {
                        externalServices: { totalCount: 3 },
                        repositories: { totalCount: 9 },
                        viewerSettings: {
                            final: JSON.stringify({}),
                        },
                        users: { totalCount: 2 },
                        currentUser: {
                            usageStatistics: {
                                searchQueries: 171,
                                findReferencesActions: 14,
                                codeIntelligenceActions: 670,
                            },
                        },
                    } as IQuery,
                    errors: undefined,
                },
                SiteFlags: {
                    data: {
                        site: {
                            needsRepositoryConfiguration: false,
                            freeUsersExceeded: false,
                            alerts: [] as IAlert[],
                            authProviders: {
                                nodes: [
                                    {
                                        serviceType: 'builtin',
                                        serviceID: '',
                                        clientID: '',
                                        displayName: 'Builtin username-password authentication',
                                        isBuiltin: true,
                                        authenticationURL: null,
                                    },
                                ],
                            },
                            disableBuiltInSearches: false,
                            sendsEmailVerificationEmails: true,
                            updateCheck: {
                                pending: false,
                                checkedAt: '2020-07-07T12:31:16+02:00',
                                errorMessage: null,
                                updateVersionAvailable: null,
                            },
                            productSubscription: {
                                license: { expiresAt: '3021-05-28T16:06:40Z' },
                                noLicenseWarningUserCount: null,
                            },
                            productVersion: '0.0.0+dev',
                        },
                    } as IQuery,
                    errors: undefined,
                },
                logEvent: {
                    data: {
                        logEvent: {
                            alwaysNil: null,
                        },
                    } as IMutation,
                    errors: undefined,
                },
                logUserEvent: {
                    data: {
                        logUserEvent: {
                            alwaysNil: null,
                        },
                    } as IMutation,
                    errors: undefined,
                },
            })

            const getSettingsEditorContent = async (): Promise<string | null | undefined> => {
                await driver.page.waitForSelector('.e2e-settings-file .monaco-editor .view-lines')
                return driver.page.evaluate(
                    () =>
                        document
                            .querySelector<HTMLElement>('.e2e-settings-file .monaco-editor .view-lines')
                            ?.textContent?.replace(/\u00A0/g, ' ') // Monaco replaces all spaces with &nbsp;
                )
            }

            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings')

            await driver.page.waitForSelector('.e2e-settings-file .monaco-editor')
            await driver.page.waitForSelector('.e2e-save-toolbar-save')

            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.e2e-save-toolbar-save')?.disabled
                ),
                true,
                'Expected save button to be disabled'
            )

            // Replace with new settings
            const newSettings = '{ /* These are new settings */}'
            await driver.replaceText({
                selector: '.e2e-settings-file .monaco-editor .view-lines',
                newText: newSettings,
                selectMethod: 'keyboard',
                enterTextMethod: 'type',
            })
            await retry(async () => {
                const currentSettings = await getSettingsEditorContent()
                assert.strictEqual(currentSettings, newSettings)
            })

            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.e2e-save-toolbar-save')?.disabled
                ),
                false,
                'Expected save button to not be disabled'
            )

            // Assert mutation is done when save button is clicked
            const overrideSettingsVariables = await waitForGraphQLRequest(async () => {
                await driver.findElementWithText('Save changes', { action: 'click' })
            }, 'OverwriteSettings')

            assert.deepStrictEqual(overrideSettingsVariables, {
                contents: newSettings,
                lastID: settingsID,
                subject: testUserID,
            })
        })
    })
})
