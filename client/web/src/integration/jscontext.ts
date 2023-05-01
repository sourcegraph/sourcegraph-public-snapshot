import { currentUserMock } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import { SourcegraphContext } from '../jscontext'

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
    extsvcConfigAllowEdits: false,
    extsvcConfigFileExists: false,
    codeIntelAutoIndexingEnabled: true,
    codeIntelAutoIndexingAllowGlobalPolicies: true,
    externalServicesUserMode: 'disabled',
    productResearchPageEnabled: true,
    assetsRoot: new URL('/.assets', sourcegraphBaseUrl).href,
    deployType: 'dev',
    debug: true,
    emailEnabled: false,
    experimentalFeatures: {},
    isAuthenticatedUser: true,
    needServerRestart: false,
    needsSiteInit: false,
    needsRepositoryConfiguration: false,
    resetPasswordEnabled: false,
    sentryDSN: null,
    site: {},
    siteID,
    siteGQLID,
    sourcegraphDotComMode: false,
    sourcegraphAppMode: false,
    userAgentIsBot: false,
    version: '0.0.0',
    xhrHeaders: {},
    authProviders: [builtinAuthProvider],
    authMinPasswordLength: 12,
    embeddingsEnabled: false,
    runningOnMacOS: true,
    localFilePickerAvailable: false,
    srcServeGitUrl: 'http://127.0.0.1:3434',
    primaryLoginProvidersCount: 5,
    batchChangesRolloutWindows: null,
})
