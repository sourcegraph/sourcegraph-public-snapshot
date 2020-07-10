import { GraphQLOverrides } from './helpers'
import { IQuery, StatusMessage, IOrg, IAlert, IMutation } from '../../../shared/src/graphql/schema'
import { builtinAuthProvider, siteID, siteGQLID } from './jscontext'

export const testUserID = 'TestUserID'
export const settingsID = 123

/**
 * Predefined results for GraphQL requests that are made on almost every page.
 */
export const commonGraphQlResults: GraphQLOverrides = {
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
                        id: siteGQLID,
                        siteID,
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
    SiteFlags: {
        data: {
            site: {
                needsRepositoryConfiguration: false,
                freeUsersExceeded: false,
                alerts: [] as IAlert[],
                authProviders: {
                    nodes: [builtinAuthProvider],
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
}
