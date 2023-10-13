import { currentUserMock } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import type { SourcegraphContext } from '../jscontext'

export const siteID = 'TestSiteID'
export const siteGQLID = 'TestGQLSiteID'

export const builtinAuthProvider = {
    serviceType: 'builtin' as const,
    serviceID: '',
    clientID: '',
    displayName: 'Builtin username-password authentication',
    isBuiltin: true,
    authenticationURL: '',
}

export const createJsContext = ({ sourcegraphBaseUrl }: { sourcegraphBaseUrl: string }): SourcegraphContext => ({
    currentUser: currentUserMock,
    temporarySettings: null,
    externalURL: sourcegraphBaseUrl,
    accessTokensAllow: 'all-users-create',
    allowSignup: false,
    batchChangesEnabled: true,
    batchChangesDisableWebhooksWarning: false,
    batchChangesWebhookLogsEnabled: true,
    codeInsightsEnabled: true,
    executorsEnabled: true,
    codyEnabled: true,
    codyEnabledForCurrentUser: true,
    codyRequiresVerifiedEmail: false,
    extsvcConfigAllowEdits: false,
    extsvcConfigFileExists: false,
    codeIntelAutoIndexingEnabled: true,
    codeIntelAutoIndexingAllowGlobalPolicies: true,
    productResearchPageEnabled: true,
    assetsRoot: new URL('/.assets', sourcegraphBaseUrl).href,
    deployType: 'dev',
    debug: true,
    emailEnabled: false,
    experimentalFeatures: {},
    isAuthenticatedUser: true,
    licenseInfo: {
        currentPlan: 'team-0',
        batchChanges: {
            maxNumChangesets: -1,
            unrestricted: true,
        },
    },
    needServerRestart: false,
    needsSiteInit: false,
    needsRepositoryConfiguration: false,
    resetPasswordEnabled: false,
    sentryDSN: null,
    site: {},
    siteID,
    siteGQLID,
    sourcegraphDotComMode: false,
    codyAppMode: false,
    userAgentIsBot: false,
    version: '0.0.0',
    xhrHeaders: {},
    authProviders: [builtinAuthProvider],
    authMinPasswordLength: 12,
    embeddingsEnabled: false,
    runningOnMacOS: true,
    srcServeGitUrl: 'http://127.0.0.1:3434',
    primaryLoginProvidersCount: 5,
})
