import { SourcegraphContext } from '../../src/jscontext'

import { getSiteConfig } from './get-site-config'

// TODO: share with `client/web/src/integration/jscontext` which is not included into `tsconfig.json` now.
export const builtinAuthProvider = {
    serviceType: 'builtin' as const,
    serviceID: '',
    clientID: '',
    displayName: 'Builtin username-password authentication',
    isBuiltin: true,
    authenticationURL: '',
}

export const createJsContext = ({ sourcegraphBaseUrl }: { sourcegraphBaseUrl: string }): SourcegraphContext => {
    const siteConfig = getSiteConfig()

    if (siteConfig?.authProviders) {
        siteConfig.authProviders.unshift(builtinAuthProvider)
    }

    return {
        externalURL: sourcegraphBaseUrl,
        accessTokensAllow: 'all-users-create',
        allowSignup: true,
        batchChangesEnabled: true,
        codeIntelAutoIndexingEnabled: false,
        externalServicesUserModeEnabled: true,
        productResearchPageEnabled: true,
        csrfToken: 'qwerty',
        assetsRoot: '/.assets',
        deployType: 'dev',
        debug: true,
        emailEnabled: false,
        experimentalFeatures: {},
        isAuthenticatedUser: true,
        likelyDockerOnMac: false,
        needServerRestart: false,
        needsSiteInit: false,
        resetPasswordEnabled: true,
        sentryDSN: null,
        site: {
            'update.channel': 'release',
        },
        siteID: 'TestSiteID',
        siteGQLID: 'TestGQLSiteID',
        sourcegraphDotComMode: true,
        userAgentIsBot: false,
        version: '0.0.0',
        xhrHeaders: {},
        authProviders: [builtinAuthProvider],
        // Site-config overrides default JS context
        ...siteConfig,
    }
}
