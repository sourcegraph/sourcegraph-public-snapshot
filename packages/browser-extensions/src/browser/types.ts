export interface RepoLocations {
    [key: string]: string
}

export interface PhabricatorMapping {
    callsign: string
    path: string
}

/**
 * The feature flags available.
 */
export interface FeatureFlags {
    newTooltips: boolean
    newInject: boolean
}

export const featureFlagDefaults: FeatureFlags = {
    newTooltips: true,
    newInject: false,
}

export interface AccessToken {
    id: string
    token: string
}

/** A map where the key is the server URL and the value is the token. */
export interface AccessTokens {
    [url: string]: AccessToken
}

// TODO(chris) Switch to Partial<StorageItems> to eliminate bugs caused by
// missing items.
export interface StorageItems {
    sourcegraphURL: string
    /**
     * The current users access tokens the different sourcegraphUrls they have
     * had configured.
     */
    accessTokens: AccessTokens

    gitHubEnterpriseURL: string
    phabricatorURL: string
    inlineSymbolSearchEnabled: boolean
    renderMermaidGraphsEnabled: boolean
    identity: string
    serverUrls: string[]
    enterpriseUrls: string[]
    serverUserId: string
    hasSeenServerModal: boolean
    repoLocations: RepoLocations
    phabricatorMappings: PhabricatorMapping[]
    openFileOnSourcegraph: boolean
    sourcegraphAnonymousUid: string
    disableExtension: boolean
    /**
     * Enable the use of Sourcegraph extensions.
     */
    useExtensions: boolean
    /**
     * Storage for feature flags
     */
    featureFlags: FeatureFlags
    clientConfiguration: ClientConfigurationDetails
    /**
     * Overrides settings from Sourcegraph.
     */
    clientSettings: string
}

interface ClientConfigurationDetails {
    contentScriptUrls: string[]
    parentSourcegraph: {
        url: string
    }
}

export const defaultStorageItems: StorageItems = {
    sourcegraphURL: 'https://sourcegraph.com',
    accessTokens: {},

    serverUrls: ['https://sourcegraph.com'],
    gitHubEnterpriseURL: '',
    phabricatorURL: '',
    inlineSymbolSearchEnabled: true,
    renderMermaidGraphsEnabled: false,
    identity: '',
    enterpriseUrls: [],
    serverUserId: '',
    hasSeenServerModal: false,
    repoLocations: {},
    phabricatorMappings: [],
    openFileOnSourcegraph: true,
    sourcegraphAnonymousUid: '',
    disableExtension: false,
    useExtensions: false,
    featureFlags: featureFlagDefaults,
    clientConfiguration: {
        contentScriptUrls: [],
        parentSourcegraph: {
            url: 'https://sourcegraph.com',
        },
    },
    clientSettings: '',
}

export type StorageChange = { [key in keyof StorageItems]: chrome.storage.StorageChange }
